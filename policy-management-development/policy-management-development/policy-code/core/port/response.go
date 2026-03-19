package port

import "io"

// ============================================================================
// Port Layer — Common Response Structures
// Source: template.md Section 6 (Port Layer)
// ============================================================================

// Standard status messages for all PM operations.
// Use these constants in all response structs.
var (
	ListSuccess   = StatusCodeAndMessage{StatusCode: 200, Message: "list retrieved successfully", Success: true}
	FetchSuccess  = StatusCodeAndMessage{StatusCode: 200, Message: "data retrieved successfully", Success: true}
	CreateSuccess = StatusCodeAndMessage{StatusCode: 201, Message: "resource created successfully", Success: true}
	UpdateSuccess = StatusCodeAndMessage{StatusCode: 200, Message: "resource updated successfully", Success: true}
	DeleteSuccess = StatusCodeAndMessage{StatusCode: 200, Message: "resource deleted successfully", Success: true}

	// PM-specific success messages
	AcceptedSuccess  = StatusCodeAndMessage{StatusCode: 202, Message: "request accepted for processing", Success: true}
	WithdrawnSuccess = StatusCodeAndMessage{StatusCode: 200, Message: "request withdrawn successfully", Success: true}
)

// StatusCodeAndMessage is embedded in all response structs.
// Provides consistent status code, success flag, and message.
type StatusCodeAndMessage struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

// Status returns HTTP status code (interface compliance).
func (s StatusCodeAndMessage) Status() int {
	return s.StatusCode
}

// ResponseType returns the response type identifier.
func (s StatusCodeAndMessage) ResponseType() string {
	return "standard"
}

// GetContentType returns the content type for this response.
func (s StatusCodeAndMessage) GetContentType() string {
	return "application/json"
}

// GetContentDisposition returns empty (not a file download).
func (s StatusCodeAndMessage) GetContentDisposition() string {
	return ""
}

// Object returns nil (not a binary payload).
func (s StatusCodeAndMessage) Object() []byte {
	return nil
}

// FileResponse for file downloads/uploads.
type FileResponse struct {
	ContentDisposition string
	ContentType        string
	Data               []byte        // Memory-based payload
	Reader             io.ReadCloser // Optional streaming source
}

// GetContentType returns the content type.
func (s FileResponse) GetContentType() string {
	return s.ContentType
}

// GetContentDisposition returns the content disposition header.
func (s FileResponse) GetContentDisposition() string {
	return s.ContentDisposition
}

// ResponseType identifies this as a file response.
func (s FileResponse) ResponseType() string {
	return "file"
}

// Status returns 200 for file responses.
func (s FileResponse) Status() int {
	return 200
}

// Object returns the raw data bytes.
func (s FileResponse) Object() []byte {
	return s.Data
}

// Stream copies Reader to w if available; else writes Data.
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
	Skip                 uint64 `json:"skip,default=0"`
	Limit                uint64 `json:"limit,default=10"`
	OrderBy              string `json:"order_by,omitempty"`
	SortType             string `json:"sort_type,omitempty"`
	TotalRecordsCount    int    `json:"total_records_count,omitempty"`
	ReturnedRecordsCount uint64 `json:"returned_records_count"`
}

// NewMetaDataResponse creates a pagination metadata response.
func NewMetaDataResponse(skip, limit, total uint64) MetaDataResponse {
	return MetaDataResponse{
		Skip:                 skip,
		Limit:                limit,
		TotalRecordsCount:    int(total),
		ReturnedRecordsCount: limit,
	}
}
