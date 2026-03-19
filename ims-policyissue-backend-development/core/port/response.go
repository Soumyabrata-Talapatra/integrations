package port

import "time"

// StatusCodeAndMessage is the standard response structure
type StatusCodeAndMessage struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

// MetaDataResponse for pagination metadata
type MetaDataResponse struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// NewMetaDataResponse creates pagination metadata
func NewMetaDataResponse(total int64, page, limit int) MetaDataResponse {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return MetaDataResponse{
		Total:       total,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
}

// ErrorResponse for API errors
type ErrorResponse struct {
	StatusCodeAndMessage
	ErrorCode string            `json:"error_code,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

// SuccessResponse for successful operations
type SuccessResponse struct {
	StatusCodeAndMessage
	Data interface{} `json:"data,omitempty"`
}

// ListResponse for paginated list results
type ListResponse struct {
	StatusCodeAndMessage
	MetaData MetaDataResponse `json:"metadata"`
	Data     interface{}      `json:"data"`
}

// CreatedResponse for resource creation
type CreatedResponse struct {
	StatusCodeAndMessage
	ID        interface{} `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
}

// UpdatedResponse for resource updates
type UpdatedResponse struct {
	StatusCodeAndMessage
	ID        interface{} `json:"id"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// DeletedResponse for resource deletion
type DeletedResponse struct {
	StatusCodeAndMessage
	ID interface{} `json:"id"`
}

// WorkflowStateResponse for workflow status
type WorkflowStateResponse struct {
	CurrentStep string                 `json:"current_step"`
	NextStep    string                 `json:"next_step"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
