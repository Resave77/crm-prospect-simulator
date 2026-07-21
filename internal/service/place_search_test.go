package service

import (
	"strings"
	"testing"
)

func TestSearchRequestLocationModes(t *testing.T) {
	nearby := nearbySearchRequest(-6.2, 106.8, 3000, nil)
	if _, ok := nearby["locationRestriction"]; !ok {
		t.Fatal("Nearby Search must use locationRestriction")
	}
	if _, ok := nearby["locationBias"]; ok {
		t.Fatal("Nearby Search must not use locationBias")
	}

	textSearch := textSearchRequest("fried chicken", -6.2, 106.8, 3000)
	if _, ok := textSearch["locationBias"]; !ok {
		t.Fatal("Text Search must use locationBias for a circle")
	}
	if _, ok := textSearch["locationRestriction"]; ok {
		t.Fatal("Text Search circle must not use locationRestriction")
	}
}

func TestNearbyCategoryUsesPrimaryTypes(t *testing.T) {
	request := nearbySearchRequest(-6.2, 106.8, 3000, []string{"restaurant", "cafe"})
	if _, ok := request["includedPrimaryTypes"]; !ok {
		t.Fatal("category search must filter includedPrimaryTypes")
	}
	if _, ok := request["includedTypes"]; ok {
		t.Fatal("secondary type filtering can mix unrelated CRM categories")
	}
}

func TestSearchResultFieldMaskContainsProspectCardFields(t *testing.T) {
	for _, field := range []string{"places.businessStatus", "places.priceLevel", "places.rating", "places.userRatingCount", "places.nationalPhoneNumber", "places.websiteUri"} {
		if !strings.Contains(searchResultFieldMask, field) {
			t.Errorf("search result field mask is missing %s", field)
		}
	}
}
