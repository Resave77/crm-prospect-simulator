package server

import (
	"context"
	"time"

	adminmodel "crm-prospect-simulator/backend/internal/admin/model"
	adminrepo "crm-prospect-simulator/backend/internal/admin/repository"
	"github.com/google/uuid"
)

func (r *adminRepoStub) ListSalesRoles(_ context.Context) ([]adminmodel.SalesRole, error) {
	return []adminmodel.SalesRole{}, nil
}
func (r *adminRepoStub) FindSalesRole(_ context.Context, id uuid.UUID) (adminmodel.SalesRole, error) {
	return adminmodel.SalesRole{}, adminrepo.ErrNotFound
}
func (r *adminRepoStub) CreateSalesRole(_ context.Context, _ uuid.UUID, _ adminmodel.CreateSalesRoleInput, _ uuid.UUID) error {
	return nil
}
func (r *adminRepoStub) UpdateSalesRole(_ context.Context, _ uuid.UUID, _ adminmodel.UpdateSalesRoleInput, _ uuid.UUID) error {
	return nil
}
func (r *adminRepoStub) UpdateSalesRoleStatus(_ context.Context, _ uuid.UUID, _ bool, _ uuid.UUID) error {
	return nil
}
func (r *adminRepoStub) DeleteSalesRole(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (r *adminRepoStub) SalesRoleNameExists(_ context.Context, _ string, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *adminRepoStub) SalesRoleHasAssignments(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (r *adminRepoStub) CreateSalesAssignment(_ context.Context, _ uuid.UUID, _ adminmodel.CreateAssignmentInput, _ uuid.UUID) error {
	return nil
}
func (r *adminRepoStub) MoveSalesAssignment(_ context.Context, _, _ uuid.UUID, _ adminmodel.MoveAssignmentInput, _ uuid.UUID) error {
	return nil
}
func (r *adminRepoStub) FindSalesAssignment(_ context.Context, _ uuid.UUID) (adminmodel.SalesStructureAssignment, error) {
	return adminmodel.SalesStructureAssignment{}, adminrepo.ErrNotFound
}
func (r *adminRepoStub) FindEffectiveSalesAssignment(_ context.Context, _ uuid.UUID, _ time.Time) (adminmodel.SalesStructureAssignment, adminmodel.SalesRole, error) {
	return adminmodel.SalesStructureAssignment{}, adminmodel.SalesRole{}, adminrepo.ErrNotFound
}
func (r *adminRepoStub) SalesAssignmentOverlaps(_ context.Context, _ uuid.UUID, _ time.Time, _ *time.Time, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *adminRepoStub) UserExists(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil }
func (r *adminRepoStub) ListSalesStructure(_ context.Context, _ time.Time) ([]adminmodel.SalesStructureItem, error) {
	return []adminmodel.SalesStructureItem{}, nil
}
func (r *adminRepoStub) ListSalesAssignmentHistory(_ context.Context, _ uuid.UUID) ([]adminmodel.AssignmentHistoryItem, error) {
	return []adminmodel.AssignmentHistoryItem{}, nil
}

func (r *patchAdminRepo) ListSalesRoles(_ context.Context) ([]adminmodel.SalesRole, error) {
	return []adminmodel.SalesRole{}, nil
}
func (r *patchAdminRepo) FindSalesRole(_ context.Context, id uuid.UUID) (adminmodel.SalesRole, error) {
	return adminmodel.SalesRole{}, adminrepo.ErrNotFound
}
func (r *patchAdminRepo) CreateSalesRole(_ context.Context, _ uuid.UUID, _ adminmodel.CreateSalesRoleInput, _ uuid.UUID) error {
	return nil
}
func (r *patchAdminRepo) UpdateSalesRole(_ context.Context, _ uuid.UUID, _ adminmodel.UpdateSalesRoleInput, _ uuid.UUID) error {
	return nil
}
func (r *patchAdminRepo) UpdateSalesRoleStatus(_ context.Context, _ uuid.UUID, _ bool, _ uuid.UUID) error {
	return nil
}
func (r *patchAdminRepo) DeleteSalesRole(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (r *patchAdminRepo) SalesRoleNameExists(_ context.Context, _ string, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *patchAdminRepo) SalesRoleHasAssignments(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (r *patchAdminRepo) CreateSalesAssignment(_ context.Context, _ uuid.UUID, _ adminmodel.CreateAssignmentInput, _ uuid.UUID) error {
	return nil
}
func (r *patchAdminRepo) MoveSalesAssignment(_ context.Context, _, _ uuid.UUID, _ adminmodel.MoveAssignmentInput, _ uuid.UUID) error {
	return nil
}
func (r *patchAdminRepo) FindSalesAssignment(_ context.Context, _ uuid.UUID) (adminmodel.SalesStructureAssignment, error) {
	return adminmodel.SalesStructureAssignment{}, adminrepo.ErrNotFound
}
func (r *patchAdminRepo) FindEffectiveSalesAssignment(_ context.Context, _ uuid.UUID, _ time.Time) (adminmodel.SalesStructureAssignment, adminmodel.SalesRole, error) {
	return adminmodel.SalesStructureAssignment{}, adminmodel.SalesRole{}, adminrepo.ErrNotFound
}
func (r *patchAdminRepo) SalesAssignmentOverlaps(_ context.Context, _ uuid.UUID, _ time.Time, _ *time.Time, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *patchAdminRepo) UserExists(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil }
func (r *patchAdminRepo) ListSalesStructure(_ context.Context, _ time.Time) ([]adminmodel.SalesStructureItem, error) {
	return []adminmodel.SalesStructureItem{}, nil
}
func (r *patchAdminRepo) ListSalesAssignmentHistory(_ context.Context, _ uuid.UUID) ([]adminmodel.AssignmentHistoryItem, error) {
	return []adminmodel.AssignmentHistoryItem{}, nil
}
