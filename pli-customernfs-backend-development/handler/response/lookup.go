// Package response contains HTTP response DTOs for lookup and validation endpoints.
//
// Phase 3: Supporting Endpoints
// LU-001..008 + VA-001..003 response types.
package response

import "customer-nfs-service/core/port"

// ─── LU-001: GET /lookup/states ──────────────────────────────────────────────

// StateItem represents a single Indian state/UT.
type StateItem struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

// ListStatesData is the payload returned on a successful LU-001 call.
type ListStatesData struct {
	States []StateItem `json:"states"`
	Total  int         `json:"total"`
}

// ListStatesResponse is the top-level success response for LU-001.
type ListStatesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ListStatesData `json:"data"`
}

// NewListStatesResponse constructs the LU-001 success response.
func NewListStatesResponse(states []StateItem) *ListStatesResponse {
	return &ListStatesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 ListStatesData{States: states, Total: len(states)},
	}
}

// ─── LU-002: GET /lookup/pincode/:pincode/validate ────────────────────────────

// PincodeValidationData is the payload returned on a successful LU-002 call.
// BR-NFS-006: state mismatch indicated via StateMatch field.
type PincodeValidationData struct {
	Pincode      string  `json:"pincode"`
	IsValid      bool    `json:"is_valid"`
	City         string  `json:"city,omitempty"`
	District     string  `json:"district,omitempty"`
	State        string  `json:"state,omitempty"`
	StateMatch   *bool   `json:"state_match,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
}

// PincodeValidationResponse is the top-level success response for LU-002.
type PincodeValidationResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      PincodeValidationData `json:"data"`
}

// NewPincodeValidationResponse constructs the LU-002 response.
func NewPincodeValidationResponse(pincode string, isValid bool, city, district, state string, stateMatch *bool, errMsg *string) *PincodeValidationResponse {
	return &PincodeValidationResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: PincodeValidationData{
			Pincode:      pincode,
			IsValid:      isValid,
			City:         city,
			District:     district,
			State:        state,
			StateMatch:   stateMatch,
			ErrorMessage: errMsg,
		},
	}
}

// ─── LU-003: GET /lookup/address-types ───────────────────────────────────────

// AddressTypeItem represents a single address type option.
type AddressTypeItem struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// ListAddressTypesData is the payload returned on a successful LU-003 call.
type ListAddressTypesData struct {
	AddressTypes []AddressTypeItem `json:"address_types"`
}

// ListAddressTypesResponse is the top-level success response for LU-003.
type ListAddressTypesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ListAddressTypesData `json:"data"`
}

// NewListAddressTypesResponse constructs the LU-003 success response.
func NewListAddressTypesResponse(addressTypes []AddressTypeItem) *ListAddressTypesResponse {
	return &ListAddressTypesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 ListAddressTypesData{AddressTypes: addressTypes},
	}
}

// ─── LU-004: GET /lookup/salutations ─────────────────────────────────────────

// SalutationItem represents a single salutation option.
// VR-NFS-008: {Mr, Mrs, Ms, Shri, Smt, Dr}
type SalutationItem struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// ListSalutationsData is the payload returned on a successful LU-004 call.
type ListSalutationsData struct {
	Salutations []SalutationItem `json:"salutations"`
}

// ListSalutationsResponse is the top-level success response for LU-004.
type ListSalutationsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ListSalutationsData `json:"data"`
}

// NewListSalutationsResponse constructs the LU-004 success response.
func NewListSalutationsResponse(salutations []SalutationItem) *ListSalutationsResponse {
	return &ListSalutationsResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 ListSalutationsData{Salutations: salutations},
	}
}

// ─── LU-005: GET /lookup/document-types/:request_type ────────────────────────

// DocumentTypeItem represents a single document type and its applicability.
type DocumentTypeItem struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	IsRequired  bool   `json:"is_required"`
	RequestType string `json:"request_type"` // ADDRESS_CHANGE or NAME_CHANGE
}

// ListDocumentTypesData is the payload returned on a successful LU-005 call.
type ListDocumentTypesData struct {
	DocumentTypes []DocumentTypeItem `json:"document_types"`
	RequestType   string             `json:"request_type"`
}

// ListDocumentTypesResponse is the top-level success response for LU-005.
type ListDocumentTypesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ListDocumentTypesData `json:"data"`
}

// NewListDocumentTypesResponse constructs the LU-005 success response.
func NewListDocumentTypesResponse(requestType string, docTypes []DocumentTypeItem) *ListDocumentTypesResponse {
	return &ListDocumentTypesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 ListDocumentTypesData{DocumentTypes: docTypes, RequestType: requestType},
	}
}

// ─── LU-006: GET /lookup/offices ─────────────────────────────────────────────

// OfficeItem represents a CPC office or servicing unit.
type OfficeItem struct {
	OfficeCode string  `json:"office_code"`
	OfficeName string  `json:"office_name"`
	OfficeType string  `json:"office_type"` // SATELLITE, REGIONAL, CENTRAL
	CircleCode string  `json:"circle_code"`
	Address    *string `json:"address,omitempty"`
	IsActive   bool    `json:"is_active"`
}

// ListOfficesData is the payload returned on a successful LU-006 call.
type ListOfficesData struct {
	Offices []OfficeItem `json:"offices"`
	Total   int          `json:"total"`
}

// ListOfficesResponse is the top-level success response for LU-006.
type ListOfficesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ListOfficesData `json:"data"`
}

// NewListOfficesResponse constructs the LU-006 success response.
func NewListOfficesResponse(offices []OfficeItem) *ListOfficesResponse {
	return &ListOfficesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 ListOfficesData{Offices: offices, Total: len(offices)},
	}
}

// ─── LU-007: GET /lookup/channels ────────────────────────────────────────────

// ChannelItem represents a valid submission channel.
// VR-NFS-016: channels — Portal, Mobile, Branch, Satellite.
type ChannelItem struct {
	Code          string `json:"code"`
	Description   string `json:"description"`
	AllowsAadhaar bool   `json:"allows_aadhaar"` // BR-NFS-015
}

// ListChannelsData is the payload returned on a successful LU-007 call.
type ListChannelsData struct {
	Channels []ChannelItem `json:"channels"`
}

// ListChannelsResponse is the top-level success response for LU-007.
type ListChannelsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ListChannelsData `json:"data"`
}

// NewListChannelsResponse constructs the LU-007 success response.
func NewListChannelsResponse(channels []ChannelItem) *ListChannelsResponse {
	return &ListChannelsResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 ListChannelsData{Channels: channels},
	}
}

// ─── LU-008: GET /lookup/status-reasons ──────────────────────────────────────

// StatusReasonItem represents a pre-defined reason code.
type StatusReasonItem struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	ActionType  string `json:"action_type"` // REJECT, SEND_BACK, WITHDRAW
}

// ListStatusReasonsData is the payload returned on a successful LU-008 call.
type ListStatusReasonsData struct {
	Reasons    []StatusReasonItem `json:"reasons"`
	ActionType string             `json:"action_type,omitempty"`
}

// ListStatusReasonsResponse is the top-level success response for LU-008.
type ListStatusReasonsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ListStatusReasonsData `json:"data"`
}

// NewListStatusReasonsResponse constructs the LU-008 success response.
func NewListStatusReasonsResponse(reasons []StatusReasonItem, actionType string) *ListStatusReasonsResponse {
	return &ListStatusReasonsResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 ListStatusReasonsData{Reasons: reasons, ActionType: actionType},
	}
}

// ─── VA-001: POST /nfs/validate/address ──────────────────────────────────────

// ValidationError represents a single validation rule failure.
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidateAddressData is the payload returned on a successful VA-001 call.
type ValidateAddressData struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// ValidateAddressResponse is the top-level success response for VA-001.
type ValidateAddressResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ValidateAddressData `json:"data"`
}

// NewValidateAddressResponse constructs the VA-001 response.
func NewValidateAddressResponse(isValid bool, errors []ValidationError) *ValidateAddressResponse {
	return &ValidateAddressResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 ValidateAddressData{IsValid: isValid, Errors: errors},
	}
}

// ─── VA-002: POST /nfs/validate/name ─────────────────────────────────────────

// ValidateNameFieldsData is the payload returned on a successful VA-002 call.
// VR-NFS-014: if date_of_birth was detected in the raw payload (checked by middleware),
// IsValid will be false with IMMUTABLE_FIELD error.
type ValidateNameFieldsData struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// ValidateNameFieldsResponse is the top-level success response for VA-002.
type ValidateNameFieldsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ValidateNameFieldsData `json:"data"`
}

// NewValidateNameFieldsResponse constructs the VA-002 response.
func NewValidateNameFieldsResponse(isValid bool, errors []ValidationError) *ValidateNameFieldsResponse {
	return &ValidateNameFieldsResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 ValidateNameFieldsData{IsValid: isValid, Errors: errors},
	}
}

// ─── VA-003: GET /nfs/validate/duplicate-check ───────────────────────────────

// DuplicateCheckData is the payload returned on a successful VA-003 call.
// VR-NFS-015: one active pending request per customer+type.
type DuplicateCheckData struct {
	IsDuplicate    bool    `json:"is_duplicate"`
	ExistingTicket *string `json:"existing_ticket,omitempty"`
	Status         *string `json:"status,omitempty"`
}

// DuplicateCheckResponse is the top-level success response for VA-003.
type DuplicateCheckResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DuplicateCheckData `json:"data"`
}

// NewDuplicateCheckResponse constructs the VA-003 response.
func NewDuplicateCheckResponse(isDuplicate bool, existingTicket, status *string) *DuplicateCheckResponse {
	return &DuplicateCheckResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: DuplicateCheckData{
			IsDuplicate:    isDuplicate,
			ExistingTicket: existingTicket,
			Status:         status,
		},
	}
}
