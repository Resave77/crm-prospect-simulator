package server

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"crm-prospect-simulator/backend/config"
)

func TestHealthRoutesReturnJSON(t *testing.T) {
	app := New(config.Config{AllowedOrigins: "http://localhost:5173"}, nil, nil, nil)
	for _, path := range []string{"/api/health", "/api/v1/health"} {
		response, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		body, _ := io.ReadAll(response.Body)
		if response.StatusCode != 200 || !strings.Contains(response.Header.Get("Content-Type"), "application/json") {
			t.Fatalf("%s returned status=%d content-type=%q", path, response.StatusCode, response.Header.Get("Content-Type"))
		}
		if strings.Contains(strings.ToLower(string(body)), "<html") {
			t.Fatalf("%s returned HTML", path)
		}
	}
}

func TestUnknownAPIRouteReturnsJSON404(t *testing.T) {
	app := New(config.Config{AllowedOrigins: "http://localhost:5173"}, nil, nil, nil)
	response, err := app.Test(httptest.NewRequest("GET", "/api/unknown", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 404 || !strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		t.Fatalf("status=%d content-type=%q", response.StatusCode, response.Header.Get("Content-Type"))
	}
}
