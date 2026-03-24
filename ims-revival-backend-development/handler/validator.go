package handler

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	validation "gitlab.cept.gov.in/it-2.0-common/api-validation"
)

func CommunityEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "UR", "OBC", "SC", "ST", "EWS", "NA":
		return true
	default:
		return false
	}
}
func ValidateEightDigitsInt64(fl validator.FieldLevel) bool {
	value := fl.Field().Int()

	// Check if the value is in the range of exactly 8 digits
	return value >= 10000000 && value <= 99999999
}
func GenderEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "Male", "Female", "Transgender-Male", "Transgender-Female", "NA":
		return true
	default:
		return false
	}
}

func GroupEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "Group A", "Group B Gazetted", "Group B Non Gazetted", "Group C", "GDS":
		return true
	default:
		return false
	}
}

func MaritalEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "Married", "UnMarried", "Divorcee", "Widower", "NA":
		return true
	default:
		return false
	}
}

func EmpTypeEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "DOP", "GDS":
		return true
	default:
		return false
	}
}

func RecruitEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "Sports", "DR", "DP", "Compassionate":
		return true
	default:
		return false
	}
}

func TaxEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "Old", "New":
		return true
	default:
		return false
	}
}

func DepTypeEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "APS", "DOP", "IPPB", "OtherDepartment":
		return true
	default:
		return false
	}
}

func PensionSchemeEnumValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "GPF", "NPS", "UPS", "SDBS", "Non-SDBS", "NA":
		return true
	default:
		return false
	}
}

func AddressFieldValidator(fl validator.FieldLevel) bool {
	addressPattern := regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9\s,./#-]{0,148}[A-Za-z0-9]$`)
	value := fl.Field().String()
	return addressPattern.MatchString(value)
}

func RemarksValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Normalize non-breaking spaces
	value = strings.ReplaceAll(value, "\u00A0", " ")
	remarksPattern := regexp.MustCompile(`^[A-Za-z0-9\s,./()\-]{1,200}$`)
	return remarksPattern.MatchString(value)
}

func CreatedByValidator(fl validator.FieldLevel) bool {
	val, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// Step 1: Validate exact 8-digit format
	if matched, _ := regexp.MatchString(`^\d{8}$`, val); !matched {
		return false
	}

	// Step 2: Convert to integer
	num, err := strconv.Atoi(val)
	if err != nil {
		return false
	}

	// Step 3: Check numeric range
	return num >= 10000000 && num <= 99999999
}

func NewValidatorService() error {

	err := validation.Create()
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("community_enum", CommunityEnumValidator, "invalid value for %s, must be one of [UR, OBC, SC, ST, EWS, NA] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("gender_enum", GenderEnumValidator, "invalid value for %s, must be one of [Male, Female, Transgender-Male, Transgender-Female, NA] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("group_enum", GroupEnumValidator, "invalid value for %s, must be one of [Group A, Group B Gazetted, Group B Non Gazetted, Group C, GDS] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("marital_enum", MaritalEnumValidator, "invalid value for %s, must be one of [Married, UnMarried, Divorcee, Widower, NA] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("emp_type_enum", EmpTypeEnumValidator, "invalid value for %s, must be one of [DOP, GDS] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("recruit_enum", RecruitEnumValidator, "invalid value for %s, must be one of [Sports, DR, DP, Compassionate] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("tax_enum", TaxEnumValidator, "invalid value for %s, must be one of [Old, New] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("dep_type_enum", DepTypeEnumValidator, "invalid value for %s, must be one of [APS, DOP, IPPB, OtherDepartment] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("pension_scheme_enum", PensionSchemeEnumValidator, "invalid value for %s, must be one of [GPF,NPS,UPS,SDBS,Non-SDBS] but received %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("address_field", AddressFieldValidator, "field %s must start and end with an alphanumeric character, and may contain letters, digits, spaces, commas, periods, and hyphens in between. The total length should be between 3 and 150 characters, but received %v")
	if err != nil {
		return err
	}
	err = validation.RegisterCustomValidation("remarks", RemarksValidator, "field %s must start and end with an alphanumeric character, and may contain letters, digits, spaces, commas, periods, parentheses, and hyphens. The total length should be between 1 and 200 characters, but received %v")
	if err != nil {
		return err
	}
	err = validation.RegisterCustomValidation("created_by", CreatedByValidator, "field %s must be a valid employee ID, but got %v")
	if err != nil {
		return err
	}

	err = validation.RegisterCustomValidation("eightdigitidint64", ValidateEightDigitsInt64, "field %s must consist of 8 digits, but received %v")
	if err != nil {
		return nil
	}
	return nil
}
