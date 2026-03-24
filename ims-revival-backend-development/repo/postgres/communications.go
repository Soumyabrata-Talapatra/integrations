package repo

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"plirevival/core/domain"

	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/volatiletech/null/v9"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/api-log"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

// EmpCommunicationDetailsRepository implements port.EmpCommunicationDetailsRepository interface
// and provides access to the postgres database for employee communication details-related operations
type CommunicationRepository struct {
	db    *dblib.DB
	cfg   *config.Config
	minio *minio.Client
}

// NewEmpCommunicationDetailsRepository creates a new EmployeeCommunicationDetailsRepository instance
func NewCommunicationRepository(db *dblib.DB, cfg *config.Config, minio *minio.Client) *CommunicationRepository {
	return &CommunicationRepository{
		db,
		cfg,
		minio,
	}
}
func (cr *CommunicationRepository) CreateCommunicationsQuery(ctx context.Context, comm *domain.EmpCommunicationDetails, ext string, file multipart.File, fileType string, header *multipart.FileHeader) (*domain.EmpDetailsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutLow))
	defer cancel()
	approverOfficeID, err := cr.GetApproverOfficeID(ctx, comm.OfficeID.Int64, comm.EmpPostID.Int64, comm.OfficeTypeCode.String)
	if err != nil || approverOfficeID == 0 {
		log.Error(ctx, "Unable to fetch approver office ID from MDM Office Master: %v", err)
		return nil, fmt.Errorf("unable to fetch approver office ID from MDM Office Master: %w", err)
	}
	batch := &pgx.Batch{}
	type CountResult struct {
		Count int64 `db:"count"`
	}

	var countResult CountResult
	checkQuery := dblib.Psql.Select("count(*)").
		From(commMakerTable).
		Where(sq.And{sq.Eq{"employee_id": comm.EmployeeID}, sq.Eq{"status": forwarded}})

	countResult, err = dblib.SelectOne(ctx, cr.db, checkQuery, pgx.RowToStructByName[CountResult])
	if err != nil {
		log.Error(ctx, "Failed to fetch count:", err)
		return nil, err
	}
	if countResult.Count > 0 {
		log.Error(ctx, "Communication details have already been forwarded. New entries cannot be added.")
		return nil, fmt.Errorf("communication details already forwarded for employee_id %d", comm.EmployeeID.Int64)
	}

	// Cancel the existing record
	cancelQuery := dblib.Psql.Update(commMakerTable).
		Set("updated_date", currDateTime).
		Set("status", cancelled).
		Set("updated_by", strconv.Itoa(int(comm.EmployeeID.Int64))).
		Where(sq.And{
			sq.Eq{"employee_id": comm.EmployeeID},
			sq.Eq{"status": pending},
		})
	err = dblib.QueueExecRow(batch, cancelQuery)
	if err != nil {
		log.Error(ctx, failedToExecQuery, err)
		return nil, err
	}

	var commResponse domain.EmpDetailsResponse
	// Insert data into pis.employee_communication_detail table
	insertQuery := dblib.Psql.Insert(commMakerTable).
		Columns(
			"employee_id",
			"communication_address_1",
			"communication_address_2",
			"communication_address_3",
			"communication_pin",
			"india_post_email_id",
			"personal_email_id",
			"mobile_no",
			"aadhaar_ref_number",
			"pan_number",
			"status",
			"user_remarks",
			"created_by",
			"created_date",
			"admin_office",
			"fwd_auth_remarks").
		Values(
			comm.EmployeeID,
			comm.CommunicationAddr1,
			comm.CommunicationAddr2,
			comm.CommunicationAddr3,
			comm.CommunicationPIN,
			comm.IndiaPostEmailID,
			comm.PersonalEmailID,
			comm.MobileNo,
			comm.AadhaarRefNumber,
			comm.PANNumber,
			comm.Status,
			comm.UserRemarks,
			comm.CreatedBy,
			currDateTime,
			approverOfficeID,
			comm.FwdAuthRemarks).
		Suffix("RETURNING communication_id AS details_id, employee_id, status,COALESCE(NULLIF(user_remarks, ''), fwd_auth_remarks,'') AS remarks")

	err = dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByName[domain.EmpDetailsResponse], &commResponse)
	if err != nil {
		log.Error(ctx, failedToExecQuery, err)
		return nil, err
	}
	// Send batch to the database
	err = cr.db.SendBatch(ctx, batch).Close()
	if err != nil {
		log.Error(ctx, err)
		log.Debug(ctx, errorPrefix, err)
		return nil, err
	}

	// Check if file is provided for upload and handle it
	var document domain.Document
	if file != nil {
		// Generate unique filename for the document
		uniqueID := generateUniqueID() // Implement this function to generate a unique number or UUID
		filename := fmt.Sprintf("%s_%s_%s_%s%s",
			strconv.FormatInt(comm.EmployeeID.Int64, 10),
			strconv.FormatInt(comServiceTypeID, 10),
			strconv.FormatUint(uint64(commResponse.DetailsID), 10), // Convert uint64 to string
			uniqueID,
			ext,
		)

		// Upload the file
		err = cr.UploadFile(file, filename, fileType, header.Size)
		if err != nil {
			log.Error(ctx, "File upload failed for %s: %s", filename, err.Error())
			// Revert the communication record insertion
			revertQuery := dblib.Psql.Update(commMakerTable).
				Set("status", failed).
				Set("remarks", "File upload failed").
				Set("updated_date", time.Now()).
				Where(sq.Eq{"communication_id": commResponse.DetailsID})
			// Convert the query to SQL
			revertSQL, args, err := revertQuery.ToSql()
			if err != nil {
				log.Debug(ctx, "Error building revert SQL query:", err)
				return nil, err
			}

			// Execute the query
			_, revertErr := cr.db.Exec(ctx, revertSQL, args...)
			if revertErr != nil {
				log.Debug(ctx, "Error reverting communication record:", revertErr)
				return nil, revertErr
			}
			return nil, err
		}

		// Document details
		document = domain.Document{
			EmployeeID:           comm.EmployeeID.Int64,
			FileNameID:           commResponse.DetailsID,
			DocumentName:         filename,
			DocumentType:         fileType,
			DocumentSize:         header.Size,
			DocumentFilePath:     filename,
			DocumentUploadStatus: uploaded,
			DocumentUploadedBy:   comm.CreatedBy.String,
			DocumentUploadedDate: time.Now(),
		}
		insertDocQuery := dblib.Psql.Insert(docTable).
			Columns(
				"file_name_id", "employee_id", "service_type_id", "document_name", "document_type", "document_size",
				"document_approver_post_id", "document_upload_status", "document_uploaded_by",
				"document_uploaded_date", "remarks", "document_file_path", "approve_status").
			Values(
				document.FileNameID, comm.EmployeeID, comServiceTypeID, document.DocumentName, document.DocumentType,
				document.DocumentSize, document.DocumentApproverPostID, document.DocumentUploadStatus, document.DocumentUploadedBy,
				document.DocumentUploadedDate, document.Remarks, document.DocumentFilePath, pending).
			Suffix("RETURNING document_id")

		sql, args, err := insertDocQuery.ToSql()
		if err != nil {
			log.Debug(ctx, "Error building SQL query for document:", err)
			return nil, err
		}

		// Execute document insertion query
		var docID int
		err = cr.db.QueryRow(ctx, sql, args...).Scan(&docID)
		if err != nil {
			log.Debug(ctx, "Error inserting document:", err)
			return nil, err
		}

		// Log document uploaded
		log.Debug(ctx, "Document uploaded successfully:", document)
	}

	// Return the communication response and document (if uploaded)
	return &commResponse, nil
}

func generateUniqueID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10) // Example using current time in nanoseconds
}
func (cr *CommunicationRepository) UploadFile(file multipart.File, objectName, contentType string, size int64) error {

	bucketName := cr.cfg.GetString(minioBucketName)

	if cr.minio == nil {

		return fmt.Errorf("MinIO client is not initialized")
	}
	_, err := cr.minio.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}

func (cr *CommunicationRepository) WithdrawCommunications(ctx context.Context, commID int32, empID int64, remarks string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutLow))
	defer cancel()
	query := dblib.Psql.Update(commMakerTable).
		Set("user_remarks", remarks).
		Set("updated_date", currDateTime).
		Set("status", cancelled).
		Set("updated_by", strconv.Itoa(int(empID))).
		Where(sq.And{
			sq.Eq{"employee_id": empID},
			sq.Eq{"communication_id": commID},
			sq.Eq{"status": pending},
		})

	commandTag, err := dblib.Update(ctx, cr.db, query)
	if err != nil {
		log.Debug(ctx, errorPrefix, err)
		return 0, err
	}

	rowsAffected := commandTag.RowsAffected()
	return rowsAffected, nil
}

func (cr *CommunicationRepository) ForwardCommunications(ctx context.Context, commID int32, FwdAuthID int64, fwdAuthRemarks string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutLow))
	defer cancel()
	query := dblib.Psql.Update(commMakerTable).
		Set("updated_date", currDateTime).
		Set("fwd_auth_remarks", fwdAuthRemarks).
		Set("status", forwarded).
		Set("updated_by", strconv.Itoa(int(FwdAuthID))).
		Where(sq.And{
			sq.Eq{"communication_id": commID},
			sq.Eq{"status": pending},
			sq.Or{sq.Eq{"fwd_auth_remarks": ""}, sq.Expr("fwd_auth_remarks IS NULL")}})

	commandTag, err := dblib.Update(ctx, cr.db, query)
	if err != nil {
		log.Debug(ctx, errorPrefix, err)
		return 0, err
	}

	rowsAffected := commandTag.RowsAffected()
	return rowsAffected, nil
}

func (cr *CommunicationRepository) ApproveCommunicationsMaker(ctx context.Context, employeeids []int, approvedBy string, approveStatus string, appRemarks string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutMed))
	defer cancel()

	log.Debug(ctx, "Starting batch processing for approving/rejecting communication details")

	// Validate approveStatus
	if approveStatus != approved && approveStatus != rejected {
		return "", fmt.Errorf("invalid approve status: %s", approveStatus)
	}

	// Batch for combined update/insert operations of employee communication details
	updateBatch := &pgx.Batch{}
	for _, employeeID := range employeeids {
		// Add the query to the batch
		if approveStatus == approved {

			updateBatch.Queue(`
            WITH updated_maker AS (
                UPDATE pis.employee_communication_detail_maker
                SET status = $5, approved_by = $1, approved_date = $2, approve_auth_remarks = $3 
                WHERE employee_id = $4 AND status IN ('Pending', 'Forwarded')
                RETURNING employee_id, communication_address_1, communication_address_2, communication_address_3, communication_pin, 
                india_post_email_id, personal_email_id, mobile_no, aadhaar_ref_number, pan_number, approved_by, approved_date,
				approver_post_id, admin_office
            )
            INSERT INTO pis.employee_communication_detail (
                employee_id, communication_address_1, communication_address_2, communication_address_3, communication_pin, 
                india_post_email_id, personal_email_id, mobile_no, aadhaar_ref_number, pan_number, updated_by, updated_date,
				status, approved_by, approved_date, approver_post_id, admin_office
            )
            SELECT 
                $4, communication_address_1, communication_address_2, communication_address_3, communication_pin, india_post_email_id, 
                personal_email_id, mobile_no, aadhaar_ref_number, pan_number, $1, $2, $5, $1, $2, approver_post_id,
				admin_office
            FROM updated_maker
            ON CONFLICT (employee_id) DO UPDATE SET
                communication_address_1 = EXCLUDED.communication_address_1,
                communication_address_2 = EXCLUDED.communication_address_2,
                communication_address_3 = EXCLUDED.communication_address_3,
                communication_pin = EXCLUDED.communication_pin,
                india_post_email_id = EXCLUDED.india_post_email_id,
                personal_email_id = EXCLUDED.personal_email_id,
                mobile_no = EXCLUDED.mobile_no,
                aadhaar_ref_number = EXCLUDED.aadhaar_ref_number,
                pan_number = EXCLUDED.pan_number,
                updated_date = EXCLUDED.updated_date,
                approved_by = EXCLUDED.approved_by,
                approved_date = EXCLUDED.approved_date,
                status = EXCLUDED.status,
				approver_post_id=EXCLUDED.approver_post_id,
				admin_office=EXCLUDED.admin_office
        `, approvedBy, time.Now(), appRemarks, employeeID, approveStatus)
		} else {
			// If the status is "Rejected", only update the maker table
			updateBatch.Queue(`
                UPDATE pis.employee_communication_detail_maker
                SET status = $5, approved_by = $1, approved_date = $2, approve_auth_remarks = $3 
                WHERE employee_id = $4 AND status IN ('Pending', 'Forwarded')
            `, approvedBy, time.Now(), appRemarks, employeeID, approveStatus)
		}
	}

	// Execute the batch
	batchResults := cr.db.SendBatch(ctx, updateBatch)
	defer batchResults.Close()

	for _, employeeID := range employeeids {
		// Execute the next result in the batch
		commandTag, err := batchResults.Exec()
		if err != nil {
			log.Error(ctx, "Failed to update records", "employeeID", employeeID, "error", err)
			return "", err
		}
		// Check if no rows were affected
		if commandTag.RowsAffected() == 0 {
			return "", errors.New("no rows were updated, check the input")
		}
	}

	successMessage := fmt.Sprintf("Successfully approved %d record(s)", len(employeeids))
	return successMessage, nil
}

func communicationBaseQuery(table string, condition sq.Sqlizer) sq.SelectBuilder {
	return dblib.Psql.Select(
		"cd.communication_id",
		"cd."+conEmpID,
		"cd.communication_address_1",
		"cd.communication_address_2",
		"cd.communication_address_3",
		"cd.communication_pin",
		"cd.india_post_email_id",
		"cd.personal_email_id",
		"cd.mobile_no",
		"cd.aadhaar_ref_number",
		"cd.pan_number",
		"cd."+conApprPostID,
		"cd.admin_office",
		"cd."+conStatus,
		"cd.remarks",
		"om."+conOfficeName,
		"em."+conEmpDesignation,
		empFullName).
		From(table + " cd").
		Join(masterTable + " em ON em.employee_id = cd.employee_id and em.approve_status='Approved'").
		LeftJoin(officeTable + joinOfficeMasterCondition).
		Where(condition)
}

func resourceCommunicationQuery(table string, condition sq.Sqlizer) sq.SelectBuilder {
	return dblib.Psql.Select(
		"em.maker_id AS communication_id",
		"em.resource_id AS employee_id",
		"TRIM(COALESCE(em.first_name, '') || ' ' ||COALESCE(em.middle_name, '') || ' ' || COALESCE(em.last_name, '')) AS employee_name",
		"em.address1 AS communication_address_1",
		"em.address2 AS communication_address_2",
		"em.address3 AS communication_address_3",
		"em.pin_code AS communication_pin",
		"em.personal_email_id",
		"em.mobile_no",
		"em.aadhaar_number AS aadhaar_ref_number",
		"em.pan_number",
		"em.approver_post_id",
		"em.remarks",
		"em.office_id",
		"om.office_name AS office_of_working",
		"em.resource_status").
		From(table + " em").
		LeftJoin(officeTable + joinOfficeMasterCondition).
		Where(condition)
}

func (cr *CommunicationRepository) GetCommunications(ctx context.Context, table string, condition sq.Sqlizer, status string, skip uint64, limit uint64) ([]domain.EmpCommunicationDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutMed))
	defer cancel()

	query := communicationBaseQuery(table, condition)

	if table == commMakerTable {
		query = query.Column("cd.user_remarks").
			Column("cd.fwd_auth_remarks").
			Column("cd.approve_auth_remarks")
	}

	switch status {
	case unApproved:
		query = query.Where(sq.Or{
			sq.Eq{"cd." + conStatus: pending},
			sq.Eq{"cd." + conStatus: forwarded}})
	case "":
		// No filter — return all
	default:
		query = query.Where(sq.Eq{"cd." + conStatus: status})
	}

	query = query.OrderBy("cd.created_date DESC").
		Offset(uint64(skip * limit)).
		Limit(uint64(limit))

	return dblib.SelectRows(ctx, cr.db, query, pgx.RowToStructByNameLax[domain.EmpCommunicationDetails])
}

func (cr *CommunicationRepository) GetMakerCommunicationsByEmpID(ctx context.Context, empID int64, status string, skip uint64, limit uint64) ([]domain.EmpCommunicationDetails, error) {

	condition := sq.Eq{"cd." + conEmpID: empID}
	return cr.GetCommunications(ctx, commMakerTable, condition, status, skip, limit)

}

func (cr *CommunicationRepository) GetMakerCommunicationsByAOID(ctx context.Context, adminID int64, status string, skip uint64, limit uint64) ([]domain.EmpCommunicationDetails, error) {

	condition := sq.Eq{"cd.admin_office": adminID}
	return cr.GetCommunications(ctx, commMakerTable, condition, status, skip, limit)

}

func (cr *CommunicationRepository) GetMakerCommunicationsByAppPostID(ctx context.Context, postID string, status string, skip uint64, limit uint64) ([]domain.EmpCommunicationDetails, error) {

	condition := sq.Eq{"cd." + conApprPostID: postID}
	return cr.GetCommunications(ctx, commMakerTable, condition, status, skip, limit)

}

func (cr *CommunicationRepository) GetCommunicationsByEmpID(ctx context.Context, empID int64) (domain.EmpCommunicationDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutLow))
	defer cancel()
	var query sq.SelectBuilder
	if empID < 70000000 && empID >= 60000000 {
		condition := sq.And{sq.Eq{"em.resource_id": empID}, sq.Eq{"em.resource_status": active}}
		query = resourceCommunicationQuery(resourceTable, condition)
	} else {
		condition := sq.And{sq.Eq{"cd." + conEmpID: empID}, sq.Eq{"cd.status": approved}}
		query = communicationBaseQuery(commTable, condition)
	}
	return dblib.SelectOne(ctx, cr.db, query, pgx.RowToStructByNameLax[domain.EmpCommunicationDetails])
}

func (cr *CommunicationRepository) GetAllCommunicationsByEmpID(ctx context.Context, empID int64) (domain.EmpCommunicationDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutLow))
	defer cancel()
	condition := sq.Eq{"cd." + conEmpID: empID}
	query := communicationBaseQuery(commTable, condition)
	return dblib.SelectOne(ctx, cr.db, query, pgx.RowToStructByNameLax[domain.EmpCommunicationDetails])
}

func (cr *CommunicationRepository) GetCommunicationsByOTP(ctx context.Context, empID int64) (domain.EmpCommunicationDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutLow))
	defer cancel()
	query := dblib.Psql.Select(
		conEmpID,
		"personal_email_id",
		"mobile_no").
		From(commTable).
		Where(sq.And{sq.Eq{conEmpID: empID}, sq.Eq{"status": approved}})
	return dblib.SelectOne(ctx, cr.db, query, pgx.RowToStructByNameLax[domain.EmpCommunicationDetails])
}

func (cr *CommunicationRepository) GetCommunicationsByAppPostID(ctx context.Context, postID string, skip uint64, limit uint64) ([]domain.EmpCommunicationDetails, error) {

	condition := sq.And{sq.Eq{"cd." + conApprPostID: postID}, sq.Eq{"cd.status": approved}}
	return cr.GetCommunications(ctx, commTable, condition, approved, skip, limit)
}

func (cr *CommunicationRepository) GetCommunicationsByAdharRefNo(ctx context.Context, adharNo string, skip uint64, limit uint64) (domain.EmpCommunicationDetails, error) {

	condition := sq.And{sq.Eq{"cd.aadhaar_ref_number": adharNo}, sq.Eq{"cd.status": approved}, sq.Eq{"em.employment_status": active}}
	// return cr.GetCommunications(ctx, commTable, condition, approved, skip, limit)
	query := communicationBaseQuery(commTable, condition)
	return dblib.SelectOne(ctx, cr.db, query, pgx.RowToStructByNameLax[domain.EmpCommunicationDetails])
}

func (cr *CommunicationRepository) GetLeaveCommunications(ctx context.Context, empID int64) (domain.CombinedStructForLMS, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutMed))
	defer cancel()
	batch := &pgx.Batch{}
	var commResp domain.EmpCommunicationDetails
	var empResp domain.FetchEmployee

	condition := sq.Eq{"cd." + conEmpID: empID}

	commQuery := communicationBaseQuery(commTable, condition)

	err := dblib.QueueReturnRow(batch, commQuery, pgx.RowToStructByNameLax[domain.EmpCommunicationDetails], &commResp)
	if err != nil {
		log.Error(ctx, failedToExecQuery, err)
		return domain.CombinedStructForLMS{}, err
	}
	empQuery := dblib.Psql.Select(
		"em."+conEmpID,
		"TRIM(em.employee_first_name || ' ' || em.employee_middle_name || ' ' || em.employee_last_name) AS employee_first_name",
		"em."+conGroupPost,
		"em."+conCadre,
		"em."+conEmpDesignation,
		"oc."+conOfficeName,
		"em."+conPostID,
		"em."+conOfficeID,
		"em."+empStatus,
		"em."+conEmpType,
		"em.gender",
		"oc."+conCircleID,
		"em."+conDOJinDept,
		"em."+conMaritalStatus,
		"em."+conRrecruitMode,
		"oc."+conOfficeType,
		"oc."+conReportingOfficeID,
		"em."+conReportingAuthPostID,
		"em."+conCadreID,
		"oc.sub_division_office_id",
		"oc.ddo_office_id",
		"oc."+conDivisionID,
		"oc."+conRegionID,
		"oc.division_name",
		"oc.region_name",
		"oc.circle_name",
		"em."+conGroupID).
		From(masterTable + " em").
		LeftJoin("pis.kafka_office_master_composite oc ON em.office_id = oc.office_id").
		Where(sq.Eq{"em." + conEmpID: empID}).
		Where(sq.Eq{"em." + empStatus: active})

	err = dblib.QueueReturnRow(batch, empQuery, pgx.RowToStructByNameLax[domain.FetchEmployee], &empResp)
	if err != nil {
		log.Error(ctx, failedToExecQuery, err)
		return domain.CombinedStructForLMS{}, err
	}
	err = cr.db.SendBatch(ctx, batch).Close()
	if err != nil {
		log.Error(ctx, err)
		log.Debug(ctx, errorPrefix, err)
		return domain.CombinedStructForLMS{}, err
	}
	combinedResp := domain.CombinedStructForLMS{
		CommResp: commResp,
		EmpResp:  empResp,
	}
	// Return the response
	return combinedResp, nil
}
func (cr *CommunicationRepository) GetMobileStatusReportQuery(ctx context.Context) ([]domain.MobileStatusReport, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutMed))
	defer cancel()

	query := dblib.Psql.Select(
		"ht.circle_office_id",
		"ht.circle_name",
		"COUNT(em.employee_id) AS total_emp_count",
		"COUNT(cd.employee_id) FILTER (WHERE cd.mobile_no IS NOT NULL AND cd.mobile_no::text ~ '^[6-9][0-9]{9}$') AS emp_with_valid_mobile",
		"(COUNT(em.employee_id) - COUNT(cd.employee_id) FILTER (WHERE cd.mobile_no IS NOT NULL AND cd.mobile_no::text ~ '^[6-9][0-9]{9}$')) AS yet_to_update",
		"ROUND(COUNT(cd.employee_id) FILTER (WHERE cd.mobile_no IS NOT NULL AND cd.mobile_no::text ~ '^[6-9][0-9]{9}$') * 100.0 / NULLIF(COUNT(em.employee_id), 0), 2) AS percentage_completed",
	).
		From(masterTable+" em").
		Join(hierarchyTable+joinHierarchyMasterCondition).
		LeftJoin(commTable+" cd ON em.employee_id = cd.employee_id").
		Where("em.employment_status = ?", active).
		GroupBy("ht.circle_office_id", "ht.circle_name").
		OrderBy("ht.circle_name")

	reports, err := dblib.SelectRows(ctx, cr.db, query, pgx.RowToStructByNameLax[domain.MobileStatusReport])
	if err != nil {
		return nil, err
	}

	return reports, nil
}

func (cr *CommunicationRepository) NonUpdatedMobileNoReport(ctx context.Context, officeID int64, officeType string, skip, limit uint64) ([]domain.EmpCommunicationDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, cr.cfg.GetDuration(dbQueryTimeoutMed))
	defer cancel()
	var condition sq.Eq
	divOfficeTypes := map[string]struct{}{
		"PDN": {}, "RDN": {}, "GPO": {},
	}
	if _, ok := divOfficeTypes[officeType]; ok {
		condition = sq.Eq{"ht.division_office_id": officeID}
	} else {
		condition = sq.Eq{"em.office_id": officeID}
	}

	query := dblib.Psql.Select(
		"cd."+conEmpID,
		empFullName,
		"ht."+conOfficeName,
		"em."+conEmpDesignation,
		"cd.mobile_no").
		From(commTable + " cd").
		Join(masterTable + " em ON em.employee_id = cd.employee_id").
		Join(hierarchyTable + joinHierarchyMasterCondition).
		Where(condition).
		Where(sq.Or{
			sq.Expr("cd.mobile_no IS NULL"),
			sq.Expr("cd.mobile_no = 0"),
			sq.Expr("length(cd.mobile_no::text)!= 10"),
			sq.Expr("cd.mobile_no::text !~ '^[6-9][0-9]{9}$'"),
		})

	reports, err := dblib.SelectRows(ctx, cr.db, query, pgx.RowToStructByNameLax[domain.EmpCommunicationDetails])
	if err != nil {
		return nil, err
	}

	return reports, nil
}

func (ar *CommunicationRepository) GetApproverOfficeID(ctx context.Context, officeID int64, empPostID int64, officeTypeCode string) (int64, error) {
	type OfficeInfo struct {
		DivisionOfficeID  null.Int64 `db:"division_office_id"`
		OfficeID          null.Int64 `db:"office_id"`
		HeadOfTheOffice   null.Int64 `db:"head_of_the_office"`
		ReportingOfficeID null.Int64 `db:"reporting_office_id"`
	}

	query := dblib.Psql.Select("division_office_id", "office_id", "head_of_the_office", "reporting_office_id").
		From("pis.kafka_office_master_composite").
		Where(sq.And{sq.Eq{"office_id": officeID},
			sq.Eq{"office_type_code": officeTypeCode}})

	result, err := dblib.SelectOne(ctx, ar.db, query, pgx.RowToStructByName[OfficeInfo])
	if err != nil {
		log.Error(ctx, "Failed to retrieve office details: %v", err)
		return 0, fmt.Errorf("Failed to retrieve office details: %w", err)
	}

	officeTypeSet1 := map[string]struct{}{
		"BPO": {}, "SPO": {}, "HPO": {}, "ICH": {}, "LPC": {}, "NPH": {}, "NSH": {}, "NPO": {}, "RMO": {}, "SRO": {},
		"HRO": {}, "BNP": {}, "BPC": {}, "NDC": {}, "PPP": {}, "WTC": {}, "CBO": {}, "CRC": {}, "DBO": {}, "TMO": {},
		"FPO": {}, "LGH": {}, "OOE": {}, "SPC": {}, "PTC": {},
	}
	officeTypeSet2 := map[string]struct{}{
		"PDN": {}, "PSD": {}, "RDN": {}, "GPO": {}, "MMS": {}, "PAO": {}, "RLO": {}, "ABP": {}, "APS": {}, "CPT": {},
		"NPA": {}, "OTH": {}, "ACR": {}, "ADN": {}, "AFP": {}, "ARG": {}, "ASB": {}, "ASD": {}, "ATC": {}, "CBP": {},
		"FFP": {}, "MHL": {}, "PCC": {}, "PCE": {}, "PEC": {}, "RGN": {}, "BDT": {}, "CPR": {}, "CRL": {}, "DTE": {},
		"DTP": {}, "PAW": {}, "PDT": {}, "PCD": {}, "PCS": {}, "SDO": {}, "PED": {}, "PES": {}, "CSD": {},
	}

	if _, ok := officeTypeSet1[officeTypeCode]; ok {
		return result.DivisionOfficeID.Int64, nil
	}
	if _, ok := officeTypeSet2[officeTypeCode]; ok {
		if result.HeadOfTheOffice.Valid && result.HeadOfTheOffice.Int64 != 0 {
			if empPostID != result.HeadOfTheOffice.Int64 {
				return result.OfficeID.Int64, nil
			} else {
				return result.ReportingOfficeID.Int64, nil
			}
		} else {
			return 0, nil
		}
	}

	log.Error(ctx, "Failed to retrieve office details: %d, %s", officeID, officeTypeCode)
	return 0, fmt.Errorf("Failed to retrieve office details: %d, %s", officeID, officeTypeCode)
}
