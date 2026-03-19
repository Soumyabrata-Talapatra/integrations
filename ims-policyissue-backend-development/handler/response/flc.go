package response

import (
	"policy-issue-service/core/port"
)

// FLCInitiateResponse represents the response for Initiate FLC API
type FLCInitiateResponse struct {
	port.StatusCodeAndMessage
	FLCRequestID string `json:"flc_request_id"`
	Status       string `json:"status"`
}

// FLCApproveResponse represents the response for Approve FLC API
type FLCApproveResponse struct {
	port.StatusCodeAndMessage
	Status string `json:"status"`
}

// FLCRejectResponse represents the response for Reject FLC API
type FLCRejectResponse struct {
	port.StatusCodeAndMessage
	Status string `json:"status"`
}

// GetFLCStatusResponse represents the response for Get FLC Status API
type GetFLCStatusResponse struct {
	port.StatusCodeAndMessage
	Data interface{} `json:"data"`
}

// FLCQueueResponse represents the response for Get FLC Queue API
type FLCQueueResponse struct {
	port.StatusCodeAndMessage
	port.MetaDataResponse
	Data []interface{} `json:"data"`
}
