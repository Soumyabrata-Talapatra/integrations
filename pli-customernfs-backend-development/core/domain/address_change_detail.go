package domain

import "time"

// AddressChangeDetail stores old and new address data for address change requests.
// Entity ID: E-2 (per Customer_NFS_Service_Analysis.md Section 7.2)
//
// Business Rules applied:
//   - BR-NFS-001: Aadhaar path — aadhaar_txn_id is set, immediate completion
//   - BR-NFS-002: Manual path — CPC approval required, documents uploaded
//   - BR-NFS-003: If address_type=COMMUNICATION, propagates to downstream services
//   - BR-NFS-004: Address versioning — old_address_version + new_address_version
//   - BR-NFS-005: address_update_for must be INSURED (others delegated to Policy Service)
//   - BR-NFS-006: Pincode must match selected state (VR-NFS-001)
//
// Validation Rules:
//   - VR-NFS-001: pincode 6 digits AND matches state
//   - VR-NFS-002: state must be valid Indian state
//   - VR-NFS-003: address_line1 not empty, max 200 chars
//   - VR-NFS-004: city not empty, max 100 chars
//   - VR-NFS-005: district not empty, max 100 chars
//   - VR-NFS-013: address_update_for must be INSURED from NFS service
//
// BATCH NOTE: Created together with nfs.service_request in a single batch TX:
//   batch := pgx.Batch{}
//   → INSERT nfs.service_request (RETURNING request_id + workflow_id)
//   → INSERT nfs.address_change_detail (using returned request_id)
//   → INSERT nfs.audit_log (CREATED action, BR-NFS-016)
// This ensures atomicity — no partial state can exist.
// See: repo/postgres/address_change.go CreateWithServiceRequest()
type AddressChangeDetail struct {
	DetailID         string `json:"detail_id" db:"detail_id"`
	RequestID        string `json:"request_id" db:"request_id"`
	AddressUpdateFor string `json:"address_update_for" db:"address_update_for"`
	AddressType      string `json:"address_type" db:"address_type"`
	// Old Address (captured for audit, BR-NFS-004)
	OldAddressLine1 *string `json:"old_address_line1,omitempty" db:"old_address_line1"`
	OldAddressLine2 *string `json:"old_address_line2,omitempty" db:"old_address_line2"`
	OldVillage      *string `json:"old_village,omitempty" db:"old_village"`
	OldTaluka       *string `json:"old_taluka,omitempty" db:"old_taluka"`
	OldCity         *string `json:"old_city,omitempty" db:"old_city"`
	OldDistrict     *string `json:"old_district,omitempty" db:"old_district"`
	OldState        *string `json:"old_state,omitempty" db:"old_state"`
	OldPincode      *string `json:"old_pincode,omitempty" db:"old_pincode"`
	// New Address (proposed values)
	NewAddressLine1 string  `json:"new_address_line1" db:"new_address_line1"`
	NewAddressLine2 *string `json:"new_address_line2,omitempty" db:"new_address_line2"`
	NewVillage      *string `json:"new_village,omitempty" db:"new_village"`
	NewTaluka       *string `json:"new_taluka,omitempty" db:"new_taluka"`
	NewCity         string  `json:"new_city" db:"new_city"`
	NewDistrict     string  `json:"new_district" db:"new_district"`
	NewState        string  `json:"new_state" db:"new_state"`
	NewPincode      string  `json:"new_pincode" db:"new_pincode"`
	// Aadhaar reference (BR-NFS-001: present if auth_method=AADHAAR)
	AadhaarTxnID *string `json:"aadhaar_txn_id,omitempty" db:"aadhaar_txn_id"`
	// Version tracking (BR-NFS-004: address versioning — never overwrite)
	OldAddressVersion *int      `json:"old_address_version,omitempty" db:"old_address_version"`
	NewAddressVersion *int      `json:"new_address_version,omitempty" db:"new_address_version"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy         string    `json:"created_by" db:"created_by"`
	UpdatedBy         *string   `json:"updated_by,omitempty" db:"updated_by"`
	Version           int       `json:"version" db:"version"`
}

// AddressVersionHistory stores each version of an address for audit and rollback.
// Used by BR-NFS-004 (Address Versioning — Never Overwrite).
// Compensation: CompensateAddressUpdate restores from this entity.
//
// BATCH NOTE: UpdateAddressActivity writes 3 records in one TX:
//   1. UPDATE nfs.address_version_history SET is_active=false (old version)
//   2. INSERT nfs.address_version_history (new version, is_active=true)
//   3. INSERT nfs.audit_log (APPROVED action)
// See: repo/postgres/address_change.go CreateAddressVersion()
type AddressVersionHistory struct {
	VersionID     string     `json:"version_id" db:"version_id"`
	CustomerID    string     `json:"customer_id" db:"customer_id"`
	RequestID     *string    `json:"request_id,omitempty" db:"request_id"`
	AddressType   string     `json:"address_type" db:"address_type"`
	AddressLine1  string     `json:"address_line1" db:"address_line1"`
	AddressLine2  *string    `json:"address_line2,omitempty" db:"address_line2"`
	Village       *string    `json:"village,omitempty" db:"village"`
	Taluka        *string    `json:"taluka,omitempty" db:"taluka"`
	City          string     `json:"city" db:"city"`
	District      string     `json:"district" db:"district"`
	State         string     `json:"state" db:"state"`
	Pincode       string     `json:"pincode" db:"pincode"`
	VersionNumber int        `json:"version_number" db:"version_number"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	EffectiveFrom time.Time  `json:"effective_from" db:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty" db:"effective_to"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	CreatedBy     string     `json:"created_by" db:"created_by"`
}
