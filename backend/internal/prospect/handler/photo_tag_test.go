package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"crm-prospect-simulator/backend/internal/shared/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TestLegacyPhotoTagJSONResponseDoesNotFail(t *testing.T) {
	index := 5
	app := fiber.New()
	app.Get("/photo-tags", func(c *fiber.Ctx) error {
		return response.Data(c, fiber.StatusOK, []prospectmodel.ProspectPhotoTag{{
			ID:         uuid.New(),
			ProspectID: uuid.New(),
			PhotoIndex: &index,
			Category:   prospectmodel.PhotoCategoryMenu,
		}})
	})

	res, err := app.Test(httptest.NewRequest("GET", "/photo-tags", nil))
	if err != nil {
		t.Fatalf("request legacy photo tags: %v", err)
	}
	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("legacy photo-tag response status=%d, want 200", res.StatusCode)
	}
	defer res.Body.Close()
	var payload struct {
		Data []struct {
			PhotoName  *string `json:"photoName"`
			PhotoIndex *int    `json:"photoIndex"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode legacy photo-tag response: %v", err)
	}
	if len(payload.Data) != 1 || payload.Data[0].PhotoName != nil || payload.Data[0].PhotoIndex == nil || *payload.Data[0].PhotoIndex != index {
		t.Fatalf("unexpected legacy response: %+v", payload.Data)
	}
}
