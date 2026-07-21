package handler

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"crm-prospect-prototype/config"
	"crm-prospect-prototype/internal/service"

	"github.com/gofiber/fiber/v2"
)

type NearbyProspectRow struct {
	Name           string
	Category       string
	Address        string
	Distance       string
	Rating         string
	BusinessStatus string
	Latitude       string
	Longitude      string
	PlaceID        string
	Phone          string
	Website        string
	GoogleMapsURL  string
	MarkerCategory string
	MarkerColor    string
	MarkerIcon     string
	LeadStatus     string
}

type ProspectCategoryOption struct {
	Slug, Label, Icon, Color string
	Selected                 bool
}

var prospectCategoryOptions = []ProspectCategoryOption{
	{Slug: "food_drink", Label: "Food & Drinks", Icon: "🍴", Color: "#f97316"},
	{Slug: "business", Label: "Business", Icon: "💼", Color: "#475467"},
	{Slug: "culture", Label: "Culture", Icon: "🏛", Color: "#8b5cf6"},
	{Slug: "education", Label: "Education", Icon: "🎓", Color: "#2563eb"},
	{Slug: "entertainment", Label: "Entertainment & Recreation", Icon: "🎟", Color: "#ec4899"},
	{Slug: "health", Label: "Health & Wellness", Icon: "✚", Color: "#ef4444"},
	{Slug: "shopping", Label: "Shopping", Icon: "🛍", Color: "#06b6d4"},
	{Slug: "lodging", Label: "Lodging", Icon: "🛏", Color: "#14b8a6"},
	{Slug: "services", Label: "Services", Icon: "⚙", Color: "#eab308"},
}

var prospectCategoryTypes = map[string][]string{
	"food_drink":    {"restaurant", "cafe", "coffee_shop", "bakery", "bar"},
	"business":      {"accounting", "insurance_agency", "real_estate_agency", "corporate_office"},
	"culture":       {"museum", "art_gallery", "library", "cultural_landmark"},
	"education":     {"school", "university", "primary_school", "secondary_school"},
	"entertainment": {"amusement_park", "movie_theater", "bowling_alley", "night_club"},
	"health":        {"hospital", "pharmacy", "doctor", "dentist", "gym", "beauty_salon"},
	"shopping":      {"shopping_mall", "supermarket", "convenience_store", "clothing_store"},
	"lodging":       {"hotel", "hostel", "guest_house", "bed_and_breakfast"},
	"services":      {"bank", "atm", "gas_station", "car_repair", "laundry"},
}

func NearbyProspectFinder(c *fiber.Ctx) error {
	latitudeText := strings.TrimSpace(c.Query("latitude"))
	longitudeText := strings.TrimSpace(c.Query("longitude"))
	radiusText := strings.TrimSpace(c.Query("radius", "3000"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	selectedCategories := parseProspectCategories(c.Query("categories"))
	searchRequested := latitudeText != "" || longitudeText != ""
	if !searchRequested {
		latitudeText = "-6.229561"
		longitudeText = "106.848651"
	}
	data := fiber.Map{
		"Latitude": latitudeText, "Longitude": longitudeText,
		"Radius": radiusText, "Keyword": keyword, "SelectedCategories": strings.Join(selectedCategories, ","),
		"CategoryOptions":  categoryOptionsForView(selectedCategories),
		"GoogleMapsAPIKey": config.GoogleMapsBrowserAPIKey(),
	}

	if !searchRequested {
		return c.Render("nearby_prospects", data)
	}
	latitude, longitude, radius, validationError := validateProspectSearch(latitudeText, longitudeText, radiusText)
	if validationError != "" {
		data["Error"] = validationError
		return c.Status(fiber.StatusBadRequest).Render("nearby_prospects", data)
	}

	startedAt := time.Now()
	var places []service.PlaceSummary
	var err error
	if keyword == "" {
		places, err = searchNearbyProspectCategories(latitude, longitude, radius, selectedCategories)
		if len(selectedCategories) == 0 {
			data["SearchMode"] = "All places within selected radius"
		} else {
			data["SearchMode"] = "Places matching selected categories"
		}
	} else {
		// Places API (New) has no free-text keyword in searchNearby. Text Search
		// with a circular location restriction provides the expected CRM UX.
		places, err = service.TextSearchPlaces(keyword, latitude, longitude, radius)
		data["SearchMode"] = "Text Search dalam radius"
	}
	data["ExecutionTime"] = time.Since(startedAt).Round(time.Millisecond).String()
	if err != nil {
		data["Error"] = "Prospect belum dapat dimuat. Periksa koneksi lalu coba kembali."
		return c.Status(fiber.StatusBadGateway).Render("nearby_prospects", data)
	}
	if keyword != "" {
		places = filterPlacesWithinRadius(places, latitude, longitude, radius)
		// For text searches, category filtering is needed because TextSearch is
		// not type-restricted. For category-based nearby searches the API already
		// uses includedPrimaryTypes, so re-filtering would drop places whose
		// primaryType is a subtype that doesn't round-trip through
		// markerCategoryForType (e.g. "physical_therapist" for health).
		places = filterPlacesByCategories(places, selectedCategories)
	}

	rows := make([]NearbyProspectRow, 0, len(places))
	for _, place := range places {
		markerCategory := markerCategoryForType(place.PrimaryType)
		markerStyle := categoryOption(markerCategory)
		rows = append(rows, NearbyProspectRow{
			Name: place.DisplayName.Text, Category: fallback(place.PrimaryType, "Tidak diketahui"),
			Address:  fallback(place.FormattedAddress, "Tidak tersedia"),
			Distance: formatDistance(latitude, longitude, place.Location.Latitude, place.Location.Longitude),
			Rating:   formatRating(place.Rating), BusinessStatus: formatBusinessStatus(place.BusinessStatus),
			Latitude:  strconv.FormatFloat(place.Location.Latitude, 'f', 6, 64),
			Longitude: strconv.FormatFloat(place.Location.Longitude, 'f', 6, 64),
			PlaceID:   place.ID, Phone: place.NationalPhoneNumber, Website: place.WebsiteURI,
			GoogleMapsURL: place.GoogleMapsURI, MarkerCategory: markerCategory,
			MarkerColor: markerStyle.Color, MarkerIcon: markerStyle.Icon, LeadStatus: "Prospect",
		})
	}
	data["Searched"] = true
	data["Rows"] = rows
	data["ResultCount"] = len(rows)
	data["CenterMapsURL"] = fmt.Sprintf("https://www.google.com/maps?q=%s,%s", url.QueryEscape(latitudeText), url.QueryEscape(longitudeText))
	return c.Render("nearby_prospects", data)
}

func searchNearbyProspectCategories(latitude, longitude, radius float64, categories []string) ([]service.PlaceSummary, error) {
	if len(categories) == 0 {
		return service.NearbyPlaces(latitude, longitude, radius, nil)
	}
	type categoryResult struct {
		index  int
		places []service.PlaceSummary
		err    error
	}
	results := make(chan categoryResult, len(categories))
	for index, category := range categories {
		go func(index int, category string) {
			places, err := service.NearbyPlaces(latitude, longitude, radius, prospectCategoryTypes[category])
			results <- categoryResult{index: index, places: places, err: err}
		}(index, category)
	}
	grouped := make([][]service.PlaceSummary, len(categories))
	for range categories {
		result := <-results
		if result.err != nil {
			return nil, result.err
		}
		grouped[result.index] = result.places
	}
	seen := map[string]bool{}
	merged := make([]service.PlaceSummary, 0)
	for _, group := range grouped {
		for _, place := range group {
			if place.ID == "" || seen[place.ID] {
				continue
			}
			seen[place.ID] = true
			merged = append(merged, place)
		}
	}
	return merged, nil
}

func parseProspectCategories(value string) []string {
	seen := map[string]bool{}
	result := make([]string, 0)
	for _, category := range strings.Split(value, ",") {
		category = strings.TrimSpace(category)
		if _, valid := prospectCategoryTypes[category]; valid && !seen[category] {
			seen[category] = true
			result = append(result, category)
		}
	}
	return result
}

func categoryOptionsForView(selected []string) []ProspectCategoryOption {
	selectedSet := map[string]bool{}
	for _, category := range selected {
		selectedSet[category] = true
	}
	options := make([]ProspectCategoryOption, len(prospectCategoryOptions))
	copy(options, prospectCategoryOptions)
	for index := range options {
		options[index].Selected = selectedSet[options[index].Slug]
	}
	return options
}

func includedTypesForCategories(categories []string) []string {
	types := make([]string, 0)
	for _, category := range categories {
		types = append(types, prospectCategoryTypes[category]...)
	}
	return types
}

func filterPlacesByCategories(places []service.PlaceSummary, categories []string) []service.PlaceSummary {
	if len(categories) == 0 {
		return places
	}
	allowed := map[string]bool{}
	for _, category := range categories {
		allowed[category] = true
	}
	filtered := make([]service.PlaceSummary, 0, len(places))
	for _, place := range places {
		if allowed[markerCategoryForType(place.PrimaryType)] {
			filtered = append(filtered, place)
		}
	}
	return filtered
}

func markerCategoryForType(placeType string) string {
	for category, types := range prospectCategoryTypes {
		for _, candidate := range types {
			if placeType == candidate {
				return category
			}
		}
	}
	switch {
	case strings.Contains(placeType, "restaurant"), strings.Contains(placeType, "cafe"), strings.Contains(placeType, "food"), strings.Contains(placeType, "coffee"), strings.Contains(placeType, "tea_"), strings.Contains(placeType, "dessert"), strings.Contains(placeType, "ice_cream"), strings.Contains(placeType, "juice"), strings.Contains(placeType, "bakery"), strings.Contains(placeType, "bar"):
		return "food_drink"
	case strings.Contains(placeType, "school"), strings.Contains(placeType, "college"):
		return "education"
	case strings.Contains(placeType, "store"), strings.Contains(placeType, "market"):
		return "shopping"
	case strings.Contains(placeType, "hotel"), strings.Contains(placeType, "lodging"):
		return "lodging"
	case strings.Contains(placeType, "clinic"), strings.Contains(placeType, "medical"):
		return "health"
	case strings.Contains(placeType, "museum"), strings.Contains(placeType, "gallery"), strings.Contains(placeType, "landmark"), strings.Contains(placeType, "library"):
		return "culture"
	case strings.Contains(placeType, "theater"), strings.Contains(placeType, "park"), strings.Contains(placeType, "entertainment"), strings.Contains(placeType, "bowling"), strings.Contains(placeType, "club"):
		return "entertainment"
	default:
		return "business"
	}
}

func categoryOption(slug string) ProspectCategoryOption {
	for _, option := range prospectCategoryOptions {
		if option.Slug == slug {
			return option
		}
	}
	return prospectCategoryOptions[1]
}

// ProspectDetail exposes a curated CRM view model, never the raw Places payload.
func ProspectDetail(c *fiber.Ctx) error {
	placeID := strings.TrimSpace(c.Params("id"))
	if placeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Prospect tidak valid."})
	}
	place, err := service.CRMPlace(placeID)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "message": "Detail prospect belum dapat dimuat. Silakan coba lagi."})
	}
	businessType := place.PrimaryTypeDisplayName.Text
	if businessType == "" {
		businessType = place.PrimaryType
	}
	return c.JSON(fiber.Map{
		"success": true,
		"prospect": fiber.Map{
			"id": place.ID, "name": place.DisplayName.Text, "category": place.PrimaryType,
			"businessType": businessType, "businessStatus": formatBusinessStatus(place.BusinessStatus),
			"rating": formatRating(place.Rating), "address": place.FormattedAddress,
			"latitude": place.Location.Latitude, "longitude": place.Location.Longitude,
			"phone": place.NationalPhoneNumber, "website": place.WebsiteURI,
			"googleMapsUrl": place.GoogleMapsURI, "openingHours": place.CurrentOpeningHours.WeekdayDescriptions,
			"openNow": place.CurrentOpeningHours.OpenNow,
		},
	})
}

// AssignExistingCustomer simulates the final CRM workflow without persistence.
func AssignExistingCustomer(c *fiber.Ctx) error {
	var request struct {
		PlaceID, CustomerName, SalesExecutive, StartDate, EndDate string
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Data assignment tidak valid."})
	}
	if strings.TrimSpace(request.PlaceID) == "" || strings.TrimSpace(request.CustomerName) == "" || strings.TrimSpace(request.SalesExecutive) == "" || strings.TrimSpace(request.StartDate) == "" || strings.TrimSpace(request.EndDate) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Lengkapi semua field wajib sebelum assign."})
	}
	startDate, startErr := time.Parse("2006-01-02", request.StartDate)
	endDate, endErr := time.Parse("2006-01-02", request.EndDate)
	if startErr != nil || endErr != nil || endDate.Before(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Rentang tanggal assignment tidak valid."})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":    true,
		"message":    fmt.Sprintf("%s berhasil dikonversi dan di-assign ke %s.", request.CustomerName, request.SalesExecutive),
		"assignment": fiber.Map{"status": "ACTIVE", "simulated": true, "startDate": request.StartDate, "endDate": request.EndDate},
	})
}

func filterPlacesWithinRadius(places []service.PlaceSummary, latitude, longitude, radius float64) []service.PlaceSummary {
	filtered := make([]service.PlaceSummary, 0, len(places))
	for _, place := range places {
		if distanceMeters(latitude, longitude, place.Location.Latitude, place.Location.Longitude) <= radius {
			filtered = append(filtered, place)
		}
	}
	return filtered
}

func SimulateCreateProspect(c *fiber.Ctx) error {
	var request struct {
		PlaceID   string  `json:"placeId"`
		Name      string  `json:"name"`
		Category  string  `json:"category"`
		Address   string  `json:"address"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Payload prospect tidak valid"})
	}
	request.PlaceID = strings.TrimSpace(request.PlaceID)
	request.Name = strings.TrimSpace(request.Name)
	if request.PlaceID == "" || request.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Place ID dan nama wajib tersedia"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Prospect %s berhasil dibuat sebagai simulasi", request.Name),
		"prospect": fiber.Map{
			"source": "GOOGLE_PLACES", "sourceId": request.PlaceID,
			"name": request.Name, "category": request.Category, "address": request.Address,
			"latitude": request.Latitude, "longitude": request.Longitude,
			"status": "DRAFT", "simulated": true, "createdAt": time.Now().Format(time.RFC3339),
		},
	})
}

func validateProspectSearch(latitudeText, longitudeText, radiusText string) (float64, float64, float64, string) {
	latitude, err := strconv.ParseFloat(latitudeText, 64)
	if err != nil || latitude < -90 || latitude > 90 {
		return 0, 0, 0, "Latitude harus berupa angka antara -90 dan 90."
	}
	longitude, err := strconv.ParseFloat(longitudeText, 64)
	if err != nil || longitude < -180 || longitude > 180 {
		return 0, 0, 0, "Longitude harus berupa angka antara -180 dan 180."
	}
	radius, err := strconv.ParseFloat(radiusText, 64)
	if err != nil || radius < 1 || radius > 50000 {
		return 0, 0, 0, "Radius harus antara 1 dan 50.000 meter."
	}
	return latitude, longitude, radius, ""
}

func formatBusinessStatus(status string) string {
	switch status {
	case "OPERATIONAL":
		return "Operational"
	case "CLOSED_TEMPORARILY":
		return "Closed Temporarily"
	case "CLOSED_PERMANENTLY":
		return "Closed Permanently"
	case "":
		return "Tidak tersedia"
	default:
		return status
	}
}

func formatRating(rating float64) string {
	if rating == 0 {
		return "-"
	}
	return strconv.FormatFloat(rating, 'f', 1, 64)
}

func formatDistance(fromLat, fromLng, toLat, toLng float64) string {
	if toLat == 0 && toLng == 0 {
		return ""
	}
	meters := distanceMeters(fromLat, fromLng, toLat, toLng)
	if meters < 1000 {
		return fmt.Sprintf("%.0f m", meters)
	}
	return fmt.Sprintf("%.1f km", meters/1000)
}

func distanceMeters(fromLat, fromLng, toLat, toLng float64) float64 {
	const earthRadius = 6371000.0
	lat1, lat2 := fromLat*math.Pi/180, toLat*math.Pi/180
	dLat, dLng := (toLat-fromLat)*math.Pi/180, (toLng-fromLng)*math.Pi/180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}
