package repo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"plirevival/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
)

type DocRepository struct {
	db    *dblib.DB
	cfg   *config.Config
	minio *minio.Client
}

// / NewDocRepository creates a new user repository instance
func NewDocRepository(db *dblib.DB, cfg *config.Config, minio *minio.Client) *DocRepository {
	return &DocRepository{
		db,
		cfg,
		minio,
	}
}

func (dr *DocRepository) UploadFile(file multipart.File, objectName, contentType string, size int64) error {
	// bucketName := config.GetBucketName()
	bucketName := dr.cfg.GetString(minioBucketName)

	if dr.minio == nil {

		return fmt.Errorf("MinIO client is not initialized")
	}
	_, err := dr.minio.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}

func (dr *DocRepository) DownloadFile(objectName string) (io.Reader, error) {
	bucketName := dr.cfg.GetString(minioBucketName)
	object, err := dr.minio.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	_, err = object.Stat()
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	return object, nil
}

// GetDocumentsByFileNameID retrieves documents by their file_name_id from the database
func (r *DocRepository) GetDocumentsByFileNameID(ctx context.Context, employeeID int, fileNameID int) ([]domain.Document, error) {
	query := dblib.Psql.Select("employee_id", "file_name_id", "document_name", "document_file_path").
		From(docTable).
		Where(sq.And{
			sq.Eq{"employee_id": employeeID},
			sq.Eq{"file_name_id": fileNameID}})
	return dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Document])
}

func (dr *DocRepository) InsertFile(ctx context.Context, document domain.Document) error {
	// Prepare the SQL query using Squirrel
	query := dblib.Psql.Insert(docTable).
		Columns(
			"file_name_id", "employee_id", "service_type_id", "document_name", "document_type", "document_size",
			"document_approver_post_id", "document_upload_status", "document_uploaded_by",
			"document_uploaded_date", "document_updated_by", "document_updated_date", "document_approved_by",
			"document_approved_date", "remarks", "document_file_path", "approve_status").
		Values(
			document.FileNameID, document.EmployeeID, document.ServiceTypeID, document.DocumentName, document.DocumentType, document.DocumentSize,
			document.DocumentApproverPostID, document.DocumentUploadStatus, document.DocumentUploadedBy, document.DocumentUploadedDate,
			document.DocumentUpdatedBy, document.DocumentUpdatedDate, document.DocumentApprovedBy, document.DocumentApprovedDate,
						document.Remarks, document.DocumentFilePath, pending).
		Suffix("RETURNING document_id") // Specify to return document_id

	sql, args, err := query.ToSql()
	if err != nil {
		log.Debug(nil, "Error building SQL query:", err)
		return err
	}

	// Execute the query and scan the result
	var documentID int
	err = dr.db.QueryRow(ctx, sql, args...).Scan(&documentID)
	if err != nil {
		log.Debug(nil, "Error executing query:", err)
		return err
	}

	// Log the document uploaded
	log.Debug(nil, "Document uploaded successfully with ID:", documentID)

	return nil // Return nil after successful insertion and obtaining document_id
}

func (dr *DocRepository) GetSingleDocumentByFileNameID(ctx context.Context, employeeID int, fileNameID int, serviceTypeID int) (*domain.Document, error) {
	// psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Build the query to select a single document
	query := dblib.Psql.Select("employee_id", "file_name_id", "service_type_id", "document_name", "document_file_path").
		From(docTable).
		Where(sq.And{
				sq.Eq{"employee_id": employeeID},
				sq.Eq{"file_name_id": fileNameID},
				sq.Eq{"service_type_id": serviceTypeID}}).
		Limit(1) // Limit the result to a single document

	// Fetch the document using SelectRow
	document, err := dblib.SelectOne(ctx, dr.db, query, pgx.RowToStructByNameLax[domain.Document])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No record is not an error
	}
	if err != nil {
		return nil, err
	}

	return &document, nil
}
func (dr *DocRepository) GetLatestDocumentByEmpAndServiceType(ctx context.Context, employeeID int, serviceTypeID int) (*domain.Document, error) {
	// Build the query to fetch the latest uploaded document for the employee and service type
	query := dblib.Psql.
		Select("employee_id", "service_type_id", "document_name", "document_file_path").
		From(docTable).
		Where(sq.And{
			sq.Eq{"employee_id": employeeID},
			sq.Eq{"service_type_id": serviceTypeID},
		}).
		OrderBy("document_uploaded_date DESC", "document_id DESC").
		Limit(1)

	// Execute the query and scan into Document struct
	document, err := dblib.SelectOne(ctx, dr.db, query, pgx.RowToStructByNameLax[domain.Document])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No record is not an error
	}
	if err != nil {
		return nil, err
	}
	fmt.Println("Doc Name:", document.DocumentName)

	return &document, nil
}
