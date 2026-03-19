package domain

import (
	"context"
	"fmt"
	"time"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
)

//Age Calculation for all PLI/RPLI products

type JointLifeAgeLookup func(ctx context.Context, ageDiff int) (int, error)

func CalculateAgeAtEntry(ctx context.Context, productCode string, proposerDOB string,
	spouseDOB *string, dateOfCalculation string, jointLifeLookup JointLifeAgeLookup) (int, error) {

	layout := "2006-01-02"

	calcDate, err := time.Parse(layout, dateOfCalculation)
	if err != nil {
		log.Error(ctx, "invalid date_of_calculation format: %v", err)
		return 0, err
	}

	calculateANB := func(dob time.Time) int {
		years := calcDate.Year() - dob.Year()
		if calcDate.Month() < dob.Month() ||
			(calcDate.Month() == dob.Month() && calcDate.Day() < dob.Day()) {
			years--
		}
		return years + 1
	}

	// ---------------------------------
	// CHILD POLICY (1006 / 5006)
	// ---------------------------------
	// if productCode == "1006" || productCode == "5006" {

	// 	if childDOB == nil {
	// 		return 0, fmt.Errorf("child_dob required for product %s", productCode)
	// 	}

	// 	childDate, err := time.Parse(layout, *childDOB)
	// 	if err != nil {
	// 		return 0, fmt.Errorf("invalid child_dob format")
	// 	}

	// 	if calcDate.Before(childDate) {
	// 		return 0, fmt.Errorf("calculation date cannot be before child_dob")
	// 	}

	// 	return calculateANB(childDate), nil
	// }

	// ---------------------------------
	// PROPOSER
	// ---------------------------------
	proposerDate, err := time.Parse(layout, proposerDOB)
	if err != nil {
		return 0, fmt.Errorf("invalid proposer dob format")
	}

	if calcDate.Before(proposerDate) {
		return 0, fmt.Errorf("calculation date cannot be before proposer dob")
	}

	proposerANB := calculateANB(proposerDate)
	ageAtEntry := proposerANB

	// ---------------------------------
	// JOINT LIFE (1005)
	// ---------------------------------
	if productCode == "1005" {

		if spouseDOB == nil {
			return 0, fmt.Errorf("spouse_dob required for product 1005")
		}

		spouseDate, err := time.Parse(layout, *spouseDOB)
		if err != nil {
			return 0, fmt.Errorf("invalid spouse_dob format")
		}

		if calcDate.Before(spouseDate) {
			return 0, fmt.Errorf("calculation date cannot be before spouse_dob")
		}

		spouseANB := calculateANB(spouseDate)

		if proposerANB < 21 || proposerANB > 45 {
			return 0, fmt.Errorf("proposer age must be between 21 and 45 years for joint life policy")
		}

		if spouseANB < 21 || spouseANB > 45 {
			return 0, fmt.Errorf("spouse age must be between 21 and 45 years for joint life policy")
		}

		elder := proposerANB
		if spouseANB > elder {
			elder = spouseANB
		}

		if elder > 45 {
			return 0, fmt.Errorf("maximum age of elder policy holder cannot exceed 45 years")
		}

		lowerAge := proposerANB
		higherAge := spouseANB

		if spouseANB < proposerANB {
			lowerAge = spouseANB
			higherAge = proposerANB
		}

		ageDiff := higherAge - lowerAge

		if jointLifeLookup == nil {
			return 0, fmt.Errorf("joint life lookup not provided")
		}

		addition, err := jointLifeLookup(ctx, ageDiff)
		if err != nil {
			return 0, err
		}

		ageAtEntry = lowerAge + addition
	}

	return ageAtEntry, nil
}

func ValidateProductAge(productCode string, age int, term int) error {

	switch productCode {

	case "1001", "1003", "5001", "5005":
		if age < 19 || age > 55 {
			return fmt.Errorf("age must be between 19 and 55 years for product %s", productCode)
		}

	case "1004", "5004":
		if age < 19 || age > 50 {
			return fmt.Errorf("age must be between 19 and 50 years for product %s", productCode)
		}

	case "5002":
		if age < 20 || age > 45 {
			return fmt.Errorf("age must be between 20 and 45 years for product %s", productCode)
		}

	case "1002", "5003":

		if age < 19 {
			return fmt.Errorf("minimum age is 19 for product %s", productCode)
		}

		if term == 20 && age > 40 {
			return fmt.Errorf("maximum age is 40 for 20 year term policy")
		}

		if term == 15 && age > 45 {
			return fmt.Errorf("maximum age is 45 for 15 year term policy")
		}

	case "1006", "5006":

		if age < 5 || age > 20 {
			return fmt.Errorf("child age must be between 5 and 20 years for product %s", productCode)
		}

	}

	return nil
}
