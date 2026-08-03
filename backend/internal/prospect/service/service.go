package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"crm-prospect-simulator/backend/internal/prospect/repository"
	"github.com/google/uuid"
)

var (
	ErrForbidden        = errors.New("prospect operation forbidden")
	ErrTransition       = errors.New("invalid prospect stage transition")
	ErrNotesRequired    = errors.New("a win note or loss reason is required")
	ErrProspectStatus   = errors.New("prospect is not eligible for this operation")
	ErrFinderInput      = errors.New("prospect finder query is invalid")
	ErrPlacesDisabled   = errors.New("Google Places server key is not configured")
	ErrVisitCoordinates = errors.New("visit GPS coordinates are invalid")
)

type Actor struct {
	UserID uuid.UUID
	Role   authmodel.Role
}

type Service struct {
	repository repository.Repository
	places     Places
}

type Places interface {
	Search(context.Context, prospectmodel.PlaceSearchInput) ([]prospectmodel.PlaceResult, error)
	Detail(context.Context, string) (prospectmodel.PlaceResult, error)
	DetailFull(context.Context, string) (prospectmodel.PlaceDetails, error)
}

func New(repo repository.Repository, placeClients ...Places) *Service {
	var places Places
	if len(placeClients) > 0 {
		places = placeClients[0]
	}
	return &Service{repository: repo, places: places}
}

func (s *Service) MyProspects(ctx context.Context, actor Actor) ([]prospectmodel.Prospect, error) {
	if actor.Role != authmodel.RoleSalesExecutive {
		return nil, ErrForbidden
	}
	return s.repository.ListAssigned(ctx, actor.UserID)
}

func (s *Service) Transition(ctx context.Context, actor Actor, id uuid.UUID, status prospectmodel.Status, notes string) (prospectmodel.Prospect, error) {
	if actor.Role != authmodel.RoleSalesExecutive {
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
	if actor.Role != authmodel.RoleSalesExecutive {
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
	if actor.Role != authmodel.RoleSalesExecutive {
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
	if actor.Role != authmodel.RoleSalesExecutive {
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

func (s *Service) MentionUsers(ctx context.Context, actor Actor) ([]prospectmodel.SalesExecutive, error) {
	if !actor.Role.IsAdminRole() && actor.Role != authmodel.RoleSalesExecutive {
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
	return s.places.Search(ctx, input)
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

func (s *Service) Save(ctx context.Context, actor Actor, input prospectmodel.SaveProspectInput) (prospectmodel.Prospect, error) {
	if !actor.Role.IsAdminRole() {
		return prospectmodel.Prospect{}, ErrForbidden
	}
	input.IndustryGroup = strings.TrimSpace(input.IndustryGroup)
	if input.Place.GooglePlaceID == "" || input.Place.PlaceName == "" || input.Place.FormattedAddress == "" || input.IndustryGroup == "" || input.AssignedSalesExecutiveID == uuid.Nil {
		return prospectmodel.Prospect{}, ErrFinderInput
	}
	return s.repository.Create(ctx, input, actor.UserID)
}

func (s *Service) ListVisitMonitoring(ctx context.Context, actor Actor, filter prospectmodel.VisitMonitoringFilter) ([]prospectmodel.VisitMonitoringItem, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListVisitMonitoring(ctx, filter)
}

func (s *Service) ListMyVisits(ctx context.Context, actor Actor, filter prospectmodel.VisitMonitoringFilter) ([]prospectmodel.VisitMonitoringItem, error) {
	if actor.Role != authmodel.RoleSalesExecutive {
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
	if actor.Role != authmodel.RoleSalesExecutive {
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
	if actor.Role != authmodel.RoleSalesExecutive {
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
	if actor.Role != authmodel.RoleSalesExecutive {
		return ErrForbidden
	}
	ownerID, err := s.repository.FindProspectOwner(ctx, prospectID)
	if err != nil {
		return err
	}
	if ownerID != actor.UserID {
		return ErrForbidden
	}
	return nil
}
