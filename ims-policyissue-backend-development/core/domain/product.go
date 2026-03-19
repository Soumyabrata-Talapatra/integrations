package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ProductCategory represents the product category enum
type ProductCategory string

const (
	ProductCategoryWLA     ProductCategory = "WLA"      // Whole Life Assurance
	ProductCategoryCWLA    ProductCategory = "CWLA"     // Convertible Whole Life Assurance
	ProductCategoryEA      ProductCategory = "EA"       // Endowment Assurance
	ProductCategoryAEA     ProductCategory = "AEA"      // Anticipated Endowment Assurance
	ProductCategoryJLA     ProductCategory = "JLA"      // Joint Life Assurance
	ProductCategoryChild   ProductCategory = "CHILD"    // Children's Policy
	ProductCategoryTenYear ProductCategory = "TEN_YEAR" // Ten Year Policy
)

// PolicyType represents PLI or RPLI
type PolicyType string

const (
	PolicyTypePLI  PolicyType = "PLI"
	PolicyTypeRPLI PolicyType = "RPLI"
)

// PremiumFrequency represents payment frequency
type PremiumFrequency string

const (
	FrequencyMonthly    PremiumFrequency = "MONTHLY"
	FrequencyQuarterly  PremiumFrequency = "QUARTERLY"
	FrequencyHalfYearly PremiumFrequency = "HALF_YEARLY"
	FrequencyYearly     PremiumFrequency = "YEARLY"
)

// StringArray for JSONB array storage
type StringArray []string

// Value implements driver.Valuer
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner with safe type handling
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into StringArray", value)
	}

	return json.Unmarshal(bytes, a)
}

// IntArray for JSONB array storage
type IntArray []int

// Value implements driver.Valuer
func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner with safe type handling
func (a *IntArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into IntArray", value)
	}

	return json.Unmarshal(bytes, a)
}

// Product represents a PLI/RPLI product from the catalog
type Product struct {
	ProductCode              string          `db:"product_code" json:"product_code"`
	ProductName              string          `db:"product_name" json:"product_name"`
	ProductType              PolicyType      `db:"product_type" json:"product_type"`
	ProductCategory          ProductCategory `db:"product_category" json:"product_category"`
	MinSumAssured            float64         `db:"min_sum_assured" json:"min_sum_assured"`
	MaxSumAssured            *float64        `db:"max_sum_assured" json:"max_sum_assured,omitempty"`
	MinEntryAge              int             `db:"min_entry_age" json:"min_entry_age"`
	MaxEntryAge              int             `db:"max_entry_age" json:"max_entry_age"`
	MaxMaturityAge           *int            `db:"max_maturity_age" json:"max_maturity_age,omitempty"`
	MinTerm                  int             `db:"min_term" json:"min_term"`
	PremiumCeasingAgeOptions IntArray        `db:"premium_ceasing_age_options" json:"premium_ceasing_age_options,omitempty"`
	AvailableFrequencies     StringArray     `db:"available_frequencies" json:"available_frequencies"`
	MedicalSAThreshold       *float64        `db:"medical_sa_threshold" json:"medical_sa_threshold,omitempty"`
	IsSADecreaseAllowed      bool            `db:"is_sa_decrease_allowed" json:"is_sa_decrease_allowed"`
	IsActive                 bool            `db:"is_active" json:"is_active"`
	EffectiveFrom            time.Time       `db:"effective_from" json:"effective_from"`
	EffectiveTo              *time.Time      `db:"effective_to" json:"effective_to,omitempty"`
	Description              *string         `db:"description" json:"description,omitempty"`
	CreatedAt                time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt                *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
	PlanCode                 *string         `db:"plan_code" json:"plan_code,omitempty"`
	Status                   *string         `db:"status" json:"status,omitempty"`
	CloseDate                *string         `db:"close_date" json:"close_date,omitempty"`
	CreatedBy                *string         `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy                *string         `db:"updated_by" json:"updated_by,omitempty"`
	ProductCodeOld           *int            `db:"product_code_old" json:"product_code_old,omitempty"`
	EditAllowed              *bool           `db:"edit_allowed" json:"edit_allowed,omitempty"`
}

// IsEligibleAge checks if the given age is within product entry age limits
func (p *Product) IsEligibleAge(age int) bool {
	return age >= p.MinEntryAge && age <= p.MaxEntryAge
}

// IsEligibleSA checks if the given sum assured is within product limits
func (p *Product) IsEligibleSA(sa float64) bool {
	if sa < p.MinSumAssured {
		return false
	}
	if p.MaxSumAssured != nil && sa > *p.MaxSumAssured {
		return false
	}
	return true
}

// IsFrequencyAllowed checks if the payment frequency is allowed for this product
func (p *Product) IsFrequencyAllowed(freq PremiumFrequency) bool {
	for _, f := range p.AvailableFrequencies {
		if f == string(freq) {
			return true
		}
	}
	return false
}

// IsMedicalRequired checks if medical examination is required based on SA
func (p *Product) IsMedicalRequired(sa float64) bool {
	if p.MedicalSAThreshold == nil {
		return false
	}
	return sa >= *p.MedicalSAThreshold
}
