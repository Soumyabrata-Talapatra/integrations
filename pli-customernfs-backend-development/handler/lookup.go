// Package handler provides HTTP handlers for the Customer NFS service.
// Phase: Phase 3 — Supporting Endpoints
// Handler: LookupHandler — covers LU-001..008 and VA-001..003
//
// LU-001: GET /lookup/states — lists all Indian states/UTs
// LU-002: GET /lookup/pincode/:pincode/validate — pincode + state validation
// LU-003: GET /lookup/address-types — static list
// LU-004: GET /lookup/salutations — static list (VR-NFS-008)
// LU-005: GET /lookup/document-types/:request_type — per request type
// LU-006: GET /lookup/offices — CPC offices (type + circle filter)
// LU-007: GET /lookup/channels — available submission channels
// LU-008: GET /lookup/status-reasons — rejection/send-back reason codes
// VA-001: POST /nfs/validate/address — field-level address validation
// VA-002: POST /nfs/validate/name — field-level name validation
// VA-003: GET /nfs/validate/duplicate-check/:customer_id/:request_type
package handler

import (
	"fmt"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	resp "customer-nfs-service/handler/response"
	repo "customer-nfs-service/repo/postgres"
)

// LookupHandler handles lookup and lightweight validation HTTP endpoints.
// No mutations — all reads. FR-NFS-012 (Lookup/Reference Data).
type LookupHandler struct {
	*serverHandler.Base

	srRepo   *repo.ServiceRequestRepository
	addrRepo *repo.AddressChangeRepository
}

// NewLookupHandler constructs the handler and wires dependencies via Uber FX.
func NewLookupHandler(
	srRepo *repo.ServiceRequestRepository,
	addrRepo *repo.AddressChangeRepository,
) *LookupHandler {
	base := serverHandler.New("Lookup").
		SetPrefix("/v1").
		AddPrefix("")
	return &LookupHandler{
		Base:     base,
		srRepo:   srRepo,
		addrRepo: addrRepo,
	}
}

// Routes registers all lookup and validation HTTP routes.
func (h *LookupHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/lookup/states", h.ListStates).Name("List States"),
		serverRoute.GET("/lookup/pincode/:pincode/validate", h.ValidatePincode).Name("Validate Pincode"),
		serverRoute.GET("/lookup/address-types", h.ListAddressTypes).Name("List Address Types"),
		serverRoute.GET("/lookup/salutations", h.ListSalutations).Name("List Salutations"),
		serverRoute.GET("/lookup/document-types/:request_type", h.ListDocumentTypes).Name("List Document Types"),
		serverRoute.GET("/lookup/offices", h.ListOffices).Name("List Offices"),
		serverRoute.GET("/lookup/channels", h.ListChannels).Name("List Channels"),
		serverRoute.GET("/lookup/status-reasons", h.ListStatusReasons).Name("List Status Reasons"),
		serverRoute.POST("/nfs/validate/address", h.ValidateAddress).Name("Validate Address"),
		serverRoute.POST("/nfs/validate/name", h.ValidateNameFields).Name("Validate Name Fields"),
		serverRoute.GET("/nfs/validate/duplicate-check/:customer_id/:request_type", h.DuplicateCheck).Name("Duplicate Request Check"),
	}
}

// ─── LU-001 ───────────────────────────────────────────────────────────────────

// ListStates returns all Indian states/UTs from the lookup table.
// FR-NFS-012, VR-NFS-002.
func (h *LookupHandler) ListStates(
	sctx *serverRoute.Context,
	req ListStatesRequest,
) (*resp.ListStatesResponse, error) {
	activeOnly := true
	if req.ActiveOnly != nil {
		activeOnly = *req.ActiveOnly
	}

	states, err := h.addrRepo.ListStates(sctx.Ctx, activeOnly)
	if err != nil {
		log.Error(sctx.Ctx, "ListStates: failed to query states: %v", err)
		return nil, err
	}

	items := make([]resp.StateItem, len(states))
	for i, s := range states {
		items[i] = resp.StateItem{Code: s.StateCode, Name: s.StateName, IsActive: s.IsActive}
	}

	r := resp.NewListStatesResponse(items)
	return r, nil
}

// ─── LU-002 ───────────────────────────────────────────────────────────────────

// ValidatePincode validates a 6-digit pincode and optionally checks against a claimed state.
// VR-NFS-001: pincode must be 6 numeric digits (validated at DTO layer).
// BR-NFS-006: pincode-state mismatch raises validation error.
func (h *LookupHandler) ValidatePincode(
	sctx *serverRoute.Context,
	req ValidatePincodeRequest,
) (*resp.PincodeValidationResponse, error) {
	pincodeData, err := h.addrRepo.LookupPincode(sctx.Ctx, req.Pincode)
	if err != nil {
		// Non-fatal: pincode not found in master = invalid
		log.Error(sctx.Ctx, "ValidatePincode: failed to look up pincode %s: %v", req.Pincode, err)
		errMsg := fmt.Sprintf("pincode %s not found in master table (VR-NFS-001)", req.Pincode)
		r := resp.NewPincodeValidationResponse(req.Pincode, false, "", "", "", nil, &errMsg)
		return r, nil
	}

	// BR-NFS-006: check state mismatch if state was supplied.
	var stateMatch *bool
	if req.State != nil && *req.State != "" {
		match := pincodeData.StateName == *req.State
		stateMatch = &match
		if !match {
			errMsg := fmt.Sprintf("pincode %s belongs to %s, not %s (BR-NFS-006)", req.Pincode, pincodeData.StateName, *req.State)
			r := resp.NewPincodeValidationResponse(req.Pincode, false, pincodeData.City, pincodeData.District, pincodeData.StateName, stateMatch, &errMsg)
			return r, nil
		}
	}

	r := resp.NewPincodeValidationResponse(req.Pincode, true, pincodeData.City, pincodeData.District, pincodeData.StateName, stateMatch, nil)
	return r, nil
}

// ─── LU-003 ───────────────────────────────────────────────────────────────────

// ListAddressTypes returns the static list of address type codes.
// FR-NFS-003: PERMANENT, COMMUNICATION, BOTH.
func (h *LookupHandler) ListAddressTypes(
	sctx *serverRoute.Context,
	_ ListAddressTypesRequest,
) (*resp.ListAddressTypesResponse, error) {
	r := resp.NewListAddressTypesResponse([]resp.AddressTypeItem{
		{Code: "PERMANENT", Description: "Permanent/Registered Address"},
		{Code: "COMMUNICATION", Description: "Communication/Mailing Address"},
		{Code: "BOTH", Description: "Both Permanent and Communication"},
	})
	return r, nil
}

// ─── LU-004 ───────────────────────────────────────────────────────────────────

// ListSalutations returns the static list of allowed salutations.
// VR-NFS-008: {Mr, Mrs, Ms, Shri, Smt, Dr}.
func (h *LookupHandler) ListSalutations(
	sctx *serverRoute.Context,
	_ ListSalutationsRequest,
) (*resp.ListSalutationsResponse, error) {
	r := resp.NewListSalutationsResponse([]resp.SalutationItem{
		{Code: "Mr", Description: "Mister"},
		{Code: "Mrs", Description: "Missus"},
		{Code: "Ms", Description: "Miss"},
		{Code: "Shri", Description: "Shri (Male)"},
		{Code: "Smt", Description: "Smt (Female)"},
		{Code: "Dr", Description: "Doctor"},
	})
	return r, nil
}

// ─── LU-005 ───────────────────────────────────────────────────────────────────

// ListDocumentTypes returns required and optional document types for a given request type.
// ADDRESS_CHANGE: UTILITY_BILL, BANK_STATEMENT, AADHAAR_COPY, DRIVING_LICENSE...
// NAME_CHANGE: GAZETTE_NOTIFICATION, NEWSPAPER_NOTIFICATION, NAME_CHANGE_FORM...
// BR-NFS-008: at least one of the required document types must be uploaded.
func (h *LookupHandler) ListDocumentTypes(
	sctx *serverRoute.Context,
	req ListDocumentTypesRequest,
) (*resp.ListDocumentTypesResponse, error) {
	docTypes, err := h.addrRepo.ListDocumentTypes(sctx.Ctx, req.RequestType)
	if err != nil {
		log.Error(sctx.Ctx, "ListDocumentTypes: failed to query document types for %s: %v", req.RequestType, err)
		return nil, err
	}

	items := make([]resp.DocumentTypeItem, len(docTypes))
	for i, dt := range docTypes {
		items[i] = resp.DocumentTypeItem{
			Code:        dt.DocTypeCode,
			Description: dt.Description,
			IsRequired:  dt.IsMandatory,
			RequestType: req.RequestType,
		}
	}

	r := resp.NewListDocumentTypesResponse(req.RequestType, items)
	return r, nil
}

// ─── LU-006 ───────────────────────────────────────────────────────────────────

// ListOffices returns CPC offices filtered by type and circle code.
// Used by CPC screens to assign requests (CPC-003) and by the bootstrap screen.
func (h *LookupHandler) ListOffices(
	sctx *serverRoute.Context,
	req ListOfficesRequest,
) (*resp.ListOfficesResponse, error) {
	offices, err := h.addrRepo.ListOffices(sctx.Ctx, req.Type, req.CircleCode)
	if err != nil {
		log.Error(sctx.Ctx, "ListOffices: failed to query offices: %v", err)
		return nil, err
	}

	items := make([]resp.OfficeItem, len(offices))
	for i, o := range offices {
		items[i] = resp.OfficeItem{
			OfficeCode: o.OfficeCode,
			OfficeName: o.OfficeName,
			OfficeType: o.OfficeType,
			CircleCode: *o.CircleCode,
			Address:    &o.Address,
			IsActive:   o.IsActive,
		}
	}

	r := resp.NewListOfficesResponse(items)
	return r, nil
}

// ─── LU-007 ───────────────────────────────────────────────────────────────────

// ListChannels returns the available submission channels.
// VR-NFS-016: Portal, Mobile, Branch, Satellite.
// BR-NFS-015: AllowsAadhaar=true only for Portal and Mobile.
func (h *LookupHandler) ListChannels(
	sctx *serverRoute.Context,
	_ ListChannelsRequest,
) (*resp.ListChannelsResponse, error) {
	r := resp.NewListChannelsResponse([]resp.ChannelItem{
		{Code: "Portal", Description: "Web Portal (Self-service)", AllowsAadhaar: true},
		{Code: "Mobile", Description: "Mobile Application (Self-service)", AllowsAadhaar: true},
		{Code: "Branch", Description: "Branch Office (Assisted)", AllowsAadhaar: false},       // BR-NFS-015
		{Code: "Satellite", Description: "Satellite Office (Assisted)", AllowsAadhaar: false}, // BR-NFS-015
	})
	return r, nil
}

// ─── LU-008 ───────────────────────────────────────────────────────────────────

// ListStatusReasons returns pre-defined rejection/send-back reason codes.
// Used by CPC for structured decision recording.
func (h *LookupHandler) ListStatusReasons(
	sctx *serverRoute.Context,
	req ListStatusReasonsRequest,
) (*resp.ListStatusReasonsResponse, error) {
	reasons, err := h.addrRepo.ListStatusReasons(sctx.Ctx, req.ActionType)
	if err != nil {
		log.Error(sctx.Ctx, "ListStatusReasons: failed to query reasons: %v", err)
		return nil, err
	}

	items := make([]resp.StatusReasonItem, len(reasons))
	for i, r := range reasons {
		items[i] = resp.StatusReasonItem{
			Code:        r.ReasonCode,
			Description: r.ReasonText,
			ActionType:  r.ActionType,
		}
	}

	actionType := ""
	if req.ActionType != nil {
		actionType = *req.ActionType
	}
	r := resp.NewListStatusReasonsResponse(items, actionType)
	return r, nil
}

// ─── VA-001 ───────────────────────────────────────────────────────────────────

// ValidateAddress validates address fields without persisting anything.
// VR-NFS-001..005, BR-NFS-006 — client-side pre-validation before submission.
func (h *LookupHandler) ValidateAddress(
	sctx *serverRoute.Context,
	req ValidateAddressRequest,
) (*resp.ValidateAddressResponse, error) {
	var errs []resp.ValidationError

	// VR-NFS-001: pincode format (6 numeric digits already validated at struct level via `validate` tag).
	// VR-NFS-003..005: field length already validated at struct level.

	// BR-NFS-006: pincode-state cross validation.
	pincodeData, err := h.addrRepo.LookupPincode(sctx.Ctx, req.Pincode)
	if err != nil {
		log.Error(sctx.Ctx, "ValidateAddress: pincode lookup failed for %s: %v", req.Pincode, err)
		errs = append(errs, resp.ValidationError{
			Field:   "pincode",
			Code:    "VR-NFS-001",
			Message: fmt.Sprintf("pincode %s is not valid", req.Pincode),
		})
	} else if pincodeData.StateName != req.State {
		errs = append(errs, resp.ValidationError{
			Field:   "state",
			Code:    "BR-NFS-006",
			Message: fmt.Sprintf("pincode %s belongs to %s, not %s", req.Pincode, pincodeData.StateName, req.State),
		})
	}

	r := resp.NewValidateAddressResponse(len(errs) == 0, errs)
	return r, nil
}

// ─── VA-002 ───────────────────────────────────────────────────────────────────

// ValidateNameFields validates name fields without persisting anything.
// VR-NFS-006..008, VR-NFS-014 — client-side pre-validation before submission.
// NOTE: VR-NFS-014 (DOB immutability) is enforced at the struct level — date_of_birth
// is absent from request.ValidateNameFieldsRequest, so any payload containing it
// will either be ignored (lenient parsing) or raise a binding error.
func (h *LookupHandler) ValidateNameFields(
	sctx *serverRoute.Context,
	req ValidateNameFieldsRequest,
) (*resp.ValidateNameFieldsResponse, error) {
	var errs []resp.ValidationError

	// VR-NFS-006: first_name rules (struct-level tag already handles max length).
	if req.FirstName != nil && len(*req.FirstName) == 0 {
		errs = append(errs, resp.ValidationError{
			Field:   "first_name",
			Code:    "VR-NFS-006",
			Message: "first_name cannot be empty",
		})
	}

	// VR-NFS-007: last_name rules.
	if req.LastName != nil && len(*req.LastName) == 0 {
		errs = append(errs, resp.ValidationError{
			Field:   "last_name",
			Code:    "VR-NFS-007",
			Message: "last_name cannot be empty",
		})
	}

	// VR-NFS-008: salutation whitelist.
	allowed := map[string]bool{"Mr": true, "Mrs": true, "Ms": true, "Shri": true, "Smt": true, "Dr": true}
	if req.Salutation != nil && !allowed[*req.Salutation] {
		errs = append(errs, resp.ValidationError{
			Field:   "salutation",
			Code:    "VR-NFS-008",
			Message: fmt.Sprintf("salutation %q not allowed; must be one of Mr/Mrs/Ms/Shri/Smt/Dr", *req.Salutation),
		})
	}

	r := resp.NewValidateNameFieldsResponse(len(errs) == 0, errs)
	return r, nil
}

// ─── VA-003 ───────────────────────────────────────────────────────────────────

// DuplicateCheck checks whether a customer has an active pending NFS request of the given type.
// VR-NFS-015: only one active request per customer+type allowed.
func (h *LookupHandler) DuplicateCheck(
	sctx *serverRoute.Context,
	req DuplicateCheckRequest,
) (*resp.DuplicateCheckResponse, error) {
	isDuplicate, existingTicket, err := h.srRepo.CheckDuplicateRequest(sctx.Ctx, req.CustomerID, req.RequestType)
	if err != nil {
		log.Error(sctx.Ctx, "DuplicateCheck: check failed for customer %d type %s: %v", req.CustomerID, req.RequestType, err)
		return nil, err
	}

	var ticketPtr, statusPtr *string
	if isDuplicate {
		ticketPtr = &existingTicket

		// Fetch current status for informational response.
		sr, err2 := h.srRepo.GetByTicketNumber(sctx.Ctx, existingTicket)
		if err2 == nil {
			s := string(sr.Status)
			statusPtr = &s
		}
	}

	r := resp.NewDuplicateCheckResponse(isDuplicate, ticketPtr, statusPtr)
	return r, nil
}
