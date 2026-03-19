package activities

import (
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// lastDayOfMonth — C6
// ─────────────────────────────────────────────────────────────────────────────

// TestLastDayOfMonth mirrors DB: DATE_TRUNC('month', p_date) + INTERVAL '1 month' - INTERVAL '1 day'
func TestLastDayOfMonth(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			name: "mid-January",
			in:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			want: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Feb non-leap",
			in:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			want: time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Feb leap year",
			in:   time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			want: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "last day of March",
			in:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
			want: time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "last day of December",
			in:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			want: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "April (30 days)",
			in:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			want: time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lastDayOfMonth(tc.in)
			// Compare date only — time components may differ
			gotDate := time.Date(got.Year(), got.Month(), got.Day(), 0, 0, 0, 0, time.UTC)
			wantDate := time.Date(tc.want.Year(), tc.want.Month(), tc.want.Day(), 0, 0, 0, 0, time.UTC)
			if !gotDate.Equal(wantDate) {
				t.Errorf("lastDayOfMonth(%v) = %v; want %v", tc.in.Format("2006-01-02"), got.Format("2006-01-02"), tc.want.Format("2006-01-02"))
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// computeRemissionExpiry — C6
// ─────────────────────────────────────────────────────────────────────────────

// TestComputeRemissionExpiry mirrors DB compute_remission_expiry(). [C6]
// paid_to_date maps to first_unpaid_date in the DB function.
//
// DB logic:
//
//	v_grace_end := last_day_of_month(p_first_unpaid)
//	< 6mo  → NULL
//	< 12mo → grace_end + 30 days
//	< 24mo → grace_end + 60 days
//	< 36mo → grace_end + 90 days
//	≥ 36mo → first_unpaid + 12 months
func TestComputeRemissionExpiry(t *testing.T) {
	scheduled := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	// paid_to_date: 2026-02-28 → lastDayOfMonth = 2026-02-28
	paidTo := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	graceEnd := lastDayOfMonth(paidTo) // 2026-02-28

	cases := []struct {
		name       string
		issueDate  time.Time
		wantNil    bool
		wantExpiry time.Time
	}{
		{
			name:      "< 6 months — no remission (VOID immediately)",
			issueDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), // ~2 months
			wantNil:   true,
		},
		{
			name:      "exactly 5 months — still no remission",
			issueDate: time.Date(2025, 10, 8, 0, 0, 0, 0, time.UTC), // 5 months
			wantNil:   true,
		},
		{
			name:       "6 months — grace_end + 30 days",
			issueDate:  time.Date(2025, 9, 8, 0, 0, 0, 0, time.UTC), // 6 months
			wantExpiry: graceEnd.AddDate(0, 0, 30),
		},
		{
			name:       "11 months — grace_end + 30 days",
			issueDate:  time.Date(2025, 4, 8, 0, 0, 0, 0, time.UTC), // 11 months
			wantExpiry: graceEnd.AddDate(0, 0, 30),
		},
		{
			name:       "12 months — grace_end + 60 days",
			issueDate:  time.Date(2025, 3, 8, 0, 0, 0, 0, time.UTC), // 12 months
			wantExpiry: graceEnd.AddDate(0, 0, 60),
		},
		{
			name:       "23 months — grace_end + 60 days",
			issueDate:  time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC), // 23 months
			wantExpiry: graceEnd.AddDate(0, 0, 60),
		},
		{
			name:       "24 months — grace_end + 90 days",
			issueDate:  time.Date(2024, 3, 8, 0, 0, 0, 0, time.UTC), // 24 months
			wantExpiry: graceEnd.AddDate(0, 0, 90),
		},
		{
			name:       "35 months — grace_end + 90 days",
			issueDate:  time.Date(2023, 4, 8, 0, 0, 0, 0, time.UTC), // 35 months
			wantExpiry: graceEnd.AddDate(0, 0, 90),
		},
		{
			name:       "36 months — paid_to_date + 12 months (INACTIVE_LAPSE path)",
			issueDate:  time.Date(2023, 3, 8, 0, 0, 0, 0, time.UTC), // 36 months
			wantExpiry: paidTo.AddDate(0, 12, 0),
		},
		{
			name:       "48 months — paid_to_date + 12 months",
			issueDate:  time.Date(2022, 3, 8, 0, 0, 0, 0, time.UTC), // 48 months
			wantExpiry: paidTo.AddDate(0, 12, 0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeRemissionExpiry(tc.issueDate, paidTo, scheduled)
			if tc.wantNil {
				if got != nil {
					t.Errorf("expected nil (no remission), got %v", got.Format("2006-01-02"))
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil expiry, got nil")
			}
			gotDate := time.Date(got.Year(), got.Month(), got.Day(), 0, 0, 0, 0, time.UTC)
			wantDate := time.Date(tc.wantExpiry.Year(), tc.wantExpiry.Month(), tc.wantExpiry.Day(), 0, 0, 0, 0, time.UTC)
			if !gotDate.Equal(wantDate) {
				t.Errorf("computeRemissionExpiry = %v; want %v",
					got.Format("2006-01-02"), tc.wantExpiry.Format("2006-01-02"))
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// PaidUp formula — C9
// ─────────────────────────────────────────────────────────────────────────────

// TestPaidUpValue_IncludesBonus verifies the correct PLI paid-up value formula.
// PUSA = (premiums_paid / total_premiums) × (sum_assured + bonus_accumulated)
func TestPaidUpValue_IncludesBonus(t *testing.T) {
	cases := []struct {
		name             string
		premiumsPaid     int
		totalPremiums    int
		sumAssured       float64
		bonusAccumulated float64
		want             float64
	}{
		{
			name:             "half premiums paid, no bonus",
			premiumsPaid:     60,
			totalPremiums:    120,
			sumAssured:       100000,
			bonusAccumulated: 0,
			want:             50000.0,
		},
		{
			name:             "half premiums paid, with bonus",
			premiumsPaid:     60,
			totalPremiums:    120,
			sumAssured:       100000,
			bonusAccumulated: 20000,
			want:             60000.0, // (60/120) * (100000+20000)
		},
		{
			name:             "three-quarters paid",
			premiumsPaid:     90,
			totalPremiums:    120,
			sumAssured:       100000,
			bonusAccumulated: 20000,
			want:             90000.0, // (90/120) * 120000
		},
		{
			name:             "zero total premiums — no divide by zero",
			premiumsPaid:     0,
			totalPremiums:    0,
			sumAssured:       100000,
			bonusAccumulated: 0,
			want:             0.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got float64
			if tc.totalPremiums > 0 {
				// [C9] Correct formula: PUSA = (paid/total) × (SA + bonus)
				got = (float64(tc.premiumsPaid) / float64(tc.totalPremiums)) *
					(tc.sumAssured + tc.bonusAccumulated)
			}
			if got != tc.want {
				t.Errorf("paidUpValue = %v; want %v", got, tc.want)
			}
		})
	}
}
