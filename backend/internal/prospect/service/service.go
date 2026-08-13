package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"crm-prospect-simulator/backend/internal/prospect/repository"
	"github.com/google/uuid"
)

var (
	ErrForbidden             = errors.New("prospect operation forbidden")
	ErrTransition            = errors.New("invalid prospect stage transition")
	ErrNotesRequired         = errors.New("a win note or loss reason is required")
	ErrProspectStatus        = errors.New("prospect is not eligible for this operation")
	ErrFinderInput           = errors.New("prospect finder query is invalid")
	ErrPlacesDisabled        = errors.New("Google Places server key is not configured")
	ErrMenuImagesDisabled    = errors.New("Google Custom Search is not configured")
	ErrPlacePhotoInvalid     = errors.New("Google Places photo resource name is invalid")
	ErrPlacePhotoUnavailable = errors.New("Google Places photo is unavailable")
	ErrAlreadyCustomer       = errors.New("place is already an existing customer")
	ErrVisitCoordinates      = errors.New("visit GPS coordinates are invalid")
	ErrPhotoTagInvalid       = errors.New("photo tag is invalid: category must be MENU or PLACE and photoName must be a valid Google photo resource name")
)

type Actor struct {
	UserID         uuid.UUID
	Role           authmodel.Role
	PermissionKeys []string
}
type ChatTurn struct {
	Role    string
	Content string
}

type Service struct {
	repository      repository.Repository
	places          Places
	initialAnalysis func(context.Context, prospectmodel.Review, *prospectmodel.PlaceDetails, []prospectmodel.ProspectComment)
	chatAI          func(context.Context, prospectmodel.Review, *prospectmodel.PlaceDetails, []prospectmodel.ProspectComment, []ChatTurn, string, string) (string, error)
}

type Places interface {
	Search(context.Context, prospectmodel.PlaceSearchInput) ([]prospectmodel.PlaceResult, error)
	Detail(context.Context, string) (prospectmodel.PlaceResult, error)
	DetailFull(context.Context, string) (prospectmodel.PlaceDetails, error)
	SearchMenuImages(context.Context, string, int) ([]prospectmodel.MenuImage, error)
	FetchPhoto(context.Context, string) ([]byte, string, error)
}

func New(repo repository.Repository, placeClients ...Places) *Service {
	var places Places
	if len(placeClients) > 0 {
		places = placeClients[0]
	}
	return &Service{repository: repo, places: places}
}

func (s *Service) SetInitialAnalysis(fn func(context.Context, prospectmodel.Review, *prospectmodel.PlaceDetails, []prospectmodel.ProspectComment)) {
	s.initialAnalysis = fn
}

func (s *Service) SetChatAI(fn func(context.Context, prospectmodel.Review, *prospectmodel.PlaceDetails, []prospectmodel.ProspectComment, []ChatTurn, string, string) (string, error)) {
	s.chatAI = fn
}

func (s *Service) ChatAI(ctx context.Context, actor Actor, id uuid.UUID, message, skill string) (string, error) {
	if !actor.can("use_prospect_ai_chat") || s.chatAI == nil {
		return "", ErrForbidden
	}
	skill = strings.ToUpper(strings.TrimSpace(skill))
	if skill == "" {
		skill = "AUTO"
	}
	supported := map[string]bool{"AUTO": true, "PROSPECT_ANALYSIS": true, "QUALIFICATION_ANALYSIS": true, "MENU_OPPORTUNITY": true, "NEXT_BEST_ACTION": true, "VISIT_PREPARATION": true, "DISCOVERY_QUESTIONS": true, "OBJECTION_HANDLING": true, "SALES_PITCH": true, "FOLLOW_UP": true, "DEAL_RISK_ANALYSIS": true}
	if !supported[skill] {
		return "", ErrFinderInput
	}
	review, err := s.repository.FindReview(ctx, id)
	if err != nil {
		return "", err
	}
	if !actor.Role.IsAdminRole() && review.Prospect.AssignedSalesExecutiveID != actor.UserID {
		return "", ErrForbidden
	}
	comments, err := s.repository.ListComments(ctx, id)
	if err != nil {
		return "", err
	}
	var details *prospectmodel.PlaceDetails
	if s.places != nil {
		if loaded, detailErr := s.places.DetailFull(ctx, review.Prospect.GooglePlaceID); detailErr == nil {
			details = &loaded
		}
	}
	chatRepo, ok := s.repository.(repository.AIChatRepository)
	if !ok {
		return "", errors.New("AI chat history repository is unavailable")
	}
	items, err := chatRepo.ListRecentAIChats(ctx, id, 8)
	if err != nil {
		return "", err
	}
	history := make([]ChatTurn, 0, len(items)*2)
	for _, item := range items {
		history = append(history, ChatTurn{Role: "user", Content: item.Message}, ChatTurn{Role: "assistant", Content: item.Answer})
	}
	result, err := s.chatAI(ctx, review, details, comments, history, message, skill)
	if err != nil {
		return "", err
	}
	var parsed struct{ Answer, Skill, Insight, Why, RecommendedAction string }
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return "", err
	}
	_, err = chatRepo.CreateAIChat(ctx, prospectmodel.ProspectAIChat{ProspectID: id, UserID: actor.UserID, Message: message, Answer: parsed.Answer, Skill: parsed.Skill, Insight: parsed.Insight, Why: parsed.Why, RecommendedAction: parsed.RecommendedAction})
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *Service) AIChatHistory(ctx context.Context, actor Actor, id uuid.UUID) ([]prospectmodel.ProspectAIChat, error) {
	if !actor.can("use_prospect_ai_chat") {
		return nil, ErrForbidden
	}
	review, err := s.repository.FindReview(ctx, id)
	if err != nil {
		return nil, err
	}
	if !actor.Role.IsAdminRole() && review.Prospect.AssignedSalesExecutiveID != actor.UserID {
		return nil, ErrForbidden
	}
	chatRepo, ok := s.repository.(repository.AIChatRepository)
	if !ok {
		return nil, errors.New("AI chat history repository is unavailable")
	}
	return chatRepo.ListAIChats(ctx, id, 100)
}

func (s *Service) MyProspects(ctx context.Context, actor Actor) ([]prospectmodel.Prospect, error) {
	if !actor.can("view_my_prospects") {
		return nil, ErrForbidden
	}
	return s.repository.ListAssigned(ctx, actor.UserID)
}

func (s *Service) TeamDashboard(ctx context.Context, actor Actor) (prospectmodel.TeamDashboard, error) {
	if !actor.can("view_team_dashboard") {
		return prospectmodel.TeamDashboard{}, ErrForbidden
	}
	return s.repository.TeamDashboard(ctx, actor.UserID)
}

func (s *Service) Transition(ctx context.Context, actor Actor, id uuid.UUID, status prospectmodel.Status, notes string) (prospectmodel.Prospect, error) {
	if !actor.can("update_prospect_pipeline") {
		return prospectmodel.Prospect{}, ErrForbidden
	}
	review, err := s.repository.FindReview(ctx, id)
	if err != nil {
		return prospectmodel.Prospect{}, err
	}
	if review.Prospect.AssignedSalesExecutiveID != actor.UserID {
		return prospectmodel.Prospect{}, ErrForbidden
	}
	if !validTransition(review.Prospect.Status, status) {
		return prospectmodel.Prospect{}, ErrTransition
	}
	if (status == prospectmodel.StatusWon || status == prospectmodel.StatusLost) && strings.TrimSpace(notes) == "" {
		return prospectmodel.Prospect{}, ErrNotesRequired
	}
	item, err := s.repository.Transition(ctx, id, actor.UserID, review.Prospect.Status, status, strings.TrimSpace(notes))
	if errors.Is(err, repository.ErrInvalidStatus) {
		return prospectmodel.Prospect{}, ErrProspectStatus
	}
	if errors.Is(err, repository.ErrNotOwner) {
		return prospectmodel.Prospect{}, ErrForbidden
	}
	return item, err
}

func validTransition(from, to prospectmodel.Status) bool {
	if to == prospectmodel.StatusLost {
		for _, active := range prospectmodel.ActiveStatuses {
			if from == active {
				return true
			}
		}
		return false
	}
	for i, active := range prospectmodel.ActiveStatuses {
		if active == from {
			if i > 0 && to == prospectmodel.ActiveStatuses[i-1] {
				return true
			}
			if i+1 < len(prospectmodel.ActiveStatuses) {
				return to == prospectmodel.ActiveStatuses[i+1]
			}
			return to == prospectmodel.StatusWon
		}
	}
	return false
}

func (s *Service) CheckIn(ctx context.Context, actor Actor, prospectID uuid.UUID, input prospectmodel.CheckInInput) (prospectmodel.Visit, error) {
	if !actor.can("check_in_prospect") {
		return prospectmodel.Visit{}, ErrForbidden
	}
	if !validCoordinates(input.Latitude, input.Longitude) {
		return prospectmodel.Visit{}, ErrVisitCoordinates
	}
	input.VisitNotes = strings.TrimSpace(input.VisitNotes)
	visit, err := s.repository.CheckIn(ctx, prospectID, actor.UserID, input)
	if errors.Is(err, repository.ErrNotOwner) {
		return prospectmodel.Visit{}, ErrForbidden
	}
	return visit, err
}

func (s *Service) CheckOut(ctx context.Context, actor Actor, prospectID, visitID uuid.UUID, input prospectmodel.CheckOutInput) (prospectmodel.Visit, error) {
	if !actor.can("check_out_prospect") {
		return prospectmodel.Visit{}, ErrForbidden
	}
	if !validCoordinates(input.Latitude, input.Longitude) {
		return prospectmodel.Visit{}, ErrVisitCoordinates
	}
	input.FollowUpNotes = strings.TrimSpace(input.FollowUpNotes)
	visit, err := s.repository.CheckOut(ctx, prospectID, visitID, actor.UserID, input)
	if errors.Is(err, repository.ErrNotOwner) {
		return prospectmodel.Visit{}, ErrForbidden
	}
	return visit, err
}

func validCoordinates(latitude, longitude float64) bool {
	return latitude >= -90 && latitude <= 90 && longitude >= -180 && longitude <= 180
}

func (s *Service) WonQueue(ctx context.Context, actor Actor) ([]prospectmodel.Prospect, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListWon(ctx)
}

func (s *Service) Review(ctx context.Context, actor Actor, id uuid.UUID) (prospectmodel.Review, error) {
	if !actor.Role.IsAdminRole() {
		return prospectmodel.Review{}, ErrForbidden
	}
	return s.repository.FindReview(ctx, id)
}

func (s *Service) MyProspect(ctx context.Context, actor Actor, id uuid.UUID) (prospectmodel.Review, error) {
	if !actor.can("view_my_prospect_detail") {
		return prospectmodel.Review{}, ErrForbidden
	}
	review, err := s.repository.FindReview(ctx, id)
	if err != nil {
		return prospectmodel.Review{}, err
	}
	if review.Prospect.AssignedSalesExecutiveID != actor.UserID {
		return prospectmodel.Review{}, ErrForbidden
	}
	return review, nil
}

func (s *Service) Pipeline(ctx context.Context, actor Actor) ([]prospectmodel.Prospect, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListAll(ctx)
}

func (s *Service) SalesExecutives(ctx context.Context, actor Actor) ([]prospectmodel.SalesExecutive, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListSalesExecutives(ctx)
}

func (s *Service) CustomerMarkers(ctx context.Context, actor Actor) ([]prospectmodel.CustomerMarker, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListCustomerMarkers(ctx)
}

func (s *Service) MentionUsers(ctx context.Context, actor Actor) ([]prospectmodel.SalesExecutive, error) {
	if !actor.Role.IsAdminRole() && !actor.can("view_my_prospects") {
		return nil, ErrForbidden
	}
	return s.repository.ListMentionUsers(ctx)
}

func (s *Service) SearchPlaces(ctx context.Context, actor Actor, input prospectmodel.PlaceSearchInput) ([]prospectmodel.PlaceResult, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	if input.Radius < 500 || input.Radius > 50000 || input.Latitude < -90 || input.Latitude > 90 || input.Longitude < -180 || input.Longitude > 180 {
		return nil, ErrFinderInput
	}
	if s.places == nil {
		return nil, ErrPlacesDisabled
	}
	results, err := s.places.Search(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		ids := make([]string, 0, len(results))
		for _, r := range results {
			ids = append(ids, r.GooglePlaceID)
		}
		existing, err := s.repository.ExistingCustomerPlaceIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		for i := range results {
			if existing[results[i].GooglePlaceID] {
				results[i].IsCustomer = true
			}
		}
	}
	return results, nil
}

func (s *Service) PlaceDetail(ctx context.Context, actor Actor, placeID string) (prospectmodel.PlaceResult, error) {
	if !actor.Role.IsAdminRole() {
		return prospectmodel.PlaceResult{}, ErrForbidden
	}
	if strings.TrimSpace(placeID) == "" {
		return prospectmodel.PlaceResult{}, ErrFinderInput
	}
	if s.places == nil {
		return prospectmodel.PlaceResult{}, ErrPlacesDisabled
	}
	return s.places.Detail(ctx, placeID)
}

func (s *Service) PlaceDetailFull(ctx context.Context, placeID string) (prospectmodel.PlaceDetails, error) {
	if strings.TrimSpace(placeID) == "" {
		return prospectmodel.PlaceDetails{}, ErrFinderInput
	}
	if s.places == nil {
		return prospectmodel.PlaceDetails{}, ErrPlacesDisabled
	}
	return s.places.DetailFull(ctx, placeID)
}

func (s *Service) MenuImages(ctx context.Context, query string, limit int) ([]prospectmodel.MenuImage, error) {
	if strings.TrimSpace(query) == "" {
		return nil, ErrFinderInput
	}
	if s.places == nil {
		return nil, ErrMenuImagesDisabled
	}
	return s.places.SearchMenuImages(ctx, query, limit)
}

func (s *Service) PlacePhoto(ctx context.Context, name string) ([]byte, string, error) {
	name = strings.TrimSpace(name)
	if !ValidGooglePlacePhotoName(name) {
		return nil, "", ErrPlacePhotoInvalid
	}
	if s.places == nil {
		return nil, "", ErrPlacesDisabled
	}
	return s.places.FetchPhoto(ctx, name)
}

func (s *Service) Save(ctx context.Context, actor Actor, input prospectmodel.SaveProspectInput) (prospectmodel.Prospect, error) {
	if !actor.Role.IsAdminRole() {
		return prospectmodel.Prospect{}, ErrForbidden
	}
	// The legacy industry_group column stores the selected place category.
	// Derive it server-side so it always matches Google Places data.
	input.IndustryGroup = strings.TrimSpace(input.Place.PlaceCategory)
	if input.Place.GooglePlaceID == "" || input.Place.PlaceName == "" || input.Place.FormattedAddress == "" || input.IndustryGroup == "" || input.AssignedSalesExecutiveID == uuid.Nil {
		return prospectmodel.Prospect{}, ErrFinderInput
	}
	existing, err := s.repository.ExistingCustomerPlaceIDs(ctx, []string{input.Place.GooglePlaceID})
	if err != nil {
		return prospectmodel.Prospect{}, err
	}
	if existing[input.Place.GooglePlaceID] {
		return prospectmodel.Prospect{}, ErrAlreadyCustomer
	}
	item, err := s.repository.Create(ctx, input, actor.UserID)
	if err == nil && s.initialAnalysis != nil {
		go func() {
			analysisCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			review, reviewErr := s.repository.FindReview(analysisCtx, item.ID)
			if reviewErr != nil {
				return
			}
			comments, _ := s.repository.ListComments(analysisCtx, item.ID)
			var details *prospectmodel.PlaceDetails
			if s.places != nil {
				if loaded, detailErr := s.places.DetailFull(analysisCtx, item.GooglePlaceID); detailErr == nil {
					details = &loaded
				}
			}
			s.initialAnalysis(analysisCtx, review, details, comments)
		}()
	}
	return item, err
}

func (s *Service) ListVisitMonitoring(ctx context.Context, actor Actor, filter prospectmodel.VisitMonitoringFilter) ([]prospectmodel.VisitMonitoringItem, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListVisitMonitoring(ctx, filter)
}

func (s *Service) Report(ctx context.Context, actor Actor, filter prospectmodel.ReportFilter) (prospectmodel.Report, error) {
	if !actor.Role.IsAdminRole() {
		return prospectmodel.Report{}, ErrForbidden
	}
	from, fromErr := time.Parse("2006-01-02", filter.DateFrom)
	to, toErr := time.Parse("2006-01-02", filter.DateTo)
	if fromErr != nil || toErr != nil || to.Before(from) {
		return prospectmodel.Report{}, ErrFinderInput
	}
	visits, err := s.repository.ListVisitMonitoring(ctx, prospectmodel.VisitMonitoringFilter{DateFrom: filter.DateFrom, DateTo: filter.DateTo, SalesExecutiveID: filter.SalesExecutiveID})
	if err != nil {
		return prospectmodel.Report{}, err
	}
	prospects, err := s.repository.ListAll(ctx)
	if err != nil {
		return prospectmodel.Report{}, err
	}
	to = to.Add(24*time.Hour - time.Nanosecond)
	territoryOf := func(address string) string {
		parts := strings.Split(address, ",")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[len(parts)-2])
		}
		return "Unspecified"
	}
	territorySet := map[string]bool{}
	filtered := make([]prospectmodel.VisitMonitoringItem, 0, len(visits))
	for _, v := range visits {
		territory := territoryOf(v.FormattedAddress)
		territorySet[territory] = true
		if filter.Territory == "" || filter.Territory == territory {
			filtered = append(filtered, v)
		}
	}
	report := prospectmodel.Report{Trends: []prospectmodel.ReportTrend{}, Stages: []prospectmodel.ReportStage{}, Performance: []prospectmodel.SalesPerformance{}, Territories: []string{}}
	for t := range territorySet {
		report.Territories = append(report.Territories, t)
	}
	sort.Strings(report.Territories)
	perf := map[string]*prospectmodel.SalesPerformance{}
	visited := map[string]bool{}
	for _, v := range filtered {
		report.KPI.TotalVisits++
		if v.RadiusStatus == "INSIDE" {
			report.KPI.WithinRadius++
		} else if v.RadiusStatus == "OUTSIDE" {
			report.KPI.OutsideRadius++
		}
		id := v.SalesExecutiveID.String()
		row := perf[id]
		if row == nil {
			row = &prospectmodel.SalesPerformance{SalesExecutiveID: id, SalesExecutiveName: v.SalesExecutiveName, Territory: territoryOf(v.FormattedAddress)}
			perf[id] = row
		}
		row.Visits++
		if v.RadiusStatus == "INSIDE" {
			row.WithinRadius++
		}
		visited[v.ProspectID.String()] = true
	}
	stageCounts := map[prospectmodel.Status]int{}
	wonBySales := map[string]int{}
	for _, p := range prospects {
		stageCounts[p.Status]++
		if p.Status == prospectmodel.StatusWon && !p.UpdatedAt.Before(from) && !p.UpdatedAt.After(to) && (filter.SalesExecutiveID == "" || p.AssignedSalesExecutiveID.String() == filter.SalesExecutiveID) {
			report.KPI.WonProspects++
			wonBySales[p.AssignedSalesExecutiveID.String()]++
		}
	}
	labels := []struct {
		name  string
		count int
	}{{"Discovered", len(prospects)}, {"Saved", len(prospects) - stageCounts[prospectmodel.StatusLost]}, {"Assigned", len(prospects) - stageCounts[prospectmodel.StatusLost]}, {"Visited", len(visited)}, {"Won", report.KPI.WonProspects}}
	for _, v := range labels {
		report.Stages = append(report.Stages, prospectmodel.ReportStage{Label: v.name, Count: v.count})
	}
	weeks := 1
	if !from.IsZero() && !to.IsZero() {
		weeks = int(to.Sub(from).Hours()/24/7) + 1
	}
	if weeks > 12 {
		weeks = 12
	}
	for i := 0; i < weeks; i++ {
		report.Trends = append(report.Trends, prospectmodel.ReportTrend{Label: fmt.Sprintf("W%d", i+1)})
	}
	for _, v := range filtered {
		idx := int(v.CheckInAt.Sub(from).Hours() / 24 / 7)
		if idx < 0 {
			continue
		}
		if idx >= weeks {
			idx = weeks - 1
		}
		if v.RadiusStatus == "INSIDE" {
			report.Trends[idx].WithinRadius++
		} else if v.RadiusStatus == "OUTSIDE" {
			report.Trends[idx].OutsideRadius++
		}
	}
	for id, row := range perf {
		row.ProspectsWon = wonBySales[id]
		if row.Visits > 0 {
			row.Conversion = float64(row.ProspectsWon) / float64(row.Visits) * 100
			row.Performance = int((float64(row.WithinRadius)/float64(row.Visits)*.7 + min(1, float64(row.ProspectsWon)/5)*.3) * 100)
		}
		report.Performance = append(report.Performance, *row)
	}
	sort.Slice(report.Performance, func(i, j int) bool { return report.Performance[i].Performance > report.Performance[j].Performance })
	return report, nil
}

func (s *Service) ListMyVisits(ctx context.Context, actor Actor, filter prospectmodel.VisitMonitoringFilter) ([]prospectmodel.VisitMonitoringItem, error) {
	if !actor.can("view_own_visits") {
		return nil, ErrForbidden
	}
	return s.repository.ListMyVisits(ctx, actor.UserID, filter)
}

func (s *Service) ListProspectVisits(ctx context.Context, prospectID uuid.UUID) ([]prospectmodel.Visit, error) {
	return s.repository.ListProspectVisits(ctx, prospectID)
}

func (s *Service) DeleteVisit(ctx context.Context, actor Actor, visitID uuid.UUID) (prospectmodel.Visit, error) {
	if actor.Role.IsAdminRole() {
		return s.repository.DeleteVisit(ctx, visitID, uuid.Nil)
	}
	if !actor.can("delete_visit") {
		return prospectmodel.Visit{}, ErrForbidden
	}
	return s.repository.DeleteVisit(ctx, visitID, actor.UserID)
}

func (s *Service) DeleteProspect(ctx context.Context, actor Actor, id uuid.UUID) ([]string, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.DeleteProspect(ctx, id)
}

func (s *Service) RequestDeletion(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.can("request_prospect_deletion") {
		return ErrForbidden
	}
	ownerID, err := s.repository.FindProspectOwner(ctx, id)
	if err != nil {
		return fmt.Errorf("find prospect owner: %w", err)
	}
	if ownerID != actor.UserID {
		return ErrForbidden
	}
	return s.repository.RequestDeletion(ctx, id, actor.UserID)
}

func (s *Service) ApproveDeletion(ctx context.Context, actor Actor, id uuid.UUID) ([]string, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ApproveDeletion(ctx, id)
}

func (s *Service) RejectDeletion(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.Role.IsAdminRole() {
		return ErrForbidden
	}
	return s.repository.RejectDeletion(ctx, id)
}

func (s *Service) ListComments(ctx context.Context, actor Actor, prospectID uuid.UUID) ([]prospectmodel.ProspectComment, error) {
	if err := s.ensureCommentAccess(ctx, actor, prospectID); err != nil {
		return nil, err
	}
	return s.repository.ListComments(ctx, prospectID)
}

func (s *Service) CreateComment(ctx context.Context, actor Actor, prospectID uuid.UUID, content string, attachments []prospectmodel.CommentAttachment) (prospectmodel.ProspectComment, error) {
	if err := s.ensureCommentAccess(ctx, actor, prospectID); err != nil {
		return prospectmodel.ProspectComment{}, err
	}
	content = strings.TrimSpace(content)
	if content == "" && len(attachments) == 0 {
		return prospectmodel.ProspectComment{}, ErrNotesRequired
	}
	return s.repository.CreateComment(ctx, prospectID, actor.UserID, content, attachments)
}

func (s *Service) DeleteComment(ctx context.Context, actor Actor, prospectID, commentID uuid.UUID) ([]prospectmodel.CommentAttachment, error) {
	if err := s.ensureCommentAccess(ctx, actor, prospectID); err != nil {
		return nil, err
	}
	ownerID := actor.UserID
	if actor.Role.IsAdminRole() {
		ownerID = uuid.Nil
	}
	return s.repository.DeleteComment(ctx, prospectID, commentID, ownerID)
}

func (s *Service) CommentAttachment(ctx context.Context, actor Actor, prospectID, attachmentID uuid.UUID) (prospectmodel.CommentAttachment, error) {
	if err := s.ensureCommentAccess(ctx, actor, prospectID); err != nil {
		return prospectmodel.CommentAttachment{}, err
	}
	return s.repository.FindCommentAttachment(ctx, prospectID, attachmentID)
}

func (s *Service) ensureCommentAccess(ctx context.Context, actor Actor, prospectID uuid.UUID) error {
	if actor.Role.IsAdminRole() {
		_, err := s.repository.FindProspectOwner(ctx, prospectID)
		return err
	}
	if !actor.can("view_my_prospect_detail") {
		return ErrForbidden
	}
	accessible, err := s.repository.ProspectAccessibleTo(ctx, prospectID, actor.UserID)
	if err != nil {
		return err
	}
	if !accessible {
		return ErrForbidden
	}
	return nil
}

func (s *Service) ListPhotoTags(ctx context.Context, actor Actor, prospectID uuid.UUID) ([]prospectmodel.ProspectPhotoTag, error) {
	if err := s.ensurePhotoTagAccess(ctx, actor, prospectID); err != nil {
		return nil, err
	}
	return s.repository.ListPhotoTags(ctx, prospectID)
}

func (s *Service) SetPhotoTag(ctx context.Context, actor Actor, prospectID uuid.UUID, photoName string, category prospectmodel.PhotoCategory) (prospectmodel.ProspectPhotoTag, error) {
	if err := s.ensurePhotoTagAccess(ctx, actor, prospectID); err != nil {
		return prospectmodel.ProspectPhotoTag{}, err
	}
	if category != prospectmodel.PhotoCategoryMenu && category != prospectmodel.PhotoCategoryPlace {
		return prospectmodel.ProspectPhotoTag{}, ErrPhotoTagInvalid
	}
	if !ValidGooglePlacePhotoName(photoName) {
		return prospectmodel.ProspectPhotoTag{}, ErrPhotoTagInvalid
	}
	return s.repository.UpsertPhotoTag(ctx, prospectID, photoName, category, actor.UserID)
}

func (s *Service) ensurePhotoTagAccess(ctx context.Context, actor Actor, prospectID uuid.UUID) error {
	if actor.Role.IsAdminRole() {
		_, err := s.repository.FindProspectOwner(ctx, prospectID)
		return err
	}
	if !actor.can("view_my_prospect_detail") {
		return ErrForbidden
	}
	accessible, err := s.repository.ProspectAccessibleTo(ctx, prospectID, actor.UserID)
	if err != nil {
		return err
	}
	if !accessible {
		return ErrForbidden
	}
	return nil
}

func (actor Actor) can(permissionKey string) bool {
	if actor.Role == authmodel.RoleSalesExecutive && actor.PermissionKeys == nil {
		return true
	}
	for _, key := range actor.PermissionKeys {
		if key == permissionKey {
			return true
		}
	}
	return false
}
