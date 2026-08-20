package usage

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type activityRecorderStub struct{ activities []Activity }

func (r *activityRecorderStub) RecordActivity(_ context.Context, activity Activity) {
	r.activities = append(r.activities, activity)
}

func TestActivityMiddlewareKeepsUnauthenticatedUserIDNull(t *testing.T) {
	recorder := &activityRecorderStub{}
	app := fiber.New()
	app.Use(ActivityMiddleware(recorder))
	app.Get("/public", func(c *fiber.Ctx) error { SetTrace(c.UserContext(), "action", "LOGIN"); return c.SendStatus(fiber.StatusOK) })

	resp, err := app.Test(httptest.NewRequest("GET", "/public", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if len(recorder.activities) != 1 {
		t.Fatalf("activity rows=%d, want 1", len(recorder.activities))
	}
	if recorder.activities[0].UserID != "" {
		t.Fatalf("unauthenticated user id=%q, want empty for SQL NULL", recorder.activities[0].UserID)
	}
}

func TestActivityMiddlewareUsesAuthenticatedUserID(t *testing.T) {
	recorder := &activityRecorderStub{}
	userID := uuid.New()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.SetUserContext(WithUser(c.UserContext(), userID))
		return c.Next()
	})
	app.Use(ActivityMiddleware(recorder))
	app.Get("/private", func(c *fiber.Ctx) error { SetTrace(c.UserContext(), "action", "VIEW_PROSPECT_DETAIL"); return c.SendStatus(fiber.StatusOK) })

	resp, err := app.Test(httptest.NewRequest("GET", "/private", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if len(recorder.activities) != 1 {
		t.Fatalf("activity rows=%d, want 1", len(recorder.activities))
	}
	if recorder.activities[0].UserID != userID.String() {
		t.Fatalf("authenticated user id=%q, want %q", recorder.activities[0].UserID, userID)
	}
}
