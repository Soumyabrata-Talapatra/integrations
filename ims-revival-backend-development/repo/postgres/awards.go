package repo

import (
	"context"
	"fmt"
	"mime/multipart"
	"path"
	"plirevival/core/domain"

	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
)

type AwardRepository struct {
	db       *dblib.DB
	cfg      *config.Config
	minio    *minio.Client
	commRepo *CommunicationRepository
}

// / NewAwardsRepository creates a new awards repository instance
func NewAwardsRepository(db *dblib.DB, cfg *config.Config, minio *minio.Client, commRepo *CommunicationRepository) *AwardRepository {
	return &AwardRepository{
		db,
		cfg,
		minio,
		commRepo,
	}
}

func (ar *AwardRepository) CreateAwardsBulkQuery(ctx context.Context, awards []domain.EmpAwardDetails,
	files []multipart.File, fileHeaders []*multipart.FileHeader) ([]domain.AwardsCreateResponse, error) {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(ctx, ar.cfg.GetDuration(dbQueryTimeoutMed))
	defer cancel()

	approverOfficeID, err := ar.commRepo.GetApproverOfficeID(ctx, awards[0].OfficeID.Int64, awards[0].EmpPostID.Int64, awards[0].OfficeTypeCode.String)
	if err != nil || approverOfficeID == 0 {
		log.Error(ctx, "Unable to fetch approver office ID from MDM Office Master: %v", err)
		return nil, fmt.Errorf("unable to fetch approver office ID from MDM Office Master: %w", err)
	}

	type CountResult struct {
		Count int64 `db:"count"`
	}

	var result CountResult
	empID := awards[0].EmployeeID
	checkQuery := dblib.Psql.Select("count(*)").
		From(awardMakerTable).
		Where(sq.And{sq.Eq{"employee_id": empID}, sq.Eq{"status": forwarded}})

	result, err = dblib.SelectOne(ctx, ar.db, checkQuery, pgx.RowToStructByName[CountResult])
	if err != nil {
		log.Error(ctx, "Failed to fetch count:", err)
		return nil, err
	}
	if result.Count > 0 {
		log.Error(ctx, "Award details have already been forwarded. New entries cannot be added.")
		return nil, fmt.Errorf("award details already forwarded for employee_id %d", empID.Int64)
	}

	currentTime := time.Now()
	batch := &pgx.Batch{}
	var awardResponses []domain.AwardsCreateResponse

	// Create the insert query for the new award details
	insertQuery := dblib.Psql.Insert(awardMakerTable).
		Columns("employee_id",
			"award_name",
			"award_type",
			"award_category",
			"award_description",
			"certificate_no",
			"monetary_benefit",
			"status",
			"created_by",
			"created_date",
			"user_remarks",
			"award_issue_date",
			"award_for_year",
			"admin_office",
			"fwd_auth_remarks")

	// Queue cancel queries for each award detail
	for _, detail := range awards {

		cancelQuery := dblib.Psql.Update(awardMakerTable).
			Set("updated_date", currentTime).
			Set("status", cancelled).
			Set("updated_by", strconv.Itoa(int(detail.EmployeeID.Int64))).
			Where(sq.And{sq.Eq{"employee_id": detail.EmployeeID},
				sq.Eq{"status": pending},
				sq.Eq{"award_details_id": detail.AwardID}})

		// Queue the cancel query for execution
		err := dblib.QueueExecRow(batch, cancelQuery)
		if err != nil {
			log.Error(ctx, failedToExecQuery, err)
			return nil, err
		}

		// Queue insert query for the current award detail
		query := insertQuery.Values(
			detail.EmployeeID,
			detail.AwardName,
			detail.AwardType,
			detail.AwardCategory,
			detail.AwardDescription,
			detail.CertificateNo,
			detail.MonetaryBenefit,
			detail.Status,
			detail.CreatedBy,
			currentTime,
			detail.UserRemarks,
			detail.AwardIssueDate,
			detail.AwardForYear,
			approverOfficeID,
			detail.FwdAuthRemarks).
			Suffix("RETURNING award_details_id, employee_id, award_name, award_type, COALESCE(NULLIF(user_remarks, ''), fwd_auth_remarks,'') AS user_remarks, status")

		err = dblib.QueueReturnBulk(batch, query, pgx.RowToStructByNameLax[domain.AwardsCreateResponse], &awardResponses)
		if err != nil {
			log.Error(ctx, failedToExecQuery, err)
			return nil, err
		}
	}

	batchResults := ar.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	if err := batchResults.Close(); err != nil {
		log.Error(ctx, "Error closing batch results: %s", err.Error())
		return nil, err
	}

	if len(awardResponses) != len(files) {
		log.Error(ctx, "Mismatch between award details and files")
		return nil, fmt.Errorf("number of award details and files do not match")
	}

	// Process file uploads and document records
	for i, awardResponse := range awardResponses {
		file := files[i]
		header := fileHeaders[i]
		award := awards[i]
		ext := path.Ext(header.Filename)
		uniqueID := generateUniqueID()
		var document domain.Document
		filename := fmt.Sprintf("%d_%d_%d_%s%s", awardResponse.EmployeeID.Int64, awardServiceTypeID,
			awardResponse.AwardDetailsID.Int, uniqueID, ext)

		// Upload the file
		err := ar.UploadFile(file, filename, header.Header.Get(contentType), header.Size)
		if err != nil {
			log.Error(ctx, "File upload failed for %s: %s", filename, err.Error())
			// Revert the award record insertion
			revertQuery := dblib.Psql.Update(awardMakerTable).
				Set("status", failed).
				Set("remarks", "File upload failed").
				Set("updated_date", time.Now()).
				Where(sq.Eq{"award_details_id": awardResponse.AwardDetailsID})
			// Convert the query to SQL
			revertSQL, args, err := revertQuery.ToSql()
			if err != nil {
				log.Debug(ctx, "Error building revert SQL query:", err)
				return nil, err
			}

			// Execute the query
			_, revertErr := ar.db.Exec(ctx, revertSQL, args...)
			if revertErr != nil {
				log.Debug(ctx, "Error reverting awards record", revertErr)
				return nil, revertErr
			}
			return nil, err
		}
		document = domain.Document{
			DocumentName:           filename,
			DocumentType:           header.Header.Get(contentType),
			DocumentSize:           header.Size,
			DocumentApproverPostID: award.ApproverPostID.String,
			DocumentFilePath:       filename,
			DocumentUploadStatus:   uploaded,
			DocumentUploadedBy:     award.CreatedBy.String,
			DocumentUploadedDate:   time.Now(),
		}

		// Insert document details
		insertDocQuery := dblib.Psql.Insert(docTable).
			Columns("file_name_id", "employee_id", "service_type_id", "document_name", "document_type",
				"document_size", "document_approver_post_id", "document_upload_status", "document_uploaded_by",
				"document_uploaded_date", "remarks", "document_file_path", "approve_status").
			Values(
				awardResponse.AwardDetailsID, awardResponse.EmployeeID, awardServiceTypeID, document.DocumentName,
				document.DocumentType, document.DocumentSize, document.DocumentApproverPostID, document.DocumentUploadStatus,
				document.DocumentUploadedBy, document.DocumentUploadedDate, document.Remarks, document.DocumentFilePath, pending)

		query, args, err := insertDocQuery.ToSql() // Generate the SQL query and arguments
		if err != nil {
			log.Error(ctx, "Failed to build SQL query: %s", err.Error())
			return nil, err
		}

		_, err = ar.db.Exec(ctx, query, args...) // Use the generated SQL and args
		if err != nil {
			log.Error(ctx, "Failed to execute SQL query: %s", err.Error())
			return nil, err
		}
	}
	log.Debug(ctx, "Bulk operation completed successfully")
	return awardResponses, nil
}

func (ar *AwardRepository) UploadFile(file multipart.File, objectName, contentType string, size int64) error {
	// The function is used to upload files to MinIO for document management.
	bucketName := ar.cfg.GetString(minioBucketName)

	if ar.minio == nil {
		return fmt.Errorf("MinIO client is not initialized")
	}
	_, err := ar.minio.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		size,
		minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (ar *AwardRepository) WithdrawAwards(ctx context.Context, awardID int32, empID int64, remarks string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, ar.cfg.GetDuration(dbQueryTimeoutLow))
	defer cancel()
	query := dblib.Psql.Update(awardMakerTable).
		Set("user_remarks", remarks).
		Set("updated_date", currDateTime).
		Set("status", cancelled).
		Set("updated_by", strconv.Itoa(int(empID))).
		Where(sq.And{sq.Eq{"employee_id": empID},
			sq.Eq{"award_details_id": awardID},
			sq.Eq{"status": pending}})

	commandTag, err := dblib.Update(ctx, ar.db, query)
	if err != nil {
		log.Debug(ctx, errorPrefix, err)
		return 0, err
	}

	rowsAffected := commandTag.RowsAffected()
	return rowsAffected, nil
}

func (ar *AwardRepository) ForwardAwards(ctx context.Context, awardID int32, fwdAuthID int64, remarks string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, ar.cfg.GetDuration(dbQueryTimeoutLow))
	defer cancel()
	query := dblib.Psql.Update(awardMakerTable).
		Set("updated_date", currDateTime).
		Set("fwd_auth_remarks", remarks).
		Set("status", forwarded).
		Set("updated_by", strconv.Itoa(int(fwdAuthID))).
		Where(sq.And{sq.Eq{"award_details_id": awardID},
			sq.Eq{"status": pending},
			sq.Or{sq.Eq{"fwd_auth_remarks": ""}, sq.Expr("fwd_auth_remarks IS NULL")}})

	commandTag, err := dblib.Update(ctx, ar.db, query)
	if err != nil {
		log.Debug(ctx, errorPrefix, err)
		return 0, err
	}

	rowsAffected := commandTag.RowsAffected()
	return rowsAffected, nil
}

func (ar *AwardRepository) ApproveAwardsQry(ctx context.Context, awardIDs []int, approvedBy, approveStatus, appRemarks string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ar.cfg.GetDuration(dbQueryTimeoutMed))
	defer cancel()

	// Validate approveStatus
	if approveStatus != approved && approveStatus != rejected {
		return "", fmt.Errorf("invalid approve status: %s", approveStatus)
	}

	// Batch operations for combined update/insert of employee award details
	batch := &pgx.Batch{}
	for _, awardID := range awardIDs {
		if approveStatus == approved {
			batch.Queue(`
			WITH updated_maker AS (
				UPDATE pis.employee_award_detail_maker
				SET status = $5, approved_by = $1, approved_date = $2, approve_auth_remarks = $3
				WHERE award_details_id = $4 AND status IN ('Pending', 'Forwarded')
				RETURNING employee_id, award_name, award_type, award_category, award_description, certificate_no, 
				monetary_benefit, approver_post_id, award_issue_date, award_for_year, admin_office,
				created_by, created_date
			)
			INSERT INTO pis.employee_award_detail (
				employee_id, award_name, award_type, award_category, award_description, certificate_no, 
				monetary_benefit, approver_post_id, award_issue_date, award_for_year,
				status, admin_office, updated_date, created_by, created_date
			)
			SELECT employee_id, award_name, award_type, award_category, award_description, certificate_no, 
				monetary_benefit, approver_post_id, award_issue_date, award_for_year, $5, admin_office, $2,
				created_by, $2
			FROM updated_maker
			ON CONFLICT (award_details_id) DO UPDATE SET
				award_name = EXCLUDED.award_name,
				award_type = EXCLUDED.award_type,
				award_category = EXCLUDED.award_category,
				award_description = EXCLUDED.award_description,
				certificate_no = EXCLUDED.certificate_no,
				monetary_benefit = EXCLUDED.monetary_benefit,
				approver_post_id = EXCLUDED.approver_post_id,
				admin_office = EXCLUDED.admin_office,
				award_issue_date=EXCLUDED.award_issue_date,
				award_for_year=EXCLUDED.award_for_year,
				updated_date = EXCLUDED.updated_date,
				created_by= EXCLUDED.created_by,
				created_date= EXCLUDED.created_date
			`, approvedBy, time.Now(), appRemarks, awardID, approveStatus)
		} else {
			// For "Rejected" status, update only in maker table
			batch.Queue(`
			UPDATE employee_award_detail_maker
			SET status = $5, approved_by = $1, approved_date = $2, approve_auth_remarks = $3
			WHERE award_details_id = $4 AND status IN ('Pending', 'Forwarded')
			`, approvedBy, time.Now(), appRemarks, awardID, approveStatus)
		}
	}
	// Execute the batch

	batchResults := ar.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	// Check if there were any updates
	for _, awardID := range awardIDs {
		// Execute the next result in the batch
		commandTag, err := batchResults.Exec()
		if err != nil {
			log.Error(ctx, "Failed to update records", "award_details_ids", awardID, "error", err)
			return "", err
		}
		// Check if no rows were affected
		if commandTag.RowsAffected() == 0 {
			return "", fmt.Errorf("no rows were updated, check the input")
		}
	}
	successMessage := fmt.Sprintf("Successfully processed award(s) %d", awardIDs)
	return successMessage, nil
}

func (ar *AwardRepository) GetAwards(ctx context.Context, table string, condition sq.Sqlizer, status string, skip uint64, limit uint64) ([]domain.EmpAwardDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, ar.cfg.GetDuration(dbQueryTimeoutMed))
	defer cancel()
	query := dblib.Psql.Select(
		"ad."+conEmpID,
		"ad.award_name",
		"ad.award_type",
		"ad.award_category",
		"ad.award_description",
		"ad.certificate_no",
		"ad.monetary_benefit",
		"ad."+conApprPostID,
		"ad."+conStatus,
		"ad.remarks",
		"ad.award_issue_date",
		"ad.award_for_year",
		"ad.award_details_id",
		"ad.admin_office",
		empFullName,
		"om."+conOfficeName,
		"em."+conEmpDesignation).
		From(table + " ad").
		Join(masterTable + " em ON em.employee_id = ad.employee_id").
		Join(officeTable + joinOfficeMasterCondition).
		Where(condition).
		OrderBy("ad.created_date DESC").
		Offset(uint64(skip * limit)).
		Limit(uint64(limit))

	if table == awardMakerTable {
		query = query.Column("ad.user_remarks").
			Column("ad.fwd_auth_remarks").
			Column("ad.approve_auth_remarks")
	}

	switch status {
	case unApproved:
		query = query.Where(sq.Or{sq.Eq{"ad." + conStatus: pending},
			sq.Eq{"ad." + conStatus: forwarded}})
	case "":
		// No filter — return all
	default:
		query = query.Where(sq.Eq{"ad." + conStatus: status})
	}

	return dblib.SelectRows(ctx, ar.db, query, pgx.RowToStructByNameLax[domain.EmpAwardDetails])
}

func (ar *AwardRepository) GetMakerAwardsByEmpID(ctx context.Context, empID int64, status string, skip uint64, limit uint64) ([]domain.EmpAwardDetails, error) {
	condition := sq.Eq{"ad." + conEmpID: empID}
	return ar.GetAwards(ctx, awardMakerTable, condition, status, skip, limit)
}

func (ar *AwardRepository) GetMakerAwardsByAOID(ctx context.Context, adminID int64, status string, skip uint64, limit uint64) ([]domain.EmpAwardDetails, error) {
	condition := sq.Eq{"ad.admin_office": adminID}
	return ar.GetAwards(ctx, awardMakerTable, condition, status, skip, limit)
}

func (ar *AwardRepository) GetMakerAwardsByAppPostID(ctx context.Context, postID string, status string, skip uint64, limit uint64) ([]domain.EmpAwardDetails, error) {
	condition := sq.Eq{"ad." + conApprPostID: postID}
	return ar.GetAwards(ctx, awardMakerTable, condition, status, skip, limit)
}

func (ar *AwardRepository) GetAwardsByEmpID(ctx context.Context, empID int64, skip uint64, limit uint64) ([]domain.EmpAwardDetails, error) {
	condition := sq.Eq{"ad." + conEmpID: empID}
	return ar.GetAwards(ctx, awardTable, condition, approved, skip, limit)
}

func (ar *AwardRepository) GetAwardsByAppPostID(ctx context.Context, postID string, skip uint64, limit uint64) ([]domain.EmpAwardDetails, error) {
	condition := sq.Eq{"ad." + conApprPostID: postID}
	return ar.GetAwards(ctx, awardTable, condition, approved, skip, limit)
}
