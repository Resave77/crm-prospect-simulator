package service

import (
	"context"
	"crm-prospect-simulator/backend/internal/prospect/model"
	"crm-prospect-simulator/backend/internal/usage"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLegacySearchTypeDoesNotNarrowMultiTypeCategories(t *testing.T) {
	if got := legacySearchType([]string{"restaurant", "cafe"}); got != "" {
		t.Fatalf("multi-type search filter=%q, want empty", got)
	}
	if got := legacySearchType([]string{"restaurant", "restaurant"}); got != "restaurant" {
		t.Fatalf("single-type search filter=%q, want restaurant", got)
	}
}

func TestSearchTextPreservesMultiplePlacesInOneRequestWithoutFanout(t *testing.T) {
	previous := placesBaseURL
	defer func() { placesBaseURL = previous }()
	searchCalls := 0
	detailCalls := 0
	photoCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/textsearch/json"):
			searchCalls++
			if r.URL.Query().Get("query") != "cafe" {
				t.Fatalf("query=%q, want cafe", r.URL.Query().Get("query"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[
				{"place_id":"p1","name":"Cafe One","formatted_address":"Address One","types":["cafe"],"geometry":{"location":{"lat":-6.2,"lng":106.8}}},
				{"place_id":"p2","name":"Cafe Two","formatted_address":"Address Two","types":["cafe"],"geometry":{"location":{"lat":-6.201,"lng":106.801}}
			}],"next_page_token":"should-not-be-followed","status":"OK"}`))
		case strings.HasSuffix(r.URL.Path, "/media"):
			photoCalls++
			http.Error(w, "unexpected photo request", http.StatusInternalServerError)
		default:
			detailCalls++
			http.Error(w, "unexpected detail request", http.StatusInternalServerError)
		}
	}))
	defer server.Close()
	placesBaseURL = server.URL
	client := NewGooglePlacesClient("mock-key")
	client.http = server.Client()
	results, err := client.Search(context.Background(), model.PlaceSearchInput{Keyword: "cafe", Latitude: -6.2, Longitude: 106.8, Radius: 1000})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if searchCalls != 1 || detailCalls != 0 || photoCalls != 0 {
		t.Fatalf("requests: search=%d detail=%d photo=%d, want 1/0/0", searchCalls, detailCalls, photoCalls)
	}
	if len(results) != 2 || results[0].PlaceName != "Cafe One" || results[1].PlaceName != "Cafe Two" {
		t.Fatalf("results=%+v, want both mocked places preserved", results)
	}
}

func TestDetailCoreAndBusinessInfoUseSeparateMasks(t *testing.T) {
	previous := placesBaseURL
	defer func() { placesBaseURL = previous }()
	var masks []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("place_id") != "p1" {
			t.Fatalf("place_id=%q, want p1", r.URL.Query().Get("place_id"))
		}
		if r.URL.Query().Get("placeid") != "" {
			t.Fatalf("legacy placeid parameter must not be sent: %q", r.URL.Query().Get("placeid"))
		}
		masks = append(masks, r.URL.Query().Get("fields"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"place_id":"p1","name":"Mock","geometry":{"location":{"lat":-6.2,"lng":106.8}}},"status":"OK"}`))
	}))
	defer server.Close()
	placesBaseURL = server.URL
	client := NewGooglePlacesClient("mock-key")
	client.http = server.Client()
	if _, err := client.DetailCore(context.Background(), "p1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.DetailBusinessInfo(context.Background(), "p1"); err != nil {
		t.Fatal(err)
	}
	if len(masks) != 2 || masks[0] != detailCoreFieldMask || masks[1] != detailBusinessInfoFieldMask {
		t.Fatalf("masks=%v, want core then business info masks", masks)
	}
	if strings.Contains(masks[0], "nationalPhoneNumber") || strings.Contains(masks[0], "rating") || strings.Contains(masks[0], "regularOpeningHours") {
		t.Fatalf("core mask contains expensive business fields: %s", masks[0])
	}
}

func TestSearchCacheMissThenHitSkipsProviderUsageOnHit(t *testing.T) {
	previous := placesBaseURL
	defer func() { placesBaseURL = previous }()
	outbound := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		outbound++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"place_id":"p1","name":"Cached Cafe","geometry":{"location":{"lat":-6.2,"lng":106.8}},"types":["cafe"]}],"status":"OK"}`))
	}))
	defer server.Close()
	placesBaseURL = server.URL
	recorder := &usage.MemoryRecorder{}
	client := NewGooglePlacesClient("mock-key")
	client.http = server.Client()
	client.SetCacheTTLs(25*time.Millisecond, 25*time.Millisecond, 25*time.Millisecond)
	client.SetUsageRecorder(recorder)
	uid := uuid.New()
	input := model.PlaceSearchInput{Keyword: "cached cafe", Latitude: -6.2, Longitude: 106.8, Radius: 1000}
	first := usage.WithTrace(usage.WithUser(context.Background(), uid), "request-1")
	if _, err := client.Search(first, input); err != nil {
		t.Fatal(err)
	}
	second := usage.WithTrace(usage.WithUser(context.Background(), uid), "request-2")
	if _, err := client.Search(second, input); err != nil {
		t.Fatal(err)
	}
	if outbound != 1 {
		t.Fatalf("Google outbound=%d, want 1", outbound)
	}
	if len(recorder.Events()) != 1 {
		t.Fatalf("provider usage events=%d, want 1", len(recorder.Events()))
	}
	if got := usage.GetTrace(first)["cache_status"]; got != "MISS" {
		t.Fatalf("first cache status=%v", got)
	}
	if got := usage.GetTrace(second)["cache_status"]; got != "HIT" {
		t.Fatalf("second cache status=%v", got)
	}
	if got := usage.GetTrace(second)["provider_hit_count"]; got != 0 {
		t.Fatalf("second provider hit count=%v", got)
	}
	time.Sleep(35 * time.Millisecond)
	third := usage.WithTrace(usage.WithUser(context.Background(), uid), "request-3")
	if _, err := client.Search(third, input); err != nil {
		t.Fatal(err)
	}
	if outbound != 2 || len(recorder.Events()) != 2 {
		t.Fatalf("after expiry outbound=%d events=%d, want 2/2", outbound, len(recorder.Events()))
	}
	if got := usage.GetTrace(third)["cache_status"]; got != "MISS" {
		t.Fatalf("expired cache status=%v", got)
	}
}

func TestCoreAndBusinessInfoCachesAreSeparate(t *testing.T) {
	previous := placesBaseURL
	defer func() { placesBaseURL = previous }()
	outbound := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		outbound++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"place_id":"p1","name":"Mock","geometry":{"location":{"lat":-6.2,"lng":106.8}}},"status":"OK"}`))
	}))
	defer server.Close()
	placesBaseURL = server.URL
	client := NewGooglePlacesClient("mock-key")
	client.http = server.Client()
	first := usage.WithTrace(context.Background(), "core-1")
	if _, err := client.DetailCore(first, "p1"); err != nil {
		t.Fatal(err)
	}
	second := usage.WithTrace(context.Background(), "core-2")
	if _, err := client.DetailCore(second, "p1"); err != nil {
		t.Fatal(err)
	}
	third := usage.WithTrace(context.Background(), "business-1")
	if _, err := client.DetailBusinessInfo(third, "p1"); err != nil {
		t.Fatal(err)
	}
	fourth := usage.WithTrace(context.Background(), "business-2")
	if _, err := client.DetailBusinessInfo(fourth, "p1"); err != nil {
		t.Fatal(err)
	}
	if outbound != 2 {
		t.Fatalf("outbound=%d, want separate core/business misses only", outbound)
	}
	if usage.GetTrace(second)["cache_status"] != "HIT" || usage.GetTrace(fourth)["cache_status"] != "HIT" {
		t.Fatal("expected core and business hits")
	}
}

func TestBusinessInfoCacheIsolatedByFieldMask(t *testing.T) {
	previous := placesBaseURL
	defer func() { placesBaseURL = previous }()
	outbound := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		outbound++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"place_id":"p1","name":"Mock","geometry":{"location":{"lat":-6.2,"lng":106.8}}},"status":"OK"}`))
	}))
	defer server.Close()
	placesBaseURL = server.URL
	client := NewGooglePlacesClient("mock-key")
	client.http = server.Client()
	if _, err := client.DetailBusinessInfo(context.Background(), "p1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.DetailFull(context.Background(), "p1"); err != nil {
		t.Fatal(err)
	}
	if outbound != 2 {
		t.Fatalf("different field masks must not collide; outbound=%d", outbound)
	}
}

func TestGoogleOutboundRequestsProduceOneAttributedEventEach(t *testing.T) {
	previous := placesBaseURL
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/photo") {
			_, _ = w.Write([]byte("image"))
			return
		}
		if strings.Contains(r.URL.Path, "search") {
			_, _ = w.Write([]byte(`{"results":[],"status":"OK"}`))
			return
		}
		_, _ = w.Write([]byte(`{"result":{"place_id":"p1","name":"Mock","geometry":{"location":{"lat":-6.2,"lng":106.8}}},"status":"OK"}`))
	}))
	defer server.Close()
	placesBaseURL = server.URL
	defer func() { placesBaseURL = previous }()
	r := &usage.MemoryRecorder{}
	c := NewGooglePlacesClient("mock-key")
	c.http = server.Client()
	c.SetUsageRecorder(r)
	uid := uuid.New()
	ctx := usage.WithFeature(usage.WithUser(context.Background(), uid), "PROSPECT_FINDER")
	if _, err := c.Search(ctx, model.PlaceSearchInput{Latitude: -6.2, Longitude: 106.8, Radius: 1000}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Search(ctx, model.PlaceSearchInput{Keyword: "cafe", Latitude: -6.2, Longitude: 106.8, Radius: 1000}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Detail(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.FetchPhoto(ctx, "legacy:photo-a"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.FetchPhoto(ctx, "legacy:photo-b"); err != nil {
		t.Fatal(err)
	}
	if calls != 5 {
		t.Fatalf("mock Google requests=%d, want 5", calls)
	}
	events := r.Events()
	if len(events) != 5 {
		t.Fatalf("Google usage events=%d, want 5", len(events))
	}
	want := []string{"NEARBY_SEARCH", "TEXT_SEARCH", "PLACE_DETAILS", "PLACE_PHOTO", "PLACE_PHOTO"}
	for i, event := range events {
		if event.Operation != want[i] || event.UserID != uid || event.HTTPStatus != 200 || !event.Success {
			t.Fatalf("event %d=%+v", i, event)
		}
		if i < 2 && event.FieldMask == "" {
			t.Fatal("search field mask missing")
		}
		if i == 2 && event.FieldMask == "" {
			t.Fatal("detail field mask missing")
		}
		if i > 2 && event.FieldMask != "" {
			t.Fatal("photo field mask must be empty")
		}
	}
}

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
		Name:     "legacy:A_B-c",
		WidthPx:  1200,
		HeightPx: 900,
	})

	details := mapPlaceDetails(place)
	if len(details.Photos) != 1 {
		t.Fatalf("photos len=%d, want 1", len(details.Photos))
	}
	photo := details.Photos[0]
	if photo.Name != "legacy:A_B-c" {
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
		"legacy:",
		"legacy:photo/ref",
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
		gotKey = r.URL.Query().Get("key")
		if r.URL.Path != "/photo" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		if r.URL.Query().Get("maxwidth") != "800" {
			t.Fatalf("maxwidth=%q", r.URL.Query().Get("maxwidth"))
		}
		if r.URL.Query().Get("photo_reference") != "photo-ref" {
			t.Fatalf("photo_reference=%q", r.URL.Query().Get("photo_reference"))
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-data"))
	}))
	defer server.Close()

	previousBaseURL := placesBaseURL
	placesBaseURL = server.URL
	defer func() { placesBaseURL = previousBaseURL }()

	client := NewGooglePlacesClient("SECRET_KEY")
	data, contentType, err := client.FetchPhoto(context.Background(), "legacy:photo-ref")
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
		t.Fatalf("query key=%q", gotKey)
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
	_, _, err := client.FetchPhoto(context.Background(), "legacy:photo-ref")
	if !errors.Is(err, ErrPlacePhotoUnavailable) {
		t.Fatalf("err=%v, want ErrPlacePhotoUnavailable", err)
	}
	if strings.Contains(err.Error(), "SECRET_KEY") || strings.Contains(err.Error(), "key=") {
		t.Fatalf("error leaks key: %v", err)
	}
}
