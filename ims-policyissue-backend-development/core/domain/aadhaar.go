package domain

import (
	"time"
)

// AadhaarSession represents a temporary Aadhaar authentication session
type AadhaarSession struct {
	SessionID     string                 `json:"session_id" db:"session_id"`
	TransactionID string                 `json:"transaction_id" db:"transaction_id"`
	AadhaarNumber string                 `json:"aadhaar_number" db:"aadhaar_number"`
	UserData      map[string]interface{} `json:"user_data" db:"user_data"`
	OTPVerified   bool                   `json:"otp_verified" db:"otp_verified"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	ExpiresAt     time.Time              `json:"expires_at" db:"expires_at"`
}
