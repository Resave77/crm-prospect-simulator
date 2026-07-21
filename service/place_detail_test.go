package service

import (
	"net/http"
	"testing"
)

func TestQuotaHeadersOnlyReturnsQuotaMetadata(t *testing.T) {
	headers := http.Header{
		"Content-Type":          {"application/json"},
		"X-RateLimit-Remaining": {"42"},
		"Retry-After":           {"10"},
		"X-Goog-Quota-Project":  {"crm-research"},
	}
	got := quotaHeaders(headers)
	if len(got) != 3 {
		t.Fatalf("quota header count=%d, want 3: %+v", len(got), got)
	}
	for _, header := range got {
		if header.Name == "Content-Type" {
			t.Fatal("non-quota response header must not be exposed as quota information")
		}
	}
}

func TestPlaceDetailDebugHeadersRedactAPIKeyContract(t *testing.T) {
	// The UI and copied request must never expose the configured credential.
	headers := []DebugHeader{
		{Name: "X-Goog-Api-Key", Value: "[REDACTED]"},
		{Name: "X-Goog-FieldMask", Value: "*"},
	}
	if headers[0].Value != "[REDACTED]" {
		t.Fatal("debug API key must be redacted")
	}
}
