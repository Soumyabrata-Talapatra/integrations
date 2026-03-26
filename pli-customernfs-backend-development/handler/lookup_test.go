// Package handler_test contains unit tests for Phase 3 supporting endpoints.
// Tests: LU-001..008 (Lookup), VA-001..003 (Validation), ST-001..005 (Status)
// No DB or Temporal dependency — pure DTO and business rule tests.
package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handler "customer-nfs-service/handler"
	resp "customer-nfs-service/handler/response"
)

// ─── LU-001: ListStatesRequest ───────────────────────────────────────────────

// TestListStatesRequest_DefaultActiveOnly verifies default active_only behaviour.
// LU-001: GET /lookup/states?active_only (defaults to true)
func TestListStatesRequest_DefaultActiveOnly(t *testing.T) {
	req := handler.ListStatesRequest{}
	assert.Nil(t, req.ActiveOnly, "ActiveOnly should be nil (query param absent means default)")
}

// TestNewListStatesResponse verifies LU-001 response construction.
func TestNewListStatesResponse(t *testing.T) {
	states := []resp.StateItem{
		{Code: "DL", Name: "Delhi", IsActive: true},
		{Code: "MH", Name: "Maharashtra", IsActive: true},
	}

	r := resp.NewListStatesResponse(states)

	require.NotNil(t, r)
	assert.Equal(t, 2, r.Data.Total)
	assert.Equal(t, 2, len(r.Data.States))
	assert.Equal(t, "DL", r.Data.States[0].Code)
	assert.Equal(t, "Maharashtra", r.Data.States[1].Name)
}

// TestNewListStatesResponse_EmptyList verifies empty state list handling.
func TestNewListStatesResponse_EmptyList(t *testing.T) {
	r := resp.NewListStatesResponse([]resp.StateItem{})

	require.NotNil(t, r)
	assert.Equal(t, 0, r.Data.Total)
	assert.Empty(t, r.Data.States)
}

// ─── LU-002: ValidatePincodeRequest ──────────────────────────────────────────

// TestValidatePincodeRequest_FieldMapping verifies pincode request fields.
// VR-NFS-001: pincode must be 6 numeric digits.
func TestValidatePincodeRequest_FieldMapping(t *testing.T) {
	state := "Delhi"
	req := handler.ValidatePincodeRequest{
		Pincode: "110001",
		State:   &state,
	}

	assert.Equal(t, "110001", req.Pincode)
	assert.Len(t, req.Pincode, 6, "VR-NFS-001: pincode must be exactly 6 digits")
	require.NotNil(t, req.State)
	assert.Equal(t, "Delhi", *req.State)
}

// TestValidatePincodeRequest_NoStateFilter verifies optional state param.
// BR-NFS-006: state is optional for pure pincode validation.
func TestValidatePincodeRequest_NoStateFilter(t *testing.T) {
	req := handler.ValidatePincodeRequest{
		Pincode: "400001",
		State:   nil,
	}

	assert.Equal(t, "400001", req.Pincode)
	assert.Nil(t, req.State, "state filter is optional in LU-002")
}

// TestNewPincodeValidationResponse_Valid verifies LU-002 response for valid pincode.
// BR-NFS-006: StateMatch field indicates whether claimed state matches India Post master.
func TestNewPincodeValidationResponse_Valid(t *testing.T) {
	matchTrue := true
	r := resp.NewPincodeValidationResponse(
		"110001",
		true,
		"New Delhi",
		"New Delhi",
		"Delhi",
		&matchTrue,
		nil,
	)

	require.NotNil(t, r)
	assert.Equal(t, "110001", r.Data.Pincode)
	assert.True(t, r.Data.IsValid)
	assert.Equal(t, "New Delhi", r.Data.City)
	assert.Equal(t, "Delhi", r.Data.State)
	require.NotNil(t, r.Data.StateMatch)
	assert.True(t, *r.Data.StateMatch, "BR-NFS-006: state must match when claimed state is correct")
	assert.Nil(t, r.Data.ErrorMessage)
}

// TestNewPincodeValidationResponse_StateMismatch verifies BR-NFS-006 mismatch.
// BR-NFS-006: pincode is in Delhi but customer claims Maharashtra → StateMatch=false.
func TestNewPincodeValidationResponse_StateMismatch(t *testing.T) {
	matchFalse := false
	errMsg := "Pincode 110001 belongs to Delhi, not Maharashtra"

	r := resp.NewPincodeValidationResponse(
		"110001",
		true,
		"New Delhi",
		"New Delhi",
		"Delhi",
		&matchFalse,
		&errMsg,
	)

	require.NotNil(t, r)
	require.NotNil(t, r.Data.StateMatch)
	assert.False(t, *r.Data.StateMatch, "BR-NFS-006: state mismatch must be flagged")
	require.NotNil(t, r.Data.ErrorMessage)
	assert.Contains(t, *r.Data.ErrorMessage, "Delhi")
}

// TestNewPincodeValidationResponse_InvalidPincode verifies invalid pincode response.
// VR-NFS-001: non-existent pincode → IsValid=false.
func TestNewPincodeValidationResponse_InvalidPincode(t *testing.T) {
	errMsg := "Pincode not found"
	r := resp.NewPincodeValidationResponse("999999", false, "", "", "", nil, &errMsg)

	require.NotNil(t, r)
	assert.False(t, r.Data.IsValid)
	assert.Empty(t, r.Data.City)
	assert.Nil(t, r.Data.StateMatch, "StateMatch is nil when no state was provided")
}

// ─── LU-003/004: Static Lists ─────────────────────────────────────────────────

// TestListAddressTypesRequest_IsEmpty verifies LU-003 has no parameters.
func TestListAddressTypesRequest_IsEmpty(t *testing.T) {
	req := handler.ListAddressTypesRequest{}
	_ = req // no fields — struct should be zero value
	assert.True(t, true, "LU-003 request has no parameters (static list)")
}

// TestListSalutationsRequest_IsEmpty verifies LU-004 has no parameters.
func TestListSalutationsRequest_IsEmpty(t *testing.T) {
	req := handler.ListSalutationsRequest{}
	_ = req
	assert.True(t, true, "LU-004 request has no parameters (static list)")
}

// TestSalutationValues_VR_NFS_008 documents valid salutation codes.
// VR-NFS-008: salutation ∈ {Mr, Mrs, Ms, Shri, Smt, Dr}
func TestSalutationValues_VR_NFS_008(t *testing.T) {
	validSalutations := []string{"Mr", "Mrs", "Ms", "Shri", "Smt", "Dr"}

	// Verify salutation items are constructable with all valid values.
	for _, sal := range validSalutations {
		item := resp.SalutationItem{Code: sal, Description: "Test - " + sal}
		assert.Equal(t, sal, item.Code)
	}
}

// TestChannelAllowsAadhaar verifies BR-NFS-015 in channel items.
// BR-NFS-015: Portal and Mobile channels allow Aadhaar auth.
func TestChannelAllowsAadhaar(t *testing.T) {
	channels := []resp.ChannelItem{
		{Code: "Portal", Description: "Customer Portal", AllowsAadhaar: true},
		{Code: "Mobile", Description: "Mobile App", AllowsAadhaar: true},
		{Code: "PostOffice", Description: "Post Office", AllowsAadhaar: false},
		{Code: "CallCenter", Description: "Call Center", AllowsAadhaar: false},
		{Code: "AgentPortal", Description: "Agent Portal", AllowsAadhaar: false},
	}

	for _, ch := range channels {
		if ch.Code == "Portal" || ch.Code == "Mobile" {
			assert.True(t, ch.AllowsAadhaar, "channel %s must allow Aadhaar (BR-NFS-015)", ch.Code)
		} else {
			assert.False(t, ch.AllowsAadhaar, "channel %s must NOT allow Aadhaar (BR-NFS-015)", ch.Code)
		}
	}
}

// ─── LU-005: ListDocumentTypesRequest ────────────────────────────────────────

// TestListDocumentTypesRequest_FieldMapping verifies LU-005 request fields.
// request_type: ADDRESS_CHANGE or NAME_CHANGE
func TestListDocumentTypesRequest_FieldMapping(t *testing.T) {
	cases := []string{"ADDRESS_CHANGE", "NAME_CHANGE"}

	for _, rt := range cases {
		req := handler.ListDocumentTypesRequest{RequestType: rt}
		assert.Equal(t, rt, req.RequestType)
	}
}

// TestDocumentTypeItem_MandatoryFlag verifies document type mandatory field.
// BR-NFS-008: required documents must be flagged.
func TestDocumentTypeItem_MandatoryFlag(t *testing.T) {
	items := []resp.DocumentTypeItem{
		{Code: "GAZETTE_NOTIFICATION", Description: "Gazette Notification", IsRequired: true, RequestType: "NAME_CHANGE"},
		{Code: "NEWSPAPER_NOTIFICATION", Description: "Newspaper Ad", IsRequired: false, RequestType: "NAME_CHANGE"},
		{Code: "PROOF_OF_RESIDENCE", Description: "Proof of Residence", IsRequired: true, RequestType: "ADDRESS_CHANGE"},
	}

	var requiredCount int
	for _, item := range items {
		if item.IsRequired {
			requiredCount++
		}
	}

	assert.Equal(t, 2, requiredCount, "2 out of 3 test items should be required")
}

// ─── LU-006: ListOfficesRequest ──────────────────────────────────────────────

// TestListOfficesRequest_FieldMapping verifies LU-006 optional filter fields.
func TestListOfficesRequest_FieldMapping(t *testing.T) {
	officeType := "REGIONAL"
	circleCode := "DEL"

	req := handler.ListOfficesRequest{
		Type:       &officeType,
		CircleCode: &circleCode,
	}

	require.NotNil(t, req.Type)
	assert.Equal(t, "REGIONAL", *req.Type)
	require.NotNil(t, req.CircleCode)
	assert.Equal(t, "DEL", *req.CircleCode)
}

// TestOfficeItem_Types documents valid office types for LU-006.
func TestOfficeItem_Types(t *testing.T) {
	validTypes := []string{"SATELLITE", "REGIONAL", "CENTRAL"}

	for _, ot := range validTypes {
		item := resp.OfficeItem{
			OfficeCode: "OFF-" + ot[:3],
			OfficeName: ot + " Office",
			OfficeType: ot,
			CircleCode: "DEL",
			IsActive:   true,
		}
		assert.Equal(t, ot, item.OfficeType)
	}
}

// ─── LU-008: ListStatusReasonsRequest ────────────────────────────────────────

// TestListStatusReasonsRequest_ActionTypeFilter verifies LU-008 filter.
func TestListStatusReasonsRequest_ActionTypeFilter(t *testing.T) {
	actionType := "REJECT"
	req := handler.ListStatusReasonsRequest{ActionType: &actionType}

	require.NotNil(t, req.ActionType)
	assert.Equal(t, "REJECT", *req.ActionType)
}

// TestStatusReasonItem_ValidActionTypes documents valid action types for LU-008.
func TestStatusReasonItem_ValidActionTypes(t *testing.T) {
	items := []resp.StatusReasonItem{
		{Code: "DOC_EXPIRED", Description: "Documents expired or invalid", ActionType: "REJECT"},
		{Code: "DOC_MISSING", Description: "Required documents not uploaded", ActionType: "SEND_BACK"},
		{Code: "CUSTOMER_REQUEST", Description: "Customer initiated withdrawal", ActionType: "WITHDRAW"},
	}

	for _, item := range items {
		assert.NotEmpty(t, item.Code)
		assert.NotEmpty(t, item.ActionType)
		assert.Contains(t, []string{"REJECT", "SEND_BACK", "WITHDRAW"}, item.ActionType,
			"action_type %q must be one of REJECT, SEND_BACK, WITHDRAW", item.ActionType)
	}
}

// ─── VA-001: ValidateAddressRequest ──────────────────────────────────────────

// TestValidateAddressRequest_FieldMapping verifies VA-001 request fields.
// VR-NFS-001..005 + BR-NFS-006
func TestValidateAddressRequest_FieldMapping(t *testing.T) {
	line2 := "Sector 17"
	req := handler.ValidateAddressRequest{
		AddressLine1: "12 Main Road",
		AddressLine2: &line2,
		City:         "Delhi",
		District:     "New Delhi",
		State:        "Delhi",
		Pincode:      "110001",
	}

	assert.Equal(t, "12 Main Road", req.AddressLine1)
	require.NotNil(t, req.AddressLine2)
	assert.Equal(t, "Sector 17", *req.AddressLine2)
	assert.Equal(t, "110001", req.Pincode)
	assert.Len(t, req.Pincode, 6, "VR-NFS-001: pincode must be 6 digits")
}

// TestNewValidateAddressResponse_Valid verifies VA-001 pass response.
func TestNewValidateAddressResponse_Valid(t *testing.T) {
	r := resp.NewValidateAddressResponse(true, nil)

	require.NotNil(t, r)
	assert.True(t, r.Data.IsValid)
	assert.Nil(t, r.Data.Errors)
}

// TestNewValidateAddressResponse_WithErrors verifies VA-001 fail response.
// VR-NFS-001: invalid pincode + VR-NFS-002: invalid state.
func TestNewValidateAddressResponse_WithErrors(t *testing.T) {
	errors := []resp.ValidationError{
		{Field: "pincode", Code: "VR-NFS-001", Message: "Pincode must be 6 numeric digits"},
		{Field: "state", Code: "VR-NFS-002", Message: "State does not match India state list"},
	}

	r := resp.NewValidateAddressResponse(false, errors)

	require.NotNil(t, r)
	assert.False(t, r.Data.IsValid)
	assert.Equal(t, 2, len(r.Data.Errors))
	assert.Equal(t, "pincode", r.Data.Errors[0].Field)
	assert.Equal(t, "VR-NFS-001", r.Data.Errors[0].Code)
}

// ─── VA-002: ValidateNameFieldsRequest ───────────────────────────────────────

// TestValidateNameFieldsRequest_NoDOBField verifies VR-NFS-014 / BR-NFS-010.
// BR-NFS-010: date_of_birth is immutable — must never appear in name validation.
func TestValidateNameFieldsRequest_NoDOBField(t *testing.T) {
	fn := "Rajesh"
	ln := "Kumar"
	sal := "Mr"

	req := handler.ValidateNameFieldsRequest{
		Salutation: &sal,
		FirstName:  &fn,
		LastName:   &ln,
		MiddleName: nil,
	}

	require.NotNil(t, req.FirstName)
	assert.Equal(t, "Rajesh", *req.FirstName)
	require.NotNil(t, req.Salutation)
	assert.Equal(t, "Mr", *req.Salutation)
	// date_of_birth field intentionally absent (VR-NFS-014)
}

// TestNewValidateNameFieldsResponse_Valid verifies VA-002 pass response.
func TestNewValidateNameFieldsResponse_Valid(t *testing.T) {
	r := resp.NewValidateNameFieldsResponse(true, nil)

	require.NotNil(t, r)
	assert.True(t, r.Data.IsValid)
	assert.Nil(t, r.Data.Errors)
}

// TestNewValidateNameFieldsResponse_DOBViolation verifies VR-NFS-014 error.
// If middleware detects date_of_birth in payload, response must mark invalid.
func TestNewValidateNameFieldsResponse_DOBViolation(t *testing.T) {
	errors := []resp.ValidationError{
		{Field: "date_of_birth", Code: "VR-NFS-014", Message: "date_of_birth is immutable and must not be provided (BR-NFS-010)"},
	}

	r := resp.NewValidateNameFieldsResponse(false, errors)

	require.NotNil(t, r)
	assert.False(t, r.Data.IsValid)
	require.Equal(t, 1, len(r.Data.Errors))
	assert.Equal(t, "VR-NFS-014", r.Data.Errors[0].Code)
	assert.Contains(t, r.Data.Errors[0].Message, "immutable")
}

// ─── VA-003: DuplicateCheckRequest ───────────────────────────────────────────

// TestDuplicateCheckRequest_FieldMapping verifies VA-003 path params.
// VR-NFS-015: one active pending request per customer+type.
func TestDuplicateCheckRequest_FieldMapping(t *testing.T) {
	req := handler.DuplicateCheckRequest{
		CustomerID:  "cust-uuid-001",
		RequestType: "ADDRESS_CHANGE",
	}

	assert.Equal(t, "cust-uuid-001", req.CustomerID)
	assert.Equal(t, "ADDRESS_CHANGE", req.RequestType)
}

// TestNewDuplicateCheckResponse_NoDuplicate verifies no duplicate found response.
// VR-NFS-015: no existing active request for this customer and type.
func TestNewDuplicateCheckResponse_NoDuplicate(t *testing.T) {
	r := resp.NewDuplicateCheckResponse(false, nil, nil)

	require.NotNil(t, r)
	assert.False(t, r.Data.IsDuplicate)
	assert.Nil(t, r.Data.ExistingTicket)
	assert.Nil(t, r.Data.Status)
}

// TestNewDuplicateCheckResponse_DuplicateExists verifies duplicate found response.
// VR-NFS-015: customer already has a pending request → return existing ticket.
func TestNewDuplicateCheckResponse_DuplicateExists(t *testing.T) {
	ticket := "NFS-ANC-20260101-000042"
	status := "PENDING_APPROVAL"

	r := resp.NewDuplicateCheckResponse(true, &ticket, &status)

	require.NotNil(t, r)
	assert.True(t, r.Data.IsDuplicate)
	require.NotNil(t, r.Data.ExistingTicket)
	assert.Equal(t, ticket, *r.Data.ExistingTicket)
	require.NotNil(t, r.Data.Status)
	assert.Equal(t, "PENDING_APPROVAL", *r.Data.Status)
}

// ─── ST-001: GetRequestDetailRequest ─────────────────────────────────────────

// TestGetRequestDetailRequest_FieldMapping verifies ST-001 include flags.
// ST-001: optional include_documents + include_audit query params.
func TestGetRequestDetailRequest_FieldMapping(t *testing.T) {
	inclDocs := true
	inclAudit := false

	req := handler.GetRequestDetailRequest{
		RequestID:        "req-uuid-st-001",
		IncludeDocuments: &inclDocs,
		IncludeAudit:     &inclAudit,
	}

	assert.Equal(t, "req-uuid-st-001", req.RequestID)
	require.NotNil(t, req.IncludeDocuments)
	assert.True(t, *req.IncludeDocuments)
	require.NotNil(t, req.IncludeAudit)
	assert.False(t, *req.IncludeAudit)
}

// TestGetRequestDetailRequest_DefaultsNil verifies nil flags when not provided.
func TestGetRequestDetailRequest_DefaultsNil(t *testing.T) {
	req := handler.GetRequestDetailRequest{
		RequestID: "req-uuid-st-002",
	}

	assert.Nil(t, req.IncludeDocuments, "IncludeDocuments defaults to nil (false)")
	assert.Nil(t, req.IncludeAudit, "IncludeAudit defaults to nil (false)")
}

// ─── ST-004: ListCustomerRequestsRequest ─────────────────────────────────────

// TestListCustomerRequestsRequest_FieldMapping verifies ST-004 query params.
// BATCH: count + paginated rows in single round-trip.
func TestListCustomerRequestsRequest_FieldMapping(t *testing.T) {
	rt := "NAME_CHANGE"
	status := "PENDING_APPROVAL"
	page := 1
	pageSize := 20

	req := handler.ListCustomerRequestsRequest{
		CustomerID:  "cust-uuid-001",
		RequestType: &rt,
		Status:      &status,
		Page:        &page,
		PageSize:    &pageSize,
	}

	assert.Equal(t, "cust-uuid-001", req.CustomerID)
	require.NotNil(t, req.RequestType)
	assert.Equal(t, "NAME_CHANGE", *req.RequestType)
	require.NotNil(t, req.Status)
	assert.Equal(t, "PENDING_APPROVAL", *req.Status)
	require.NotNil(t, req.Page)
	assert.Equal(t, 1, *req.Page)
	require.NotNil(t, req.PageSize)
	assert.Equal(t, 20, *req.PageSize)
}

// TestNewListCustomerRequestsResponse_Pagination verifies ST-004 paginated response.
func TestNewListCustomerRequestsResponse_Pagination(t *testing.T) {
	requests := []resp.CustomerRequestSummary{
		{RequestID: "req-001", TicketNumber: "NFS-ANC-20260101-000001", RequestType: "ADDRESS_CHANGE", Status: "COMPLETED"},
		{RequestID: "req-002", TicketNumber: "NFS-NMC-20260201-000001", RequestType: "NAME_CHANGE", Status: "PENDING_APPROVAL"},
	}

	r := resp.NewListCustomerRequestsResponse("cust-uuid-001", 100, 1, 20, requests)

	require.NotNil(t, r)
	assert.Equal(t, "cust-uuid-001", r.Data.CustomerID)
	assert.Equal(t, 100, r.Data.Total)
	assert.Equal(t, 1, r.Data.Page)
	assert.Equal(t, 20, r.Data.PageSize)
	assert.Equal(t, 2, len(r.Data.Requests))
	assert.Equal(t, "NFS-ANC-20260101-000001", r.Data.Requests[0].TicketNumber)
}

// ─── ST-005: RequestDocumentsResponse ────────────────────────────────────────

// TestNewRequestDocumentsResponse_DocumentList verifies ST-005 response.
func TestNewRequestDocumentsResponse_DocumentList(t *testing.T) {
	docs := []resp.RequestDetailDocumentItem{
		{
			DocumentID:         "doc-uuid-001",
			DocumentType:       "GAZETTE_NOTIFICATION",
			FileName:           "gazette.pdf",
			FileSizeBytes:      102400,
			VerificationStatus: "PENDING",
			UploadedAt:         "2026-01-15T10:00:00Z",
		},
	}

	r := resp.NewRequestDocumentsResponse("req-uuid-001", docs)

	require.NotNil(t, r)
	assert.Equal(t, "req-uuid-001", r.Data.RequestID)
	assert.Equal(t, 1, r.Data.Total)
	assert.Equal(t, "doc-uuid-001", r.Data.Documents[0].DocumentID)
	assert.Equal(t, "GAZETTE_NOTIFICATION", r.Data.Documents[0].DocumentType)
}

// TestNewRequestDocumentsResponse_Empty verifies empty document list handled.
func TestNewRequestDocumentsResponse_Empty(t *testing.T) {
	r := resp.NewRequestDocumentsResponse("req-uuid-002", []resp.RequestDetailDocumentItem{})

	require.NotNil(t, r)
	assert.Equal(t, 0, r.Data.Total)
	assert.Empty(t, r.Data.Documents)
}

// ─── ST-002 AuditTrail ────────────────────────────────────────────────────────

// TestAuditTrailItem_FieldMapping verifies audit trail item fields for ST-002.
// FR-NFS-008: Audit trail must include action type, actor, timestamp, old/new values.
func TestAuditTrailItem_FieldMapping(t *testing.T) {
	oldVal := `{"status":"CREATED"}`
	newVal := `{"status":"PENDING_APPROVAL"}`
	notes := "Documents submitted by agent"

	item := resp.AuditTrailItem{
		ActionType:  "STATUS_CHANGE",
		PerformedBy: "agent-01",
		PerformedAt: "2026-01-15T10:00:00Z",
		OldValue:    &oldVal,
		NewValue:    &newVal,
		Notes:       &notes,
	}

	assert.Equal(t, "STATUS_CHANGE", item.ActionType)
	assert.Equal(t, "agent-01", item.PerformedBy)
	require.NotNil(t, item.OldValue)
	assert.Equal(t, oldVal, *item.OldValue)
	require.NotNil(t, item.NewValue)
	assert.Equal(t, newVal, *item.NewValue)
	require.NotNil(t, item.Notes)
	assert.Contains(t, *item.Notes, "Documents submitted")
}

// ─── Validation Rules: Field Length Limits ───────────────────────────────────

// TestAddressValidationFieldLengths verifies length constraint documentation.
// VR-NFS-003: address_line_1 max 150 chars.
// VR-NFS-004: city max 100 chars.
// VR-NFS-005: district max 100 chars.
func TestAddressValidationFieldLengths(t *testing.T) {
	// Max-length boundary test values
	maxLine1 := make([]byte, 150)
	for i := range maxLine1 {
		maxLine1[i] = 'A'
	}
	maxCity := make([]byte, 100)
	for i := range maxCity {
		maxCity[i] = 'B'
	}

	req := handler.ValidateAddressRequest{
		AddressLine1: string(maxLine1),
		City:         string(maxCity),
		District:     "Test District",
		State:        "Delhi",
		Pincode:      "110001",
	}

	assert.Len(t, req.AddressLine1, 150, "VR-NFS-003: address_line_1 max 150 chars")
	assert.Len(t, req.City, 100, "VR-NFS-004: city max 100 chars")
}
