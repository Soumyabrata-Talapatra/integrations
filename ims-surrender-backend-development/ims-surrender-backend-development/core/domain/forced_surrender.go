package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReminderLevel represents the forced surrender reminder level
// Business Rule: BR-FS-002, BR-FS-003, BR-FS-004
type ReminderLevel string

const (
	ReminderLevelFirst  ReminderLevel = "FIRST"  // 95% threshold
	ReminderLevelSecond ReminderLevel = "SECOND" // 98% threshold
	ReminderLevelThird  ReminderLevel = "THIRD"  // 100% threshold
)

// ForcedSurrenderReminder represents a forced surrender reminder sent to policyholder
// Table: forced_surrender_reminders
// Business Rules: BR-FS-002, BR-FS-003, BR-FS-004
type ForcedSurrenderReminder struct {
	ID                      uuid.UUID              `json:"id" db:"id"`
	PolicyID                string                 `json:"policy_id" db:"policy_id"`
	ReminderNumber          ReminderLevel          `json:"reminder_number" db:"reminder_number"`
	ReminderDate            time.Time              `json:"reminder_date" db:"reminder_date"`
	LoanCapitalizationRatio float64                `json:"loan_capitalization_ratio" db:"loan_capitalization_ratio"`
	LoanPrincipal           float64                `json:"loan_principal" db:"loan_principal"`
	LoanInterest            float64                `json:"loan_interest" db:"loan_interest"`
	GrossSurrenderValue     float64                `json:"gross_surrender_value" db:"gross_surrender_value"`
	LetterSent              bool                   `json:"letter_sent" db:"letter_sent"`
	SMSSent                 bool                   `json:"sms_sent" db:"sms_sent"`
	LetterReference         *string                `json:"letter_reference" db:"letter_reference"`
	SMSReference            *string                `json:"sms_reference" db:"sms_reference"`
	CreatedAt               time.Time              `json:"created_at" db:"created_at"`
	Metadata                map[string]interface{} `json:"metadata" db:"metadata"`
}

// ForcedSurrenderPaymentWindow represents the 30-day payment window after 3rd reminder
// Table: forced_surrender_payment_windows
// Business Rules: BR-FS-006, BR-FS-007
type ForcedSurrenderPaymentWindow struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	SurrenderRequestID  uuid.UUID  `json:"surrender_request_id" db:"surrender_request_id"`
	PolicyID            string     `json:"policy_id" db:"policy_id"`
	WindowStartDate     time.Time  `json:"window_start_date" db:"window_start_date"`
	WindowEndDate       time.Time  `json:"window_end_date" db:"window_end_date"`
	PaymentReceived     bool       `json:"payment_received" db:"payment_received"`
	PaymentReceivedAt   *time.Time `json:"payment_received_at" db:"payment_received_at"`
	PaymentAmount       *float64   `json:"payment_amount" db:"payment_amount"`
	PaymentReference    *string    `json:"payment_reference" db:"payment_reference"`
	WorkflowForwarded   bool       `json:"workflow_forwarded" db:"workflow_forwarded"`
	WorkflowForwardedAt *time.Time `json:"workflow_forwarded_at" db:"workflow_forwarded_at"`
	AutoCompleted       bool       `json:"auto_completed" db:"auto_completed"`
	AutoCompletedAt     *time.Time `json:"auto_completed_at" db:"auto_completed_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
}
