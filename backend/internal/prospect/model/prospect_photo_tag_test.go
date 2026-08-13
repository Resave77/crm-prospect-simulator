package model

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestProspectPhotoTagSerializesLegacyNullPhotoName(t *testing.T) {
	index := 4
	raw, err := json.Marshal(ProspectPhotoTag{ID: uuid.New(), ProspectID: uuid.New(), PhotoIndex: &index, Category: PhotoCategoryMenu})
	if err != nil {
		t.Fatalf("marshal legacy photo tag: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode legacy photo tag: %v", err)
	}
	if value, exists := payload["photoName"]; !exists || value != nil {
		t.Fatalf("expected explicit null photoName, got %#v", payload["photoName"])
	}
	if payload["photoIndex"] != float64(index) {
		t.Fatalf("expected photoIndex %d, got %#v", index, payload["photoIndex"])
	}
}
