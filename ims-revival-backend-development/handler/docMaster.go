package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"mime/multipart"
	"path"
	"plirevival/core/domain"
	"plirevival/core/port"
	"plirevival/handler/response"

	repo "plirevival/repo/postgres"
	"strconv"
	"time"

	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

type DocHandler struct {
	*serverHandler.Base
	dr *repo.DocRepository
}

func NewDocHandler(dr *repo.DocRepository) *DocHandler {
	base := serverHandler.New("DOCS").SetPrefix("/v1").AddPrefix("/files")
	return &DocHandler{
		base,
		dr,
	}
}
func (c *DocHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("", c.FileUpload).Name("Upload File"),
		serverRoute.GET("", c.DownloadSingleFile).Name("Download Single File"),

		// Zip Files
		serverRoute.GET("/zip-files", c.DownloadFile).Name("Download Zip File"),
	}
}

func generateUniqueID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10) // Example using current time in nanoseconds
}

type CreateFileReq struct {
	FolderPath    string                `form:"folderPath" validate:"required"`
	FileNameID    string                `form:"file_name_id" validate:"required"`
	EmployeeID    string                `form:"employee_id" validate:"required"`
	ServiceTypeID string                `form:"service_type_id" validate:"required"`
	File          *multipart.FileHeader `form:"file" validate:"required"`
}

func (emh *DocHandler) FileUpload(sctx *serverRoute.Context, req CreateFileReq) (*response.DocUploadResponse, error) {

	fileNameIDInt, err := strconv.ParseUint(req.FileNameID, 10, 64)
	if err != nil {
		// ctx.String(http.StatusBadRequest, fileNameIDShouldBeAnInt)
		return nil, fmt.Errorf("file_name_id should be an integer")
	}

	// employeeIDInt, err := strconv.Atoi(employeeID)
	employeeIDInt, err := strconv.ParseInt(req.EmployeeID, 10, 64)
	if err != nil {
		// ctx.String(http.StatusBadRequest, empIDShouldBeAnIinteger)
		return nil, fmt.Errorf("employee_id should be an integer")
	}

	serviceTypeIDInt, err := strconv.Atoi(req.ServiceTypeID)
	if err != nil {
		// ctx.String(http.StatusBadRequest, "service_type_id should be an integer")
		return nil, fmt.Errorf("service_type_id should be an integer")
	}

	var header *multipart.FileHeader
	header = req.File
	file, err := req.File.Open()
	if err != nil {
		return nil, fmt.Errorf("file couldn't be opened")
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Debug(sctx.Ctx, "Error in closing file")
		}
	}(file)

	// File size validation
	const maxFileSize = 2 * 1024 * 1024 // 2MB
	log.Info(sctx.Ctx, "Validating file size... Max file size: %d bytes", maxFileSize)
	if header.Size > maxFileSize {
		// ctx.String(http.StatusBadRequest, "File size exceeds the maximum limit of 2 MB")
		return nil, fmt.Errorf("file size exceeds 2MB")
	}

	// File type validation
	log.Debug(sctx.Ctx, "Validating file type...")
	allowedMimeTypes := map[string]bool{
		"jpeg": true,
		"png":  true,
		"pdf":  true,
	}
	fileType := header.Header.Get(ContentType)
	fileExt := filepath.Ext(header.Filename)

	if !allowedMimeTypes[strings.ToLower(strings.TrimPrefix(fileExt, "."))] {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(apierrors.CustomError, "file type is not allowed", nil)
		return nil, errMsg
	}

	// Generate custom filename
	log.Debug(sctx.Ctx, "Generating custom filename...")
	ext := path.Ext(header.Filename)
	uniqueID := generateUniqueID() // Implement this function to generate a unique number or UUID

	filename := fmt.Sprintf("%s_%s_%s_%s%s", req.EmployeeID, req.ServiceTypeID, req.FileNameID, uniqueID, ext)

	objectName := path.Join(req.FolderPath, filename)
	err = emh.dr.UploadFile(file, objectName, fileType, header.Size)

	if err != nil {
		// ctx.String(http.StatusInternalServerError, "Failed to upload file")
		return nil, fmt.Errorf("Failed to upload file")
	}
	log.Debug(sctx.Ctx, "Uploading file on server is completed...")

	// Insert details into the document_master_pis table
	log.Debug(sctx.Ctx, "Inserting details into the document_master_pis table...")
	document := domain.Document{
		FileNameID:           fileNameIDInt,
		EmployeeID:           employeeIDInt,
		ServiceTypeID:        serviceTypeIDInt,
		DocumentName:         filename,
		DocumentType:         fileType,
		DocumentSize:         header.Size,
		DocumentFilePath:     objectName,
		DocumentUploadStatus: "uploaded", // Example status
		DocumentUploadedBy:   "user_id",  // Retrieve the actual user ID from context or session
		DocumentUploadedDate: time.Now(),
	}
	err = emh.dr.InsertFile(sctx.Ctx, document)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to insert document record: %v", err)
		// ctx.String(http.StatusInternalServerError, "Failed to insert document record")
		return nil, fmt.Errorf("Failed to insert document record")
	}

	// Construct response
	log.Debug(sctx.Ctx, "Constructing response...")

	// Return a success response with the created data
	rsp := response.NewDocumentResponse(document)
	resp := response.DocUploadResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 rsp,
	}
	log.Debug(sctx.Ctx, "FileUpload response:%v", resp)
	//handleSuccess(sctx.Ctx, resp)
	return &resp, nil
}

type GetFileReq struct {
	FolderPath    string `form:"folderPath"`
	FileNameID    string `form:"file_name_id" validate:"required"`
	EmployeeID    string `form:"employee_id" validate:"required"`
	ServiceTypeID string `form:"service_type_id"`
	// File          *multipart.FileHeader `form:"file" validate:"required"`
}

func (emh *DocHandler) DownloadSingleFile(sctx *serverRoute.Context, req GetFileReq) (*port.FileResponse, error) {

	// sctx.Ctx = log.WithTags(sctx.Ctx, "doc-handler", "download-single-file")

	// log.DebugEvent(sctx.Ctx).Msg("DownloadSingleFile called with parameters")
	log.Debug(sctx.Ctx, "GetFileReq: %+v", req)
	fmt.Println("GetFileReq:", req)

	// employeeIDStr := sctx.Ctx.Query("employee_id")
	if req.EmployeeID == "" {
		// sctx.Ctx.String(http.StatusBadRequest, "employee_id is required")
		return nil, fmt.Errorf("employee_id is required")
	}

	// serviceTypeIDStr := sctx.Ctx.Query("service_type_id")
	if req.ServiceTypeID == "" {
		// sctx.Ctx.String(http.StatusBadRequest, "service_type_id is required")
		return nil, fmt.Errorf("service_type_id is required")
	}

	employeeID, err := strconv.Atoi(req.EmployeeID)
	if err != nil {
		// sctx.Ctx.String(http.StatusBadRequest, "employee_id should be an integer")
		return nil, fmt.Errorf("employee_id should be an integer")
	}

	serviceTypeID, err := strconv.Atoi(req.ServiceTypeID)
	if err != nil {
		// sctx.Ctx.String(http.StatusBadRequest, "service_type_id should be an integer")
		return nil, fmt.Errorf("service_type_id should be an integer")
	}

	var fileNameID *int
	// fileNameIDStr := sctx.Ctx.Query("file_name_id")
	if req.FileNameID != "" {
		id, err := strconv.Atoi(req.FileNameID)
		if err != nil {
			// sctx.Ctx.String(http.StatusBadRequest, "file_name_id should be an integer if provided")
			return nil, fmt.Errorf("file_name_id should be an integer if provided")
		}
		fileNameID = &id
	}

	// ctx := sctx.Ctx.Request.Context()
	var document *domain.Document
	if fileNameID != nil {
		document, err = emh.dr.GetSingleDocumentByFileNameID(sctx.Ctx, employeeID, *fileNameID, serviceTypeID)
	} else {
		document, err = emh.dr.GetLatestDocumentByEmpAndServiceType(sctx.Ctx, employeeID, serviceTypeID)
	}
	if err != nil {
		log.Error(sctx.Ctx, "Failed to fetch document metadata", "error", err)
		// sctx.Ctx.String(http.StatusInternalServerError, "Failed to fetch document metadata")
		return nil, fmt.Errorf("Failed to fetch document metadata")
	}

	if document == nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(apierrors.DBErrorRecordNotFound, "No file found for the given parameters", err)
		return nil, errMsg
	}

	object, err := emh.dr.DownloadFile(document.DocumentFilePath)
	if err != nil {
		if err.Error() == "file not found" {
			// sctx.Ctx.String(http.StatusNotFound, "File not found: "+document.DocumentFilePath)
			return nil, fmt.Errorf("File not found: %s", document.DocumentFilePath)
		}
		log.Error(sctx.Ctx, "Failed to download file", "file", document.DocumentFilePath, "error", err)
		// sctx.Ctx.String(http.StatusInternalServerError, "Failed to download file: "+document.DocumentFilePath)
		return nil, fmt.Errorf("Failed to download file: %s", document.DocumentFilePath)
	}

	// Read the full object into memory as []byte
	// if rc, ok := object.(io.ReadCloser); ok {
	// 	defer rc.Close()
	// }
	// data, readErr := io.ReadAll(object)
	// if readErr != nil {
	// 	log.Error(sctx.Ctx, "Failed to read file content", "file", document.DocumentFilePath, "error", readErr)
	// 	return nil, fmt.Errorf("Failed to read file content: %s", document.DocumentFilePath)
	// }

	res := port.FileResponse{
		ContentType:        "application/octet-stream",
		ContentDisposition: fmt.Sprintf("attachment; filename=\"%s\"", path.Base(document.DocumentFilePath)),
		// Data:               data,
		Reader: io.NopCloser(object),
	}

	// sctx.Ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(document.DocumentFilePath)))
	// sctx.Ctx.Header("Content-Type", "application/octet-stream")
	// if _, err := io.Copy(sctx.Ctx.Writer, object); err != nil {
	// 	log.Error(sctx.Ctx, "Failed to write file to response", "file", document.DocumentFilePath, "error", err)
	// 	// sctx.Ctx.String(http.StatusInternalServerError, "Failed to write file to response: "+document.DocumentFilePath)
	// 	return nil, fmt.Errorf("Failed to write file to response: " + document.DocumentFilePath)
	// }

	return &res, nil

}

func (emh *DocHandler) DownloadFile(sctx *serverRoute.Context, req GetFileReq) (*port.FileResponse, error) {
	// employeeIDStr := c.Query("employee_id")
	if req.EmployeeID == "" {
		// c.String(http.StatusBadRequest, empIDRequired)
		return nil, fmt.Errorf("employee_id is required")
	}
	// fileNameIDStr := c.Query("file_name_id")
	if req.FileNameID == "" {
		// c.String(http.StatusBadRequest, fileNameIDRequired)
		return nil, fmt.Errorf("file_name_id is required")
	}
	employeeID, err := strconv.Atoi(req.EmployeeID)
	if err != nil {
		// c.String(http.StatusBadRequest, empIDShouldBeAnIinteger)
		return nil, fmt.Errorf("employee_id should be an integer")
	}

	// Parse file_name_id as integer
	fileNameID, err := strconv.Atoi(req.FileNameID)
	if err != nil {
		// c.String(http.StatusBadRequest, fileNameIDShouldBeAnInt)
		return nil, fmt.Errorf("file_name_id should be an integer")
	}

	// Fetch all document metadata from the database for the given file_name_id
	// ctx := c.Request.Context()
	documents, err := emh.dr.GetDocumentsByFileNameID(sctx.Ctx, employeeID, fileNameID)
	if err != nil {
		log.Error(sctx.Ctx, FailedToFetchDocMetaData, "error", err)
		// c.String(http.StatusInternalServerError, failedToFetchDocMetaData)
		return nil, fmt.Errorf("Failed to fetch document metadata")
	}

	if len(documents) == 0 {
		log.Debug(sctx.Ctx, "No files found for the given ", "employee_id", "file_name_id", employeeID, fileNameID)
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No files found for the given parameters", err)
		// c.String(http.StatusNotFound, "No files found for the given file_name_id")
		return nil, errMsg
	}

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add files to the zip
	for _, document := range documents {
		object, err := emh.dr.DownloadFile(document.DocumentFilePath)
		if err != nil {
			if err.Error() == "file not found" {
				log.Debug(sctx.Ctx, "File not found", "file", document.DocumentFilePath)
				continue
			} else {
				log.Error(sctx.Ctx, "Failed to download file", "file", document.DocumentFilePath, "error", err)
				// c.String(http.StatusInternalServerError, "Failed to download file: "+document.DocumentFilePath)
				return nil, fmt.Errorf("Failed to download file: %s", document.DocumentFilePath)
			}
		}

		// Create a zip file entry
		zipFileWriter, err := zipWriter.Create(path.Base(document.DocumentFilePath))
		if err != nil {
			log.Error(sctx.Ctx, "Failed to create zip entry", "file", document.DocumentFilePath, "error", err)
			// c.String(http.StatusInternalServerError, "Failed to create zip entry for file: "+document.DocumentFilePath)
			return nil, fmt.Errorf("Failed to create zip entry for file: %s", document.DocumentFilePath)
		}

		// Copy file content to zip entry
		if _, err := io.Copy(zipFileWriter, object); err != nil {
			log.Error(sctx.Ctx, "Failed to write file to zip", "file", document.DocumentFilePath, "error", err)
			// c.String(http.StatusInternalServerError, "Failed to write file to zip: "+document.DocumentFilePath)
			return nil, fmt.Errorf("Failed to write file to zip: %s", document.DocumentFilePath)
		}
	}

	// Close the zip writer to flush the buffer
	if err := zipWriter.Close(); err != nil {
		log.Error(sctx.Ctx, "Failed to close zip writer", "error", err)
		// c.String(http.StatusInternalServerError, "Failed to close zip writer")
		return nil, fmt.Errorf("Failed to close zip writer")
	}

	res := port.FileResponse{
		ContentType:        "application/zip",
		ContentDisposition: "attachment; filename=\"pisdocuments.zip\"",
		Data:               buf.Bytes(),
	}

	// Set headers and send the zip file in the response
	// sctx.Ctx.Header("Content-Disposition", "attachment; filename=\"pisdocuments.zip\"")
	// sctx.Ctx.Header(contentType, "application/zip")
	// sctx.Ctx.Data(http.StatusOK, "application/zip", buf.Bytes())

	return &res, nil
}
