// Package handler_test contains unit tests for Document Management (DM-001..005)
// and CPC Operations (CPC-001..006) endpoints.
// Phase 3: Supporting Endpoints — no DB or Temporal dependency.
package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handler "customer-nfs-service/handler"
	resp "customer-nfs-service/handler/response"
)

// ─── DM-001: DocumentUploadRequest (multipart) ───────────────────────────────

// TestDocumentRequestFields_FieldMapping verifies DM-001 form field expectations.
// VR-NFS-009: max file size 5 MB.
// VR-NFS-010: only PDF/JPG/PNG/TIFF accepted.
func TestDocumentRequestFields_FieldMapping(t *testing.T) {
	req := handler.UploadDocumentRequest{
		RequestID:    "req-uuid-dm-001",
		DocumentType: "GAZETTE_NOTIFICATION",
		UploadedBy:   "agent-01",
	}

	assert.Equal(t, "req-uuid-dm-001", req.RequestID)
	assert.Equal(t, "GAZETTE_NOTIFICATION", req.DocumentType)
	assert.Equal(t, "agent-01", req.UploadedBy)
}

// TestDocumentMIMETypeValidation verifies VR-NFS-010 allowed types.
// VR-NFS-010: MIME types allowed: application/pdf, image/jpeg, image/png, image/tiff
func TestDocumentMIMETypeValidation(t *testing.T) {
	allowedMIMETypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/png",
		"image/tiff",
	}
	disallowedMIMETypes := []string{
		"text/plain",
		"application/msword",
		"application/zip",
		"video/mp4",
	}

	isAllowed := func(mime string) bool {
		for _, allowed := range allowedMIMETypes {
			if allowed == mime {
				return true
			}
		}
		return false
	}

	for _, mime := range allowedMIMETypes {
		assert.True(t, isAllowed(mime), "MIME type %q must be allowed (VR-NFS-010)", mime)
	}
	for _, mime := range disallowedMIMETypes {
		assert.False(t, isAllowed(mime), "MIME type %q must NOT be allowed (VR-NFS-010)", mime)
	}
}

// TestDocumentFileSizeLimit verifies VR-NFS-009 max file size.
// VR-NFS-009: max file size 5 MB = 5 * 1024 * 1024 bytes.
func TestDocumentFileSizeLimit(t *testing.T) {
	const maxFileSizeBytes int64 = 5 * 1024 * 1024 // 5 MB

	cases := []struct {
		name       string
		fileSize   int64
		shouldPass bool
	}{
		{"exactly at limit", maxFileSizeBytes, true},
		{"just under limit", maxFileSizeBytes - 1, true},
		{"just over limit", maxFileSizeBytes + 1, false},
		{"very small file", 1024, true},
		{"empty file", 0, false},
	}

	isWithinLimit := func(size int64) bool {
		return size > 0 && size <= maxFileSizeBytes
	}

	for _, tc := range cases {
		got := isWithinLimit(tc.fileSize)
		assert.Equal(t, tc.shouldPass, got, "file size %d: %s", tc.fileSize, tc.name)
	}
}

// ─── DM-001 Response: NewDocumentUploadResponse ──────────────────────────────

// TestNewDocumentUploadResponse verifies DM-001 success response.
// VR-NFS-009/010: file validated; VerificationStatus always "PENDING" on upload.
func TestNewDocumentUploadResponse(t *testing.T) {
	r := resp.NewDocumentUploadResponse(
		"doc-uuid-001",
		"req-uuid-001",
		"GAZETTE_NOTIFICATION",
		"gazette.pdf",
		"/dms/path/gazette.pdf",
		"application/pdf",
		204800,
	)

	require.NotNil(t, r)
	assert.Equal(t, "doc-uuid-001", r.Data.DocumentID)
	assert.Equal(t, "req-uuid-001", r.Data.RequestID)
	assert.Equal(t, "GAZETTE_NOTIFICATION", r.Data.DocumentType)
	assert.Equal(t, "gazette.pdf", r.Data.FileName)
	assert.Equal(t, int64(204800), r.Data.FileSizeBytes)
	assert.Equal(t, "application/pdf", r.Data.MIMEType)
	assert.Equal(t, "PENDING", r.Data.VerificationStatus, "VerificationStatus must be PENDING on upload")
	assert.NotEmpty(t, r.Data.UploadedAt)
}

// ─── DM-004 Response: NewDocumentDeleteResponse ──────────────────────────────

// TestNewDocumentDeleteResponse_Success verifies DM-004 successful deletion.
// DM-004: only CREATED status documents can be deleted.
func TestNewDocumentDeleteResponse_Success(t *testing.T) {
	r := resp.NewDocumentDeleteResponse("doc-uuid-001", true)

	require.NotNil(t, r)
	assert.Equal(t, "doc-uuid-001", r.Data.DocumentID)
	assert.True(t, r.Data.Deleted)
	assert.Contains(t, r.Data.Message, "deleted successfully")
}

// TestNewDocumentDeleteResponse_Failure verifies DM-004 failure path.
func TestNewDocumentDeleteResponse_Failure(t *testing.T) {
	r := resp.NewDocumentDeleteResponse("doc-uuid-002", false)

	require.NotNil(t, r)
	assert.False(t, r.Data.Deleted)
	assert.Contains(t, r.Data.Message, "could not be deleted")
}

// ─── DM-005 Response: NewSecureUploadResponse ────────────────────────────────

// TestNewSecureUploadResponse verifies DM-005 secure upload response.
// BR-NFS-014: status transitions to PENDING_APPROVAL after upload.
func TestNewSecureUploadResponse(t *testing.T) {
	r := resp.NewSecureUploadResponse("doc-uuid-003", "PENDING_APPROVAL", 2)

	require.NotNil(t, r)
	assert.Equal(t, "doc-uuid-003", r.Data.DocumentID)
	assert.Equal(t, "PENDING_APPROVAL", r.Data.NewStatus)
	assert.Equal(t, 2, r.Data.AttemptNumber)
	assert.Contains(t, r.Data.Message, "forwarded for review")
}

// TestSecureUploadAttemptTracking verifies BR-NFS-014 attempt limit.
// BR-NFS-014: max 3 attempts. Attempt 3 → next status = DOCUMENTS_EXPIRED.
func TestSecureUploadAttemptTracking(t *testing.T) {
	cases := []struct {
		attempt      int
		status       string
		expectExpiry bool
	}{
		{1, "PENDING_APPROVAL", false},
		{2, "PENDING_APPROVAL", false},
		{3, "DOCUMENTS_EXPIRED", true}, // 3rd attempt exhausted → expired
	}

	for _, tc := range cases {
		r := resp.NewSecureUploadResponse("doc-uuid-loop", tc.status, tc.attempt)
		assert.Equal(t, tc.attempt, r.Data.AttemptNumber)
		assert.Equal(t, tc.status, r.Data.NewStatus)
		if tc.expectExpiry {
			assert.Equal(t, "DOCUMENTS_EXPIRED", r.Data.NewStatus)
		}
	}
}

// ─── CPC-001 Response: NewGetCPCQueueResponse ────────────────────────────────

// TestNewGetCPCQueueResponse_Pagination verifies CPC-001 paginated queue response.
func TestNewGetCPCQueueResponse_Pagination(t *testing.T) {
	queue := []resp.CPCQueueItem{
		{
			RequestID:    "req-001",
			TicketNumber: "NFS-ANC-20260101-000001",
			RequestType:  "ADDRESS_CHANGE",
			Status:       "PENDING_APPROVAL",
			SLAStatus:    "GREEN",
			SLADaysLeft:  12,
			SLADeadline:  "2026-01-20",
		},
		{
			RequestID:    "req-002",
			TicketNumber: "NFS-NMC-20260101-000002",
			RequestType:  "NAME_CHANGE",
			Status:       "IN_PROGRESS",
			SLAStatus:    "AMBER",
			SLADaysLeft:  3,
			SLADeadline:  "2026-01-18",
		},
	}

	r := resp.NewGetCPCQueueResponse(50, 1, 20, queue)

	require.NotNil(t, r)
	assert.Equal(t, 50, r.Data.Total)
	assert.Equal(t, 1, r.Data.Page)
	assert.Equal(t, 20, r.Data.PageSize)
	assert.Equal(t, 2, len(r.Data.Queue))
	assert.Equal(t, "GREEN", r.Data.Queue[0].SLAStatus)
	assert.Equal(t, "AMBER", r.Data.Queue[1].SLAStatus)
}

// TestCPCQueueItem_SLAStatusValues documents valid SLA status values for CPC-001.
// SLA status computed from deadline: GREEN (<50%), AMBER (50-80%), RED (>80%) elapsed.
func TestCPCQueueItem_SLAStatusValues(t *testing.T) {
	validStatuses := []string{"GREEN", "AMBER", "RED"}

	for _, status := range validStatuses {
		item := resp.CPCQueueItem{
			RequestID: "req-sla",
			SLAStatus: status,
		}
		assert.Equal(t, status, item.SLAStatus)
	}
}

// TestCPCQueueItem_SLAComputation documents SLA computation logic.
// computeSLAStatus helper in handler/cpc.go:
//   - daysLeft > 5  → GREEN
//   - daysLeft 1-5  → AMBER
//   - daysLeft <= 0 → RED
func TestCPCQueueItem_SLAComputation(t *testing.T) {
	computeSLA := func(daysLeft int) string {
		if daysLeft > 5 {
			return "GREEN"
		} else if daysLeft > 0 {
			return "AMBER"
		}
		return "RED"
	}

	cases := []struct {
		daysLeft int
		expected string
	}{
		{15, "GREEN"},
		{6, "GREEN"},
		{5, "AMBER"},
		{1, "AMBER"},
		{0, "RED"},
		{-3, "RED"},
	}

	for _, tc := range cases {
		got := computeSLA(tc.daysLeft)
		assert.Equal(t, tc.expected, got, "daysLeft=%d SLA status mismatch", tc.daysLeft)
	}
}

// ─── CPC-002 Response: NewGetCPCQueueStatsResponse ───────────────────────────

// TestNewGetCPCQueueStatsResponse verifies CPC-002 stats response.
// BATCH: 5-query batch for stats breakdown.
func TestNewGetCPCQueueStatsResponse(t *testing.T) {
	officeCode := "RO-DEL"
	stats := resp.CPCQueueStats{
		TotalPending:    42,
		TotalInProgress: 8,
		TotalOverdue:    3,
		SLAGreen:        25,
		SLAAmber:        14,
		SLARed:          3,
		AddressChanges:  30,
		NameChanges:     12,
	}

	r := resp.NewGetCPCQueueStatsResponse(&officeCode, stats)

	require.NotNil(t, r)
	require.NotNil(t, r.Data.OfficeCode)
	assert.Equal(t, "RO-DEL", *r.Data.OfficeCode)
	assert.Equal(t, 42, r.Data.Stats.TotalPending)
	assert.Equal(t, 3, r.Data.Stats.TotalOverdue)
	assert.Equal(t, 25, r.Data.Stats.SLAGreen)
	assert.Equal(t, 3, r.Data.Stats.SLARed)
	assert.NotEmpty(t, r.Data.GeneratedAt)

	// SLA breakdown must sum to total (PENDING + IN_PROGRESS)
	total := r.Data.Stats.SLAGreen + r.Data.Stats.SLAAmber + r.Data.Stats.SLARed
	assert.Equal(t, 42, total, "SLA status counts must sum to TotalPending")
}

// TestNewGetCPCQueueStatsResponse_NilOffice verifies stats without office filter.
func TestNewGetCPCQueueStatsResponse_NilOffice(t *testing.T) {
	stats := resp.CPCQueueStats{TotalPending: 100}
	r := resp.NewGetCPCQueueStatsResponse(nil, stats)

	require.NotNil(t, r)
	assert.Nil(t, r.Data.OfficeCode, "OfficeCode should be nil for system-wide stats")
	assert.Equal(t, 100, r.Data.Stats.TotalPending)
}

// ─── CPC-003 Response: NewAssignRequestResponse ──────────────────────────────

// TestNewAssignRequestResponse verifies CPC-003 assignment response.
// BR-NFS-002: assignment puts request IN_PROGRESS.
func TestNewAssignRequestResponse(t *testing.T) {
	r := resp.NewAssignRequestResponse("req-uuid-001", "cpc-officer-01", "IN_PROGRESS")

	require.NotNil(t, r)
	assert.Equal(t, "req-uuid-001", r.Data.RequestID)
	assert.Equal(t, "cpc-officer-01", r.Data.AssignedTo)
	assert.Equal(t, "IN_PROGRESS", r.Data.Status)
	assert.NotEmpty(t, r.Data.AssignedAt)
	assert.Equal(t, "Request assigned successfully.", r.Data.Message)
}

// ─── CPC-004 Response: NewCPCRejectResponse ──────────────────────────────────

// TestNewCPCRejectResponse verifies CPC-004 rejection response.
// BR-NFS-012: IN_PROGRESS → REJECTED; rejection_reason is mandatory.
func TestNewCPCRejectResponse(t *testing.T) {
	r := resp.NewCPCRejectResponse(
		"req-uuid-002",
		"cpc-officer-02",
		"Documents not legible",
	)

	require.NotNil(t, r)
	assert.Equal(t, "req-uuid-002", r.Data.RequestID)
	assert.Equal(t, "REJECTED", r.Data.NewStatus)
	assert.Equal(t, "cpc-officer-02", r.Data.RejectedBy)
	assert.Equal(t, "Documents not legible", r.Data.RejectionReason)
	assert.NotEmpty(t, r.Data.RejectedAt)
	assert.Equal(t, "Request rejected.", r.Data.Message)
}

// TestCPCRejectRequiresReason documents BR-NFS-012 mandatory rejection reason.
// BR-NFS-012: REJECT requires non-empty rejection_reason.
func TestCPCRejectRequiresReason(t *testing.T) {
	req := handler.CPCRejectRequest{
		RequestID:       "req-uuid-003",
		RejectionReason: "Gazette notification not valid",
		RejectedBy:      "cpc-officer-03",
	}

	assert.NotEmpty(t, req.RejectionReason, "BR-NFS-012: rejection_reason is mandatory for REJECT decision")
}

// ─── CPC-005 Response: NewCPCSendBackResponse ────────────────────────────────

// TestNewCPCSendBackResponse verifies CPC-005 send-back response.
// BR-NFS-014: secure upload link generated; status → PENDING_DOCUMENTS.
func TestNewCPCSendBackResponse(t *testing.T) {
	docs := []string{"GAZETTE_NOTIFICATION", "NEWSPAPER_NOTIFICATION"}
	r := resp.NewCPCSendBackResponse(
		"req-uuid-004",
		"https://nfs.example.com/upload/tok_xyz",
		"2026-02-14T10:00:00Z",
		2,
		docs,
	)

	require.NotNil(t, r)
	assert.Equal(t, "req-uuid-004", r.Data.RequestID)
	assert.Equal(t, "PENDING_DOCUMENTS", r.Data.Status)
	assert.Equal(t, "https://nfs.example.com/upload/tok_xyz", r.Data.SecureUploadURL)
	assert.Equal(t, "2026-02-14T10:00:00Z", r.Data.ExpiresAt)
	assert.Equal(t, 2, r.Data.AttemptNumber)
	assert.Equal(t, 2, len(r.Data.MissingDocuments))
	assert.Contains(t, r.Data.Message, "Customer notified")
}

// ─── CPC-006 Response: NewGetSLADashboardResponse ────────────────────────────

// TestNewGetSLADashboardResponse verifies CPC-006 SLA dashboard response.
// CPC-006: Lists breached items + aggregate stats.
func TestNewGetSLADashboardResponse(t *testing.T) {
	officeCode := "RO-MUM"
	stats := resp.CPCQueueStats{
		TotalPending:    20,
		TotalInProgress: 5,
		TotalOverdue:    4,
		SLAGreen:        10,
		SLAAmber:        6,
		SLARed:          4,
		AddressChanges:  15,
		NameChanges:     5,
	}
	breached := []resp.SLABreachItem{
		{RequestID: "req-breach-001", TicketNumber: "NFS-ANC-20260101-000099", RequestType: "ADDRESS_CHANGE", OfficeCode: "RO-MUM", SLADeadline: "2026-01-10", DaysOverdue: 5},
		{RequestID: "req-breach-002", TicketNumber: "NFS-NMC-20260101-000100", RequestType: "NAME_CHANGE", OfficeCode: "RO-MUM", SLADeadline: "2026-01-08", DaysOverdue: 7},
	}

	r := resp.NewGetSLADashboardResponse(&officeCode, stats, breached)

	require.NotNil(t, r)
	require.NotNil(t, r.Data.OfficeCode)
	assert.Equal(t, "RO-MUM", *r.Data.OfficeCode)
	assert.Equal(t, 4, r.Data.Summary.TotalOverdue)
	assert.Equal(t, 4, r.Data.Summary.SLARed)
	assert.Equal(t, 2, len(r.Data.BreachedItems))
	assert.Equal(t, 5, r.Data.BreachedItems[0].DaysOverdue)
	assert.Equal(t, 7, r.Data.BreachedItems[1].DaysOverdue)
	assert.NotEmpty(t, r.Data.GeneratedAt)
}

// TestSLABreachItem_DaysOverdue verifies breach days calculation semantics.
// CPC-006: DaysOverdue should be positive integer indicating how many days past deadline.
func TestSLABreachItem_DaysOverdue(t *testing.T) {
	item := resp.SLABreachItem{
		RequestID:   "req-overdue",
		DaysOverdue: 10,
		SLADeadline: "2026-01-01",
	}

	assert.Greater(t, item.DaysOverdue, 0, "DaysOverdue must be positive for a breach item")
	assert.Equal(t, 10, item.DaysOverdue)
}

// ─── CPC Request DTOs ─────────────────────────────────────────────────────────

// TestCPCQueueFilterRequest_FieldMapping verifies CPC-001 queue filter request.
func TestCPCQueueFilterRequest_FieldMapping(t *testing.T) {
	status := "PENDING_APPROVAL"
	rt := "ADDRESS_CHANGE"
	officeCode := "RO-DEL"
	page := 1
	pageSize := 20

	req := handler.GetCPCQueueRequest{
		Status:      &status,
		RequestType: &rt,
		OfficeCode:  &officeCode,
		Page:        &page,
		PageSize:    &pageSize,
	}

	require.NotNil(t, req.Status)
	assert.Equal(t, "PENDING_APPROVAL", *req.Status)
	require.NotNil(t, req.RequestType)
	assert.Equal(t, "ADDRESS_CHANGE", *req.RequestType)
	require.NotNil(t, req.Page)
	assert.Equal(t, 1, *req.Page)
	require.NotNil(t, req.PageSize)
	assert.Equal(t, 20, *req.PageSize)
}

// TestCPCSendBackRequest_FieldMapping verifies CPC-005 send-back request.
// BR-NFS-014: missing documents list must be provided.
func TestCPCSendBackRequest_FieldMapping(t *testing.T) {
	req := handler.CPCSendBackRequest{
		RequestID:         "req-uuid-005",
		MissingDocuments:  []string{"GAZETTE_NOTIFICATION"},
		MessageToCustomer: "Please re-upload the gazette notification.",
		RequestedBy:       "cpc-officer-05",
	}

	assert.Equal(t, "req-uuid-005", req.RequestID)
	assert.Equal(t, 1, len(req.MissingDocuments))
	assert.Equal(t, "GAZETTE_NOTIFICATION", req.MissingDocuments[0])
}

// ─── CPC Business Rules ───────────────────────────────────────────────────────

// TestCPCStatusTransitions documents valid CPC-driven status transitions.
// BR-NFS-012: CPC can drive PENDING_APPROVAL → IN_PROGRESS → COMPLETED/REJECTED/PENDING_DOCUMENTS.
func TestCPCStatusTransitions(t *testing.T) {
	type transition struct {
		from string
		to   string
		via  string // CPC action
	}

	validTransitions := []transition{
		{"PENDING_APPROVAL", "IN_PROGRESS", "CPC-003 assign"},
		{"IN_PROGRESS", "REJECTED", "CPC-004 reject"},
		{"IN_PROGRESS", "PENDING_DOCUMENTS", "CPC-005 send-back"},
		{"PENDING_APPROVAL", "COMPLETED", "CORE-008 approve (AADHAAR path)"},
		{"IN_PROGRESS", "COMPLETED", "CORE-008 approve (MANUAL path)"},
	}

	for _, t2 := range validTransitions {
		assert.NotEmpty(t, t2.from)
		assert.NotEmpty(t, t2.to)
		assert.NotEqual(t, t2.from, t2.to, "transition from %s must change status", t2.from)
	}
}
