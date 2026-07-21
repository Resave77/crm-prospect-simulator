package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"google-places-playground/config"
)

type PlaceSummary struct {
	ID                  string        `json:"id"`
	DisplayName         LocalizedText `json:"displayName"`
	FormattedAddress    string        `json:"formattedAddress"`
	PrimaryType         string        `json:"primaryType"`
	BusinessStatus      string        `json:"businessStatus"`
	PriceLevel          string        `json:"priceLevel"`
	Rating              float64       `json:"rating"`
	UserRatingCount     int           `json:"userRatingCount"`
	NationalPhoneNumber string        `json:"nationalPhoneNumber"`
	WebsiteURI          string        `json:"websiteUri"`
	GoogleMapsURI       string        `json:"googleMapsUri"`
	Location            LatLng        `json:"location"`
}

type LocalizedText struct {
	Text string `json:"text"`
}

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

const searchResultFieldMask = "places.id,places.displayName,places.formattedAddress,places.primaryType,places.businessStatus,places.priceLevel,places.rating,places.userRatingCount,places.nationalPhoneNumber,places.websiteUri,places.googleMapsUri,places.location"

func NearbyPlaces(latitude, longitude, radius float64, includedTypes []string) ([]PlaceSummary, error) {
	return searchPlaces("/places:searchNearby", nearbySearchRequest(latitude, longitude, radius, includedTypes))
}

func nearbySearchRequest(latitude, longitude, radius float64, includedTypes []string) map[string]any {
	body := map[string]any{
		"maxResultCount": 20,
		"rankPreference": "DISTANCE",
		"languageCode":   "id",
		"locationRestriction": map[string]any{
			"circle": map[string]any{
				"center": map[string]float64{"latitude": latitude, "longitude": longitude},
				"radius": radius,
			},
		},
	}
	if len(includedTypes) > 0 {
		// Primary-type filtering keeps broad CRM categories mutually exclusive.
		// A secondary type (for example a cafe inside an office) must not make an
		// unrelated business appear in the Food & Drinks result set.
		body["includedPrimaryTypes"] = includedTypes
	}
	return body
}

func TextSearchPlaces(query string, latitude, longitude, radius float64) ([]PlaceSummary, error) {
	return searchPlaces("/places:searchText", textSearchRequest(query, latitude, longitude, radius))
}

func textSearchRequest(query string, latitude, longitude, radius float64) map[string]any {
	return map[string]any{
		"textQuery":    query,
		"pageSize":     20,
		"languageCode": "id",
		"locationBias": map[string]any{
			"circle": map[string]any{
				"center": map[string]float64{"latitude": latitude, "longitude": longitude},
				"radius": radius,
			},
		},
	}
}

func searchPlaces(path string, requestBody map[string]any) ([]PlaceSummary, error) {
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, config.GoogleBaseURL()+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", config.GoogleAPIKey())
	req.Header.Set("X-Goog-FieldMask", searchResultFieldMask)

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("pencarian Google Places: %w", err)
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("membaca hasil pencarian: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, googlePlacesResponseError(resp, response)
	}

	var result struct {
		Places []PlaceSummary `json:"places"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("parse hasil pencarian: %w", err)
	}
	return result.Places, nil
}
