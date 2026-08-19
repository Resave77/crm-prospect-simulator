package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"crm-prospect-simulator/backend/internal/prospect/repository"
	"github.com/google/uuid"
)

type fakeProspectRepository struct {
	prospect      prospectmodel.Prospect
	accessible    *bool
	history       []prospectmodel.StatusHistory
	teamDashboard prospectmodel.TeamDashboard
	photoTags     []prospectmodel.ProspectPhotoTag
	upsertedTag   *prospectmodel.ProspectPhotoTag
}

type fakeAIChatRepository struct {
	*fakeProspectRepository
	chats []prospectmodel.ProspectAIChat
}

func (f *fakeAIChatRepository) CreateAIChat(_ context.Context, item prospectmodel.ProspectAIChat) (prospectmodel.ProspectAIChat, error) {
	item.ID = uuid.New()
	f.chats = append(f.chats, item)
	return item, nil
}

func (f *fakeAIChatRepository) ListAIChats(_ context.Context, prospectID uuid.UUID, limit int) ([]prospectmodel.ProspectAIChat, error) {
	items := make([]prospectmodel.ProspectAIChat, 0, len(f.chats))
	for _, item := range f.chats {
		if item.ProspectID == prospectID {
			items = append(items, item)
		}
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (f *fakeAIChatRepository) ListRecentAIChats(ctx context.Context, prospectID uuid.UUID, limit int) ([]prospectmodel.ProspectAIChat, error) {
	items, err := f.ListAIChats(ctx, prospectID, 0)
	if err != nil || limit <= 0 || len(items) <= limit {
		return items, err
	}
	return items[len(items)-limit:], nil
}

type fakePlaces struct {
	calls         int
	detailFullErr error
	detailFull    prospectmodel.PlaceDetails
}

func (f *fakePlaces) Search(_ context.Context, _ prospectmodel.PlaceSearchInput) ([]prospectmodel.PlaceResult, error) {
	f.calls++
	return []prospectmodel.PlaceResult{{GooglePlaceID: "place-1"}}, nil
}
func (f *fakePlaces) Detail(_ context.Context, _ string) (prospectmodel.PlaceResult, error) {
	return prospectmodel.PlaceResult{GooglePlaceID: "place-1"}, nil
}
func (f *fakePlaces) DetailCore(_ context.Context, _ string) (prospectmodel.PlaceDetails, error) {
	return f.DetailFull(context.Background(), "place-1")
}
func (f *fakePlaces) DetailBusinessInfo(_ context.Context, _ string) (prospectmodel.PlaceDetails, error) {
	return f.DetailFull(context.Background(), "place-1")
}
func (f *fakePlaces) DetailFull(_ context.Context, _ string) (prospectmodel.PlaceDetails, error) {
	if f.detailFull.GooglePlaceID == "" {
		f.detailFull.GooglePlaceID = "place-1"
	}
	return f.detailFull, f.detailFullErr
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
func (f *fakeProspectRepository) ListPhotoTags(_ context.Context, prospectID uuid.UUID) ([]prospectmodel.ProspectPhotoTag, error) {
	if f.prospect.ID != prospectID {
		return []prospectmodel.ProspectPhotoTag{}, nil
	}
	return f.photoTags, nil
}
func (f *fakeProspectRepository) UpsertPhotoTag(_ context.Context, prospectID uuid.UUID, photoName string, photoIndex *int, category prospectmodel.PhotoCategory, userID uuid.UUID) (prospectmodel.ProspectPhotoTag, error) {
	item := prospectmodel.ProspectPhotoTag{ProspectID: prospectID, PhotoName: &photoName, PhotoIndex: photoIndex, Category: category, UpdatedBy: &userID}
	f.upsertedTag = &item
	for i, existing := range f.photoTags {
		if existing.PhotoName != nil && *existing.PhotoName == photoName {
			f.photoTags[i] = item
			return item, nil
		}
	}
	f.photoTags = append(f.photoTags, item)
	return item, nil
}
func (f *fakeProspectRepository) ProspectAccessibleTo(_ context.Context, prospectID uuid.UUID, userID uuid.UUID) (bool, error) {
	if f.accessible != nil {
		return *f.accessible, nil
	}
	return f.prospect.ID == prospectID && f.prospect.AssignedSalesExecutiveID == userID, nil
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

const validTestPhotoName = "places/ChIJTestPlace/photos/AUyValidPhotoRef"

func TestSetPhotoTagStoresPhotoName(t *testing.T) {
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: uuid.New()}}
	actor := Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}
	photoIndex := 2
	item, err := New(repo).SetPhotoTag(context.Background(), actor, repo.prospect.ID, validTestPhotoName, &photoIndex, prospectmodel.PhotoCategoryMenu)
	if err != nil {
		t.Fatalf("set photo tag: %v", err)
	}
	if repo.upsertedTag == nil {
		t.Fatal("expected repository upsert to be called")
	}
	if repo.upsertedTag.PhotoName == nil || *repo.upsertedTag.PhotoName != validTestPhotoName {
		t.Fatalf("expected photoName %q to be stored, got %v", validTestPhotoName, repo.upsertedTag.PhotoName)
	}
	if repo.upsertedTag.Category != prospectmodel.PhotoCategoryMenu {
		t.Fatalf("expected MENU category, got %q", repo.upsertedTag.Category)
	}
	if item.PhotoName == nil || *item.PhotoName != validTestPhotoName {
		t.Fatalf("returned tag photoName mismatch: %v", item.PhotoName)
	}
	if item.PhotoIndex == nil || *item.PhotoIndex != photoIndex {
		t.Fatalf("returned tag photoIndex mismatch: %v", item.PhotoIndex)
	}
}

func TestSetPhotoTagUpsertsSamePhotoName(t *testing.T) {
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: uuid.New()}}
	actor := Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}
	svc := New(repo)
	photoIndex := 2
	if _, err := svc.SetPhotoTag(context.Background(), actor, repo.prospect.ID, validTestPhotoName, &photoIndex, prospectmodel.PhotoCategoryMenu); err != nil {
		t.Fatalf("first tag: %v", err)
	}
	if _, err := svc.SetPhotoTag(context.Background(), actor, repo.prospect.ID, validTestPhotoName, &photoIndex, prospectmodel.PhotoCategoryPlace); err != nil {
		t.Fatalf("second tag: %v", err)
	}
	if len(repo.photoTags) != 1 {
		t.Fatalf("expected a single upserted row for the same photoName, got %d", len(repo.photoTags))
	}
	if repo.photoTags[0].Category != prospectmodel.PhotoCategoryPlace {
		t.Fatalf("expected PLACE after re-tag, got %q", repo.photoTags[0].Category)
	}
}

func TestSetPhotoTagRejectsInvalidCategory(t *testing.T) {
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: uuid.New()}}
	photoIndex := 0
	_, err := New(repo).SetPhotoTag(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, repo.prospect.ID, validTestPhotoName, &photoIndex, prospectmodel.PhotoCategory("FOOD"))
	if !errors.Is(err, ErrPhotoTagInvalid) {
		t.Fatalf("expected ErrPhotoTagInvalid for bad category, got %v", err)
	}
	if repo.upsertedTag != nil {
		t.Fatal("repository must not be called for an invalid category")
	}
}

func TestSetPhotoTagRejectsInvalidPhotoName(t *testing.T) {
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: uuid.New()}}
	photoIndex := 0
	for _, name := range []string{"", "not-a-photo", "places/ChIJTestPlace", "places//photos/ref"} {
		_, err := New(repo).SetPhotoTag(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, repo.prospect.ID, name, &photoIndex, prospectmodel.PhotoCategoryMenu)
		if !errors.Is(err, ErrPhotoTagInvalid) {
			t.Fatalf("photoName %q: expected ErrPhotoTagInvalid, got %v", name, err)
		}
	}
	if repo.upsertedTag != nil {
		t.Fatal("repository must not be called for an invalid photoName")
	}
}

func TestSetPhotoTagRequiresValidPhotoIndex(t *testing.T) {
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: uuid.New()}}
	actor := Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}
	if _, err := New(repo).SetPhotoTag(context.Background(), actor, repo.prospect.ID, validTestPhotoName, nil, prospectmodel.PhotoCategoryMenu); !errors.Is(err, ErrPhotoTagInvalid) {
		t.Fatalf("expected missing photoIndex to be rejected, got %v", err)
	}
	invalidIndex := -1
	if _, err := New(repo).SetPhotoTag(context.Background(), actor, repo.prospect.ID, validTestPhotoName, &invalidIndex, prospectmodel.PhotoCategoryMenu); !errors.Is(err, ErrPhotoTagInvalid) {
		t.Fatalf("expected negative photoIndex to be rejected, got %v", err)
	}
	if repo.upsertedTag != nil {
		t.Fatal("repository must not be called for an invalid photoIndex")
	}
}

func TestListPhotoTagsReturnsStoredPhotoName(t *testing.T) {
	photoName := validTestPhotoName
	repo := &fakeProspectRepository{
		prospect:  prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: uuid.New()},
		photoTags: []prospectmodel.ProspectPhotoTag{{ID: uuid.New(), ProspectID: uuid.New(), PhotoName: &photoName, Category: prospectmodel.PhotoCategoryMenu}},
	}
	items, err := New(repo).ListPhotoTags(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, repo.prospect.ID)
	if err != nil {
		t.Fatalf("list photo tags: %v", err)
	}
	if len(items) != 1 || items[0].PhotoName == nil || *items[0].PhotoName != validTestPhotoName || items[0].Category != prospectmodel.PhotoCategoryMenu {
		t.Fatalf("expected stored photoName/category to round-trip, got %+v", items)
	}
}

func TestListPhotoTagsReturnsLegacyNullPhotoName(t *testing.T) {
	photoIndex := 1
	repo := &fakeProspectRepository{
		prospect:  prospectmodel.Prospect{ID: uuid.New(), AssignedSalesExecutiveID: uuid.New()},
		photoTags: []prospectmodel.ProspectPhotoTag{{ID: uuid.New(), ProspectID: uuid.New(), PhotoIndex: &photoIndex, Category: prospectmodel.PhotoCategoryMenu}},
	}
	items, err := New(repo).ListPhotoTags(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, repo.prospect.ID)
	if err != nil {
		t.Fatalf("list legacy photo tag: %v", err)
	}
	if len(items) != 1 || items[0].PhotoName != nil || items[0].PhotoIndex == nil || *items[0].PhotoIndex != photoIndex {
		t.Fatalf("expected legacy null photoName and retained photoIndex, got %+v", items)
	}
}

func TestSetPhotoTagAccessControl(t *testing.T) {
	prospectID := uuid.New()
	allowed := true
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: prospectID, AssignedSalesExecutiveID: uuid.New()}, accessible: &allowed}
	admin := Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}
	photoIndex := 0
	if _, err := New(repo).SetPhotoTag(context.Background(), admin, prospectID, validTestPhotoName, &photoIndex, prospectmodel.PhotoCategoryMenu); err != nil {
		t.Fatalf("admin must be allowed to tag: %v", err)
	}
	salesWithoutPermission := Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive, PermissionKeys: []string{}}
	if _, err := New(repo).SetPhotoTag(context.Background(), salesWithoutPermission, prospectID, validTestPhotoName, &photoIndex, prospectmodel.PhotoCategoryMenu); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden for sales without permission, got %v", err)
	}
	salesWithPermission := Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive, PermissionKeys: []string{"view_my_prospect_detail"}}
	if _, err := New(repo).SetPhotoTag(context.Background(), salesWithPermission, prospectID, validTestPhotoName, &photoIndex, prospectmodel.PhotoCategoryMenu); err != nil {
		t.Fatalf("expected sales with view permission to tag, got %v", err)
	}
	if _, err := New(repo).ListPhotoTags(context.Background(), salesWithoutPermission, prospectID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden read for sales without permission, got %v", err)
	}
	if _, err := New(repo).ListPhotoTags(context.Background(), salesWithPermission, prospectID); err != nil {
		t.Fatalf("expected sales with view permission to read tags, got %v", err)
	}
}

func TestFindMenuRunsOnlyThroughExplicitServiceCall(t *testing.T) {
	prospectID := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: prospectID, GooglePlaceID: "place-1"}}
	svc := New(repo, &fakePlaces{})
	calls := 0
	svc.SetFindMenu(func(_ context.Context, _ prospectmodel.Review, _ *prospectmodel.PlaceDetails) (json.RawMessage, error) {
		calls++
		return json.RawMessage(`{"status":"MENU_SOURCE_NOT_AVAILABLE","sources":[],"categories":[]}`), nil
	})
	result, err := svc.FindMenu(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator, PermissionKeys: []string{"view_ai_menu_profiling"}}, prospectID)
	if err != nil {
		t.Fatalf("find menu: %v", err)
	}
	if calls != 1 || !json.Valid(result) {
		t.Fatalf("expected one explicit finder call and valid result, calls=%d result=%s", calls, result)
	}
}

func TestFindMenuContinuesWithStoredIdentityWhenPlaceDetailsFails(t *testing.T) {
	prospectID := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: prospectID, GooglePlaceID: "place-1"}}
	places := &fakePlaces{detailFullErr: errors.New("places unavailable")}
	svc := New(repo, places)
	calls := 0
	svc.SetFindMenu(func(_ context.Context, _ prospectmodel.Review, details *prospectmodel.PlaceDetails) (json.RawMessage, error) {
		calls++
		if details != nil {
			t.Fatal("details must be optional on enrichment failure")
		}
		return json.RawMessage(`{"status":"NOT_FOUND","sources":[],"categories":[]}`), nil
	})
	_, err := svc.FindMenu(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator, PermissionKeys: []string{"view_ai_menu_profiling"}}, prospectID)
	if err != nil || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestProfileMenuPrefersStoredStructuredMenu(t *testing.T) {
	prospectID := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: prospectID, GooglePlaceID: "place-1"}}
	svc := New(repo, &fakePlaces{})
	visionCalls := 0
	svc.SetMenuAI(func(context.Context, prospectmodel.Review, *prospectmodel.PlaceDetails, []MenuImageInput) (json.RawMessage, error) {
		visionCalls++
		return nil, nil
	})
	forces := make([]bool, 0, 2)
	svc.SetStructuredMenuAI(func(_ context.Context, _ prospectmodel.Review, _ *prospectmodel.PlaceDetails, force bool) (json.RawMessage, bool, error) {
		forces = append(forces, force)
		return json.RawMessage(`{"menuOpportunity":"HIGH","yoghurtFit":"MEDIUM","topOpportunity":"Pairing","why":"Evidence","recommendedAction":"Validate","confidence":"MEDIUM"}`), true, nil
	})
	actor := Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator, PermissionKeys: []string{"view_ai_menu_profiling"}}
	if _, err := svc.ProfileMenu(context.Background(), actor, prospectID, false); err != nil {
		t.Fatalf("profile structured menu: %v", err)
	}
	if _, err := svc.ProfileMenu(context.Background(), actor, prospectID, true); err != nil {
		t.Fatalf("re-profile structured menu: %v", err)
	}
	if len(forces) != 2 || forces[0] || !forces[1] {
		t.Fatalf("structured profiling force propagation=%v", forces)
	}
	if visionCalls != 0 {
		t.Fatalf("photo Vision fallback must not run when a structured finding exists, calls=%d", visionCalls)
	}
}

func TestProfileMenuUsesLegacyIndexWithoutDereferencingPhotoName(t *testing.T) {
	prospectID := uuid.New()
	photoIndex := 0
	repo := &fakeProspectRepository{
		prospect: prospectmodel.Prospect{ID: prospectID, GooglePlaceID: "place-1"},
		photoTags: []prospectmodel.ProspectPhotoTag{{
			ID: uuid.New(), ProspectID: prospectID, PhotoIndex: &photoIndex, Category: prospectmodel.PhotoCategoryMenu,
		}},
	}
	places := &fakePlaces{detailFull: prospectmodel.PlaceDetails{
		GooglePlaceID: "place-1",
		Photos:        []prospectmodel.PlacePhoto{{Name: "places/example/photos/legacy"}},
	}}
	svc := New(repo, places)
	calls := 0
	svc.SetMenuAI(func(_ context.Context, _ prospectmodel.Review, _ *prospectmodel.PlaceDetails, images []MenuImageInput) (json.RawMessage, error) {
		calls++
		if len(images) != 1 || images[0].Name != "places/example/photos/legacy" {
			t.Fatalf("expected the indexed legacy photo, got %+v", images)
		}
		return json.RawMessage(`{"menus":[]}`), nil
	})
	actor := Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator, PermissionKeys: []string{"view_ai_menu_profiling"}}
	if _, err := svc.ProfileMenu(context.Background(), actor, prospectID, false); err != nil {
		t.Fatalf("profile legacy indexed menu tag: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one menu analysis call, got %d", calls)
	}
}

func TestGenerateSummaryUsesExistingProspectAuthorizationAndOneCallback(t *testing.T) {
	prospectID := uuid.New()
	ownerID := uuid.New()
	repo := &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: prospectID, AssignedSalesExecutiveID: ownerID, GooglePlaceID: "place-1"}}
	svc := New(repo, &fakePlaces{})
	calls := 0
	svc.SetSummaryAI(func(context.Context, prospectmodel.Review, *prospectmodel.PlaceDetails, []prospectmodel.ProspectComment) (json.RawMessage, error) {
		calls++
		return json.RawMessage(`{"summary":"ready","potential":"HIGH"}`), nil
	})
	actor := Actor{UserID: ownerID, Role: authmodel.RoleSalesExecutive, PermissionKeys: []string{"view_ai_summary"}}
	result, err := svc.GenerateSummary(context.Background(), actor, prospectID)
	if err != nil {
		t.Fatalf("generate summary: %v", err)
	}
	if calls != 1 || !json.Valid(result) {
		t.Fatalf("expected one callback and valid summary, calls=%d result=%s", calls, result)
	}
	if _, err := svc.GenerateSummary(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSuperAdmin}, prospectID); err != nil {
		t.Fatalf("SUPER_ADMIN without a sales role must be allowed: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected SUPER_ADMIN request to reach summary service, calls=%d", calls)
	}
	if _, err := svc.GenerateSummary(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, prospectID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("ADMINISTRATOR without explicit permission must be denied, got %v", err)
	}
	other := Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive, PermissionKeys: []string{"view_ai_summary"}}
	if _, err := svc.GenerateSummary(context.Background(), other, prospectID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected cross-prospect access to be forbidden, got %v", err)
	}
}

func TestSharedProspectAccessAllowsAdminAndOwnerOnly(t *testing.T) {
	prospectID := uuid.New()
	ownerID := uuid.New()
	svc := New(&fakeProspectRepository{prospect: prospectmodel.Prospect{ID: prospectID, AssignedSalesExecutiveID: ownerID}})
	if err := svc.AuthorizeProspectAccess(context.Background(), Actor{UserID: ownerID, Role: authmodel.RoleSalesExecutive}, prospectID); err != nil {
		t.Fatalf("owner access: %v", err)
	}
	if err := svc.AuthorizeProspectAccess(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, prospectID); err != nil {
		t.Fatalf("admin access: %v", err)
	}
	if err := svc.AuthorizeProspectAccess(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}, prospectID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected unassigned sales to be forbidden, got %v", err)
	}
}

func TestAIChatHistoryIsSharedPerProspectWithAuthorAndBoundedContext(t *testing.T) {
	prospectID := uuid.New()
	otherProspectID := uuid.New()
	ownerID := uuid.New()
	adminID := uuid.New()
	repo := &fakeAIChatRepository{fakeProspectRepository: &fakeProspectRepository{prospect: prospectmodel.Prospect{ID: prospectID, AssignedSalesExecutiveID: ownerID}}}
	for i := 0; i < 9; i++ {
		repo.chats = append(repo.chats, prospectmodel.ProspectAIChat{ProspectID: prospectID, UserID: ownerID, Message: fmt.Sprintf("q%d", i), Answer: fmt.Sprintf("a%d", i), AuthorName: "Nurdin", AuthorRole: "SALES_EXECUTIVE"})
	}
	repo.chats = append(repo.chats, prospectmodel.ProspectAIChat{ProspectID: otherProspectID, UserID: ownerID, Message: "other", Answer: "other"})
	svc := New(repo)
	contextTurns := 0
	svc.SetChatAI(func(_ context.Context, _ prospectmodel.Review, _ *prospectmodel.PlaceDetails, _ []prospectmodel.ProspectComment, history []ChatTurn, _ string, _ string) (string, error) {
		contextTurns = len(history)
		return `{"Answer":"shared","Skill":"AUTO","Insight":"","Why":"","RecommendedAction":""}`, nil
	})
	sales := Actor{UserID: ownerID, Role: authmodel.RoleSalesExecutive, PermissionKeys: []string{"use_prospect_ai_chat"}}
	admin := Actor{UserID: adminID, Role: authmodel.RoleAdministrator, PermissionKeys: []string{"use_prospect_ai_chat"}}
	if _, err := svc.ChatAI(context.Background(), sales, prospectID, "latest", "AUTO"); err != nil {
		t.Fatalf("sales chat: %v", err)
	}
	if contextTurns != 16 {
		t.Fatalf("shared context turns=%d, want 16 from the latest 8 interactions", contextTurns)
	}
	items, err := svc.AIChatHistory(context.Background(), admin, prospectID)
	if err != nil || len(items) != 10 {
		t.Fatalf("admin shared history len=%d err=%v", len(items), err)
	}
	if items[0].AuthorName != "Nurdin" || items[0].UserID != ownerID {
		t.Fatalf("author metadata was not preserved: %+v", items[0])
	}
	unauthorized := Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive, PermissionKeys: []string{"use_prospect_ai_chat"}}
	denied := false
	repo.accessible = &denied
	if _, err := svc.AIChatHistory(context.Background(), unauthorized, prospectID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected unauthorized history read to be forbidden, got %v", err)
	}
}

func TestConvertedCustomerOwnerCanReadCanonicalProspectAIHistory(t *testing.T) {
	prospectID, originalOwner, customerOwner := uuid.New(), uuid.New(), uuid.New()
	allowed := true
	repo := &fakeAIChatRepository{
		fakeProspectRepository: &fakeProspectRepository{
			prospect:   prospectmodel.Prospect{ID: prospectID, AssignedSalesExecutiveID: originalOwner},
			accessible: &allowed,
		},
		chats: []prospectmodel.ProspectAIChat{{ID: uuid.New(), ProspectID: prospectID, UserID: originalOwner, Message: "Produk apa yang cocok?", Answer: "Yoghurt."}},
	}
	svc := New(repo, nil)
	items, err := svc.AIChatHistory(context.Background(), Actor{UserID: customerOwner, Role: authmodel.RoleSalesExecutive, PermissionKeys: []string{"use_prospect_ai_chat"}}, prospectID)
	if err != nil {
		t.Fatalf("customer owner history: %v", err)
	}
	if len(items) != 1 || items[0].ProspectID != prospectID {
		t.Fatalf("history=%+v", items)
	}

	denied := false
	repo.accessible = &denied
	if _, err := svc.AIChatHistory(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive, PermissionKeys: []string{"use_prospect_ai_chat"}}, prospectID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("unrelated sales err=%v, want forbidden", err)
	}
}
