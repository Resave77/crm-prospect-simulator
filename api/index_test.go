package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerRestoresRewrittenCRMPath(t *testing.T) {
	request := httptest.NewRequest("GET", "http://example.test/api?__path=/", nil)
	response := httptest.NewRecorder()
	Handler(response, request)
	if response.Code != 200 {
		t.Fatalf("status=%d, want=200", response.Code)
	}
	if !strings.Contains(response.Body.String(), "Atlas CRM") {
		t.Fatal("Vercel handler did not render the CRM workspace")
	}
}
