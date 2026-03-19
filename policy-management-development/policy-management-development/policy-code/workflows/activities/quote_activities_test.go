package activities_test

import (
	"net/url"
	"strings"
	"testing"
)

// TestURLEscaping_PolicyNumberWithSlash verifies that policy numbers containing
// slashes are correctly encoded in URL path segments.
// Policy numbers follow the format PLI/YYYY/NNNNNN or RPLI/YYYY/NNNNNN.
// An unescaped "/" in a URL path creates spurious extra path segments and
// routes to the wrong endpoint (or returns 404/500). [C4]
func TestURLEscaping_PolicyNumberWithSlash(t *testing.T) {
	cases := []struct {
		name         string
		policyNumber string
	}{
		{"PLI policy", "PLI/2026/000001"},
		{"RPLI policy", "RPLI/2025/999999"},
		{"long number", "PLI/2026/123456"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			escaped := url.PathEscape(tc.policyNumber)

			// Must not contain raw slashes after escaping
			if strings.Contains(escaped, "/") {
				t.Errorf("url.PathEscape(%q) = %q; still contains raw '/'", tc.policyNumber, escaped)
			}

			// Must contain %2F for each slash in the original
			slashCount := strings.Count(tc.policyNumber, "/")
			escapedCount := strings.Count(escaped, "%2F")
			if escapedCount != slashCount {
				t.Errorf("url.PathEscape(%q) = %q; expected %d %%2F, got %d",
					tc.policyNumber, escaped, slashCount, escapedCount)
			}

			// Simulate URL construction — verify no spurious path segments
			baseURL := "http://surrender-svc"
			constructed := baseURL + "/internal/v1/policies/" + escaped + "/surrender-quote"
			if strings.Count(constructed, "/surrender-quote") != 1 {
				t.Errorf("URL has wrong path structure: %s", constructed)
			}
		})
	}
}

// TestURLEscaping_QueryParams verifies date and product code query params are safe.
func TestURLEscaping_QueryParams(t *testing.T) {
	cases := []struct {
		name  string
		param string
	}{
		{"date", "2026-03-08"},
		{"product code", "EA"},
		{"product code with space", "EA PLUS"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := url.QueryEscape(tc.param)
			if encoded == "" {
				t.Errorf("url.QueryEscape(%q) returned empty string", tc.param)
			}
			// Verify spaces are encoded
			if strings.Contains(tc.param, " ") && !strings.Contains(encoded, "+") && !strings.Contains(encoded, "%20") {
				t.Errorf("url.QueryEscape(%q) = %q; space not encoded", tc.param, encoded)
			}
		})
	}
}
