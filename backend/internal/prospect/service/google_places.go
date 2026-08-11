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
		AuthorAttribution struct {
			DisplayName string `json:"displayName"`
			PhotoURI    string `json:"photoUri"`
		} `json:"authorAttribution"`
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
	Places        []googlePlace `json:"places"`
	NextPageToken string        `json:"nextPageToken"`
}

const (
	searchFieldMask  = "places.id,places.displayName,places.formattedAddress,places.primaryTypeDisplayName,places.types,places.businessStatus,places.rating,places.userRatingCount,places.nationalPhoneNumber,places.websiteUri,places.googleMapsUri,places.location"
	singleTileRadius   = 5000.0
	defaultTileSize    = 3500.0
	maxTiles           = 16
	maxResults         = 150
	maxTypesPerRequest = 45
)

type latLng struct {
	Lat float64
	Lng float64
}

func (c *GooglePlacesClient) Search(ctx context.Context, input prospectmodel.PlaceSearchInput) ([]prospectmodel.PlaceResult, error) {
	if c.key == "" {
		return nil, ErrPlacesDisabled
	}
	keyword := strings.TrimSpace(input.Keyword)
	types := categoryTypes(input.Categories)
	seen := make(map[string]bool)
	results := make([]prospectmodel.PlaceResult, 0, 40)
	collect := func(mapped prospectmodel.PlaceResult) {
		if len(results) >= maxResults {
			return
		}
		if mapped.GooglePlaceID == "" || seen[mapped.GooglePlaceID] {
			return
		}
		seen[mapped.GooglePlaceID] = true
		results = append(results, mapped)
	}

	if keyword != "" {
		if err := c.searchText(ctx, input, keyword, types, collect); err != nil {
			return nil, err
		}
		return results, nil
	}

	for _, chunk := range chunkTypes(types, maxTypesPerRequest) {
		for _, tile := range coverDisk(input.Latitude, input.Longitude, input.Radius) {
			if len(results) >= maxResults {
				break
			}
			if err := c.searchNearbyCircle(ctx, input, tile, chunk, collect); err != nil {
				return nil, err
			}
		}
		if len(results) >= maxResults {
			break
		}
	}
	return results, nil
}

// chunkTypes splits types into chunks of at most size each. Google Places
// Nearby Search rejects includedTypes with more than 50 entries, so large
// category selections must be split across multiple requests.
func chunkTypes(types []string, size int) [][]string {
	if len(types) == 0 {
		return [][]string{nil}
	}
	chunks := make([][]string, 0, (len(types)+size-1)/size)
	for len(types) > size {
		chunks = append(chunks, types[:size])
		types = types[size:]
	}
	return append(chunks, types)
}

func (c *GooglePlacesClient) searchNearbyCircle(ctx context.Context, input prospectmodel.PlaceSearchInput, tile latLng, types []string, collect func(prospectmodel.PlaceResult)) error {
	body := map[string]any{
		"maxResultCount": 20,
		"languageCode":   "id",
		"locationRestriction": map[string]any{"circle": map[string]any{
			"center": map[string]float64{"latitude": tile.Lat, "longitude": tile.Lng},
			"radius": tileRadiusFor(input.Radius),
		}},
	}
	if len(types) > 0 {
		body["includedTypes"] = types
	}
	places, _, err := c.postPlaces(ctx, placesBaseURL+"/places:searchNearby", body)
	if err != nil {
		return err
	}
	for _, place := range places {
		mapped := mapGooglePlace(place, input.Latitude, input.Longitude)
		if mapped.Distance > input.Radius {
			continue
		}
		collect(mapped)
	}
	return nil
}

func (c *GooglePlacesClient) searchText(ctx context.Context, input prospectmodel.PlaceSearchInput, keyword string, types []string, collect func(prospectmodel.PlaceResult)) error {
	pageToken := ""
	for page := 0; page < 3; page++ {
		body := map[string]any{
			"textQuery":    keyword,
			"languageCode": "id",
			"pageSize":     20,
			"locationBias": map[string]any{"circle": map[string]any{"center": map[string]float64{"latitude": input.Latitude, "longitude": input.Longitude}, "radius": input.Radius}},
		}
		if pageToken != "" {
			body["pageToken"] = pageToken
		}
		places, nextToken, err := c.postPlaces(ctx, placesBaseURL+"/places:searchText", body)
		if err != nil {
			return err
		}
		for _, place := range places {
			mapped := mapGooglePlace(place, input.Latitude, input.Longitude)
			if mapped.Distance > input.Radius {
				continue
			}
			if len(types) > 0 && !hasAnyType(place.Types, types) {
				continue
			}
			collect(mapped)
		}
		pageToken = nextToken
		if pageToken == "" {
			break
		}
	}
	return nil
}

func (c *GooglePlacesClient) postPlaces(ctx context.Context, endpoint string, body map[string]any) ([]googlePlace, string, error) {
	encoded, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return nil, "", fmt.Errorf("create Places request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.key)
	req.Header.Set("X-Goog-FieldMask", searchFieldMask)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("Google Places request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("Google Places returned HTTP %d", resp.StatusCode)
	}
	var payload googleResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, "", fmt.Errorf("decode Google Places response: %w", err)
	}
	return payload.Places, payload.NextPageToken, nil
}

func hasAnyType(placeTypes, allowed []string) bool {
	for _, t := range placeTypes {
		for _, a := range allowed {
			if t == a {
				return true
			}
		}
	}
	return false
}

// tileRadiusFor returns the sub-circle radius used for the search area. If the
// requested radius is small enough a single circle covers it.
func tileRadiusFor(radius float64) float64 {
	if radius <= singleTileRadius {
		return radius
	}
	tile := math.Max(defaultTileSize, radius/3.2)
	if tile > 8000 {
		tile = 8000
	}
	for 1.2092*(radius/tile)*(radius/tile) > maxTiles {
		tile *= 1.25
	}
	return tile
}

// coverDisk returns the centers of overlapping sub-circles (hexagonal grid)
// whose union fully covers the disk of the given radius around (lat, lng).
// Nearby Search (New) caps each request at 20 results, so large radii must be
// split into several smaller circles to cover the whole area.
func coverDisk(lat, lng, radius float64) []latLng {
	if radius <= singleTileRadius {
		return []latLng{{Lat: lat, Lng: lng}}
	}
	tile := tileRadiusFor(radius)
	spacing := tile * math.Sqrt(3)
	cosLat := math.Cos(lat * math.Pi / 180)
	metersPerDeg := 111320.0
	cover := radius + tile
	stepLat := spacing / metersPerDeg
	stepLng := spacing / (metersPerDeg * cosLat)
	limit := int(math.Ceil((cover / metersPerDeg) / stepLat))
	centers := make([]latLng, 0, 24)
	for q := -limit; q <= limit; q++ {
		for r := -limit; r <= limit; r++ {
			tLat := lat + stepLat*1.5*float64(r)
			tLng := lng + stepLng*(float64(q)+float64(r)*0.5)
			if haversine(lat, lng, tLat, tLng) <= cover {
				centers = append(centers, latLng{Lat: tLat, Lng: tLng})
			}
		}
	}
	return centers
}

func (c *GooglePlacesClient) Detail(ctx context.Context, placeID string) (prospectmodel.PlaceResult, error) {
	if c.key == "" {
		return prospectmodel.PlaceResult{}, ErrPlacesDisabled
	}
	endpoint := placesBaseURL + "/places/" + url.PathEscape(placeID) + "?languageCode=id"
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
	endpoint := placesBaseURL + "/places/" + url.PathEscape(placeID) + "?languageCode=id"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return prospectmodel.PlaceDetails{}, err
	}
	req.Header.Set("X-Goog-Api-Key", c.key)
	req.Header.Set("X-Goog-FieldMask", "id,displayName,formattedAddress,primaryTypeDisplayName,types,businessStatus,rating,userRatingCount,nationalPhoneNumber,internationalPhoneNumber,websiteUri,googleMapsUri,location,priceLevel,utcOffsetMinutes,editorialSummary,photos,regularOpeningHours,reviews.authorAttribution,reviews.rating,reviews.text,reviews.originalText,reviews.relativePublishTimeDescription,delivery,dineIn,takeout,curbsidePickup,parkingOptions,paymentOptions,accessibilityOptions")
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
		"food_drink": {"restaurant", "cafe", "bakery", "fast_food_restaurant", "bar"},
		"business":   {"corporate_office", "bank", "insurance_agency", "accounting", "real_estate_agency"},
		"culture":    {"museum", "art_gallery", "aquarium", "zoo"},
		"education":  {"school", "university", "primary_school", "secondary_school"},
		"entertainment": {"movie_theater", "amusement_center", "bowling_alley", "night_club"},
		"health":     {"hospital", "pharmacy", "beauty_salon", "dentist", "doctor", "spa", "gym", "physiotherapist"},
		"shopping":   {"shopping_mall", "store", "department_store", "supermarket", "convenience_store", "clothing_store", "electronics_store", "furniture_store", "home_goods_store", "shoe_store", "jewelry_store", "pet_store", "book_store"},
		"lodging":    {"hotel", "motel"},
		"services":   {"bank", "car_repair", "gas_station", "car_wash", "laundry", "plumber", "electrician", "locksmith", "moving_company"},
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
		authorName := r.AuthorAttribution.DisplayName
		authorPhoto := r.AuthorAttribution.PhotoURI
		if authorName == "" {
			authorName = r.AuthorName
		}
		if authorPhoto == "" {
			authorPhoto = r.AuthorPhoto.PhotoURI
		}
		reviews = append(reviews, prospectmodel.PlaceReview{
			AuthorName:   authorName,
			AuthorPhoto:  authorPhoto,
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
