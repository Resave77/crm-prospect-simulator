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
	if role, ok := r.roles[id]; ok {
		return role, nil
	}
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
func (r *repoStub) DeleteSalesRole(_ context.Context, _ uuid.UUID) error {
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
func (r *repoStub) EndSalesAssignment(_ context.Context, _ uuid.UUID, _ time.Time, _ uuid.UUID) error {
	return nil
}
func (r *repoStub) SetCurrentSalesAssignment(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ *uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (r *repoStub) DeleteUser(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (r *repoStub) FindSalesAssignment(_ context.Context, _ uuid.UUID) (model.SalesStructureAssignment, error) {
	return model.SalesStructureAssignment{}, repository.ErrNotFound
}
func (r *repoStub) FindEffectiveSalesAssignment(_ context.Context, userID uuid.UUID, _ time.Time) (model.SalesStructureAssignment, model.SalesRole, error) {
	if r.effectiveRoles != nil {
		if role, ok := r.effectiveRoles[userID]; ok {
			return model.SalesStructureAssignment{UserID: userID, SalesRoleID: role.ID}, role, nil
		}
	}
	return model.SalesStructureAssignment{}, model.SalesRole{}, repository.ErrNotFound
}
func (r *repoStub) SalesAssignmentOverlaps(_ context.Context, _ uuid.UUID, _ time.Time, _ *time.Time, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *repoStub) HasIncompatibleCurrentChildren(_ context.Context, _ uuid.UUID, _ int, _ time.Time) (bool, error) {
	return false, nil
}
func (r *repoStub) HasActiveChildAssignments(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return false, nil
}
func (r *repoStub) CountEffectiveLevel1Roots(_ context.Context, _ time.Time, _ *uuid.UUID) (int, error) {
	return 0, nil
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
func (r *repoStub) ListPermissions(_ context.Context, _ string) ([]model.Permission, error) {
	return nil, nil
}
func (r *repoStub) FindPermissionByKey(_ context.Context, _ string) (model.Permission, error) {
	return model.Permission{}, repository.ErrNotFound
}
func (r *repoStub) FindPermissionsByKeys(_ context.Context, _ []string) ([]model.Permission, error) {
	return nil, nil
}
func (r *repoStub) ListRolePermissions(_ context.Context, _ uuid.UUID) ([]model.Permission, error) {
	return nil, nil
}
