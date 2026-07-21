package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"crm-prospect-prototype/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func TestNearbyProspectFinderRendersSearchForm(t *testing.T) {
	app := fiber.New(fiber.Config{Views: html.New("../views", ".html")})
	app.Get("/prospects/nearby", NearbyProspectFinder)
	response, err := app.Test(httptest.NewRequest("GET", "/prospects/nearby", nil))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	page := string(body)
	if response.StatusCode != fiber.StatusOK || !strings.Contains(page, "Atlas CRM") {
		t.Fatalf("unexpected response: status=%d", response.StatusCode)
	}
	for _, field := range []string{"latitude", "longitude", "radius", "keyword"} {
		if !strings.Contains(string(body), `name="`+field+`"`) {
			t.Errorf("form field %q not rendered", field)
		}
	}
	for _, expected := range []string{"Prospect Intelligence", "Prospect → Existing", "Customer Name", "Sales Executive", "Start Date", "End Date", "Search results", "Search in results", `id="crm-app"`, "createApp", "Business profile", `window.initProspectMap`} {
		if !strings.Contains(page, expected) {
			t.Errorf("CRM workspace does not contain %q", expected)
		}
	}
	for _, stateContract := range []string{"places:initialProspects", "selectedProspect:null", "assignmentMode:false", `v-if="selectedProspect"`, `@click.stop.prevent="openAssignment"`, "this.assignmentMode=true", "renderMarkers(AdvancedMarkerElement)", "this.places.forEach"} {
		if !strings.Contains(page, stateContract) {
			t.Errorf("Vue state contract does not contain %q", stateContract)
		}
	}
	for _, forbidden := range []string{"Raw JSON", "API Inspector", "Developer Debug", "Google Places CRM Explorer", "Assignment Type", "Priority", "Notes", "Customer Code", "Visit Radius", "Assignment Date"} {
		if strings.Contains(page, forbidden) {
			t.Errorf("CRM workspace exposes forbidden debug interface %q", forbidden)
		}
	}
}

func TestAssignExistingCustomer(t *testing.T) {
	app := fiber.New()
	app.Post("/api/assignments", AssignExistingCustomer)
	body := `{"placeId":"ChIJ123","customerName":"Nusantara Coffee","salesExecutive":"Ahmad Fauzi","startDate":"2026-07-21","endDate":"2026-08-21"}`
	request := httptest.NewRequest("POST", "/api/assignments", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("status=%d, want=%d", response.StatusCode, fiber.StatusCreated)
	}
	var payload struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil || !payload.Success {
		t.Fatalf("unexpected assignment response: success=%v err=%v", payload.Success, err)
	}
}

func TestAssignExistingCustomerRejectsInvalidDate(t *testing.T) {
	app := fiber.New()
	app.Post("/api/assignments", AssignExistingCustomer)
	body := `{"placeId":"ChIJ123","customerName":"Nusantara Coffee","salesExecutive":"Ahmad Fauzi","startDate":"2026-08-21","endDate":"2026-07-21"}`
	request := httptest.NewRequest("POST", "/api/assignments", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status=%d, want=%d", response.StatusCode, fiber.StatusBadRequest)
	}
}

func TestValidateProspectSearch(t *testing.T) {
	lat, lng, radius, message := validateProspectSearch("-6.229561", "106.848651", "3000")
	if message != "" || lat != -6.229561 || lng != 106.848651 || radius != 3000 {
		t.Fatalf("valid input rejected: %q", message)
	}
	if _, _, _, message := validateProspectSearch("-91", "106", "3000"); message == "" {
		t.Fatal("invalid latitude was accepted")
	}
}

func TestSimulateCreateProspect(t *testing.T) {
	app := fiber.New()
	app.Post("/prospects/nearby/create", SimulateCreateProspect)
	request := httptest.NewRequest("POST", "/prospects/nearby/create", strings.NewReader(`{"placeId":"ChIJ123","name":"Fried Chicken Outlet","category":"restaurant","latitude":-6.2,"longitude":106.8}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("status=%d, want=%d", response.StatusCode, fiber.StatusCreated)
	}
	var payload struct {
		Success  bool `json:"success"`
		Prospect struct {
			Status    string `json:"status"`
			Simulated bool   `json:"simulated"`
		} `json:"prospect"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Success || !payload.Prospect.Simulated || payload.Prospect.Status != "DRAFT" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestFilterPlacesWithinRadius(t *testing.T) {
	places := []service.PlaceSummary{
		{ID: "near", Location: service.LatLng{Latitude: 0.001, Longitude: 0}},
		{ID: "far", Location: service.LatLng{Latitude: 0.1, Longitude: 0}},
	}
	filtered := filterPlacesWithinRadius(places, 0, 0, 3000)
	if len(filtered) != 1 || filtered[0].ID != "near" {
		t.Fatalf("unexpected filtered places: %+v", filtered)
	}
}

func TestProspectCategorySelection(t *testing.T) {
	selected := parseProspectCategories("food_drink,health,invalid,food_drink")
	if len(selected) != 2 || selected[0] != "food_drink" || selected[1] != "health" {
		t.Fatalf("unexpected categories: %#v", selected)
	}
	types := includedTypesForCategories(selected)
	if len(types) != 11 {
		t.Fatalf("included types=%d, want=11", len(types))
	}
	allTypes := includedTypesForCategories([]string{"food_drink", "lodging", "health", "education", "services", "shopping"})
	for _, expected := range []string{"restaurant", "cafe", "coffee_shop", "hotel", "hospital", "school", "bank", "atm", "gas_station", "shopping_mall", "convenience_store", "pharmacy", "gym", "beauty_salon"} {
		found := false
		for _, placeType := range allTypes {
			if placeType == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("category mapping is missing Google type %q", expected)
		}
	}
}

func TestNearbySearchRejectsUnknownCategory(t *testing.T) {
	app := fiber.New()
	app.Get("/api/nearby-search", NearbySearch)
	response, err := app.Test(httptest.NewRequest("GET", "/api/nearby-search?lat=-6.2&lng=106.8&radius=3000&category=unknown", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status=%d, want=%d", response.StatusCode, fiber.StatusBadRequest)
	}
}

func TestNoCategoryMeansAllPlaces(t *testing.T) {
	places := []service.PlaceSummary{
		{ID: "restaurant", PrimaryType: "restaurant"},
		{ID: "school", PrimaryType: "school"},
	}
	filtered := filterPlacesByCategories(places, nil)
	if len(filtered) != len(places) {
		t.Fatalf("empty category selection must keep all places: %#v", filtered)
	}
}

func TestCategoryFilterAndMarkerGrouping(t *testing.T) {
	places := []service.PlaceSummary{
		{ID: "restaurant", PrimaryType: "fast_food_restaurant"},
		{ID: "school", PrimaryType: "secondary_school"},
	}
	filtered := filterPlacesByCategories(places, []string{"food_drink"})
	if len(filtered) != 1 || filtered[0].ID != "restaurant" {
		t.Fatalf("unexpected category filter: %#v", filtered)
	}
	if markerCategoryForType("clothing_store") != "shopping" {
		t.Fatal("clothing store should use shopping marker")
	}
}
