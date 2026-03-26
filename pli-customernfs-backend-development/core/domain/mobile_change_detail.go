package domain

import "time"

// MobileChangeDetail stores old and new mobile data for mobile change requests.
type MobileChangeDetail struct {
	DetailID        string    `json:"detail_id" db:"detail_id"`
	RequestID       string    `json:"request_id" db:"request_id"`
	OldMobileNumber *string   `json:"old_mobile_number,omitempty" db:"old_mobile_number"`
	NewMobileNumber string    `json:"new_mobile_number" db:"new_mobile_number"`
	AadhaarTxnID    *string   `json:"aadhaar_txn_id,omitempty" db:"aadhaar_txn_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	UpdatedBy       *string   `json:"updated_by,omitempty" db:"updated_by"`
	Version         int       `json:"version" db:"version"`
}

// MobileVersionHistory stores each version of customer mobile for audit.
type MobileVersionHistory struct {
	VersionID     string     `json:"version_id" db:"version_id"`
	CustomerID    int64      `json:"customer_id" db:"customer_id"`
	RequestID     *string    `json:"request_id,omitempty" db:"request_id"`
	MobileNumber  string     `json:"mobile_number" db:"mobile_number"`
	VersionNumber int        `json:"version_number" db:"version_number"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	EffectiveFrom time.Time  `json:"effective_from" db:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty" db:"effective_to"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	CreatedBy     string     `json:"created_by" db:"created_by"`
}
