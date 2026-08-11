package service

import (
	"context"
	"errors"
	"strings"
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
	incompatible      bool
	activeChildren    bool
	level1Roots       int
	moveErr           error
	endErr            error
	createdAssignment uuid.UUID
	movedFrom         uuid.UUID
	movedTo           uuid.UUID
	endedAssignment   uuid.UUID
	endedTo           time.Time
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
	r.roles[id] = model.SalesRole{ID: id, Name: input.Name, Level: input.Level, Description: input.Description, IsActive: true, LandingPage: input.LandingPage}
	for _, key := range input.PermissionKeys {
		permission, err := r.FindPermissionByKey(context.Background(), key)
		if err != nil {
			return err
		}
		role := r.roles[id]
		role.Permissions = append(role.Permissions, permission)
		r.roles[id] = role
	}
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
	if input.Description != nil {
		role.Description = *input.Description
	}
	if input.LandingPage != nil {
		role.LandingPage = input.LandingPage
	}
	if input.PermissionKeys != nil {
		role.Permissions = nil
		for _, key := range input.PermissionKeys {
			permission, err := r.FindPermissionByKey(context.Background(), key)
			if err != nil {
				return err
			}
			role.Permissions = append(role.Permissions, permission)
		}
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
func (r *salesRepo) FindUserByID(_ context.Context, id uuid.UUID) (authmodel.User, error) {
	user, ok := r.users[id]
	if !ok {
		return authmodel.User{}, repository.ErrNotFound
	}
	if user.Status == "" {
		user.Status = authmodel.UserActive
	}
	return user, nil
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
func (r *salesRepo) HasIncompatibleCurrentChildren(_ context.Context, _ uuid.UUID, _ int, _ time.Time) (bool, error) {
	return r.incompatible, nil
}
func (r *salesRepo) HasActiveChildAssignments(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return r.activeChildren, nil
}
func (r *salesRepo) CountEffectiveLevel1Roots(_ context.Context, _ time.Time, _ *uuid.UUID) (int, error) {
	return r.level1Roots, nil
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
	if r.moveErr != nil {
		return r.moveErr
	}
	r.movedFrom = currentID
	r.movedTo = newID
	old := r.assignments[currentID]
	closeDate := input.EffectiveFrom.Time.AddDate(0, 0, -1)
	old.EffectiveTo = &closeDate
	r.assignments[currentID] = old
	r.assignments[newID] = model.SalesStructureAssignment{ID: newID, UserID: old.UserID, SalesRoleID: input.SalesRoleID, ParentUserID: input.ParentUserID, EffectiveFrom: input.EffectiveFrom.Time}
	return nil
}
func (r *salesRepo) EndSalesAssignment(_ context.Context, assignmentID uuid.UUID, effectiveTo time.Time, _ uuid.UUID) error {
	if r.endErr != nil {
		return r.endErr
	}
	r.endedAssignment = assignmentID
	r.endedTo = effectiveTo
	assignment, ok := r.assignments[assignmentID]
	if !ok {
		return repository.ErrNotFound
	}
	assignment.EffectiveTo = &effectiveTo
	r.assignments[assignmentID] = assignment
	return nil
}
func (r *salesRepo) ListSalesAssignmentHistory(_ context.Context, userID uuid.UUID) ([]model.AssignmentHistoryItem, error) {
	items := []model.AssignmentHistoryItem{}
	for _, assignment := range r.assignments {
		if assignment.UserID != userID {
			continue
		}
		role := r.roles[assignment.SalesRoleID]
		status := "CURRENT"
		if assignment.EffectiveTo != nil {
			status = "PAST"
		}
		items = append(items, model.AssignmentHistoryItem{
			AssignmentID:  assignment.ID,
			SalesRole:     model.SalesRoleRef{ID: role.ID, Name: role.Name, Level: role.Level},
			ParentUserID:  assignment.ParentUserID,
			EffectiveFrom: assignment.EffectiveFrom.Format(model.DateLayout),
			EffectiveTo:   formatDatePtr(assignment.EffectiveTo),
			Status:        status,
		})
	}
	for i := range items {
		for j := i + 1; j < len(items); j++ {
			if items[j].EffectiveFrom > items[i].EffectiveFrom {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items, nil
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
	if _, err := svc.CreateSalesRole(context.Background(), actor, model.CreateSalesRoleInput{Name: "X", Level: 2}); !errors.Is(err, ErrSalesRoleNameExists) {
		t.Fatalf("duplicate err=%v", err)
	}
}

func TestSalesAssignmentHierarchyAndMove(t *testing.T) {
	repo, l1, l2, l3, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	repo.users[u1] = authmodel.User{ID: u1, Role: authmodel.RoleSuperAdmin, Status: authmodel.UserActive}
	repo.users[u2] = authmodel.User{ID: u2, Status: authmodel.UserActive}
	repo.users[u3] = authmodel.User{ID: u3, Status: authmodel.UserActive}
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
	repo.users[u] = authmodel.User{ID: u, Role: authmodel.RoleSuperAdmin, Status: authmodel.UserActive}
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

func TestSalesAssignmentMoveScenariosAndHistory(t *testing.T) {
	repo, l1, l2, l3, l4 := salesTestRepo()
	svc := New(repo)
	actor := adminActor()
	root, managerA, managerB, supervisor, rep := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo.users[root] = authmodel.User{ID: root, Role: authmodel.RoleSuperAdmin, Status: authmodel.UserActive}
	for _, id := range []uuid.UUID{managerA, managerB, supervisor, rep} {
		repo.users[id] = authmodel.User{ID: id, Status: authmodel.UserActive}
	}
	repo.assignments[uuid.New()] = model.SalesStructureAssignment{ID: uuid.New(), UserID: root, SalesRoleID: l1, EffectiveFrom: mustSalesTime("2026-08-01")}
	repo.assignments[uuid.New()] = model.SalesStructureAssignment{ID: uuid.New(), UserID: managerA, SalesRoleID: l2, ParentUserID: &root, EffectiveFrom: mustSalesTime("2026-08-01")}
	repo.assignments[uuid.New()] = model.SalesStructureAssignment{ID: uuid.New(), UserID: managerB, SalesRoleID: l2, ParentUserID: &root, EffectiveFrom: mustSalesTime("2026-08-01")}
	supervisorAssignmentID := uuid.New()
	repo.assignments[supervisorAssignmentID] = model.SalesStructureAssignment{ID: supervisorAssignmentID, UserID: supervisor, SalesRoleID: l3, ParentUserID: &managerA, EffectiveFrom: mustSalesTime("2026-08-01")}
	repAssignmentID := uuid.New()
	repo.assignments[repAssignmentID] = model.SalesStructureAssignment{ID: repAssignmentID, UserID: rep, SalesRoleID: l4, ParentUserID: &supervisor, EffectiveFrom: mustSalesTime("2026-08-01")}

	moved, err := svc.MoveSalesAssignment(context.Background(), actor, supervisorAssignmentID, model.MoveAssignmentInput{SalesRoleID: l3, ParentUserID: &managerB, EffectiveFrom: salesDate("2026-09-01")})
	if err != nil {
		t.Fatalf("same-level team move: %v", err)
	}
	if moved.ParentUserID == nil || *moved.ParentUserID != managerB {
		t.Fatalf("parent after team move=%v", moved.ParentUserID)
	}
	if repo.assignments[supervisorAssignmentID].EffectiveTo == nil || repo.assignments[supervisorAssignmentID].EffectiveTo.Format(model.DateLayout) != "2026-08-31" {
		t.Fatalf("old assignment close=%v", repo.assignments[supervisorAssignmentID].EffectiveTo)
	}
	history, err := svc.ListSalesAssignmentHistory(context.Background(), actor, supervisor)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(history) != 2 || history[0].Status != "CURRENT" || history[1].Status != "PAST" || history[0].EffectiveFrom != "2026-09-01" {
		t.Fatalf("history order/status=%+v", history)
	}

	promoted, err := svc.MoveSalesAssignment(context.Background(), actor, repAssignmentID, model.MoveAssignmentInput{SalesRoleID: l3, ParentUserID: &managerB, EffectiveFrom: salesDate("2026-09-01")})
	if err != nil {
		t.Fatalf("promotion: %v", err)
	}
	demoted, err := svc.MoveSalesAssignment(context.Background(), actor, promoted.ID, model.MoveAssignmentInput{SalesRoleID: l4, ParentUserID: &supervisor, EffectiveFrom: salesDate("2026-10-01")})
	if err != nil {
		t.Fatalf("demotion: %v", err)
	}
	if demoted.EffectiveFrom.Format(model.DateLayout) != "2026-10-01" {
		t.Fatalf("demotion date=%v", demoted.EffectiveFrom)
	}
	roleChange, err := svc.MoveSalesAssignment(context.Background(), actor, demoted.ID, model.MoveAssignmentInput{SalesRoleID: l4, ParentUserID: &supervisor, EffectiveFrom: salesDate("2026-11-01")})
	if err != nil {
		t.Fatalf("same-level role change: %v", err)
	}
	if roleChange.SalesRoleID != l4 {
		t.Fatalf("role change role=%s", roleChange.SalesRoleID)
	}
}

func TestEndSalesAssignmentSucceedsAndPreservesHistory(t *testing.T) {
	repo, l1, _, _, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()
	userID := uuid.New()
	repo.users[userID] = authmodel.User{ID: userID, Status: authmodel.UserActive}
	assignmentID := uuid.New()
	repo.assignments[assignmentID] = model.SalesStructureAssignment{ID: assignmentID, UserID: userID, SalesRoleID: l1, EffectiveFrom: mustSalesTime("2026-08-01")}

	ended, err := svc.EndSalesAssignment(context.Background(), actor, assignmentID, model.EndAssignmentInput{EffectiveTo: salesDate("2026-08-31")})
	if err != nil {
		t.Fatalf("end assignment: %v", err)
	}
	if ended.ID != assignmentID || ended.EffectiveTo == nil || ended.EffectiveTo.Format(model.DateLayout) != "2026-08-31" {
		t.Fatalf("ended assignment=%+v", ended)
	}
	if repo.endedAssignment != assignmentID || repo.endedTo.Format(model.DateLayout) != "2026-08-31" {
		t.Fatalf("repo end call assignment=%s to=%s", repo.endedAssignment, repo.endedTo.Format(model.DateLayout))
	}
	history, err := svc.ListSalesAssignmentHistory(context.Background(), actor, userID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(history) != 1 || history[0].Status != "PAST" || history[0].EffectiveTo == nil || *history[0].EffectiveTo != "2026-08-31" {
		t.Fatalf("history after end=%+v", history)
	}
}

func TestEndSalesAssignmentRejectsInvalidStates(t *testing.T) {
	repo, l1, _, _, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()
	userID := uuid.New()
	repo.users[userID] = authmodel.User{ID: userID, Status: authmodel.UserActive}

	if _, err := svc.EndSalesAssignment(context.Background(), actor, uuid.New(), model.EndAssignmentInput{EffectiveTo: salesDate("2026-08-31")}); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("not found err=%v", err)
	}

	assignmentID := uuid.New()
	endedAt := mustSalesTime("2026-08-15")
	repo.assignments[assignmentID] = model.SalesStructureAssignment{ID: assignmentID, UserID: userID, SalesRoleID: l1, EffectiveFrom: mustSalesTime("2026-08-01"), EffectiveTo: &endedAt}
	if _, err := svc.EndSalesAssignment(context.Background(), actor, assignmentID, model.EndAssignmentInput{EffectiveTo: salesDate("2026-08-31")}); !errors.Is(err, ErrAssignmentAlreadyEnded) {
		t.Fatalf("already ended err=%v", err)
	}

	activeID := uuid.New()
	repo.assignments[activeID] = model.SalesStructureAssignment{ID: activeID, UserID: userID, SalesRoleID: l1, EffectiveFrom: mustSalesTime("2026-08-01")}
	if _, err := svc.EndSalesAssignment(context.Background(), actor, activeID, model.EndAssignmentInput{EffectiveTo: salesDate("2026-07-31")}); !errors.Is(err, ErrInvalidEffectiveDate) {
		t.Fatalf("before start err=%v", err)
	}

	repo.activeChildren = true
	if _, err := svc.EndSalesAssignment(context.Background(), actor, activeID, model.EndAssignmentInput{EffectiveTo: salesDate("2026-08-31")}); !errors.Is(err, ErrInvalidHierarchy) {
		t.Fatalf("active children err=%v", err)
	}
}

func TestSalesAssignmentMoveRejectsRequiredRules(t *testing.T) {
	repo, l1, l2, l3, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()
	root, manager, managerB, supervisor := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo.users[root] = authmodel.User{ID: root, Role: authmodel.RoleSuperAdmin, Status: authmodel.UserActive}
	for _, id := range []uuid.UUID{manager, managerB, supervisor} {
		repo.users[id] = authmodel.User{ID: id, Status: authmodel.UserActive}
	}
	rootAssignmentID := uuid.New()
	repo.assignments[rootAssignmentID] = model.SalesStructureAssignment{ID: rootAssignmentID, UserID: root, SalesRoleID: l1, EffectiveFrom: mustSalesTime("2026-08-01")}
	managerAssignmentID := uuid.New()
	repo.assignments[managerAssignmentID] = model.SalesStructureAssignment{ID: managerAssignmentID, UserID: manager, SalesRoleID: l2, ParentUserID: &root, EffectiveFrom: mustSalesTime("2026-08-01")}
	repo.assignments[uuid.New()] = model.SalesStructureAssignment{ID: uuid.New(), UserID: managerB, SalesRoleID: l2, ParentUserID: &root, EffectiveFrom: mustSalesTime("2026-08-01")}
	supervisorAssignmentID := uuid.New()
	repo.assignments[supervisorAssignmentID] = model.SalesStructureAssignment{ID: supervisorAssignmentID, UserID: supervisor, SalesRoleID: l3, ParentUserID: &manager, EffectiveFrom: mustSalesTime("2026-08-01")}

	if _, err := svc.MoveSalesAssignment(context.Background(), actor, supervisorAssignmentID, model.MoveAssignmentInput{SalesRoleID: l3, ParentUserID: &root, EffectiveFrom: salesDate("2026-09-01")}); !errors.Is(err, ErrInvalidHierarchy) {
		t.Fatalf("invalid parent level err=%v", err)
	}
	repo.overlap = true
	if _, err := svc.MoveSalesAssignment(context.Background(), actor, supervisorAssignmentID, model.MoveAssignmentInput{SalesRoleID: l3, ParentUserID: &manager, EffectiveFrom: salesDate("2026-09-01")}); !errors.Is(err, ErrAssignmentOverlap) {
		t.Fatalf("overlap err=%v", err)
	}
	repo.overlap = false
	repo.incompatible = true
	if _, err := svc.MoveSalesAssignment(context.Background(), actor, managerAssignmentID, model.MoveAssignmentInput{SalesRoleID: l3, ParentUserID: &managerB, EffectiveFrom: salesDate("2026-09-01")}); !errors.Is(err, ErrIncompatibleChildren) {
		t.Fatalf("children err=%v", err)
	}
	repo.incompatible = false
	if _, err := svc.MoveSalesAssignment(context.Background(), actor, rootAssignmentID, model.MoveAssignmentInput{SalesRoleID: l2, ParentUserID: &root, EffectiveFrom: salesDate("2026-09-01")}); !errors.Is(err, ErrInvalidHierarchy) {
		t.Fatalf("self parent err=%v", err)
	}
	if _, err := svc.MoveSalesAssignment(context.Background(), actor, rootAssignmentID, model.MoveAssignmentInput{SalesRoleID: l2, ParentUserID: &root, EffectiveFrom: salesDate("2026-09-01")}); !errors.Is(err, ErrInvalidHierarchy) {
		t.Fatalf("root protection setup err=%v", err)
	}
}

func TestSalesAssignmentRejectsSecondLevel1Root(t *testing.T) {
	repo, l1, _, _, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()

	root := uuid.New()
	repo.users[root] = authmodel.User{
		ID:     root,
		Role:   authmodel.RoleSuperAdmin,
		Status: authmodel.UserActive,
	}

	rootAssignmentID := uuid.New()
	repo.assignments[rootAssignmentID] = model.SalesStructureAssignment{
		ID:            rootAssignmentID,
		UserID:        root,
		SalesRoleID:   l1,
		EffectiveFrom: mustSalesTime("2026-08-01"),
	}

	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{
		UserID:        uuid.New(),
		SalesRoleID:   l1,
		EffectiveFrom: salesDate("2026-09-01"),
	}); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("missing user should short-circuit root count, err=%v", err)
	}

	secondRoot := uuid.New()
	repo.users[secondRoot] = authmodel.User{
		ID:     secondRoot,
		Role:   authmodel.RoleSuperAdmin,
		Status: authmodel.UserActive,
	}

	repo.level1Roots = 1
	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{
		UserID:        secondRoot,
		SalesRoleID:   l1,
		EffectiveFrom: salesDate("2026-09-01"),
	}); !errors.Is(err, ErrInvalidHierarchy) {
		t.Fatalf("second level 1 root err=%v, want ErrInvalidHierarchy", err)
	}

	repo.level1Roots = 0
	repo.moveErr = repository.ErrConflict
	before := repo.assignments[rootAssignmentID]
	if _, err := svc.MoveSalesAssignment(context.Background(), actor, rootAssignmentID, model.MoveAssignmentInput{
		SalesRoleID:   l1,
		EffectiveFrom: salesDate("2026-10-01"),
	}); !errors.Is(err, repository.ErrConflict) {
		t.Fatalf("move insert failure err=%v", err)
	}

	after := repo.assignments[rootAssignmentID]
	if before.EffectiveTo != after.EffectiveTo {
		t.Fatalf("failed move changed old close: before=%v after=%v", before.EffectiveTo, after.EffectiveTo)
	}
}

func TestSalesAssignmentSuperAdminRootRules(t *testing.T) {
	repo, l1, l2, _, _ := salesTestRepo()
	svc := New(repo)
	actor := adminActor()

	superAdmin := uuid.New()
	repo.users[superAdmin] = authmodel.User{
		ID:     superAdmin,
		Role:   authmodel.RoleSuperAdmin,
		Status: authmodel.UserActive,
	}

	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{
		UserID:        superAdmin,
		SalesRoleID:   l1,
		EffectiveFrom: salesDate("2026-08-01"),
	}); err != nil {
		t.Fatalf("super admin level 1 root should be allowed: %v", err)
	}

	otherSuperAdmin := uuid.New()
	repo.users[otherSuperAdmin] = authmodel.User{
		ID:     otherSuperAdmin,
		Role:   authmodel.RoleSuperAdmin,
		Status: authmodel.UserActive,
	}

	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{
		UserID:        otherSuperAdmin,
		SalesRoleID:   l2,
		ParentUserID:  &superAdmin,
		EffectiveFrom: salesDate("2026-08-01"),
	}); !errors.Is(err, ErrInvalidHierarchy) {
		t.Fatalf("super admin level 2 err=%v, want ErrInvalidHierarchy", err)
	}

	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{
		UserID:        otherSuperAdmin,
		SalesRoleID:   l1,
		ParentUserID:  &superAdmin,
		EffectiveFrom: salesDate("2026-08-01"),
	}); !errors.Is(err, ErrInvalidHierarchy) {
		t.Fatalf("super admin level 1 with parent err=%v, want ErrInvalidHierarchy", err)
	}

	regularUser := uuid.New()
	repo.users[regularUser] = authmodel.User{
		ID:     regularUser,
		Status: authmodel.UserActive,
	}

	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{
		UserID:        regularUser,
		SalesRoleID:   l1,
		EffectiveFrom: salesDate("2026-08-01"),
	}); !errors.Is(err, ErrInvalidHierarchy) {
		t.Fatalf("non-super-admin level 1 err=%v, want ErrInvalidHierarchy", err)
	}
}

func TestSalesAssignmentUnauthorizedRoleForbidden(t *testing.T) {
	repo, _, _, _, _ := salesTestRepo()
	svc := New(repo)
	actor := Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}
	if _, err := svc.CreateSalesAssignment(context.Background(), actor, model.CreateAssignmentInput{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("create forbidden err=%v", err)
	}
	if _, err := svc.MoveSalesAssignment(context.Background(), actor, uuid.New(), model.MoveAssignmentInput{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("move forbidden err=%v", err)
	}
	if _, err := svc.EndSalesAssignment(context.Background(), actor, uuid.New(), model.EndAssignmentInput{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("end forbidden err=%v", err)
	}
	if _, err := svc.ListSalesAssignmentHistory(context.Background(), actor, uuid.New()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("history forbidden err=%v", err)
	}
}

func mustSalesTime(s string) time.Time {
	t, err := time.Parse(model.DateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func (r *salesRepo) ListPermissions(_ context.Context, search string) ([]model.Permission, error) {
	items := auditTestPermissions()
	if search == "" {
		return items, nil
	}
	filtered := []model.Permission{}
	for _, item := range items {
		if strings.Contains(item.Key, search) || strings.Contains(item.Name, search) {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}
func (r *salesRepo) FindPermissionByKey(_ context.Context, key string) (model.Permission, error) {
	for _, item := range auditTestPermissions() {
		if item.Key == key {
			return item, nil
		}
	}
	return model.Permission{}, repository.ErrNotFound
}
func (r *salesRepo) FindPermissionsByKeys(_ context.Context, keys []string) ([]model.Permission, error) {
	items := []model.Permission{}
	for _, key := range keys {
		for _, item := range auditTestPermissions() {
			if item.Key == key {
				items = append(items, item)
			}
		}
	}
	return items, nil
}
func (r *salesRepo) ListRolePermissions(_ context.Context, roleID uuid.UUID) ([]model.Permission, error) {
	role := r.roles[roleID]
	return role.Permissions, nil
}

func auditTestPermissions() []model.Permission {
	menuDashboard := "menu_admin_dashboard"
	menuRoles := "menu_roles"
	menuSales := "menu_sales_dashboard"
	return []model.Permission{
		{ID: uuid.New(), Key: "menu_admin_dashboard", Name: "Admin Dashboard", GroupKey: "dashboard", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/dashboard"), IsActive: true, SortOrder: 1},
		{ID: uuid.New(), Key: "view_admin_dashboard", Name: "View Admin Dashboard", GroupKey: "dashboard", ParentKey: &menuDashboard, NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/dashboard"), IsActive: true, SortOrder: 2},
		{ID: uuid.New(), Key: "menu_roles", Name: "Roles", GroupKey: "roles", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/role-management"), IsActive: true, SortOrder: 3},
		{ID: uuid.New(), Key: "view_roles", Name: "View Roles", GroupKey: "roles", ParentKey: &menuRoles, NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/role-management"), IsActive: true, SortOrder: 4},
		{ID: uuid.New(), Key: "update_role", Name: "Update Role", GroupKey: "roles", ParentKey: &menuRoles, NodeType: model.PermissionNodeAction, IsActive: true, SortOrder: 5},
		{ID: uuid.New(), Key: "menu_sales_dashboard", Name: "Sales Dashboard", GroupKey: "dashboard", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/sales/dashboard"), IsActive: true, SortOrder: 6},
		{ID: uuid.New(), Key: "view_sales_dashboard", Name: "View Sales Dashboard", GroupKey: "dashboard", ParentKey: &menuSales, NodeType: model.PermissionNodeAction, RoutePath: strPtr("/sales/dashboard"), IsActive: true, SortOrder: 7},
	}
}

func permissionKeys(perms []model.Permission) map[string]bool {
	set := map[string]bool{}
	for _, permission := range perms {
		set[permission.Key] = true
	}
	return set
}

func TestPermissionCatalogAndRolePermissionContracts(t *testing.T) {
	repo, _, _, _, _ := salesTestRepo()
	svc := New(repo)
	items, err := svc.ListPermissions(context.Background(), adminActor(), "")
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	if len(items) < 7 || items[0].Key != "menu_admin_dashboard" {
		t.Fatalf("catalog order/items=%+v", items)
	}
	if _, err := svc.ListPermissions(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}, ""); !errors.Is(err, ErrForbidden) {
		t.Fatalf("sales catalog err=%v", err)
	}
}

func TestRolePermissionsCreateLandingAndAncestors(t *testing.T) {
	repo, _, _, _, _ := salesTestRepo()
	svc := New(repo)
	role, err := svc.CreateSalesRole(context.Background(), adminActor(), model.CreateSalesRoleInput{Name: "Ops", Level: 2, LandingPage: strPtr("/admin/dashboard"), PermissionKeys: []string{"view_admin_dashboard"}})
	if err != nil {
		t.Fatalf("create valid landing: %v", err)
	}
	keys := permissionKeys(role.Permissions)
	if !keys["view_admin_dashboard"] || !keys["menu_admin_dashboard"] {
		t.Fatalf("ancestor normalization failed: %+v", role.Permissions)
	}
	if _, err := svc.CreateSalesRole(context.Background(), adminActor(), model.CreateSalesRoleInput{Name: "BadPerm", Level: 2, PermissionKeys: []string{"missing_key"}}); !errors.Is(err, ErrPermissionNotFound) {
		t.Fatalf("invalid permission err=%v", err)
	}
	if _, err := svc.CreateSalesRole(context.Background(), adminActor(), model.CreateSalesRoleInput{Name: "BadLanding", Level: 2, LandingPage: strPtr("/admin/dashboard"), PermissionKeys: []string{"menu_roles"}}); !errors.Is(err, ErrLandingPagePermissionNeeded) {
		t.Fatalf("landing permission err=%v", err)
	}
}

func TestRolePermissionUpdateSemantics(t *testing.T) {
	repo, _, l2, _, _ := salesTestRepo()
	repo.roles[l2] = model.SalesRole{ID: l2, Name: "L2", Level: 2, IsActive: true, Permissions: []model.Permission{{Key: "view_roles"}}}
	svc := New(repo)
	updated, err := svc.UpdateSalesRole(context.Background(), adminActor(), l2, model.UpdateSalesRoleInput{Description: strPtr("kept")})
	if err != nil {
		t.Fatalf("update preserve: %v", err)
	}
	if len(updated.Permissions) != 1 || updated.Permissions[0].Key != "view_roles" {
		t.Fatalf("permissions not preserved: %+v", updated.Permissions)
	}
	updated, err = svc.UpdateSalesRole(context.Background(), adminActor(), l2, model.UpdateSalesRoleInput{PermissionKeys: []string{"view_admin_dashboard"}})
	if err != nil {
		t.Fatalf("replace permissions: %v", err)
	}
	keys := permissionKeys(updated.Permissions)
	if !keys["view_admin_dashboard"] || !keys["menu_admin_dashboard"] || keys["view_roles"] {
		t.Fatalf("replace keys=%v", keys)
	}
	updated, err = svc.UpdateSalesRole(context.Background(), adminActor(), l2, model.UpdateSalesRoleInput{PermissionKeys: []string{}})
	if err != nil {
		t.Fatalf("clear permissions: %v", err)
	}
	if len(updated.Permissions) != 0 {
		t.Fatalf("permissions not cleared: %+v", updated.Permissions)
	}
}
