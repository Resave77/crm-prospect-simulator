package handler

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	prospectservice "crm-prospect-simulator/backend/internal/prospect/service"
	"github.com/gofiber/fiber/v2"
)

func TestWriteErrorMapsGoogleProviderFailuresWithoutLeakingBody(t *testing.T) {
	for _, test := range []struct {
		name, body         string
		status, wantStatus int
		wantCode           string
	}{
		{"forbidden", "secret provider body", fiber.StatusForbidden, fiber.StatusServiceUnavailable, "PLACES_PROVIDER_FORBIDDEN"},
		{"unauthorized", "secret provider body", fiber.StatusUnauthorized, fiber.StatusServiceUnavailable, "PLACES_PROVIDER_FORBIDDEN"},
		{"rate limited", "secret provider body", fiber.StatusTooManyRequests, fiber.StatusBadGateway, "PLACES_PROVIDER_ERROR"},
		{"upstream", "secret provider body", fiber.StatusBadGateway, fiber.StatusBadGateway, "PLACES_PROVIDER_ERROR"},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				return writeError(c, errors.Join(&prospectservice.GooglePlacesProviderError{StatusCode: test.status}, errors.New(test.body)))
			})
			res, err := app.Test(httptest.NewRequest("GET", "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != test.wantStatus {
				t.Fatalf("status=%d, want %d", res.StatusCode, test.wantStatus)
			}
			var payload struct {
				Error struct{ Code, Message string } `json:"error"`
			}
			if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload.Error.Code != test.wantCode {
				t.Fatalf("code=%q, want %q", payload.Error.Code, test.wantCode)
			}
			if payload.Error.Message == "" || payload.Error.Message == test.body || containsSecret(payload.Error.Message, test.body) {
				t.Fatalf("provider body leaked: %q", payload.Error.Message)
			}
		})
	}
}

func containsSecret(message, secret string) bool {
	return secret != "" && len(secret) > 0 && message == secret
}
