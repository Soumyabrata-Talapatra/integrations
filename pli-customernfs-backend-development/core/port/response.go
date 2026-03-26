package port

import "io"

// Standard status messages for all NFS operations
var (
	ListSuccess   = StatusCodeAndMessage{StatusCode: 200, Message: "list retrieved successfully", Success: true}
	FetchSuccess  = StatusCodeAndMessage{StatusCode: 200, Message: "data retrieved successfully", Success: true}
	CreateSuccess = StatusCodeAndMessage{StatusCode: 201, Message: "resource created successfully", Success: true}
	UpdateSuccess = StatusCodeAndMessage{StatusCode: 200, Message: "resource updated successfully", Success: true}
	DeleteSuccess = StatusCodeAndMessage{StatusCode: 200, Message: "resource deleted successfully", Success: true}

	// NFS-specific status constants
	// FR-NFS-001, FR-NFS-004: Aadhaar OTP operations
	OTPSuccess     = StatusCodeAndMessage{StatusCode: 200, Message: "OTP generated successfully", Success: true}
	OTPAuthSuccess = StatusCodeAndMessage{StatusCode: 200, Message: "OTP authenticated successfully", Success: true}

	// FR-NFS-007: Withdrawal
	WithdrawSuccess = StatusCodeAndMessage{StatusCode: 200, Message: "request withdrawn successfully", Success: true}

	// FR-NFS-006: Request initiated
	RequestInitiatedSuccess = StatusCodeAndMessage{StatusCode: 201, Message: "service request created successfully", Success: true}

	// FR-NFS-011: CPC action
	ActionSuccess = StatusCodeAndMessage{StatusCode: 200, Message: "action completed successfully", Success: true}
)

// StatusCodeAndMessage is embedded in all response structs.
// Provides consistent status code, success flag, and message.
type StatusCodeAndMessage struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

// Status returns HTTP status code (interface compliance)
func (s StatusCodeAndMessage) Status() int {
	return s.StatusCode
}

func (s StatusCodeAndMessage) ResponseType() string {
	return "standard"
}

func (s StatusCodeAndMessage) GetContentType() string {
	return "application/json"
}

func (s StatusCodeAndMessage) GetContentDisposition() string {
	return ""
}

func (s StatusCodeAndMessage) Object() []byte {
	return nil
}

// FileResponse is for file downloads (acknowledgment receipts, document downloads)
// Used by: GET /nfs/requests/{request_id}/receipt, GET /nfs/documents/{document_id}/download
type FileResponse struct {
	ContentDisposition string
	ContentType        string
	Data               []byte
	Reader             io.ReadCloser
}

func (s FileResponse) GetContentType() string {
	return s.ContentType
}

func (s FileResponse) GetContentDisposition() string {
	return s.ContentDisposition
}

func (s FileResponse) ResponseType() string {
	return "file"
}

func (s FileResponse) Status() int {
	return 200
}

func (s FileResponse) Object() []byte {
	return s.Data
}

func (s FileResponse) Stream(w io.Writer) error {
	if s.Reader == nil {
		if len(s.Data) > 0 {
			_, err := w.Write(s.Data)
			return err
		}
		return nil
	}
	defer s.Reader.Close()
	_, err := io.Copy(w, s.Reader)
	return err
}

// MetaDataResponse provides pagination metadata.
// Embed this in list response structs.
type MetaDataResponse struct {
	Skip                 uint64 `json:"skip"`
	Limit                uint64 `json:"limit"`
	OrderBy              string `json:"order_by,omitempty"`
	SortType             string `json:"sort_type,omitempty"`
	TotalRecordsCount    int    `json:"total_records_count,omitempty"`
	ReturnedRecordsCount uint64 `json:"returned_records_count"`
}

// NewMetaDataResponse creates a MetaDataResponse for list endpoints
func NewMetaDataResponse(skip, limit uint64, total int) MetaDataResponse {
	return MetaDataResponse{
		Skip:                 skip,
		Limit:                limit,
		TotalRecordsCount:    total,
		ReturnedRecordsCount: limit,
	}
}
