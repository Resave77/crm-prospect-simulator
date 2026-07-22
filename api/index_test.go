package handler

import (
	"net/http/httptest"
	"testing"
)

func TestRestoreRewrittenAPIPath(t *testing.T) {
	request := httptest.NewRequest("GET", "https://crm.test/api?__api_path=v1%2Fauth%2Fme&filter=active", nil)
	restoreRewrittenAPIPath(request)

	if request.URL.Path != "/api/v1/auth/me" {
		t.Fatalf("path=%q, want /api/v1/auth/me", request.URL.Path)
	}
	if request.URL.Query().Get("filter") != "active" {
		t.Fatal("public query parameter was not preserved")
	}
	if request.URL.Query().Has("__api_path") {
		t.Fatal("internal rewrite parameter reached Fiber")
	}
}

func TestRestoreRewrittenAPIPathRejectsTraversal(t *testing.T) {
	request := httptest.NewRequest("GET", "https://crm.test/api?__api_path=..%2Fadmin", nil)
	restoreRewrittenAPIPath(request)
	if request.URL.Path != "/api" {
		t.Fatalf("unsafe route changed path to %q", request.URL.Path)
	}
}
