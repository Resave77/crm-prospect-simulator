package usage

import "testing"

func TestClassifyGoogleSKUUsesOperationAndFieldMask(t *testing.T) {
	cases := []struct{ operation, mask, want string }{
		{"NEARBY_SEARCH", "places.displayName,places.location", "99F9-A108-83A6"},
		{"TEXT_SEARCH", "places.displayName,places.location", "4FDA-34B1-A910"},
		{"CORE_DETAIL", "displayName,businessStatus,googleMapsUri", "4ED6-464A-2AFC"},
		{"BUSINESS_INFO", "nationalPhoneNumber,regularOpeningHours", "2D9A-3DE0-3766"},
		{"PLACE_PHOTO", "", "DCD1-FE97-8C71"},
	}
	for _, tc := range cases {
		sku, ok := ClassifyGoogleSKU(tc.operation, tc.mask)
		if !ok || sku.ID != tc.want || !sku.Verified {
			t.Fatalf("%s classified as %+v, ok=%v", tc.operation, sku, ok)
		}
	}
}

func TestResolveGoogleSKULegacyPlaceDetailsOnlyWithDeterministicMask(t *testing.T) {
	if operation, _, ok := ResolveGoogleSKU("PLACE_DETAILS", "", "displayName,googleMapsUri"); !ok || operation != "CORE_DETAIL" {
		t.Fatalf("legacy core resolution=%q, ok=%v", operation, ok)
	}
	if operation, _, ok := ResolveGoogleSKU("PLACE_DETAILS", "", "displayName"); ok || operation != "" {
		t.Fatalf("ambiguous legacy resolution=%q, ok=%v", operation, ok)
	}
}
