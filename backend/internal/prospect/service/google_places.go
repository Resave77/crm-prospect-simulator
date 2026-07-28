package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
)

const placesBaseURL = "https://places.googleapis.com/v1"

type GooglePlacesClient struct {
	key  string
	http *http.Client
}

func NewGooglePlacesClient(key string) *GooglePlacesClient {
	return &GooglePlacesClient{key: strings.TrimSpace(key), http: &http.Client{Timeout: 15e9}}
}

type googlePlace struct {
	ID          string `json:"id"`
	DisplayName struct {
		Text string `json:"text"`
	} `json:"displayName"`
	FormattedAddress       string `json:"formattedAddress"`
	PrimaryTypeDisplayName struct {
		Text string `json:"text"`
	} `json:"primaryTypeDisplayName"`
	Types                   []string `json:"types"`
	BusinessStatus          string   `json:"businessStatus"`
	Rating                  float64  `json:"rating"`
	UserRatingCount         int      `json:"userRatingCount"`
	NationalPhoneNumber     string   `json:"nationalPhoneNumber"`
	InternationalPhoneNumber string `json:"internationalPhoneNumber"`
	WebsiteURI              string   `json:"websiteUri"`
	GoogleMapsURI           string   `json:"googleMapsUri"`
	Location                struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	PriceLevel       string `json:"priceLevel"`
	UTCOffsetMinutes int    `json:"utcOffsetMinutes"`
	EditorialSummary struct {
		Overview string `json:"overview"`
	} `json:"editorialSummary"`
	Photos []struct {
		Name        string `json:"name"`
		WidthPx     int    `json:"widthPx"`
		HeightPx    int    `json:"heightPx"`
		Attributions []struct {
			DisplayName string `json:"displayName"`
			URI         string `json:"uri"`
		} `json:"attributions"`
	} `json:"photos"`
	RegularOpeningHours struct {
		OpenNow              bool     `json:"openNow"`
		WeekdayDescriptions  []string `json:"weekdayDescriptions"`
	} `json:"regularOpeningHours"`
	Reviews []struct {
		AuthorName  string `json:"authorName"`
		AuthorPhoto struct {
			PhotoURI string `json:"photoUri"`
		} `json:"authorPhoto"`
		Rating       float64 `json:"rating"`
		Text         struct {
			Text string `json:"text"`
		} `json:"text"`
		RelativePublishTimeDescription string `json:"relativePublishTimeDescription"`
		LanguageCode                   string `json:"languageCode"`
	} `json:"reviews"`
	Delivery     bool `json:"delivery"`
	DineIn       bool `json:"dineIn"`
	Takeout      bool `json:"takeout"`
	CurbsidePickup bool `json:"curbsidePickup"`
	ParkingOptions struct {
		PaidStreetParking bool `json:"paidStreetParking"`
		PaidParkingLot    bool `json:"paidParkingLot"`
		FreeStreetParking bool `json:"freeStreetParking"`
		FreeParkingLot    bool `json:"freeParkingLot"`
		ValetParking      bool `json:"valetParking"`
		GarageParking     bool `json:"garageParking"`
	} `json:"parkingOptions"`
	PaymentOptions struct {
		CashOnly       bool `json:"cashOnly"`
		CreditCardOnly bool `json:"creditCardOnly"`
		DebitCardOnly  bool `json:"debitCardOnly"`
		NfcOnly        bool `json:"nfcOnly"`
	} `json:"paymentOptions"`
	AccessibilityOptions struct {
		WheelchairAccessibleEntrance bool `json:"wheelchairAccessibleEntrance"`
		WheelchairAccessibleParking  bool `json:"wheelchairAccessibleParking"`
		WheelchairAccessibleRestroom bool `json:"wheelchairAccessibleRestroom"`
		WheelchairAccessibleSeating  bool `json:"wheelchairAccessibleSeating"`
	} `json:"accessibilityOptions"`
}

type googleResponse struct {
	Places []googlePlace `json:"places"`
}

func (c *GooglePlacesClient) Search(ctx context.Context, input prospectmodel.PlaceSearchInput) ([]prospectmodel.PlaceResult, error) {
	if c.key == "" {
		return nil, ErrPlacesDisabled
	}
	body := map[string]any{
		"maxResultCount":      20,
		"locationRestriction": map[string]any{"circle": map[string]any{"center": map[string]float64{"latitude": input.Latitude, "longitude": input.Longitude}, "radius": input.Radius}},
	}
	endpoint := placesBaseURL + "/places:searchNearby"
	if strings.TrimSpace(input.Keyword) != "" {
		endpoint = placesBaseURL + "/places:searchText"
		body["textQuery"] = strings.TrimSpace(input.Keyword)
		delete(body, "locationRestriction")
		body["locationBias"] = map[string]any{"circle": map[string]any{"center": map[string]float64{"latitude": input.Latitude, "longitude": input.Longitude}, "radius": input.Radius}}
	}
	if types := categoryTypes(input.Categories); len(types) > 0 {
		body["includedTypes"] = types
	}
	encoded, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("create Places request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.key)
	req.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.formattedAddress,places.primaryTypeDisplayName,places.types,places.businessStatus,places.rating,places.userRatingCount,places.nationalPhoneNumber,places.websiteUri,places.googleMapsUri,places.location")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Google Places request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Google Places returned HTTP %d", resp.StatusCode)
	}
	var payload googleResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode Google Places response: %w", err)
	}
	items := make([]prospectmodel.PlaceResult, 0, len(payload.Places))
	for _, place := range payload.Places {
		items = append(items, mapGooglePlace(place, input.Latitude, input.Longitude))
	}
	return items, nil
}

func (c *GooglePlacesClient) Detail(ctx context.Context, placeID string) (prospectmodel.PlaceResult, error) {
	if c.key == "" {
		return prospectmodel.PlaceResult{}, ErrPlacesDisabled
	}
	endpoint := placesBaseURL + "/places/" + url.PathEscape(placeID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return prospectmodel.PlaceResult{}, err
	}
	req.Header.Set("X-Goog-Api-Key", c.key)
	req.Header.Set("X-Goog-FieldMask", "id,displayName,formattedAddress,primaryTypeDisplayName,types,businessStatus,rating,userRatingCount,nationalPhoneNumber,websiteUri,googleMapsUri,location")
	resp, err := c.http.Do(req)
	if err != nil {
		return prospectmodel.PlaceResult{}, fmt.Errorf("Google Place detail request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return prospectmodel.PlaceResult{}, fmt.Errorf("Google Place detail returned HTTP %d", resp.StatusCode)
	}
	var place googlePlace
	if err := json.NewDecoder(resp.Body).Decode(&place); err != nil {
		return prospectmodel.PlaceResult{}, err
	}
	return mapGooglePlace(place, place.Location.Latitude, place.Location.Longitude), nil
}

func (c *GooglePlacesClient) DetailFull(ctx context.Context, placeID string) (prospectmodel.PlaceDetails, error) {
	if c.key == "" {
		return prospectmodel.PlaceDetails{}, ErrPlacesDisabled
	}
	endpoint := placesBaseURL + "/places/" + url.PathEscape(placeID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return prospectmodel.PlaceDetails{}, err
	}
	req.Header.Set("X-Goog-Api-Key", c.key)
	req.Header.Set("X-Goog-FieldMask", "id,displayName,formattedAddress,primaryTypeDisplayName,types,businessStatus,rating,userRatingCount,nationalPhoneNumber,internationalPhoneNumber,websiteUri,googleMapsUri,location,priceLevel,utcOffsetMinutes,editorialSummary,photos,regularOpeningHours,reviews,delivery,dineIn,takeout,curbsidePickup,parkingOptions,paymentOptions,accessibilityOptions")
	resp, err := c.http.Do(req)
	if err != nil {
		return prospectmodel.PlaceDetails{}, fmt.Errorf("Google Place detail full request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return prospectmodel.PlaceDetails{}, fmt.Errorf("Google Place detail full returned HTTP %d", resp.StatusCode)
	}
	var place googlePlace
	if err := json.NewDecoder(resp.Body).Decode(&place); err != nil {
		return prospectmodel.PlaceDetails{}, err
	}
	return mapPlaceDetails(place, c.key), nil
}

func mapGooglePlace(place googlePlace, originLat, originLng float64) prospectmodel.PlaceResult {
	lat, lng := place.Location.Latitude, place.Location.Longitude
	category := place.PrimaryTypeDisplayName.Text
	if category == "" && len(place.Types) > 0 {
		category = strings.ReplaceAll(place.Types[0], "_", " ")
	}
	markerCategory, color, icon := markerFor(place.Types)
	return prospectmodel.PlaceResult{GooglePlaceID: place.ID, PlaceName: place.DisplayName.Text, FormattedAddress: place.FormattedAddress,
		Latitude: &lat, Longitude: &lng, PlaceCategory: category, PlaceTypes: place.Types, PhoneNumber: place.NationalPhoneNumber,
		Distance: haversine(originLat, originLng, lat, lng), Rating: place.Rating, UserRatingCount: place.UserRatingCount,
		BusinessStatus: place.BusinessStatus, WebsiteURL: place.WebsiteURI, GoogleMapsURL: place.GoogleMapsURI,
		MarkerCategory: markerCategory, MarkerColor: color, MarkerIcon: icon}
}

func categoryTypes(categories []string) []string {
	mapping := map[string][]string{
		"food_drink": {"restaurant", "cafe", "bakery"}, "business": {"corporate_office"},
		"culture": {"museum", "art_gallery"}, "education": {"school", "university"},
		"entertainment": {"movie_theater", "amusement_center"}, "health": {"hospital", "pharmacy", "beauty_salon"},
		"shopping": {"shopping_mall", "store"}, "lodging": {"hotel"}, "services": {"bank", "car_repair"},
	}
	seen := map[string]bool{}
	result := make([]string, 0)
	for _, category := range categories {
		for _, item := range mapping[category] {
			if !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
	}
	return result
}

func markerFor(types []string) (string, string, string) {
	joined := strings.Join(types, " ")
	switch {
	case strings.Contains(joined, "restaurant") || strings.Contains(joined, "cafe"):
		return "food_drink", "#f97316", "pi pi-shopping-bag"
	case strings.Contains(joined, "hotel"):
		return "lodging", "#8b5cf6", "pi pi-building"
	case strings.Contains(joined, "store") || strings.Contains(joined, "mall"):
		return "shopping", "#2563eb", "pi pi-shopping-cart"
	case strings.Contains(joined, "hospital") || strings.Contains(joined, "pharmacy"):
		return "health", "#ef4444", "pi pi-heart"
	default:
		return "business", "#0ea5e9", "pi pi-briefcase"
	}
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const radius = 6371000.0
	toRad := math.Pi / 180
	dLat, dLon := (lat2-lat1)*toRad, (lon2-lon1)*toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return math.Round(radius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a)))
}

func mapPlaceDetails(place googlePlace, apiKey string) prospectmodel.PlaceDetails {
	lat, lng := place.Location.Latitude, place.Location.Longitude
	category := place.PrimaryTypeDisplayName.Text
	if category == "" && len(place.Types) > 0 {
		category = strings.ReplaceAll(place.Types[0], "_", " ")
	}

	photos := make([]prospectmodel.PlacePhoto, 0, len(place.Photos))
	for _, p := range place.Photos {
		att := ""
		if len(p.Attributions) > 0 {
			att = p.Attributions[0].DisplayName
		}
		photoURL := fmt.Sprintf("https://places.googleapis.com/v1/%s/media?maxWidthPx=800&key=%s", p.Name, apiKey)
		photos = append(photos, prospectmodel.PlacePhoto{
			Name: p.Name, PhotoURL: photoURL, WidthPx: p.WidthPx, HeightPx: p.HeightPx, Attribution: att,
		})
	}

	var openingHours *prospectmodel.PlaceOpeningHours
	if len(place.RegularOpeningHours.WeekdayDescriptions) > 0 {
		openingHours = &prospectmodel.PlaceOpeningHours{
			OpenNow:  place.RegularOpeningHours.OpenNow,
			Weekdays: place.RegularOpeningHours.WeekdayDescriptions,
		}
	}

	reviews := make([]prospectmodel.PlaceReview, 0, len(place.Reviews))
	for _, r := range place.Reviews {
		reviews = append(reviews, prospectmodel.PlaceReview{
			AuthorName:   r.AuthorName,
			AuthorPhoto:  r.AuthorPhoto.PhotoURI,
			Rating:       r.Rating,
			Text:         r.Text.Text,
			Time:         r.RelativePublishTimeDescription,
			LanguageCode: r.LanguageCode,
		})
	}

	return prospectmodel.PlaceDetails{
		GooglePlaceID:      place.ID,
		PlaceName:          place.DisplayName.Text,
		FormattedAddress:   place.FormattedAddress,
		Latitude:           lat,
		Longitude:          lng,
		PlaceCategory:      category,
		PlaceTypes:         place.Types,
		PhoneNumber:        place.NationalPhoneNumber,
		InternationalPhone: place.InternationalPhoneNumber,
		WebsiteURL:         place.WebsiteURI,
		GoogleMapsURL:      place.GoogleMapsURI,
		Rating:             place.Rating,
		UserRatingCount:    place.UserRatingCount,
		BusinessStatus:     place.BusinessStatus,
		PriceLevel:         place.PriceLevel,
		EditorialSummary:   place.EditorialSummary.Overview,
		UTCOffsetMinutes:   place.UTCOffsetMinutes,
		Photos:             photos,
		OpeningHours:       openingHours,
		Reviews:            reviews,
		Delivery:           place.Delivery,
		DineIn:             place.DineIn,
		Takeout:            place.Takeout,
		CurbsidePickup:     place.CurbsidePickup,
		ParkingOptions: &prospectmodel.PlaceParking{
			PaidStreetParking: place.ParkingOptions.PaidStreetParking,
			PaidParkingLot:    place.ParkingOptions.PaidParkingLot,
			FreeStreetParking: place.ParkingOptions.FreeStreetParking,
			FreeParkingLot:    place.ParkingOptions.FreeParkingLot,
			ValetParking:      place.ParkingOptions.ValetParking,
			GarageParking:     place.ParkingOptions.GarageParking,
		},
		PaymentOptions: &prospectmodel.PlacePayments{
			CashOnly:       place.PaymentOptions.CashOnly,
			CreditCardOnly: place.PaymentOptions.CreditCardOnly,
			DebitCardOnly:  place.PaymentOptions.DebitCardOnly,
			NfcOnly:        place.PaymentOptions.NfcOnly,
		},
		AccessibilityOptions: &prospectmodel.PlaceAccessibility{
			WheelchairAccessibleEntrance: place.AccessibilityOptions.WheelchairAccessibleEntrance,
			WheelchairAccessibleParking:  place.AccessibilityOptions.WheelchairAccessibleParking,
			WheelchairAccessibleRestroom: place.AccessibilityOptions.WheelchairAccessibleRestroom,
			WheelchairAccessibleSeating:  place.AccessibilityOptions.WheelchairAccessibleSeating,
		},
	}
}
