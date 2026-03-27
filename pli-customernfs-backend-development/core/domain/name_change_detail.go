package domain

import "time"

// NameChangeDetail stores old and new name data for name change requests.
// Entity ID: E-3 (per Customer_NFS_Service_Analysis.md Section 7.3)
//
// Business Rules applied:
//   - BR-NFS-007: Aadhaar path — aadhaar_txn_id is set, UIDAI name is authoritative
//   - BR-NFS-008: Manual path — requires Gazette/Newspaper/Application Form (at least 1)
//   - BR-NFS-009: Name change propagates to ALL linked policies via event
//   - BR-NFS-010: DOB is IMMUTABLE — must not be included in name change request
//
// Validation Rules:
//   - VR-NFS-006: first_name not empty, alpha+spaces only, max 100 chars
//   - VR-NFS-007: last_name not empty, alpha+spaces only, max 100 chars
//   - VR-NFS-008: salutation must be in {Mr, Mrs, Ms, Shri, Smt, Dr}
//   - VR-NFS-014: DOB must NOT be present in the request
//
// BATCH NOTE: Created together with nfs.service_request in a single batch TX:
//   batch := pgx.Batch{}
//   → INSERT nfs.service_request (RETURNING request_id)
//   → INSERT nfs.name_change_detail (using returned request_id)
//   → INSERT nfs.audit_log (CREATED action)
// See: repo/postgres/name_change.go CreateWithServiceRequest()
type NameChangeDetail struct {
	DetailID  string `json:"detail_id" db:"detail_id"`
	RequestID string `json:"request_id" db:"request_id"`
	// Old Name (captured at request time for audit and rollback)
	OldSalutation *string `json:"old_salutation,omitempty" db:"old_salutation"`
	OldFirstName  *string `json:"old_first_name,omitempty" db:"old_first_name"`
	OldMiddleName *string `json:"old_middle_name,omitempty" db:"old_middle_name"`
	OldLastName   *string `json:"old_last_name,omitempty" db:"old_last_name"`
	// New Name (proposed values, validated per VR-NFS-006/007/008)
	NewSalutation *string `json:"new_salutation" db:"new_salutation"`
	NewFirstName  *string `json:"new_first_name" db:"new_first_name"`
	NewMiddleName *string `json:"new_middle_name,omitempty" db:"new_middle_name"`
	NewLastName   *string `json:"new_last_name" db:"new_last_name"`
	// Cross-policy impact count (BR-NFS-009, set on completion)
	PoliciesAffected int `json:"policies_affected" db:"policies_affected"`
	// Aadhaar reference (BR-NFS-007: present if auth_method=AADHAAR)
	AadhaarTxnID *string `json:"aadhaar_txn_id,omitempty" db:"aadhaar_txn_id"`
	// Version tracking
	OldNameVersion *int      `json:"old_name_version,omitempty" db:"old_name_version"`
	NewNameVersion *int      `json:"new_name_version,omitempty" db:"new_name_version"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	UpdatedBy      *string   `json:"updated_by,omitempty" db:"updated_by"`
	Version        int       `json:"version" db:"version"`
}

// NameVersionHistory stores each version of customer name for audit and rollback.
// Used by BR-NFS-009 (Cross-Policy Propagation) and Compensation workflows.
// Compensation: CompensateNameChange restores from this entity (Section 14.2).
//
// BATCH NOTE: UpdateIdentityActivity writes 3 records in one TX:
//   1. UPDATE nfs.name_version_history SET is_active=false (old version)
//   2. INSERT nfs.name_version_history (new version, is_active=true)
//   3. INSERT nfs.audit_log (APPROVED action, with policies_affected count)
// See: repo/postgres/name_change.go CreateNameVersion()
type NameVersionHistory struct {
	VersionID     string     `json:"version_id" db:"version_id"`
	CustomerID    int64      `json:"customer_id" db:"customer_id"`
	RequestID     *string    `json:"request_id,omitempty" db:"request_id"`
	Salutation    *string    `json:"salutation,omitempty" db:"salutation"`
	FirstName     string     `json:"first_name" db:"first_name"`
	MiddleName    *string    `json:"middle_name,omitempty" db:"middle_name"`
	LastName      string     `json:"last_name" db:"last_name"`
	VersionNumber int        `json:"version_number" db:"version_number"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	EffectiveFrom time.Time  `json:"effective_from" db:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty" db:"effective_to"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	CreatedBy     int64      `json:"created_by" db:"created_by"`
}
