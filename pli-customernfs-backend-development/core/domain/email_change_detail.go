package domain

import "time"

// EmailChangeDetail stores old and new email data for email change requests.
type EmailChangeDetail struct {
	DetailID  string    `json:"detail_id" db:"detail_id"`
	RequestID string    `json:"request_id" db:"request_id"`
	OldEmail  *string   `json:"old_email,omitempty" db:"old_email"`
	NewEmail  string    `json:"new_email" db:"new_email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	UpdatedBy *string   `json:"updated_by,omitempty" db:"updated_by"`
	Version   int       `json:"version" db:"version"`
}

// EmailVersionHistory stores each version of customer email for audit.
type EmailVersionHistory struct {
	VersionID     string     `json:"version_id" db:"version_id"`
	CustomerID    int64      `json:"customer_id" db:"customer_id"`
	RequestID     *string    `json:"request_id,omitempty" db:"request_id"`
	Email         string     `json:"email" db:"email"`
	VersionNumber int        `json:"version_number" db:"version_number"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	EffectiveFrom time.Time  `json:"effective_from" db:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty" db:"effective_to"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	CreatedBy     string     `json:"created_by" db:"created_by"`
}
