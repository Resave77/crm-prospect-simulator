package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"crm-prospect-simulator/backend/internal/admin/model"
	"crm-prospect-simulator/backend/internal/admin/repository"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"

	"github.com/google/uuid"
)

var (
	ErrSalesRoleNameRequired  = errors.New("sales role name required")
	ErrInvalidSalesRoleLevel  = errors.New("invalid sales role level")
	ErrSalesRoleNameExists    = errors.New("sales role name exists")
	ErrSalesRoleInUse         = errors.New("sales role in use")
	ErrSalesRoleInactive      = errors.New("sales role inactive")
	ErrSalesUserInactive      = errors.New("sales user inactive")
	ErrInvalidHierarchy       = errors.New("invalid sales hierarchy")
	ErrInvalidEffectiveDate   = errors.New("invalid effective date")
	ErrAssignmentOverlap      = errors.New("sales assignment overlaps")
	ErrIncompatibleChildren   = errors.New("assignment has incompatible children")
	ErrAssignmentAlreadyEnded = errors.New("sales assignment already ended")
)

var defaultSalesRoleIDs = map[uuid.UUID]bool{
	uuid.MustParse("00000000-0000-0000-0000-000000000101"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000102"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000103"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000104"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000105"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000106"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000107"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000108"): true,
}

func (s *Service) ListSalesRoles(ctx context.Context, actor Actor) ([]model.SalesRole, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repo.ListSalesRoles(ctx)
}

func (s *Service) GetSalesRole(ctx context.Context, actor Actor, id uuid.UUID) (model.SalesRole, error) {
	if !actor.Role.IsAdminRole() {
		return model.SalesRole{}, ErrForbidden
	}
	role, err := s.repo.FindSalesRole(ctx, id)
	if err != nil {
		return model.SalesRole{}, err
	}
	return s.roleWithPermissions(ctx, role)
}

func (s *Service) CreateSalesRole(ctx context.Context, actor Actor, input model.CreateSalesRoleInput) (model.SalesRole, error) {
	if !actor.Role.IsAdminRole() {
		return model.SalesRole{}, ErrForbidden
	}
	if err := s.validateSalesRole(ctx, input.Name, input.Level, nil); err != nil {
		return model.SalesRole{}, err
	}
	keys, err := s.validateRolePermissionsForCreate(ctx, &input)
	if err != nil {
		return model.SalesRole{}, err
	}
	input.PermissionKeys = keys
	id := uuid.New()
	if err := s.repo.CreateSalesRole(ctx, id, input, actor.UserID); err != nil {
		return model.SalesRole{}, err
	}
	role, err := s.repo.FindSalesRole(ctx, id)
	if err != nil {
		return model.SalesRole{}, err
	}
	return s.roleWithPermissions(ctx, role)
}

func (s *Service) UpdateSalesRole(ctx context.Context, actor Actor, id uuid.UUID, input model.UpdateSalesRoleInput) (model.SalesRole, error) {
	if !actor.Role.IsAdminRole() {
		return model.SalesRole{}, ErrForbidden
	}
	current, err := s.repo.FindSalesRole(ctx, id)
	if err != nil {
		return model.SalesRole{}, err
	}
	name := current.Name
	level := current.Level
	if input.Name != nil {
		name = *input.Name
	}
	if input.Level != nil {
		level = *input.Level
	}
	if err := s.validateSalesRole(ctx, name, level, &id); err != nil {
		return model.SalesRole{}, err
	}
	keys, replacePermissions, err := s.validateRolePermissionsForUpdate(ctx, current, &input)
	if err != nil {
		return model.SalesRole{}, err
	}
	if replacePermissions {
		input.PermissionKeys = keys
	}
	if input.Level != nil && *input.Level != current.Level {
		inUse, err := s.repo.SalesRoleHasAssignments(ctx, id)
		if err != nil {
			return model.SalesRole{}, err
		}
		if inUse {
			return model.SalesRole{}, ErrSalesRoleInUse
		}
	}
	if err := s.repo.UpdateSalesRole(ctx, id, input, actor.UserID); err != nil {
		return model.SalesRole{}, err
	}
	return s.repo.FindSalesRole(ctx, id)
}

func (s *Service) UpdateSalesRoleStatus(ctx context.Context, actor Actor, id uuid.UUID, isActive bool) (model.SalesRole, error) {
	if !actor.Role.IsAdminRole() {
		return model.SalesRole{}, ErrForbidden
	}
	if _, err := s.repo.FindSalesRole(ctx, id); err != nil {
		return model.SalesRole{}, err
	}
	if err := s.repo.UpdateSalesRoleStatus(ctx, id, isActive, actor.UserID); err != nil {
		return model.SalesRole{}, err
	}
	return s.repo.FindSalesRole(ctx, id)
}

func (s *Service) DeleteSalesRole(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.Role.IsAdminRole() {
		return ErrForbidden
	}
	if _, err := s.repo.FindSalesRole(ctx, id); err != nil {
		return err
	}
	if defaultSalesRoleIDs[id] {
		return ErrSalesRoleInUse
	}
	inUse, err := s.repo.SalesRoleHasAssignments(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return ErrSalesRoleInUse
	}
	return s.repo.DeleteSalesRole(ctx, id)
}

func (s *Service) CreateSalesAssignment(ctx context.Context, actor Actor, input model.CreateAssignmentInput) (model.SalesStructureAssignment, error) {
	if !actor.Role.IsAdminRole() {
		return model.SalesStructureAssignment{}, ErrForbidden
	}
	if err := s.validateAssignment(ctx, input.UserID, input.SalesRoleID, input.ParentUserID, input.EffectiveFrom.Time, nil); err != nil {
		return model.SalesStructureAssignment{}, err
	}
	id := uuid.New()
	if err := s.repo.CreateSalesAssignment(ctx, id, input, actor.UserID); err != nil {
		return model.SalesStructureAssignment{}, err
	}
	return s.repo.FindSalesAssignment(ctx, id)
}

func (s *Service) MoveSalesAssignment(ctx context.Context, actor Actor, currentID uuid.UUID, input model.MoveAssignmentInput) (model.SalesStructureAssignment, error) {
	if !actor.Role.IsAdminRole() {
		return model.SalesStructureAssignment{}, ErrForbidden
	}
	current, err := s.repo.FindSalesAssignment(ctx, currentID)
	if err != nil {
		return model.SalesStructureAssignment{}, err
	}
	if input.SalesRoleID == uuid.Nil {
		input.SalesRoleID = current.SalesRoleID
	}
	if input.ParentUserID == nil {
		input.ParentUserID = current.ParentUserID
	}
	if !input.EffectiveFrom.Time.After(current.EffectiveFrom) {
		return model.SalesStructureAssignment{}, ErrInvalidEffectiveDate
	}
	if err := s.validateAssignment(ctx, current.UserID, input.SalesRoleID, input.ParentUserID, input.EffectiveFrom.Time, &currentID); err != nil {
		return model.SalesStructureAssignment{}, err
	}
	newID := uuid.New()
	if err := s.repo.MoveSalesAssignment(ctx, currentID, newID, input, actor.UserID); err != nil {
		return model.SalesStructureAssignment{}, err
	}
	return s.repo.FindSalesAssignment(ctx, newID)
}

func (s *Service) EndSalesAssignment(ctx context.Context, actor Actor, assignmentID uuid.UUID, input model.EndAssignmentInput) (model.SalesStructureAssignment, error) {
	if !actor.Role.IsAdminRole() {
		return model.SalesStructureAssignment{}, ErrForbidden
	}
	effectiveTo := truncateDate(input.EffectiveTo.Time)
	if effectiveTo.IsZero() {
		return model.SalesStructureAssignment{}, ErrInvalidEffectiveDate
	}
	current, err := s.repo.FindSalesAssignment(ctx, assignmentID)
	if err != nil {
		return model.SalesStructureAssignment{}, err
	}
	if current.EffectiveTo != nil {
		return model.SalesStructureAssignment{}, ErrAssignmentAlreadyEnded
	}
	if effectiveTo.Before(truncateDate(current.EffectiveFrom)) {
		return model.SalesStructureAssignment{}, ErrInvalidEffectiveDate
	}
	hasChildren, err := s.repo.HasActiveChildAssignments(ctx, current.UserID, effectiveTo)
	if err != nil {
		return model.SalesStructureAssignment{}, err
	}
	if hasChildren {
		return model.SalesStructureAssignment{}, ErrInvalidHierarchy
	}
	if err := s.repo.EndSalesAssignment(ctx, assignmentID, effectiveTo, actor.UserID); err != nil {
		return model.SalesStructureAssignment{}, err
	}
	return s.repo.FindSalesAssignment(ctx, assignmentID)
}

func (s *Service) ListSalesStructure(ctx context.Context, actor Actor, effectiveDate time.Time) ([]model.SalesStructureItem, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repo.ListSalesStructure(ctx, truncateDate(effectiveDate))
}

func (s *Service) ListSalesAssignmentHistory(ctx context.Context, actor Actor, userID uuid.UUID) ([]model.AssignmentHistoryItem, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	exists, err := s.repo.UserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, repository.ErrNotFound
	}
	return s.repo.ListSalesAssignmentHistory(ctx, userID)
}

func (s *Service) validateSalesRole(ctx context.Context, name string, level int, excludeID *uuid.UUID) error {
	if strings.TrimSpace(name) == "" {
		return ErrSalesRoleNameRequired
	}
	if level < 1 || level > 4 {
		return ErrInvalidSalesRoleLevel
	}
	exists, err := s.repo.SalesRoleNameExists(ctx, normalizeSalesRoleName(name), excludeID)
	if err != nil {
		return err
	}
	if exists {
		return ErrSalesRoleNameExists
	}
	return nil
}

func (s *Service) validateAssignment(ctx context.Context, userID, roleID uuid.UUID, parentID *uuid.UUID, effectiveFrom time.Time, excludeID *uuid.UUID) error {
	date := truncateDate(effectiveFrom)
	if date.IsZero() || date.Day() != 1 {
		return ErrInvalidEffectiveDate
	}
	if parentID != nil && *parentID == userID {
		return ErrInvalidHierarchy
	}
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Status != authmodel.UserActive {
		return ErrSalesUserInactive
	}
	if user.Role == authmodel.RoleSuperAdmin {
		return ErrInvalidHierarchy
	}
	role, err := s.repo.FindSalesRole(ctx, roleID)
	if err != nil {
		return err
	}
	if !role.IsActive {
		return ErrSalesRoleInactive
	}
	if err := s.ensureCompatibleChildren(ctx, userID, role.Level, date); err != nil {
		return err
	}
	if role.Level == 1 {
		if parentID != nil {
			return ErrInvalidHierarchy
		}
	} else {
		if parentID == nil {
			return ErrInvalidHierarchy
		}
		parent, err := s.repo.FindUserByID(ctx, *parentID)
		if err != nil {
			return err
		}
		if parent.Status != authmodel.UserActive {
			return ErrSalesUserInactive
		}
		_, parentRole, err := s.repo.FindEffectiveSalesAssignment(ctx, *parentID, date)
		if err != nil {
			return ErrInvalidHierarchy
		}
		if parentRole.Level != role.Level-1 {
			return ErrInvalidHierarchy
		}
		if err := s.ensureNoCycle(ctx, userID, *parentID, date); err != nil {
			return err
		}
	}
	overlaps, err := s.repo.SalesAssignmentOverlaps(ctx, userID, date, nil, excludeID)
	if err != nil {
		return err
	}
	if overlaps {
		return ErrAssignmentOverlap
	}
	return nil
}

func (s *Service) ensureCompatibleChildren(ctx context.Context, userID uuid.UUID, level int, date time.Time) error {
	hasIncompatible, err := s.repo.HasIncompatibleCurrentChildren(ctx, userID, level, date)
	if err != nil {
		return err
	}
	if hasIncompatible {
		return ErrIncompatibleChildren
	}
	return nil
}

func (s *Service) ensureNoCycle(ctx context.Context, userID, parentID uuid.UUID, date time.Time) error {
	seen := map[uuid.UUID]bool{userID: true}
	current := parentID
	for {
		if seen[current] {
			return ErrInvalidHierarchy
		}
		seen[current] = true
		assignment, _, err := s.repo.FindEffectiveSalesAssignment(ctx, current, date)
		if err != nil {
			return nil
		}
		if assignment.ParentUserID == nil {
			return nil
		}
		current = *assignment.ParentUserID
	}
}

func normalizeSalesRoleName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}
func truncateDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(model.DateLayout)
	return &s
}
func previousDay(t time.Time) time.Time { return truncateDate(t).AddDate(0, 0, -1) }
func invalidHierarchyf(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalidHierarchy}, args...)...)
}
