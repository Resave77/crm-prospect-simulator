package repository

import (
	"testing"
	"time"

	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"github.com/google/uuid"
)

type photoTagRow struct {
	photoName  *string
	photoIndex *int
}

func (r photoTagRow) Scan(dest ...any) error {
	id := uuid.New()
	prospectID := uuid.New()
	category := prospectmodel.PhotoCategoryMenu
	now := time.Now().UTC()
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*uuid.UUID) = prospectID
	*dest[2].(**string) = r.photoName
	*dest[3].(**int) = r.photoIndex
	*dest[4].(*prospectmodel.PhotoCategory) = category
	*dest[5].(**uuid.UUID) = nil
	*dest[6].(*time.Time) = now
	*dest[7].(*time.Time) = now
	return nil
}

func TestScanPhotoTagAcceptsNullPhotoName(t *testing.T) {
	index := 3
	item, err := scanPhotoTag(photoTagRow{photoIndex: &index})
	if err != nil {
		t.Fatalf("scan legacy photo tag: %v", err)
	}
	if item.PhotoName != nil {
		t.Fatalf("expected null photoName, got %v", item.PhotoName)
	}
	if item.PhotoIndex == nil || *item.PhotoIndex != index {
		t.Fatalf("expected photoIndex %d, got %v", index, item.PhotoIndex)
	}
}

func TestScanPhotoTagPreservesNormalPhotoName(t *testing.T) {
	name := "places/example/photos/reference"
	index := 1
	item, err := scanPhotoTag(photoTagRow{photoName: &name, photoIndex: &index})
	if err != nil {
		t.Fatalf("scan normal photo tag: %v", err)
	}
	if item.PhotoName == nil || *item.PhotoName != name {
		t.Fatalf("expected photoName %q, got %v", name, item.PhotoName)
	}
}
