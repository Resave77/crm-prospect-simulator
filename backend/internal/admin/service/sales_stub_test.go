package service

import (
	"context"
	"time"

	"crm-prospect-simulator/backend/internal/admin/model"
	"crm-prospect-simulator/backend/internal/admin/repository"
	"github.com/google/uuid"
)

func (r *repoStub) ListSalesRoles(_ context.Context) ([]model.SalesRole, error) { return nil, nil }
func (r *repoStub) FindSalesRole(_ context.Context, id uuid.UUID) (model.SalesRole, error) {
	return model.SalesRole{}, repository.ErrNotFound
}
func (r *repoStub) CreateSalesRole(_ context.Context, _ uuid.UUID, _ model.CreateSalesRoleInput, _ uuid.UUID) error {
	return nil
}
func (r *repoStub) UpdateSalesRole(_ context.Context, _ uuid.UUID, _ model.UpdateSalesRoleInput, _ uuid.UUID) error {
	return nil
}
func (r *repoStub) UpdateSalesRoleStatus(_ context.Context, _ uuid.UUID, _ bool, _ uuid.UUID) error {
	return nil
}
func (r *repoStub) SalesRoleNameExists(_ context.Context, _ string, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *repoStub) SalesRoleHasAssignments(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (r *repoStub) CreateSalesAssignment(_ context.Context, _ uuid.UUID, _ model.CreateAssignmentInput, _ uuid.UUID) error {
	return nil
}
func (r *repoStub) MoveSalesAssignment(_ context.Context, _, _ uuid.UUID, _ model.MoveAssignmentInput, _ uuid.UUID) error {
	return nil
}
func (r *repoStub) FindSalesAssignment(_ context.Context, _ uuid.UUID) (model.SalesStructureAssignment, error) {
	return model.SalesStructureAssignment{}, repository.ErrNotFound
}
func (r *repoStub) FindEffectiveSalesAssignment(_ context.Context, _ uuid.UUID, _ time.Time) (model.SalesStructureAssignment, model.SalesRole, error) {
	return model.SalesStructureAssignment{}, model.SalesRole{}, repository.ErrNotFound
}
func (r *repoStub) SalesAssignmentOverlaps(_ context.Context, _ uuid.UUID, _ time.Time, _ *time.Time, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *repoStub) UserExists(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := r.users[id]
	return ok, nil
}
func (r *repoStub) ListSalesStructure(_ context.Context, _ time.Time) ([]model.SalesStructureItem, error) {
	return nil, nil
}
func (r *repoStub) ListSalesAssignmentHistory(_ context.Context, _ uuid.UUID) ([]model.AssignmentHistoryItem, error) {
	return nil, nil
}
