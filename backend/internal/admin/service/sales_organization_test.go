package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"crm-prospect-simulator/backend/internal/admin/model"
	"crm-prospect-simulator/backend/internal/admin/repository"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
)

type salesRepo struct {
	repoStub
	roles             map[uuid.UUID]model.SalesRole
	assignments       map[uuid.UUID]model.SalesStructureAssignment
	nameExists        bool
	roleInUse         bool
	overlap           bool
	createdAssignment uuid.UUID
	movedFrom         uuid.UUID
	movedTo           uuid.UUID
}

func (r *salesRepo) FindSalesRole(_ context.Context, id uuid.UUID) (model.SalesRole, error) {
	if v, ok := r.roles[id]; ok {
		return v, nil
	}
	return model.SalesRole{}, repository.ErrNotFound
}
func (r *salesRepo) SalesRoleNameExists(_ context.Context, _ string, _ *uuid.UUID) (bool, error) {
	return r.nameExists, nil
}
func (r *salesRepo) SalesRoleHasAssignments(_ context.Context, _ uuid.UUID) (bool, error) {
	return r.roleInUse, nil
}
func (r *salesRepo) CreateSalesRole(_ context.Context, id uuid.UUID, input model.CreateSalesRoleInput, _ uuid.UUID) error {
	r.roles[id] = model.SalesRole{ID: id, Name: input.Name, Level: input.Level, Description: input.Description, IsActive: true}
	return nil
}
func (r *salesRepo) UpdateSalesRole(_ context.Context, id uuid.UUID, input model.UpdateSalesRoleInput, _ uuid.UUID) error {
	role := r.roles[id]
	if input.Name != nil {
		role.Name = *input.Name
	}
	if input.Level != nil {
		role.Level = *input.Level
	}
	r.roles[id] = role
	return nil
}
func (r *salesRepo) UpdateSalesRoleStatus(_ context.Context, id uuid.UUID, active bool, _ uuid.UUID) error {
	role := r.roles[id]
	role.IsActive = active
	r.roles[id] = role
	return nil
}
func (r *salesRepo) DeleteSalesRole(_ context.Context, id uuid.UUID) error {
	delete(r.roles, id)
	return nil
}
func (r *salesRepo) UserExists(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := r.users[id]
	return ok, nil
}
func (r *salesRepo) FindEffectiveSalesAssignment(_ context.Context, userID uuid.UUID, date time.Time) (model.SalesStructureAssignment, model.SalesRole, error) {
	for _, a := range r.assignments {
		if a.UserID == userID && !date.Before(a.EffectiveFrom) && (a.EffectiveTo == nil || !date.After(*a.EffectiveTo)) {
			return a, r.roles[a.SalesRoleID], nil
		}
	}
	return model.SalesStructureAssignment{}, model.SalesRole{}, repository.ErrNotFound
}
func (r *salesRepo) SalesAssignmentOverlaps(_ context.Context, _ uuid.UUID, _ time.Time, _ *time.Time, _ *uuid.UUID) (bool, error) {
	return r.overlap, nil
}
func (r *salesRepo) CreateSalesAssignment(_ context.Context, id uuid.UUID, input model.CreateAssignmentInput, _ uuid.UUID) error {
	r.createdAssignment = id
	r.assignments[id] = model.SalesStructureAssignment{ID: id, UserID: input.UserID, SalesRoleID: input.SalesRoleID, ParentUserID: input.ParentUserID, EffectiveFrom: input.EffectiveFrom.Time}
	return nil
}
func (r *salesRepo) FindSalesAssignment(_ context.Context, id uuid.UUID) (model.SalesStructureAssignment, error) {
	if v, ok := r.assignments[id]; ok {
		return v, nil
	}
	return model.SalesStructureAssignment{}, repository.ErrNotFound
}
func (r *salesRepo) MoveSalesAssignment(_ context.Context, currentID, newID uuid.UUID, input model.MoveAssignmentInput, _ uuid.UUID) error {
	r.movedFrom = currentID
	r.movedTo = newID
	old := r.assignments[currentID]
	closeDate := input.EffectiveFrom.Time.AddDate(0, 0, -1)
	old.EffectiveTo = &closeDate
	r.assignments[currentID] = old
	r.assignments[newID] = model.SalesStructureAssignment{ID: newID, UserID: old.UserID, SalesRoleID: input.SalesRoleID, ParentUserID: input.ParentUserID, EffectiveFrom: input.EffectiveFrom.Time}
	return nil
}

func salesDate(s string) model.SalesStructureDate {
	d, _ := time.Parse(model.DateLayout, s)
	return model.SalesStructureDate{Time: d}
}
func salesTestRepo() (*salesRepo, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	l1, l2, l3, l4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	return &salesRepo{roles: map[uuid.UUID]model.SalesRole{l1: {ID: l1, Level: 1, IsActive: true, Name: "L1"}, l2: {ID: l2, Level: 2, IsActive: true, Name: "L2"}, l3: {ID: l3, Level: 3, IsActive: true, Name: "L3"}, l4: {ID: l4, Level: 4, IsActive: true, Name: "L4"}}, repoStub: repoStub{users: map[uuid.UUID]authmodel.User{}}, assignments: map[uuid.UUID]model.SalesStructureAssignment{}}, l1, l2, l3, l4
}

func TestSalesRoleValidation(t *testing.T) {
	repo, _, _, _, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()
	for _, level := range []int{0, 5} {
		if _, err := svc.CreateSalesRole(context.Background(), actor, model.CreateSalesRoleInput{Name: "X", Level: level}); !errors.Is(err, ErrInvalidSalesRoleLevel) {
			t.Fatalf("level %d err=%v", level, err)
		}
	}
	if _, err := svc.CreateSalesRole(context.Background(), actor, model.CreateSalesRoleInput{Name: " ", Level: 1}); !errors.Is(err, ErrSalesRoleNameRequired) {
		t.Fatalf("empty name err=%v", err)
	}
	repo.nameExists = true
	if _, err := svc.CreateSalesRole(context.Background(), actor, model.CreateSalesRoleInput{Name: "X", Level: 1}); !errors.Is(err, ErrSalesRoleNameExists) {
		t.Fatalf("duplicate err=%v", err)
	}
}

func TestSalesAssignmentHierarchyAndMove(t *testing.T) {
	repo, l1, l2, l3, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	repo.users[u1] = authmodel.User{ID: u1}
	repo.users[u2] = authmodel.User{ID: u2}
	repo.users[u3] = authmodel.User{ID: u3}
	a1, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{UserID: u1, SalesRoleID: l1, EffectiveFrom: salesDate("2026-08-01")})
	if err != nil {
		t.Fatalf("l1 assignment: %v", err)
	}
	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{UserID: u2, SalesRoleID: l2, ParentUserID: &u1, EffectiveFrom: salesDate("2026-08-01")}); err != nil {
		t.Fatalf("l2 assignment: %v", err)
	}
	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{UserID: u3, SalesRoleID: l3, ParentUserID: &u2, EffectiveFrom: salesDate("2026-08-01")}); err != nil {
		t.Fatalf("l3 assignment: %v", err)
	}
	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{UserID: uuid.New(), SalesRoleID: l2, ParentUserID: &u2, EffectiveFrom: salesDate("2026-08-01")}); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("missing user err=%v", err)
	}
	moved, err := svc.MoveSalesAssignment(context.Background(), actor, a1.ID, model.MoveAssignmentInput{SalesRoleID: l1, EffectiveFrom: salesDate("2026-09-01")})
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if moved.EffectiveFrom.Format(model.DateLayout) != "2026-09-01" {
		t.Fatalf("move date=%v", moved.EffectiveFrom)
	}
	if repo.assignments[a1.ID].EffectiveTo == nil || repo.assignments[a1.ID].EffectiveTo.Format(model.DateLayout) != "2026-08-31" {
		t.Fatalf("old not closed: %+v", repo.assignments[a1.ID])
	}
}

func TestSalesAssignmentRejectsInactiveOverlapAndBadMonth(t *testing.T) {
	repo, l1, _, _, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()
	u := uuid.New()
	repo.users[u] = authmodel.User{ID: u}
	role := repo.roles[l1]
	role.IsActive = false
	repo.roles[l1] = role
	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{UserID: u, SalesRoleID: l1, EffectiveFrom: salesDate("2026-08-01")}); !errors.Is(err, ErrSalesRoleInactive) {
		t.Fatalf("inactive err=%v", err)
	}
	role.IsActive = true
	repo.roles[l1] = role
	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{UserID: u, SalesRoleID: l1, EffectiveFrom: salesDate("2026-08-02")}); !errors.Is(err, ErrInvalidEffectiveDate) {
		t.Fatalf("month err=%v", err)
	}
	repo.overlap = true
	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{UserID: u, SalesRoleID: l1, EffectiveFrom: salesDate("2026-08-01")}); !errors.Is(err, ErrAssignmentOverlap) {
		t.Fatalf("overlap err=%v", err)
	}
}
