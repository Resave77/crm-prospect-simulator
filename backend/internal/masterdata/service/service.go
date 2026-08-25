package service

import (
	"context"
	"errors"
	"strings"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/masterdata/model"
	"crm-prospect-simulator/backend/internal/masterdata/repository"
	"github.com/google/uuid"
)

var ErrForbidden = errors.New("master data operation forbidden")

type ValidationError struct{ message string }

func (e *ValidationError) Error() string { return e.message }

func validationError(message string) error {
	return &ValidationError{message: message}
}

type Actor struct {
	UserID uuid.UUID
	Role   authmodel.Role
}

type Service struct {
	repository repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) ListSegments(ctx context.Context, actor Actor, search, status string) ([]model.Segment, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListSegments(ctx, search, normalizeStatus(status))
}

func (s *Service) CreateSegment(ctx context.Context, actor Actor, input model.SegmentInput) (model.Segment, error) {
	if !actor.Role.IsAdminRole() {
		return model.Segment{}, ErrForbidden
	}
	normalizeSegmentInput(&input)
	if err := validateName(input.Name); err != nil {
		return model.Segment{}, err
	}
	if !validStatus(input.Status) {
		return model.Segment{}, validationError("Status must be ACTIVE or INACTIVE.")
	}
	return s.repository.CreateSegment(ctx, input)
}

func (s *Service) UpdateSegment(ctx context.Context, actor Actor, id uuid.UUID, input model.SegmentInput) (model.Segment, error) {
	if !actor.Role.IsAdminRole() {
		return model.Segment{}, ErrForbidden
	}
	normalizeSegmentInput(&input)
	if err := validateName(input.Name); err != nil {
		return model.Segment{}, err
	}
	if !validStatus(input.Status) {
		return model.Segment{}, validationError("Status must be ACTIVE or INACTIVE.")
	}
	return s.repository.UpdateSegment(ctx, id, input)
}

func (s *Service) DeleteSegment(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.Role.IsAdminRole() {
		return ErrForbidden
	}
	return s.repository.DeleteSegment(ctx, id)
}

func (s *Service) ListTrashedSegments(ctx context.Context, actor Actor) ([]model.Segment, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListTrashedSegments(ctx)
}

func (s *Service) RestoreSegment(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.Role.IsAdminRole() {
		return ErrForbidden
	}
	return s.repository.RestoreSegment(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context, actor Actor, search, segmentID, status string) ([]model.Category, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(segmentID) != "" {
		if _, err := uuid.Parse(segmentID); err != nil {
			return nil, validationError("Segment filter is invalid.")
		}
	}
	return s.repository.ListCategories(ctx, search, strings.TrimSpace(segmentID), normalizeStatus(status))
}

func (s *Service) ListCategoriesBySegment(ctx context.Context, actor Actor, segmentID uuid.UUID) ([]model.Category, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListCategoriesBySegment(ctx, segmentID)
}

func (s *Service) CreateCategory(ctx context.Context, actor Actor, input model.CategoryInput) (model.Category, error) {
	if !actor.Role.IsAdminRole() {
		return model.Category{}, ErrForbidden
	}
	if err := s.prepareCategory(ctx, &input); err != nil {
		return model.Category{}, err
	}
	return s.repository.CreateCategory(ctx, input)
}

func (s *Service) UpdateCategory(ctx context.Context, actor Actor, id uuid.UUID, input model.CategoryInput) (model.Category, error) {
	if !actor.Role.IsAdminRole() {
		return model.Category{}, ErrForbidden
	}
	if err := s.prepareCategory(ctx, &input); err != nil {
		return model.Category{}, err
	}
	return s.repository.UpdateCategory(ctx, id, input)
}

func (s *Service) DeleteCategory(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.Role.IsAdminRole() {
		return ErrForbidden
	}
	return s.repository.DeleteCategory(ctx, id)
}

func (s *Service) ListTrashedCategories(ctx context.Context, actor Actor) ([]model.Category, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repository.ListTrashedCategories(ctx)
}

func (s *Service) RestoreCategory(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.Role.IsAdminRole() {
		return ErrForbidden
	}
	return s.repository.RestoreCategory(ctx, id)
}

func (s *Service) prepareCategory(ctx context.Context, input *model.CategoryInput) error {
	input.SegmentID = strings.TrimSpace(input.SegmentID)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.PlaceAPI = normalizePlaceAPI(input.PlaceAPI)
	input.Status = normalizeStatus(input.Status)
	if err := validateName(input.Name); err != nil {
		return err
	}
	segmentID, err := uuid.Parse(input.SegmentID)
	if err != nil {
		return validationError("Segment is required.")
	}
	if _, err := s.repository.GetSegment(ctx, segmentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return validationError("Selected segment does not exist.")
		}
		return err
	}
	if !validStatus(input.Status) {
		return validationError("Status must be ACTIVE or INACTIVE.")
	}
	return nil
}

func normalizeSegmentInput(input *model.SegmentInput) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = normalizeStatus(input.Status)
}

func validateName(name string) error {
	if name == "" {
		return validationError("Name is required.")
	}
	if len(name) > 100 {
		return validationError("Name must be 100 characters or fewer.")
	}
	return nil
}

func validStatus(status string) bool {
	return status == model.StatusActive || status == model.StatusInactive
}

func normalizePlaceAPI(placeAPI string) string {
	parts := strings.Split(placeAPI, ",")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			normalized = append(normalized, strings.ToLower(trimmed))
		}
	}
	return strings.Join(normalized, ", ")
}

func normalizeStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case model.StatusInactive:
		return model.StatusInactive
	case "":
		return ""
	default:
		return model.StatusActive
	}
}
