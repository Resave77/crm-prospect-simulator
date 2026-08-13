package service

import (
	"encoding/json"
	"strings"
	"testing"

	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
)

func TestUsableMenuDataRejectsGeneralAndTaggedGooglePhotos(t *testing.T) {
	for _, details := range []*prospectmodel.PlaceDetails{
		nil,
		{Photos: []prospectmodel.PlacePhoto{{Name: "places/ChIJ123/photos/general", IsMenu: false}}},
		{Photos: []prospectmodel.PlacePhoto{{Name: "places/ChIJ123/photos/tagged", IsMenu: true}}},
	} {
		if usableMenuData(details) {
			t.Fatalf("usableMenuData(%+v)=true, want false", details)
		}
	}
}

func TestMenuDataUnavailablePayloadIsValidClearState(t *testing.T) {
	if !json.Valid(menuDataNotAvailableJSON) {
		t.Fatal("menuDataNotAvailableJSON must be valid JSON")
	}
	if string(menuDataNotAvailableJSON) == "" || !strings.Contains(string(menuDataNotAvailableJSON), menuDataNotAvailable) {
		t.Fatalf("menu unavailable payload=%s, want clear state", menuDataNotAvailableJSON)
	}
}
