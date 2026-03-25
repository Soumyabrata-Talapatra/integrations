package domain

// import (
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )

// // TestSurrenderStatus_String tests status enum string conversion
// func TestSurrenderStatus_String(t *testing.T) {
// 	tests := []struct {
// 		status   SurrenderStatus
// 		expected string
// 	}{
// 		{SurrenderStatusPendingDocumentUpload, "PENDING_DOCUMENT_UPLOAD"},
// 		{SurrenderStatusPendingVerification, "PENDING_VERIFICATION"},
// 		{SurrenderStatusPendingApproval, "PENDING_APPROVAL"},
// 		{SurrenderStatusApproved, "APPROVED"},
// 		{SurrenderStatusRejected, "REJECTED"},
// 		{SurrenderStatusTerminated, "TERMINATED"},
// 	}

// 	for _, tt := range tests {
// 		t.Run(string(tt.status), func(t *testing.T) {
// 			assert.Equal(t, tt.expected, string(tt.status))
// 		})
// 	}
// }

// // TestSurrenderRequestType_Validation tests request type validation
// func TestSurrenderRequestType_Validation(t *testing.T) {
// 	validTypes := []SurrenderRequestType{
// 		SurrenderRequestTypeVoluntary,
// 		SurrenderRequestTypeForced,
// 	}

// 	for _, rt := range validTypes {
// 		assert.NotEmpty(t, string(rt))
// 		assert.Contains(t, []string{"VOLUNTARY", "FORCED"}, string(rt))
// 	}
// }

// // TestPolicySurrenderRequest_Validation tests domain model validation
// func TestPolicySurrenderRequest_Validation(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		request PolicySurrenderRequest
// 		isValid bool
// 	}{
// 		{
// 			name: "valid voluntary surrender",
// 			request: PolicySurrenderRequest{
// 				PolicyID:                "0000000000001",
// 				RequestNumber:           "SUR-TEST-001",
// 				RequestType:             SurrenderRequestTypeVoluntary,
// 				RequestDate:             time.Now(),
// 				GrossSurrenderValue:     50000,
// 				NetSurrenderValue:       45000,
// 				PaidUpValue:             40000,
// 				SurrenderFactor:         0.75,
// 				UnpaidPremiumsDeduction: 3000,
// 				LoanDeduction:           2000,
// 				DisbursementMethod:      DisbursementMethodCheque,
// 				DisbursementAmount:      45000,
// 				Status:                  SurrenderStatusPendingDocumentUpload,
// 				Owner:                   RequestOwnerCustomer,
// 				CreatedBy:               uuid.New(),
// 			},
// 			isValid: true,
// 		},
// 		{
// 			name: "invalid - negative values",
// 			request: PolicySurrenderRequest{
// 				PolicyID:            "0000000000002",
// 				RequestNumber:       "SUR-TEST-002",
// 				RequestType:         SurrenderRequestTypeVoluntary,
// 				GrossSurrenderValue: -1000, // Invalid
// 				NetSurrenderValue:   45000,
// 				Status:              SurrenderStatusPendingDocumentUpload,
// 				CreatedBy:           uuid.New(),
// 			},
// 			isValid: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Validate required fields
// 			if tt.isValid {
// 				assert.NotEqual(t, uuid.Nil, tt.request.PolicyID)
// 				assert.NotEmpty(t, tt.request.RequestNumber)
// 				assert.NotEmpty(t, tt.request.RequestType)
// 				assert.NotEmpty(t, tt.request.Status)
// 				assert.NotEqual(t, uuid.Nil, tt.request.CreatedBy)
// 				assert.GreaterOrEqual(t, tt.request.GrossSurrenderValue, float64(0))
// 				assert.GreaterOrEqual(t, tt.request.NetSurrenderValue, float64(0))
// 			}
// 		})
// 	}
// }

// // TestDocumentType_Validation tests document type enum
// func TestDocumentType_Validation(t *testing.T) {
// 	validTypes := []DocumentType{
// 		DocumentTypeWrittenConsent,
// 		DocumentTypePolicyBond,
// 		DocumentTypePremiumReceiptBook,
// 		DocumentTypeLoanBond,
// 		DocumentTypeAssignmentDeed,
// 		DocumentTypeIdentityProof,
// 		DocumentTypeBankDetails,
// 	}

// 	for _, dt := range validTypes {
// 		assert.NotEmpty(t, string(dt))
// 	}
// }

// // TestSurrenderDocument_Validation tests document model
// func TestSurrenderDocument_Validation(t *testing.T) {
// 	doc := SurrenderDocument{
// 		SurrenderRequestID: uuid.New(),
// 		DocumentType:       DocumentTypeWrittenConsent,
// 		DocumentName:       "consent.pdf",
// 		DocumentPath:       "/documents/surrender/consent.pdf",
// 		Verified:           false,
// 		UploadedDate:       time.Now(),
// 		Metadata:           map[string]interface{}{"size": 1024},
// 	}

// 	assert.NotEqual(t, uuid.Nil, doc.SurrenderRequestID)
// 	assert.NotEmpty(t, doc.DocumentType)
// 	assert.NotEmpty(t, doc.DocumentName)
// 	assert.NotEmpty(t, doc.DocumentPath)
// 	assert.NotZero(t, doc.UploadedDate)
// }

// // TestTaskStatus_Lifecycle tests task status transitions
// func TestTaskStatus_Lifecycle(t *testing.T) {
// 	// Valid lifecycle: PENDING -> RESERVED -> IN_PROGRESS -> COMPLETED
// 	lifecycle := []TaskStatus{
// 		TaskStatusPending,
// 		TaskStatusReserved,
// 		TaskStatusInProgress,
// 		TaskStatusCompleted,
// 	}

// 	for i, status := range lifecycle {
// 		assert.NotEmpty(t, string(status))
// 		if i > 0 {
// 			// Each status should be different
// 			assert.NotEqual(t, lifecycle[i-1], status)
// 		}
// 	}
// }

// // TestTaskPriority_Ordering tests priority levels
// func TestTaskPriority_Ordering(t *testing.T) {
// 	priorities := []TaskPriority{
// 		TaskPriorityLow,
// 		TaskPriorityNormal,
// 		TaskPriorityHigh,
// 		TaskPriorityCritical,
// 	}

// 	// Verify all priorities are defined
// 	for _, priority := range priorities {
// 		assert.NotEmpty(t, string(priority))
// 	}

// 	// Critical should have different value than Low
// 	assert.NotEqual(t, TaskPriorityLow, TaskPriorityCritical)
// }

// // TestApprovalWorkflowTask_Validation tests approval task model
// func TestApprovalWorkflowTask_Validation(t *testing.T) {
// 	task := ApprovalWorkflowTask{
// 		SurrenderRequestID: uuid.New(),
// 		Status:             TaskStatusPending,
// 		Priority:           TaskPriorityNormal,
// 		CreatedBy:          uuid.New(),
// 		CreatedAt:          time.Now(),
// 		Metadata:           map[string]interface{}{},
// 	}

// 	assert.NotEqual(t, uuid.Nil, task.SurrenderRequestID)
// 	assert.NotEmpty(t, task.Status)
// 	assert.NotEmpty(t, task.Priority)
// 	assert.NotEqual(t, uuid.Nil, task.CreatedBy)
// 	assert.NotZero(t, task.CreatedAt)
// }

// // TestDisbursementMethod_Validation tests disbursement methods
// func TestDisbursementMethod_Validation(t *testing.T) {
// 	methods := []DisbursementMethod{
// 		DisbursementMethodCash,
// 		DisbursementMethodCheque,
// 		DisbursementMethodBankTransfer,
// 	}

// 	for _, method := range methods {
// 		assert.NotEmpty(t, string(method))
// 		assert.Contains(t, []string{"CASH", "CHEQUE", "BANK_TRANSFER"}, string(method))
// 	}
// }

// // TestMetadata_JsonMarshaling tests metadata field
// func TestMetadata_JsonMarshaling(t *testing.T) {
// 	metadata := map[string]interface{}{
// 		"policy_number":     "PLI/2020/123456",
// 		"policyholder_name": "John Doe",
// 		"product_code":      "EA",
// 		"nested": map[string]interface{}{
// 			"key1": "value1",
// 			"key2": 123,
// 		},
// 	}

// 	// Verify metadata can store various types
// 	assert.Equal(t, "PLI/2020/123456", metadata["policy_number"])
// 	assert.Equal(t, "John Doe", metadata["policyholder_name"])

// 	nested, ok := metadata["nested"].(map[string]interface{})
// 	assert.True(t, ok)
// 	assert.Equal(t, "value1", nested["key1"])
// 	assert.Equal(t, 123, nested["key2"])
// }

// // TestForcedSurrenderReminder_Validation tests forced surrender reminder
// func TestForcedSurrenderReminder_Validation(t *testing.T) {
// 	reminder := ForcedSurrenderReminder{
// 		PolicyID:       "0000000000001",
// 		ReminderNumber: ReminderLevelFirst,
// 		ReminderDate:   time.Now(),
// 		Metadata:       map[string]interface{}{},
// 	}

// 	assert.NotEmpty(t, reminder.PolicyID)
// 	assert.Equal(t, ReminderLevelFirst, reminder.ReminderNumber)
// }

// // TestPaymentWindowStatus_Validation tests payment window status
// func TestPaymentWindowStatus_Validation(t *testing.T) {
// 	statuses := []PaymentWindowStatus{
// 		PaymentWindowStatusActive,
// 		PaymentWindowStatusExpired,
// 		PaymentWindowStatusPaid,
// 		PaymentWindowStatusCancelled,
// 	}

// 	for _, status := range statuses {
// 		assert.NotEmpty(t, string(status))
// 	}
// }
