package domain

import (
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Request Type Constants — Customer NFR additions
// ─────────────────────────────────────────────────────────────────────────────

func TestRequestTypeMobileChange_Constant(t *testing.T) {
	const want = "MOBILE_CHANGE"
	if RequestTypeMobileChange != want {
		t.Errorf("RequestTypeMobileChange = %q, want %q", RequestTypeMobileChange, want)
	}
}

func TestRequestTypeEmailChange_Constant(t *testing.T) {
	const want = "EMAIL_CHANGE"
	if RequestTypeEmailChange != want {
		t.Errorf("RequestTypeEmailChange = %q, want %q", RequestTypeEmailChange, want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DownstreamTaskQueueForType — Customer NFR routing
// ─────────────────────────────────────────────────────────────────────────────

func TestDownstreamTaskQueueForType_MobileChange(t *testing.T) {
	got := DownstreamTaskQueueForType("MOBILE_CHANGE")
	if got != "nfs-tq" {
		t.Errorf("DownstreamTaskQueueForType(MOBILE_CHANGE) = %q, want %q", got, "nfs-tq")
	}
}

func TestDownstreamTaskQueueForType_EmailChange(t *testing.T) {
	got := DownstreamTaskQueueForType("EMAIL_CHANGE")
	if got != "nfs-tq" {
		t.Errorf("DownstreamTaskQueueForType(EMAIL_CHANGE) = %q, want %q", got, "nfs-tq")
	}
}
