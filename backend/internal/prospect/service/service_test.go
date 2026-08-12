package service

import (
	"context"
	"errors"
	"testing"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"crm-prospect-simulator/backend/internal/prospect/repository"
	"github.com/google/uuid"
)

type fakeProspectRepository struct {
	prospect      prospectmodel.Prospect
	history       []prospectmodel.StatusHistory
	teamDashboard prospectmodel.TeamDashboard
}

type fakePlaces struct{ calls int }

func (f *fakePlaces) Search(_ context.Context, _ prospectmodel.PlaceSearchInput) ([]prospectmodel.PlaceResult, error) {
	f.calls++
	return []prospectmodel.PlaceResult{{GooglePlaceID: "place-1"}}, nil
}
func (f *fakePlaces) Detail(_ context.Context, _ string) (prospectmodel.PlaceResult, error) {
	return prospectmodel.PlaceResult{GooglePlaceID: "place-1"}, nil
}
func (f *fakePlaces) DetailFull(_ context.Context, _ string) (prospectmodel.PlaceDetails, error) {
	return prospectmodel.PlaceDetails{GooglePlaceID: "place-1"}, nil
}
func (f *fakePlaces) SearchMenuImages(_ context.Context, _ string, _ int) ([]prospectmodel.MenuImage, error) {
	return nil, nil
}

func (f *fakePlaces) FetchPhoto(_ context.Context, _ string) ([]byte, string, error) {
	return []byte("photo"), "image/jpeg", nil
}

func (f *fakeProspectRepository) ListAssigned(_ context.Context, owner uuid.UUID) ([]prospectmodel.Prospect, error) {
	if f.prospect.AssignedSalesExecutiveID != owner {
		return []prospectmodel.Prospect{}, nil
	}
	return []prospectmodel.Prospect{f.prospect}, nil
}

func (f *fakeProspectRepository) TeamDashboard(_ context.Context, _ uuid.UUID) (prospectmodel.TeamDashboard, error) {
	return f.teamDashboard, nil
}

func (f *fakeProspectRepository) ListWon(context.Context) ([]prospectmodel.Prospect, error) {
	if f.prospect.Status == prospectmodel.StatusWon {
		return []prospectmodel.Prospect{f.prospect}, nil
	}
	return []prospectmodel.Prospect{}, nil
}

func (f *fakeProspectRepository) ListAll(context.Context) ([]prospectmodel.Prospect, error) {
	return []prospectmodel.Prospect{f.prospect}, nil
}
func (f *fakeProspectRepository) ListSalesExecutives(context.Context) ([]prospectmodel.SalesExecutive, error) {
	return []prospectmodel.SalesExecutive{}, nil
}
func (f *fakeProspectRepository) ListMentionUsers(context.Context) ([]prospectmodel.SalesExecutive, error) {
	return []prospectmodel.SalesExecutive{}, nil
}
func (f *fakeProspectRepository) Create(_ context.Context, _ prospectmodel.SaveProspectInput, _ uuid.UUID) (prospectmodel.Prospect, error) {
	return f.prospect, nil
}
func (f *fakeProspectRepository) CheckIn(_ context.Context, prospectID, owner uuid.UUID, input prospectmodel.CheckInInput) (prospectmodel.Visit, error) {
	if f.prospect.ID != prospectID || f.prospect.AssignedSalesExecutiveID != owner {
		return prospectmodel.Visit{}, repository.ErrNotOwner
	}
	return prospectmodel.Visit{ID: uuid.New(), ProspectID: prospectID, SalesExecutiveID: owner, CheckInLatitude: input.Latitude, CheckInLongitude: input.Longitude}, nil
}
func (f *fakeProspectRepository) CheckOut(_ context.Context, prospectID, visitID, owner uuid.UUID, input prospectmodel.CheckOutInput) (prospectmodel.Visit, error) {
	if f.prospect.ID != prospectID || f.prospect.AssignedSalesExecutiveID != owner {
		return prospectmodel.Visit{}, repository.ErrNotOwner
	}
	return prospectmodel.Visit{ID: visitID, ProspectID: prospectID, SalesExecutiveID: owner}, nil
}

func (f *fakeProspectRepository) FindReview(_ context.Context, id uuid.UUID) (prospectmodel.Review, error) {
	if f.prospect.ID != id {
		return prospectmodel.Review{}, repository.ErrNotFound
	}
	return prospectmodel.Review{Prospect: f.prospect, History: f.history}, nil
}

func (f *fakeProspectRepository) Transition(_ context.Context, id, owner uuid.UUID, expected, status prospectmodel.Status, notes string) (prospectmodel.Prospect, error) {
	if f.prospect.ID != id {
		return prospectmodel.Prospect{}, repository.ErrNotFound
	}
	if f.prospect.AssignedSalesExecutiveID != owner {
		return prospectmodel.Prospect{}, repository.ErrNotOwner
	}
	if f.prospect.Status != expected {
		return prospectmodel.Prospect{}, repository.ErrInvalidStatus
	}
	previous := f.prospect.Status
	f.prospect.Status = status
	f.history = append(f.history, prospectmodel.StatusHistory{FromStatus: &previous, ToStatus: status, Notes: notes})
	return f.prospect, nil
}

func (f *fakeProspectRepository) ListVisitMonitoring(_ context.Context, _ prospectmodel.VisitMonitoringFilter) ([]prospectmodel.VisitMonitoringItem, error) {
	return []prospectmodel.VisitMonitoringItem{}, nil
}

func (f *fakeProspectRepository) ListMyVisits(_ context.Context, _ uuid.UUID, _ prospectmodel.VisitMonitoringFilter) ([]prospectmodel.VisitMonitoringItem, error) {
	return []prospectmodel.VisitMonitoringItem{}, nil
}

func (f *fakeProspectRepository) ListProspectVisits(_ context.Context, _ uuid.UUID) ([]prospectmodel.Visit, error) {
	return []prospectmodel.Visit{}, nil
}

func (f *fakeProspectRepository) DeleteVisit(_ context.Context, _ uuid.UUID, _ uuid.UUID) (prospectmodel.Visit, error) {
	return prospectmodel.Visit{}, nil
}

func (f *fakeProspectRepository) DeleteProspect(_ context.Context, _ uuid.UUID) ([]string, error) {
	return nil, nil
}

func (f *fakeProspectRepository) RequestDeletion(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (f *fakeProspectRepository) ApproveDeletion(_ context.Context, _ uuid.UUID) ([]string, error) {
	return nil, nil
}
func (f *fakeProspectRepository) RejectDeletion(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (f *fakeProspectRepository) ListComments(_ context.Context, _ uuid.UUID) ([]prospectmodel.ProspectComment, error) {
	return []prospectmodel.ProspectComment{}, nil
}
func (f *fakeProspectRepository) CreateComment(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string, _ []prospectmodel.CommentAttachment) (prospectmodel.ProspectComment, error) {
	return prospectmodel.ProspectComment{}, nil
}
func (f *fakeProspectRepository) DeleteComment(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) ([]prospectmodel.CommentAttachment, error) {
	return nil, nil
}
func (f *fakeProspectRepository) FindCommentAttachment(_ context.Context, _ uuid.UUID, _ uuid.UUID) (prospectmodel.CommentAttachment, error) {
	return prospectmodel.CommentAttachment{}, nil
}
func (f *fakeProspectRepository) ListPhotoTags(_ context.Context, _ uuid.UUID) ([]prospectmodel.ProspectPhotoTag, error) {
	return []prospectmodel.ProspectPhotoTag{}, nil
}
func (f *fakeProspectRepository) UpsertPhotoTag(_ context.Context, prospectID uuid.UUID, photoIndex int, category prospectmodel.PhotoCategory, userID uuid.UUID) (prospectmodel.ProspectPhotoTag, error) {
	return prospectmodel.ProspectPhotoTag{ProspectID: prospectID, PhotoIndex: photoIndex, Category: category, UpdatedBy: &userID}, nil
}
func (f *fakeProspectRepository) ProspectAccessibleTo(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	return true, nil
}
func (f *fakeProspectRepository) FindProspectOwner(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (f *fakeProspectRepository) ExistingCustomerPlaceIDs(_ context.Context, _ []string) (map[string]bool, error) {
	return map[string]bool{}, nil
}

func (f *fakeProspectRepository) ListCustomerMarkers(_ context.Context) ([]prospectmodel.CustomerMarker, error) {
	return []prospectmodel.CustomerMarker{}, nil
}

func TestSalesExecutiveCanMarkOwnNegotiationProspectWon(t *testing.T) {
	owner := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: owner, Status: prospectmodel.StatusNegotiation}}
	result, err := New(repo).Transition(context.Background(), Actor{UserID: owner, Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, prospectmodel.StatusWon, "Commercial terms accepted")
	if err != nil || result.Status != prospectmodel.StatusWon {
		t.Fatalf("expected WON decision, result=%+v err=%v", result, err)
	}
}

func TestTeamDashboardRequiresPermission(t *testing.T) {
	repo := &fakeProspectRepository{}
	_, err := New(repo).TeamDashboard(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSalesManager})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden without view_team_dashboard, got %v", err)
	}
}

func TestTeamDashboardAllowsNoSubordinateActorWithPermission(t *testing.T) {
	actorID := uuid.New()
	repo := &fakeProspectRepository{teamDashboard: prospectmodel.TeamDashboard{
		Lead:           prospectmodel.TeamLeadInfo{ID: actorID, FullName: "Leaf Manager"},
		HasTeam:        false,
		PipelineCounts: map[prospectmodel.Status]int{},
	}}
	item, err := New(repo).TeamDashboard(context.Background(), Actor{UserID: actorID, Role: authmodel.RoleSalesManager, PermissionKeys: []string{"view_team_dashboard"}})
	if err != nil {
		t.Fatalf("team dashboard: %v", err)
	}
	if item.HasTeam || item.TotalDescendantCount != 0 {
		t.Fatalf("expected no-subordinate dashboard, got %+v", item)
	}
}

func TestSalesExecutiveCanMarkAnyActiveStageLost(t *testing.T) {
	for _, stage := range prospectmodel.ActiveStatuses {
		owner := uuid.New()
		repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: owner, Status: stage}}
		result, err := New(repo).Transition(context.Background(), Actor{UserID: owner, Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, prospectmodel.StatusLost, "Prospect declined")
		if err != nil || result.Status != prospectmodel.StatusLost {
			t.Fatalf("stage=%s result=%+v err=%v", stage, result, err)
		}
	}
}

func TestPipelineDoesNotAllowSkippingOrEarlyWon(t *testing.T) {
	owner := uuid.New()
	for _, target := range []prospectmodel.Status{prospectmodel.StatusQualified, prospectmodel.StatusWon} {
		repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: owner, Status: prospectmodel.StatusNewLead}}
		_, err := New(repo).Transition(context.Background(), Actor{UserID: owner, Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, target, "note")
		if !errors.Is(err, ErrTransition) {
			t.Fatalf("target=%s expected invalid transition, got %v", target, err)
		}
	}
}

func TestTerminalDecisionsRequireNotes(t *testing.T) {
	owner := uuid.New()
	for _, tc := range []struct {
		from prospectmodel.Status
		to   prospectmodel.Status
	}{{prospectmodel.StatusNegotiation, prospectmodel.StatusWon}, {prospectmodel.StatusInterested, prospectmodel.StatusLost}} {
		repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: owner, Status: tc.from}}
		_, err := New(repo).Transition(context.Background(), Actor{UserID: owner, Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, tc.to, " ")
		if !errors.Is(err, ErrNotesRequired) {
			t.Fatalf("transition %s -> %s expected notes error, got %v", tc.from, tc.to, err)
		}
	}
}

func TestPipelineAdvancesOneStage(t *testing.T) {
	owner := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: owner, Status: prospectmodel.StatusNewLead}}
	result, err := New(repo).Transition(context.Background(), Actor{UserID: owner, Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, prospectmodel.StatusContacted, "")
	if err != nil || result.Status != prospectmodel.StatusContacted {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestPipelineMovesBackOneStage(t *testing.T) {
	owner := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: owner, Status: prospectmodel.StatusQualified}}
	result, err := New(repo).Transition(context.Background(), Actor{UserID: owner, Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, prospectmodel.StatusInterested, "Correction")
	if err != nil || result.Status != prospectmodel.StatusInterested {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestWonProspectIsTerminalForSalesPipeline(t *testing.T) {
	owner := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: owner, Status: prospectmodel.StatusWon}}
	_, err := New(repo).Transition(context.Background(), Actor{UserID: owner, Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, prospectmodel.StatusNegotiation, "Reopen")
	if !errors.Is(err, ErrTransition) {
		t.Fatalf("expected WON to be terminal, got %v", err)
	}
}

func TestVisitRequiresOwnerAndValidCoordinates(t *testing.T) {
	owner := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: owner}}
	_, err := New(repo).CheckIn(context.Background(), Actor{UserID: owner, Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, prospectmodel.CheckInInput{Latitude: 100, Longitude: 106})
	if !errors.Is(err, ErrVisitCoordinates) {
		t.Fatalf("expected coordinates error, got %v", err)
	}
	_, err = New(repo).CheckIn(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, prospectmodel.CheckInInput{Latitude: -6.2, Longitude: 106.8})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestSalesExecutiveCannotDecideAnotherOwnersProspect(t *testing.T) {
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: uuid.New(), Status: prospectmodel.StatusNegotiation}}
	_, err := New(repo).Transition(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}, repo.prospect.ID, prospectmodel.StatusWon, "Won")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestWonProspectCanBeReviewedByAdministrator(t *testing.T) {
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), Status: prospectmodel.StatusWon}}
	review, err := New(repo).Review(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, repo.prospect.ID)
	if err != nil || review.Prospect.Status != prospectmodel.StatusWon {
		t.Fatalf("expected won review, result=%+v err=%v", review, err)
	}
}

func TestCategoryTypesCoverageForCRMPlaceTypes(t *testing.T) {
	categories := []string{
		"resto_cafe", "qsr_fast_food", "bakery_dessert", "hotels_accommodation", "catering_event",
		"modern_trade", "convenience_store", "general_trade", "distributor_agent",
		"industry_manufacturer", "baking_supply", "institutional",
	}
	mapped := categoryTypes(categories)
	for _, want := range []string{
		"food_court", "hamburger_restaurant", "coffee_roastery", "meal_delivery", "dessert_restaurant",
		"guest_house", "catering_service", "shopping_mall", "grocery_store", "food_store",
		"wholesaler", "manufacturer", "hospital", "general_hospital", "business_center",
	} {
		if !hasAnyType(mapped, []string{want}) {
			t.Fatalf("expected CRM Google Place type %q to be included in categoryTypes() result, got %+v", want, mapped)
		}
	}
}

func TestAppCategoryUsesPrimaryTypeDeterministically(t *testing.T) {
	cases := []struct {
		types    []string
		wantKey  string
		wantMenu bool
	}{
		{[]string{"restaurant", "food", "point_of_interest", "establishment"}, "resto_cafe", true},
		{[]string{"coffee_shop", "cafe", "food"}, "resto_cafe", true},
		{[]string{"fast_food_restaurant", "restaurant"}, "qsr_fast_food", true},
		{[]string{"cake_shop", "bakery"}, "bakery_dessert", true},
		{[]string{"catering_service", "event_venue"}, "catering_event", true},
		{[]string{"hotel", "lodging", "restaurant"}, "hotels_accommodation", false},
		{[]string{"shopping_mall", "food_court", "store"}, "modern_trade", false},
		{[]string{"grocery_store", "food_store"}, "convenience_store", false},
		{[]string{"supermarket", "grocery_store"}, "modern_trade", false},
		{[]string{"hospital", "medical_center"}, "institutional", false},
		{[]string{}, "", false},
	}
	for _, tc := range cases {
		key, label := appCategory(tc.types)
		if key != tc.wantKey {
			t.Fatalf("appCategory(%v) key = %q, want %q (label %q)", tc.types, key, tc.wantKey, label)
		}
		if tc.wantKey != "" && label != categoryLabels[tc.wantKey] {
			t.Fatalf("appCategory(%v) label = %q, want %q", tc.types, label, categoryLabels[tc.wantKey])
		}
		if got := placeHasMenu(tc.types); got != tc.wantMenu {
			t.Fatalf("placeHasMenu(%v) = %v, want %v", tc.types, got, tc.wantMenu)
		}
	}
}

func TestPlaceHasMenuExcludesNonFoodPrimaryTypes(t *testing.T) {
	if placeHasMenu([]string{"hotel", "restaurant", "food"}) {
		t.Fatal("a hotel with an in-house restaurant must not be menu-bearing by its primary type")
	}
	if placeHasMenu([]string{"shopping_mall", "food_court"}) {
		t.Fatal("a shopping mall must not be menu-bearing by its primary type")
	}
}

func TestMatchesCategoryFiltersByPrimaryTypeCategory(t *testing.T) {
	resto := selectedCategorySet([]string{"resto_cafe"})
	if !matchesCategory([]string{"restaurant", "food", "point_of_interest"}, resto) {
		t.Fatal("restaurant should match resto_cafe")
	}
	if matchesCategory([]string{"hotel", "restaurant", "food"}, resto) {
		t.Fatal("a hotel with a restaurant secondary type must not match resto_cafe")
	}
	if matchesCategory([]string{"shopping_mall", "food_court"}, resto) {
		t.Fatal("a mall with a food court must not match resto_cafe")
	}
	if !matchesCategory([]string{"shopping_mall"}, selectedCategorySet([]string{"modern_trade"})) {
		t.Fatal("mall should match modern_trade")
	}
	if !matchesCategory([]string{"hospital"}, selectedCategorySet(nil)) {
		t.Fatal("no category selected should match everything")
	}
}

func TestProspectFinderValidatesRadiusBeforeProviderCall(t *testing.T) {
	repo := &fakeProspectRepository{}
	places := &fakePlaces{}
	_, err := New(repo, places).SearchPlaces(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, prospectmodel.PlaceSearchInput{Latitude: -6.2, Longitude: 106.8, Radius: 10})
	if !errors.Is(err, ErrFinderInput) || places.calls != 0 {
		t.Fatalf("expected validation before provider call, calls=%d err=%v", places.calls, err)
	}
}

func TestSalesRoleCannotUseProspectFinder(t *testing.T) {
	_, err := New(&fakeProspectRepository{}, &fakePlaces{}).SearchPlaces(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}, prospectmodel.PlaceSearchInput{Latitude: -6.2, Longitude: 106.8, Radius: 3000})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestChunkTypesSplitsOverMaxTypesPerRequest(t *testing.T) {
	types := make([]string, 0, 60)
	for i := 0; i < 60; i++ {
		types = append(types, "type_"+string(rune('a'+i%26)))
	}
	chunks := chunkTypes(types, maxTypesPerRequest)
	joined := make([]string, 0, len(types))
	for _, chunk := range chunks {
		if len(chunk) > maxTypesPerRequest {
			t.Fatalf("chunk of size %d exceeds limit %d", len(chunk), maxTypesPerRequest)
		}
		joined = append(joined, chunk...)
	}
	if len(joined) != len(types) {
		t.Fatalf("expected %d types total, got %d", len(types), len(joined))
	}
}

func TestChunkTypesEmptyKeepsSingleChunk(t *testing.T) {
	chunks := chunkTypes(nil, maxTypesPerRequest)
	if len(chunks) != 1 || chunks[0] != nil {
		t.Fatalf("expected single nil chunk, got %+v", chunks)
	}
}
