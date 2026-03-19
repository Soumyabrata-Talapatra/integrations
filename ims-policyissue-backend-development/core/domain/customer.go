package domain

import "time"

type CustomerAddress struct {
	AddressID     int64      `db:"address_id"`
	CustomerID    string     `db:"customer_id"`
	AddressType   string     `db:"address_type"`
	Line1         string     `db:"line1"`
	Line2         *string    `db:"line2"`
	Village       *string    `db:"village"`
	Taluka        *string    `db:"taluka"`
	City          string     `db:"city"`
	District      string     `db:"district"`
	State         string     `db:"state"`
	Country       string     `db:"country"`
	PinCode       string     `db:"pin_code"`
	Version       int        `db:"version"`
	IsActive      bool       `db:"is_active"`
	EffectiveFrom time.Time  `db:"effective_from"`
	EffectiveTo   *time.Time `db:"effective_to"`
	ChangeReason  *string    `db:"change_reason"`
	ApprovedBy    *string    `db:"approved_by"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

type CustomerContact struct {
	ID           int64     `db:"contact_id"`
	CustomerID   string    `db:"customer_id"`
	ContactType  string    `db:"contact_type"`
	ContactValue string    `db:"contact_value"`
	IsPrimary    bool      `db:"is_primary"`
	IsVerified   bool      `db:"is_verified"`
	IsActive     bool      `db:"is_active"`
	ChangeReason *string   `db:"change_reason"`
	ApprovedBy   *string   `db:"approved_by"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
 
type CustomerEmployment struct {
	ID                   int64      `db:"employemnt_id"`
	CustomerID           string     `db:"customer_id"`
	Occupation           string     `db:"occupation"`
	PAODDOCode           *string    `db:"pao_ddo_code"`
	Organization         *string    `db:"organization"`
	Designation          *string    `db:"designation"`
	DateOfEntry          *time.Time `db:"date_of_entry"`
	SuperiorDesignation  *string    `db:"superior_designation"`
	MonthlyIncome        *float64   `db:"monthly_income"`
	Qualification        *string    `db:"qualification"`
	IsActive             bool       `db:"is_active"`
	ChangeReason         *string    `db:"change_reason"`
	ApprovedBy           *string    `db:"approved_by"`
	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
}
