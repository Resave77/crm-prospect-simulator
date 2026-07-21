package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestLegacyExplorerRedirectsToCRMWorkspace(t *testing.T) {
	response, err := New().Test(httptest.NewRequest("GET", "/debug/place", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusMovedPermanently || response.Header.Get("Location") != "/" {
		t.Fatalf("legacy route status=%d location=%q", response.StatusCode, response.Header.Get("Location"))
	}
}

func TestNearbySearchRouteIsRegistered(t *testing.T) {
	response, err := New().Test(httptest.NewRequest("GET", "/api/nearby-search?lat=-6.2&lng=106.8&radius=3000&category=unknown", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("nearby search route status=%d, want=%d", response.StatusCode, fiber.StatusBadRequest)
	}
}
