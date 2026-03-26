// Package domain defines reference/lookup domain entities.
// These are master-data types populated from nfs.ref_* tables.
//
// FR-NFS-001: state and pincode lookups for address change form.
// LU-001..008: all lookup endpoints read from these types.
package domain

import "time"

// SLADeadline represents the resolved SLA deadline for a service request.
// Calculated from cfg.GetInt("nfs.sladeadlineregional") and cfg.GetInt("nfs.sladeadlinemanual").
// BR-NFS-011: SLA window starts from ticket creation date.
type SLADeadline struct {
	RegionalDays int
	ManualDays   int
}

// State represents a postal/administrative state in India.
// Referenced by LU-001 and VA-001 (BR-NFS-006 pincode ↔ state cross-validation).
type State struct {
	StateCode  string    `json:"state_code" db:"state_code"`
	StateName  string    `json:"state_name" db:"state_name"`
	CircleCode *string   `json:"circle_code,omitempty" db:"circle_code"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// PincodeData represents geo data for a service pincode.
// Referenced by LU-002 and VA-001 (BR-NFS-006: pincode-state validation).
// VR-NFS-013: pincode must be exactly 6 digits.
type PincodeData struct {
	Pincode   string `json:"pincode" db:"pincode"`
	StateCode string `json:"state_code" db:"state_code"`
	StateName string `json:"state_name" db:"state_name"`
	District  string `json:"district" db:"district"`
	City      string `json:"city" db:"city"`
	IsActive  bool   `json:"is_active" db:"is_active"`
}

// DocumentType defines a required or optional document for a specific request type.
// Referenced by LU-005.
// FR-NFS-002: Address change — Address Proof mandatory.
// FR-NFS-005: Name change — at least one of Gazette/Newspaper required (BR-NFS-008).
type DocumentType struct {
	DocTypeCode string `json:"doc_type_code" db:"doc_type_code"`
	DocTypeName string `json:"doc_type_name" db:"doc_type_name"`
	RequestType string `json:"request_type" db:"request_type"`
	IsMandatory bool   `json:"is_mandatory" db:"is_mandatory"`
	Description string `json:"description" db:"description"`
	IsActive    bool   `json:"is_active" db:"is_active"`
}

// Office is a CPC/Regional/Head office record.
// Referenced by LU-006. Used for BR-NFS-002 (routing to servicing office).
type Office struct {
	OfficeCode string  `json:"office_code" db:"office_code"`
	OfficeName string  `json:"office_name" db:"office_name"`
	OfficeType string  `json:"office_type" db:"office_type"` // CPC | RO | HO
	CircleCode *string `json:"circle_code,omitempty" db:"circle_code"`
	StateCode  string  `json:"state_code" db:"state_code"`
	Address    string  `json:"address" db:"address"`
	Pincode    string  `json:"pincode" db:"pincode"`
	IsActive   bool    `json:"is_active" db:"is_active"`
}

// StatusReason is a configurable rejection/send-back reason code.
// Referenced by LU-008. Used in CPC-004 (reject) and CORE-010 / CPC-005 (send-back).
type StatusReason struct {
	ReasonCode string `json:"reason_code" db:"reason_code"`
	ReasonText string `json:"reason_text" db:"reason_text"`
	ActionType string `json:"action_type" db:"action_type"` // REJECT | SEND_BACK
	IsActive   bool   `json:"is_active" db:"is_active"`
}

// NameVersionHistory tracks versioned name data for a customer.
// Created by repo/postgres/name_change.go CreateNameVersion().
// BR-NFS-009: all linked policies updated on version commit.
// type NameVersionHistory struct {
// 	VersionID     string     `json:"version_id" db:"version_id"`
// 	CustomerID    string     `json:"customer_id" db:"customer_id"`
// 	RequestID     string     `json:"request_id" db:"request_id"`
// 	Salutation    *string    `json:"salutation,omitempty" db:"salutation"`
// 	FirstName     string     `json:"first_name" db:"first_name"`
// 	MiddleName    *string    `json:"middle_name,omitempty" db:"middle_name"`
// 	LastName      string     `json:"last_name" db:"last_name"`
// 	VersionNumber int        `json:"version_number" db:"version_number"`
// 	IsActive      bool       `json:"is_active" db:"is_active"`
// 	EffectiveFrom time.Time  `json:"effective_from" db:"effective_from"`
// 	EffectiveTo   *time.Time `json:"effective_to,omitempty" db:"effective_to"`
// 	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
// }
