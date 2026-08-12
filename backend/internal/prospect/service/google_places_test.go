package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMapPlaceDetailsPhotoURLUsesCRMProxy(t *testing.T) {
	place := googlePlace{ID: "place-1"}
	place.Photos = append(place.Photos, struct {
		Name         string `json:"name"`
		WidthPx      int    `json:"widthPx"`
		HeightPx     int    `json:"heightPx"`
		Attributions []struct {
			DisplayName string `json:"displayName"`
			URI         string `json:"uri"`
		} `json:"attributions"`
	}{
		Name:     "places/ChIJ123/photos/A_B-c",
		WidthPx:  1200,
		HeightPx: 900,
	})

	details := mapPlaceDetails(place)
	if len(details.Photos) != 1 {
		t.Fatalf("photos len=%d, want 1", len(details.Photos))
	}
	photo := details.Photos[0]
	if photo.Name != "places/ChIJ123/photos/A_B-c" {
		t.Fatalf("photo name=%q", photo.Name)
	}
	if strings.Contains(photo.PhotoURL, "SECRET_KEY") || strings.Contains(photo.PhotoURL, "key=") {
		t.Fatalf("photo URL exposes key: %s", photo.PhotoURL)
	}
	if !strings.HasPrefix(photo.PhotoURL, "/api/v1/places/photo?name=") {
		t.Fatalf("photo URL=%q, want CRM proxy", photo.PhotoURL)
	}
	if strings.Contains(photo.PhotoURL, "places.googleapis.com") {
		t.Fatalf("photo URL points to Google: %s", photo.PhotoURL)
	}
}

func TestPlacePhotoRejectsMalformedResourceNames(t *testing.T) {
	svc := New(&fakeProspectRepository{}, &fakePlaces{})
	for _, name := range []string{
		"",
		"https://places.googleapis.com/v1/places/ChIJ/photos/ref/media?key=SECRET",
		"places/ChIJ",
		"places/ChIJ/photos/ref?target=https://example.com",
		"../places/ChIJ/photos/ref",
	} {
		if _, _, err := svc.PlacePhoto(context.Background(), name); !errors.Is(err, ErrPlacePhotoInvalid) {
			t.Fatalf("PlacePhoto(%q) err=%v, want ErrPlacePhotoInvalid", name, err)
		}
	}
}

func TestFetchPhotoPreservesContentTypeAndUsesHeaderKey(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Goog-Api-Key")
		if strings.Contains(r.URL.RawQuery, "key=") {
			t.Fatalf("query exposes key: %s", r.URL.RawQuery)
		}
		if r.URL.Path != "/places/ChIJ123/photos/photo-ref/media" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		if r.URL.Query().Get("maxWidthPx") != "800" {
			t.Fatalf("maxWidthPx=%q", r.URL.Query().Get("maxWidthPx"))
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-data"))
	}))
	defer server.Close()

	previousBaseURL := placesBaseURL
	placesBaseURL = server.URL
	defer func() { placesBaseURL = previousBaseURL }()

	client := NewGooglePlacesClient("SECRET_KEY")
	data, contentType, err := client.FetchPhoto(context.Background(), "places/ChIJ123/photos/photo-ref")
	if err != nil {
		t.Fatalf("FetchPhoto err=%v", err)
	}
	if string(data) != "png-data" {
		t.Fatalf("data=%q", string(data))
	}
	if contentType != "image/png" {
		t.Fatalf("contentType=%q", contentType)
	}
	if gotKey != "SECRET_KEY" {
		t.Fatalf("X-Goog-Api-Key=%q", gotKey)
	}
}

func TestFetchPhotoHandlesGoogleNon2xxSafely(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream denied", http.StatusForbidden)
	}))
	defer server.Close()

	previousBaseURL := placesBaseURL
	placesBaseURL = server.URL
	defer func() { placesBaseURL = previousBaseURL }()

	client := NewGooglePlacesClient("SECRET_KEY")
	_, _, err := client.FetchPhoto(context.Background(), "places/ChIJ123/photos/photo-ref")
	if !errors.Is(err, ErrPlacePhotoUnavailable) {
		t.Fatalf("err=%v, want ErrPlacePhotoUnavailable", err)
	}
	if strings.Contains(err.Error(), "SECRET_KEY") || strings.Contains(err.Error(), "key=") {
		t.Fatalf("error leaks key: %v", err)
	}
}
