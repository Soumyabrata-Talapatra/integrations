package response

import (
	"policy-issue-service/core/port"
)

// QRApproveResponse represents the response for QR approval API
type QRApproveResponse struct {
	port.StatusCodeAndMessage
	Status string `json:"status"`
}

// QRRejectResponse represents the response for QR rejection API
type QRRejectResponse struct {
	port.StatusCodeAndMessage
	Status string `json:"status"`
}

// QRReturnResponse represents the response for QR return API
type QRReturnResponse struct {
	port.StatusCodeAndMessage
	Status string `json:"status"`
}

// ApproverApproveResponse represents the response for Approver approval API
type ApproverApproveResponse struct {
	port.StatusCodeAndMessage
	Status string `json:"status"`
}

// ApproverRejectResponse represents the response for Approver rejection API
type ApproverRejectResponse struct {
	port.StatusCodeAndMessage
	Status string `json:"status"`
}

type ResubmitResponse struct {
    port.StatusCodeAndMessage
    Status string `json:"status"`
}
