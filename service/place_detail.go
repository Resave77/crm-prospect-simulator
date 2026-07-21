package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"crm-prospect-prototype/config"
)

// CRMPlaceDetail is the presentation-safe subset of Google Places data used by
// the prospect workflow. Keeping this contract explicit prevents raw provider
// responses from leaking into the CRM interface.
type CRMPlaceDetail struct {
	ID                     string        `json:"id"`
	DisplayName            LocalizedText `json:"displayName"`
	FormattedAddress       string        `json:"formattedAddress"`
	PrimaryType            string        `json:"primaryType"`
	PrimaryTypeDisplayName LocalizedText `json:"primaryTypeDisplayName"`
	BusinessStatus         string        `json:"businessStatus"`
	Rating                 float64       `json:"rating"`
	NationalPhoneNumber    string        `json:"nationalPhoneNumber"`
	WebsiteURI             string        `json:"websiteUri"`
	GoogleMapsURI          string        `json:"googleMapsUri"`
	Location               LatLng        `json:"location"`
	CurrentOpeningHours    struct {
		OpenNow             bool     `json:"openNow"`
		WeekdayDescriptions []string `json:"weekdayDescriptions"`
	} `json:"currentOpeningHours"`
}

// CRMPlace returns only the fields rendered by the CRM prospect detail panel.
func CRMPlace(placeID string) (CRMPlaceDetail, error) {
	const fieldMask = "id,displayName,formattedAddress,primaryType,primaryTypeDisplayName,businessStatus,rating,nationalPhoneNumber,websiteUri,googleMapsUri,location,currentOpeningHours"
	var place CRMPlaceDetail
	endpoint := config.GoogleBaseURL() + "/places/" + url.PathEscape(placeID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return place, err
	}
	req.Header.Set("X-Goog-Api-Key", config.GoogleAPIKey())
	req.Header.Set("X-Goog-FieldMask", fieldMask)
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return place, fmt.Errorf("mengambil detail prospect: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return place, fmt.Errorf("membaca detail prospect: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return place, googlePlacesResponseError(resp, body)
	}
	if err := json.Unmarshal(body, &place); err != nil {
		return place, fmt.Errorf("memproses detail prospect: %w", err)
	}
	return place, nil
}

type DebugHeader struct {
	Name  string
	Value string
}

type APIDebugTrace struct {
	Endpoint       string
	Method         string
	RequestBody    string
	RequestHeaders []DebugHeader
	FieldMask      string
	ResponseTime   string
	PayloadSize    int
	StatusCode     int
	Status         string
	QuotaHeaders   []DebugHeader
	ResponseBody   []byte
}

func PlaceDetail(placeID string) ([]byte, error) {
	trace, err := PlaceDetailWithTrace(placeID)
	return trace.ResponseBody, err
}

func PlaceDetailWithTrace(placeID string) (APIDebugTrace, error) {
	const fieldMask = "*"
	trace := APIDebugTrace{
		Method:      http.MethodGet,
		RequestBody: "(empty)",
		FieldMask:   fieldMask,
		RequestHeaders: []DebugHeader{
			{Name: "X-Goog-Api-Key", Value: "[REDACTED]"},
			{Name: "X-Goog-FieldMask", Value: fieldMask},
			{Name: "Accept", Value: "application/json"},
		},
	}

	endpoint := config.GoogleBaseURL() + "/places/" + url.PathEscape(placeID)
	trace.Endpoint = endpoint

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)

	if err != nil {
		return trace, err
	}

	req.Header.Set("X-Goog-Api-Key", config.GoogleAPIKey())
	req.Header.Set("X-Goog-FieldMask", fieldMask)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	startedAt := time.Now()
	resp, err := client.Do(req)

	if err != nil {
		trace.ResponseTime = formatResponseTime(time.Since(startedAt))
		return trace, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	trace.ResponseTime = formatResponseTime(time.Since(startedAt))
	trace.StatusCode = resp.StatusCode
	trace.Status = resp.Status
	trace.ResponseBody = body
	trace.PayloadSize = len(body)
	trace.QuotaHeaders = quotaHeaders(resp.Header)
	if err != nil {
		return trace, fmt.Errorf("membaca respons Google Places: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var googleError struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		message := strings.TrimSpace(string(body))
		if json.Unmarshal(body, &googleError) == nil && googleError.Error.Message != "" {
			message = googleError.Error.Message
		}
		return trace, fmt.Errorf("Google Places API (%s): %s", resp.Status, message)
	}

	return trace, nil
}

func formatResponseTime(duration time.Duration) string {
	return fmt.Sprintf("%.2f ms", float64(duration.Microseconds())/1000)
}

func quotaHeaders(headers http.Header) []DebugHeader {
	result := make([]DebugHeader, 0)
	for name, values := range headers {
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "quota") || strings.Contains(lowerName, "rate-limit") || strings.Contains(lowerName, "ratelimit") || lowerName == "retry-after" {
			result = append(result, DebugHeader{Name: name, Value: strings.Join(values, ", ")})
		}
	}
	return result
}
