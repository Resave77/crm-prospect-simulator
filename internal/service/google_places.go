package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google-places-playground/config"
)

func TextSearch(body []byte) ([]byte, error) {
	endpoint := config.GoogleBaseURL() + "/places:searchText"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", config.GoogleAPIKey())
	req.Header.Set("X-Goog-FieldMask", "*")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("membaca respons Google Places: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, googlePlacesResponseError(resp, response)
	}
	return response, nil
}

// SearchPlaceID resolves a natural-language query to Google's most relevant Place ID.
func SearchPlaceID(query string) (string, error) {
	body, err := json.Marshal(map[string]string{"textQuery": query})
	if err != nil {
		return "", err
	}

	endpoint := config.GoogleBaseURL() + "/places:searchText"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", config.GoogleAPIKey())
	// Text Search only resolves an ID; CRM values come from Place Details.
	req.Header.Set("X-Goog-FieldMask", "places.id")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("Text Search Google Places: %w", err)
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("membaca respons Text Search: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", googlePlacesResponseError(resp, response)
	}

	var result struct {
		Places []struct {
			ID string `json:"id"`
		} `json:"places"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("parse respons Text Search: %w", err)
	}
	if len(result.Places) == 0 || result.Places[0].ID == "" {
		return "", fmt.Errorf("tempat %q tidak ditemukan", query)
	}
	return result.Places[0].ID, nil
}

func googlePlacesResponseError(resp *http.Response, body []byte) error {
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	message := strings.TrimSpace(string(body))
	if json.Unmarshal(body, &payload) == nil && payload.Error.Message != "" {
		message = payload.Error.Message
	}
	return fmt.Errorf("Google Places API (%s): %s", resp.Status, message)
}
