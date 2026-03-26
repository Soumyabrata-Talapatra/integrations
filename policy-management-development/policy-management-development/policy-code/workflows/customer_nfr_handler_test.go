package workflows

import (
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// customerNFRMetadataAllowList — allow-list entries
// ─────────────────────────────────────────────────────────────────────────────

func TestCustomerNFRMetadataAllowList_AddressChange(t *testing.T) {
	entry, ok := customerNFRMetadataAllowList["ADDRESS_CHANGE"]
	if !ok {
		t.Fatal("customerNFRMetadataAllowList missing ADDRESS_CHANGE entry")
	}
	wantKeys := []string{"address_type", "pin_code", "state", "district", "city"}
	for _, k := range wantKeys {
		if !entry[k] {
			t.Errorf("ADDRESS_CHANGE allow-list missing key %q", k)
		}
	}
}

func TestCustomerNFRMetadataAllowList_NameChange(t *testing.T) {
	entry, ok := customerNFRMetadataAllowList["NAME_CHANGE"]
	if !ok {
		t.Fatal("customerNFRMetadataAllowList missing NAME_CHANGE entry")
	}
	wantKeys := []string{"salutation", "first_name", "middle_name", "last_name"}
	for _, k := range wantKeys {
		if !entry[k] {
			t.Errorf("NAME_CHANGE allow-list missing key %q", k)
		}
	}
}

func TestCustomerNFRMetadataAllowList_MobileChange(t *testing.T) {
	entry, ok := customerNFRMetadataAllowList["MOBILE_CHANGE"]
	if !ok {
		t.Fatal("customerNFRMetadataAllowList missing MOBILE_CHANGE entry")
	}
	if !entry["mobile_number"] {
		t.Error("MOBILE_CHANGE allow-list missing key \"mobile_number\"")
	}
}

func TestCustomerNFRMetadataAllowList_EmailChange(t *testing.T) {
	entry, ok := customerNFRMetadataAllowList["EMAIL_CHANGE"]
	if !ok {
		t.Fatal("customerNFRMetadataAllowList missing EMAIL_CHANGE entry")
	}
	if !entry["email"] {
		t.Error("EMAIL_CHANGE allow-list missing key \"email\"")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// filterCustomerNFRPayload
// ─────────────────────────────────────────────────────────────────────────────

func TestFilterCustomerNFRPayload_AddressChange_ValidKeys(t *testing.T) {
	payload := map[string]interface{}{
		"address_type": "PERMANENT",
		"pin_code":     "110001",
		"state":        "Delhi",
		"district":     "Central Delhi",
		"city":         "New Delhi",
	}
	got := filterCustomerNFRPayload("ADDRESS_CHANGE", payload)
	if got == nil {
		t.Fatal("filterCustomerNFRPayload returned nil for valid ADDRESS_CHANGE payload")
	}
	for k, v := range payload {
		gv, ok := got[k]
		if !ok {
			t.Errorf("missing key %q in filtered result", k)
			continue
		}
		if gv != v {
			t.Errorf("key %q: got %v, want %v", k, gv, v)
		}
	}
}

func TestFilterCustomerNFRPayload_AddressChange_StripsDisallowed(t *testing.T) {
	payload := map[string]interface{}{
		"pin_code":      "110001",
		"state":         "Delhi",
		"secret_field":  "should-be-stripped",
		"internal_note": "also-stripped",
	}
	got := filterCustomerNFRPayload("ADDRESS_CHANGE", payload)
	if got == nil {
		t.Fatal("filterCustomerNFRPayload returned nil; expected filtered result")
	}
	if _, ok := got["secret_field"]; ok {
		t.Error("disallowed key \"secret_field\" was not stripped")
	}
	if _, ok := got["internal_note"]; ok {
		t.Error("disallowed key \"internal_note\" was not stripped")
	}
	// Allowed keys should be present.
	if _, ok := got["pin_code"]; !ok {
		t.Error("allowed key \"pin_code\" missing from result")
	}
	if _, ok := got["state"]; !ok {
		t.Error("allowed key \"state\" missing from result")
	}
}

func TestFilterCustomerNFRPayload_UnknownType_ReturnsNil(t *testing.T) {
	payload := map[string]interface{}{"foo": "bar"}
	got := filterCustomerNFRPayload("UNKNOWN_TYPE", payload)
	if got != nil {
		t.Errorf("expected nil for unknown request type, got %v", got)
	}
}

func TestFilterCustomerNFRPayload_EmptyPayload_ReturnsNil(t *testing.T) {
	got := filterCustomerNFRPayload("ADDRESS_CHANGE", map[string]interface{}{})
	if got != nil {
		t.Errorf("expected nil for empty payload, got %v", got)
	}
}

// TestFilterCustomerNFRPayload_AllTypes_MixedKeys is a table-driven test that
// covers all four NFR types with a mix of valid and invalid keys.
func TestFilterCustomerNFRPayload_AllTypes_MixedKeys(t *testing.T) {
	cases := []struct {
		name        string
		requestType string
		payload     map[string]interface{}
		wantKeys    []string
		stripKeys   []string
	}{
		{
			name:        "ADDRESS_CHANGE mixed",
			requestType: "ADDRESS_CHANGE",
			payload: map[string]interface{}{
				"pin_code": "400001",
				"city":     "Mumbai",
				"ssn":      "should-strip",
			},
			wantKeys:  []string{"pin_code", "city"},
			stripKeys: []string{"ssn"},
		},
		{
			name:        "NAME_CHANGE mixed",
			requestType: "NAME_CHANGE",
			payload: map[string]interface{}{
				"first_name": "Priya",
				"last_name":  "Patel",
				"dob":        "should-strip",
			},
			wantKeys:  []string{"first_name", "last_name"},
			stripKeys: []string{"dob"},
		},
		{
			name:        "MOBILE_CHANGE mixed",
			requestType: "MOBILE_CHANGE",
			payload: map[string]interface{}{
				"mobile_number": "9999888877",
				"carrier":       "should-strip",
			},
			wantKeys:  []string{"mobile_number"},
			stripKeys: []string{"carrier"},
		},
		{
			name:        "EMAIL_CHANGE mixed",
			requestType: "EMAIL_CHANGE",
			payload: map[string]interface{}{
				"email":    "new@example.com",
				"verified": "should-strip",
			},
			wantKeys:  []string{"email"},
			stripKeys: []string{"verified"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCustomerNFRPayload(tc.requestType, tc.payload)
			if got == nil {
				t.Fatal("filterCustomerNFRPayload returned nil; expected filtered result")
			}

			for _, k := range tc.wantKeys {
				if _, ok := got[k]; !ok {
					t.Errorf("allowed key %q missing from result", k)
				}
			}
			for _, k := range tc.stripKeys {
				if _, ok := got[k]; ok {
					t.Errorf("disallowed key %q was not stripped", k)
				}
			}

			// Verify no extra keys snuck in.
			for k := range got {
				found := false
				for _, wk := range tc.wantKeys {
					if k == wk {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("unexpected key %q in filtered result", k)
				}
			}
		})
	}
}
