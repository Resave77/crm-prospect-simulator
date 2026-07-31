package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"crm-prospect-simulator/backend/internal/admin/model"
	"crm-prospect-simulator/backend/internal/admin/repository"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type repoStub struct {
	users            map[uuid.UUID]authmodel.User
	details          map[uuid.UUID]model.UserDetail
	emails           map[string]uuid.UUID
	employeeIDs      map[string]uuid.UUID
	activeAdminCount int
	managers         []model.ManagerOption
	listResult       model.UserListResult
	lastUpdate       *model.UpdateUserInput
	createErr        error
	updateErr        error
	statusErr        error
	lastReset        *resetCapture
	revokedCount     int64
	resetErr         error
}

type resetCapture struct {
	targetID     uuid.UUID
	actorID      uuid.UUID
	passwordHash string
}

func (r *repoStub) ResetPassword(_ context.Context, targetID, actorID uuid.UUID, passwordHash string) (int64, error) {
	r.lastReset = &resetCapture{targetID: targetID, actorID: actorID, passwordHash: passwordHash}
	if r.resetErr != nil {
		return 0, r.resetErr
	}
	if _, ok := r.users[targetID]; !ok {
		return 0, repository.ErrNotFound
	}
	return r.revokedCount, nil
}

func (r *repoStub) ListUsers(_ context.Context, _ model.ListFilter) (model.UserListResult, error) {
	return r.listResult, nil
}

func (r *repoStub) FindUserDetail(_ context.Context, id uuid.UUID) (model.UserDetail, error) {
	if d, ok := r.details[id]; ok {
		return d, nil
	}
	return model.UserDetail{}, repository.ErrNotFound
}

func (r *repoStub) CreateUser(_ context.Context, id uuid.UUID, input model.CreateUserInput, _ string, _ uuid.UUID) error {
	if r.createErr != nil {
		return r.createErr
	}
	if r.details != nil {
		r.details[id] = model.UserDetail{
			ID: id, EmployeeID: input.EmployeeID, FullName: input.FullName,
			Email: input.Email, Phone: input.Phone, Role: input.Role,
			Status: authmodel.UserActive, ManagerID: input.ManagerID,
			MustChangePassword: true,
		}
	}
	return nil
}

func (r *repoStub) UpdateUser(_ context.Context, _ uuid.UUID, input model.UpdateUserInput, _ uuid.UUID) error {
	r.lastUpdate = &input
	return r.updateErr
}

func (r *repoStub) UpdateStatus(_ context.Context, _ uuid.UUID, _ authmodel.UserStatus, _ uuid.UUID) error {
	return r.statusErr
}

func (r *repoStub) ListActiveManagers(_ context.Context) ([]model.ManagerOption, error) {
	return r.managers, nil
}

func (r *repoStub) ExistsByEmail(_ context.Context, email string, excludeID *uuid.UUID) (bool, error) {
	_, exists := r.emails[email]
	if exists && excludeID != nil {
		if owner, ok := r.emails[email]; ok && owner == *excludeID {
			exists = false
		}
	}
	return exists, nil
}

func (r *repoStub) ExistsByEmployeeID(_ context.Context, employeeID string, excludeID *uuid.UUID) (bool, error) {
	_, exists := r.employeeIDs[employeeID]
	if exists && excludeID != nil {
		if owner, ok := r.employeeIDs[employeeID]; ok && owner == *excludeID {
			exists = false
		}
	}
	return exists, nil
}

func (r *repoStub) FindManagerByID(_ context.Context, id uuid.UUID) (authmodel.User, error) {
	if m, ok := r.users[id]; ok {
		return m, nil
	}
	return authmodel.User{}, repository.ErrNotFound
}

func (r *repoStub) CountActiveAdministrators(_ context.Context) (int, error) {
	return r.activeAdminCount, nil
}

func (r *repoStub) FindUserByID(_ context.Context, id uuid.UUID) (authmodel.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return authmodel.User{}, repository.ErrNotFound
}

func newTestService(stub *repoStub) *Service {
	return New(stub)
}

func adminActor() Actor {
	return Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}
}

func managerActor() Actor {
	return Actor{UserID: uuid.New(), Role: authmodel.RoleSalesManager}
}

func salesActor() Actor {
	return Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}
}

func managerUser(id uuid.UUID, status authmodel.UserStatus) authmodel.User {
	return authmodel.User{
		ID: id, Email: "manager@yummy.test", FullName: "Budi Santoso",
		Role: authmodel.RoleSalesManager, Status: status,
	}
}

func TestAdminCanListUsers(t *testing.T) {
	stub := &repoStub{listResult: model.UserListResult{
		Items: []model.UserListItem{{ID: uuid.New(), FullName: "A"}},
		Total: 1, Page: 1, Limit: 10, Pages: 1,
	}}
	svc := newTestService(stub)
	result, err := svc.ListUsers(context.Background(), adminActor(), model.ListFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected list result: %+v", result)
	}
}

func TestManagerAndSalesGetForbidden(t *testing.T) {
	stub := &repoStub{}
	svc := newTestService(stub)
	if _, err := svc.ListUsers(context.Background(), managerActor(), model.ListFilter{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("manager err=%v, want ErrForbidden", err)
	}
	if _, err := svc.ListUsers(context.Background(), salesActor(), model.ListFilter{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("sales err=%v, want ErrForbidden", err)
	}
	if _, err := svc.CreateUser(context.Background(), salesActor(), model.CreateUserInput{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("sales create err=%v, want ErrForbidden", err)
	}
}

func TestCreateSalesManagerWithoutManagerSucceeds(t *testing.T) {
	stub := &repoStub{
		details: map[uuid.UUID]model.UserDetail{},
		emails:  map[string]uuid.UUID{}, employeeIDs: map[string]uuid.UUID{},
	}
	svc := newTestService(stub)
	input := model.CreateUserInput{
		EmployeeID: "SM-0002", FullName: "Sari", Email: "sari@yummy.test",
		Role: authmodel.RoleSalesManager, TemporaryPassword: "Password123",
	}
	_, err := svc.CreateUser(context.Background(), adminActor(), input)
	if err != nil {
		t.Fatalf("create sales manager: %v", err)
	}
}

func TestCreateSalesExecutiveWithActiveManagerSucceeds(t *testing.T) {
	managerID := uuid.New()
	stub := &repoStub{
		details: map[uuid.UUID]model.UserDetail{},
		emails:  map[string]uuid.UUID{}, employeeIDs: map[string]uuid.UUID{},
		users: map[uuid.UUID]authmodel.User{
			managerID: managerUser(managerID, authmodel.UserActive),
		},
	}
	svc := newTestService(stub)
	input := model.CreateUserInput{
		EmployeeID: "SE-0010", FullName: "Rina", Email: "rina@yummy.test",
		Role: authmodel.RoleSalesExecutive, ManagerID: &managerID,
		TemporaryPassword: "Password123",
	}
	if _, err := svc.CreateUser(context.Background(), adminActor(), input); err != nil {
		t.Fatalf("create sales executive: %v", err)
	}
}

func TestCreateSalesExecutiveWithoutManagerRejected(t *testing.T) {
	stub := &repoStub{emails: map[string]uuid.UUID{}, employeeIDs: map[string]uuid.UUID{}}
	svc := newTestService(stub)
	input := model.CreateUserInput{
		EmployeeID: "SE-0011", FullName: "Rina", Email: "rina2@yummy.test",
		Role: authmodel.RoleSalesExecutive, TemporaryPassword: "Password123",
	}
	_, err := svc.CreateUser(context.Background(), adminActor(), input)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err=%v, want ErrValidation", err)
	}
}

func TestManagerWithRoleNotSalesManagerRejected(t *testing.T) {
	managerID := uuid.New()
	stub := &repoStub{
		emails: map[string]uuid.UUID{}, employeeIDs: map[string]uuid.UUID{},
		users: map[uuid.UUID]authmodel.User{
			managerID: {ID: managerID, Email: "admin2@yummy.test", FullName: "Admin", Role: authmodel.RoleAdministrator, Status: authmodel.UserActive},
		},
	}
	svc := newTestService(stub)
	input := model.CreateUserInput{
		EmployeeID: "SE-0012", FullName: "Rina", Email: "rina3@yummy.test",
		Role: authmodel.RoleSalesExecutive, ManagerID: &managerID,
		TemporaryPassword: "Password123",
	}
	_, err := svc.CreateUser(context.Background(), adminActor(), input)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err=%v, want ErrValidation", err)
	}
}

func TestInactiveManagerRejected(t *testing.T) {
	managerID := uuid.New()
	stub := &repoStub{
		emails: map[string]uuid.UUID{}, employeeIDs: map[string]uuid.UUID{},
		users: map[uuid.UUID]authmodel.User{
			managerID: managerUser(managerID, authmodel.UserInactive),
		},
	}
	svc := newTestService(stub)
	input := model.CreateUserInput{
		EmployeeID: "SE-0013", FullName: "Rina", Email: "rina4@yummy.test",
		Role: authmodel.RoleSalesExecutive, ManagerID: &managerID,
		TemporaryPassword: "Password123",
	}
	_, err := svc.CreateUser(context.Background(), adminActor(), input)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err=%v, want ErrValidation", err)
	}
}

func TestDuplicateEmailRejected(t *testing.T) {
	stub := &repoStub{
		emails: map[string]uuid.UUID{"duplicate@yummy.test": uuid.New()},
	}
	svc := newTestService(stub)
	input := model.CreateUserInput{
		EmployeeID: "SE-0014", FullName: "Rina", Email: "duplicate@yummy.test",
		Role: authmodel.RoleSalesExecutive, ManagerID: &uuid.UUID{},
		TemporaryPassword: "Password123",
	}
	_, err := svc.CreateUser(context.Background(), adminActor(), input)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err=%v, want ErrValidation", err)
	}
}

func TestDuplicateEmployeeIDRejected(t *testing.T) {
	stub := &repoStub{
		emails:      map[string]uuid.UUID{},
		employeeIDs: map[string]uuid.UUID{"SE-0001": uuid.New()},
	}
	svc := newTestService(stub)
	input := model.CreateUserInput{
		EmployeeID: "SE-0001", FullName: "Rina", Email: "rina5@yummy.test",
		Role: authmodel.RoleSalesExecutive, ManagerID: &uuid.UUID{},
		TemporaryPassword: "Password123",
	}
	_, err := svc.CreateUser(context.Background(), adminActor(), input)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err=%v, want ErrValidation", err)
	}
}

func mustUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return id
}

func salesExecutive(id uuid.UUID, managerID *uuid.UUID) authmodel.User {
	return authmodel.User{
		ID: id, Email: "se@yummy.test", FullName: "SE",
		Role: authmodel.RoleSalesExecutive, Status: authmodel.UserActive, ManagerID: managerID,
	}
}

func updateTestStub(target uuid.UUID, current authmodel.User, managerIDs ...uuid.UUID) *repoStub {
	users := map[uuid.UUID]authmodel.User{target: current}
	for _, m := range managerIDs {
		users[m] = managerUser(m, authmodel.UserActive)
	}
	return &repoStub{
		emails:      map[string]uuid.UUID{},
		employeeIDs: map[string]uuid.UUID{},
		users:       users,
		details:     map[uuid.UUID]model.UserDetail{target: {ID: target}},
	}
}

func TestUpdateOmittedManagerIDPreservesManagerForSE(t *testing.T) {
	managerA := mustUUID(t, "00000000-0000-0000-0000-00000000000a")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000aa")
	stub := updateTestStub(target, salesExecutive(target, &managerA), managerA)
	svc := newTestService(stub)

	name := "Updated"
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, model.UpdateUserInput{FullName: &name}); err != nil {
		t.Fatalf("update without managerId: %v", err)
	}
	if stub.lastUpdate == nil {
		t.Fatal("repo update not called")
	}
	if stub.lastUpdate.ManagerID.Present {
		t.Fatalf("managerId must be untouched, got %+v", stub.lastUpdate.ManagerID)
	}
}

func TestUpdateManagerIDReplacesManagerForSE(t *testing.T) {
	managerA := mustUUID(t, "00000000-0000-0000-0000-00000000000a")
	managerB := mustUUID(t, "00000000-0000-0000-0000-00000000000b")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000aa")
	stub := updateTestStub(target, salesExecutive(target, &managerA), managerA, managerB)
	svc := newTestService(stub)

	input := model.UpdateUserInput{ManagerID: model.OptionalUUID{Present: true, Value: &managerB}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); err != nil {
		t.Fatalf("replace manager: %v", err)
	}
	if stub.lastUpdate == nil || stub.lastUpdate.ManagerID.Value == nil || *stub.lastUpdate.ManagerID.Value != managerB {
		t.Fatalf("expected manager B, got %+v", stub.lastUpdate.ManagerID)
	}
}

func TestUpdatePromoteSEToSalesManagerClearsManager(t *testing.T) {
	managerA := mustUUID(t, "00000000-0000-0000-0000-00000000000a")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000aa")
	stub := updateTestStub(target, salesExecutive(target, &managerA), managerA)
	svc := newTestService(stub)

	role := authmodel.RoleSalesManager
	input := model.UpdateUserInput{Role: &role, ManagerID: model.OptionalUUID{Present: true, Value: nil}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); err != nil {
		t.Fatalf("SE->SM with null manager: %v", err)
	}
	if stub.lastUpdate == nil || !stub.lastUpdate.ManagerID.Present || stub.lastUpdate.ManagerID.Value != nil {
		t.Fatalf("expected manager cleared, got %+v", stub.lastUpdate.ManagerID)
	}
}

func TestUpdatePromoteSEToAdministratorClearsManager(t *testing.T) {
	managerA := mustUUID(t, "00000000-0000-0000-0000-00000000000a")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000aa")
	stub := updateTestStub(target, salesExecutive(target, &managerA), managerA)
	svc := newTestService(stub)

	role := authmodel.RoleAdministrator
	input := model.UpdateUserInput{Role: &role, ManagerID: model.OptionalUUID{Present: true, Value: nil}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); err != nil {
		t.Fatalf("SE->ADMIN with null manager: %v", err)
	}
	if stub.lastUpdate == nil || !stub.lastUpdate.ManagerID.Present || stub.lastUpdate.ManagerID.Value != nil {
		t.Fatalf("expected manager cleared, got %+v", stub.lastUpdate.ManagerID)
	}
}

func TestUpdatePromoteSEToSalesManagerOmittedManagerAutoclears(t *testing.T) {
	managerA := mustUUID(t, "00000000-0000-0000-0000-00000000000a")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000aa")
	stub := updateTestStub(target, salesExecutive(target, &managerA), managerA)
	svc := newTestService(stub)

	role := authmodel.RoleSalesManager
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, model.UpdateUserInput{Role: &role}); err != nil {
		t.Fatalf("SE->SM without managerId: %v", err)
	}
	if stub.lastUpdate == nil || !stub.lastUpdate.ManagerID.Present || stub.lastUpdate.ManagerID.Value != nil {
		t.Fatalf("expected manager auto-cleared, got %+v", stub.lastUpdate.ManagerID)
	}
}

func TestUpdatePromoteSEToSalesManagerWithManagerRejected(t *testing.T) {
	managerA := mustUUID(t, "00000000-0000-0000-0000-00000000000a")
	managerB := mustUUID(t, "00000000-0000-0000-0000-00000000000b")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000aa")
	stub := updateTestStub(target, salesExecutive(target, &managerA), managerA, managerB)
	svc := newTestService(stub)

	role := authmodel.RoleSalesManager
	input := model.UpdateUserInput{Role: &role, ManagerID: model.OptionalUUID{Present: true, Value: &managerB}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); !errors.Is(err, ErrValidation) {
		t.Fatalf("SE->SM with manager err=%v, want ErrValidation", err)
	}
}

func TestUpdateSalesManagerWithNonNilManagerRejected(t *testing.T) {
	managerB := mustUUID(t, "00000000-0000-0000-0000-00000000000b")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000bb")
	stub := updateTestStub(target, authmodel.User{
		ID: target, Email: "sm@yummy.test", FullName: "SM",
		Role: authmodel.RoleSalesManager, Status: authmodel.UserActive,
	}, managerB)
	svc := newTestService(stub)

	input := model.UpdateUserInput{ManagerID: model.OptionalUUID{Present: true, Value: &managerB}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); !errors.Is(err, ErrValidation) {
		t.Fatalf("SM with manager err=%v, want ErrValidation", err)
	}
}

func TestUpdateAdministratorWithNonNilManagerRejected(t *testing.T) {
	managerB := mustUUID(t, "00000000-0000-0000-0000-00000000000b")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000cc")
	stub := updateTestStub(target, authmodel.User{
		ID: target, Email: "admin2@yummy.test", FullName: "Admin",
		Role: authmodel.RoleAdministrator, Status: authmodel.UserActive,
	}, managerB)
	svc := newTestService(stub)

	input := model.UpdateUserInput{ManagerID: model.OptionalUUID{Present: true, Value: &managerB}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); !errors.Is(err, ErrValidation) {
		t.Fatalf("ADMIN with manager err=%v, want ErrValidation", err)
	}
}

func TestUpdateSMToSEWithoutManagerRejected(t *testing.T) {
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000bb")
	stub := updateTestStub(target, authmodel.User{
		ID: target, Email: "sm@yummy.test", FullName: "SM",
		Role: authmodel.RoleSalesManager, Status: authmodel.UserActive,
	})
	svc := newTestService(stub)

	role := authmodel.RoleSalesExecutive
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, model.UpdateUserInput{Role: &role}); !errors.Is(err, ErrValidation) {
		t.Fatalf("SM->SE without manager err=%v, want ErrValidation", err)
	}
}

func TestUpdateSMToSEWithValidActiveManagerSucceeds(t *testing.T) {
	managerB := mustUUID(t, "00000000-0000-0000-0000-00000000000b")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000bb")
	stub := updateTestStub(target, authmodel.User{
		ID: target, Email: "sm@yummy.test", FullName: "SM",
		Role: authmodel.RoleSalesManager, Status: authmodel.UserActive,
	}, managerB)
	svc := newTestService(stub)

	role := authmodel.RoleSalesExecutive
	input := model.UpdateUserInput{Role: &role, ManagerID: model.OptionalUUID{Present: true, Value: &managerB}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); err != nil {
		t.Fatalf("SM->SE with manager: %v", err)
	}
	if stub.lastUpdate == nil || stub.lastUpdate.ManagerID.Value == nil || *stub.lastUpdate.ManagerID.Value != managerB {
		t.Fatalf("expected manager B, got %+v", stub.lastUpdate.ManagerID)
	}
}

func TestUpdateSEExplicitNullManagerRejected(t *testing.T) {
	managerA := mustUUID(t, "00000000-0000-0000-0000-00000000000a")
	target := mustUUID(t, "00000000-0000-0000-0000-0000000000aa")
	stub := updateTestStub(target, salesExecutive(target, &managerA), managerA)
	svc := newTestService(stub)

	input := model.UpdateUserInput{ManagerID: model.OptionalUUID{Present: true, Value: nil}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); !errors.Is(err, ErrValidation) {
		t.Fatalf("SE with null manager err=%v, want ErrValidation", err)
	}
}

func TestUpdateInactiveManagerRejectedForSE(t *testing.T) {
	managerID := uuid.New()
	target := uuid.New()
	stub := &repoStub{
		emails:      map[string]uuid.UUID{},
		employeeIDs: map[string]uuid.UUID{},
		users: map[uuid.UUID]authmodel.User{
			target:    {ID: target, Email: "se@yummy.test", FullName: "SE", Role: authmodel.RoleSalesExecutive, Status: authmodel.UserActive},
			managerID: managerUser(managerID, authmodel.UserInactive),
		},
		details: map[uuid.UUID]model.UserDetail{target: {ID: target}},
	}
	svc := newTestService(stub)

	input := model.UpdateUserInput{ManagerID: model.OptionalUUID{Present: true, Value: &managerID}}
	if _, err := svc.UpdateUser(context.Background(), adminActor(), target, input); !errors.Is(err, ErrValidation) {
		t.Fatalf("inactive manager err=%v, want ErrValidation", err)
	}
}

func TestAdminCannotDeactivateSelf(t *testing.T) {
	adminID := uuid.New()
	stub := &repoStub{users: map[uuid.UUID]authmodel.User{
		adminID: {ID: adminID, Role: authmodel.RoleAdministrator, Status: authmodel.UserActive},
	}}
	svc := newTestService(stub)
	_, err := svc.UpdateStatus(context.Background(), Actor{UserID: adminID, Role: authmodel.RoleAdministrator}, adminID, authmodel.UserInactive)
	if !errors.Is(err, ErrSelfDeactivate) {
		t.Fatalf("err=%v, want ErrSelfDeactivate", err)
	}
}

func TestLastActiveAdminCannotBeDeactivated(t *testing.T) {
	adminID := uuid.New()
	actor := adminActor()
	stub := &repoStub{
		activeAdminCount: 1,
		users: map[uuid.UUID]authmodel.User{
			adminID: {ID: adminID, Role: authmodel.RoleAdministrator, Status: authmodel.UserActive},
		},
	}
	svc := newTestService(stub)
	if _, err := svc.UpdateStatus(context.Background(), actor, adminID, authmodel.UserInactive); !errors.Is(err, ErrLastAdmin) {
		t.Fatalf("err=%v, want ErrLastAdmin", err)
	}
}

func TestNullableProfileFieldsPassThrough(t *testing.T) {
	managerID := uuid.New()
	stub := &repoStub{
		details: map[uuid.UUID]model.UserDetail{},
		emails:  map[string]uuid.UUID{}, employeeIDs: map[string]uuid.UUID{},
		users: map[uuid.UUID]authmodel.User{
			managerID: managerUser(managerID, authmodel.UserActive),
		},
	}
	svc := newTestService(stub)
	input := model.CreateUserInput{
		FullName: "NoID", Email: "noid@yummy.test",
		Role: authmodel.RoleSalesExecutive, ManagerID: &managerID,
		TemporaryPassword: "Password123",
	}
	if _, err := svc.CreateUser(context.Background(), adminActor(), input); err != nil {
		t.Fatalf("create with nullable employee_id/phone: %v", err)
	}
}

func resetTargetStub(target uuid.UUID, role authmodel.Role, status authmodel.UserStatus) *repoStub {
	return &repoStub{
		users: map[uuid.UUID]authmodel.User{
			target: {ID: target, Email: "target@yummy.test", FullName: "Target", Role: role, Status: status},
		},
	}
}

func TestGenerateTemporaryPasswordMeetsRules(t *testing.T) {
	for i := 0; i < 25; i++ {
		pw, err := generateTemporaryPassword()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if utf8.RuneCountInString(pw) < 12 {
			t.Fatalf("generated password too short: %q", pw)
		}
		var groups [4]bool
		for _, r := range pw {
			switch {
			case unicode.IsUpper(r):
				groups[0] = true
			case unicode.IsLower(r):
				groups[1] = true
			case unicode.IsDigit(r):
				groups[2] = true
			case strings.ContainsRune("!@#$%&*", r):
				groups[3] = true
			}
		}
		for group, ok := range groups {
			if !ok {
				t.Fatalf("generated password missing group %d: %q", group, pw)
			}
		}
	}
}

func TestResetAutoGeneratesAndHashes(t *testing.T) {
	target := uuid.New()
	stub := resetTargetStub(target, authmodel.RoleSalesExecutive, authmodel.UserActive)
	svc := newTestService(stub)

	result, err := svc.ResetPassword(context.Background(), adminActor(), target, model.ResetPasswordInput{Mode: model.ResetPasswordModeAuto})
	if err != nil {
		t.Fatalf("auto reset: %v", err)
	}
	if stub.lastReset == nil {
		t.Fatal("repository ResetPassword not called")
	}
	if stub.lastReset.passwordHash == result.TemporaryPassword {
		t.Fatal("hash must not equal the plain temporary password")
	}
	if bcrypt.CompareHashAndPassword([]byte(stub.lastReset.passwordHash), []byte(result.TemporaryPassword)) != nil {
		t.Fatal("hash passed to repository does not match the returned temporary password")
	}
}

func TestResetManualValidPasswordHashedExactlyAsSubmitted(t *testing.T) {
	target := uuid.New()
	stub := resetTargetStub(target, authmodel.RoleSalesExecutive, authmodel.UserActive)
	svc := newTestService(stub)
	const pw = "TempPass123!"

	result, err := svc.ResetPassword(context.Background(), adminActor(), target, model.ResetPasswordInput{
		Mode: model.ResetPasswordModeManual, TemporaryPassword: pw,
	})
	if err != nil {
		t.Fatalf("manual reset: %v", err)
	}
	if result.TemporaryPassword != pw {
		t.Fatalf("returned password %q, want the exact submitted %q", result.TemporaryPassword, pw)
	}
	if stub.lastReset == nil || bcrypt.CompareHashAndPassword([]byte(stub.lastReset.passwordHash), []byte(pw)) != nil {
		t.Fatal("submitted password was altered before hashing")
	}
}

func TestResetManualMissingPasswordRejected(t *testing.T) {
	stub := &repoStub{}
	svc := newTestService(stub)
	_, err := svc.ResetPassword(context.Background(), adminActor(), uuid.New(), model.ResetPasswordInput{Mode: model.ResetPasswordModeManual})
	if !errors.Is(err, ErrTemporaryPasswordRequired) {
		t.Fatalf("err=%v, want ErrTemporaryPasswordRequired", err)
	}
}

func TestResetInvalidModeRejected(t *testing.T) {
	stub := &repoStub{}
	svc := newTestService(stub)
	for _, mode := range []model.ResetPasswordMode{"", "RANDOM", " "} {
		_, err := svc.ResetPassword(context.Background(), adminActor(), uuid.New(), model.ResetPasswordInput{Mode: mode})
		if !errors.Is(err, ErrInvalidResetMode) {
			t.Fatalf("mode %q err=%v, want ErrInvalidResetMode", mode, err)
		}
	}
	// mode is normalized to uppercase
	_, err := svc.ResetPassword(context.Background(), adminActor(), uuid.New(), model.ResetPasswordInput{Mode: "manual"})
	if !errors.Is(err, ErrTemporaryPasswordRequired) {
		t.Fatalf("lowercase manual mode err=%v, want normalization to MANUAL", err)
	}
}

func TestResetWeakPasswordsRejected(t *testing.T) {
	stub := &repoStub{}
	svc := newTestService(stub)
	weak := []string{
		"password",  // no uppercase, no digit
		"PASSWORD1", // no lowercase
		"Password",  // no digit
		"pass1234",  // no uppercase
		"        ",  // whitespace only
		"Pass1",     // too short
	}
	for _, pw := range weak {
		_, err := svc.ResetPassword(context.Background(), adminActor(), uuid.New(), model.ResetPasswordInput{
			Mode: model.ResetPasswordModeManual, TemporaryPassword: pw,
		})
		if !errors.Is(err, ErrWeakTemporaryPassword) {
			t.Fatalf("password %q err=%v, want ErrWeakTemporaryPassword", pw, err)
		}
	}
}

func TestResetNotFoundMapsFromRepository(t *testing.T) {
	stub := &repoStub{}
	svc := newTestService(stub)
	_, err := svc.ResetPassword(context.Background(), adminActor(), uuid.New(), model.ResetPasswordInput{Mode: model.ResetPasswordModeAuto})
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err=%v, want repository.ErrNotFound", err)
	}
}

func TestResetRepositoryErrorDoesNotLeakPassword(t *testing.T) {
	target := uuid.New()
	stub := resetTargetStub(target, authmodel.RoleSalesExecutive, authmodel.UserActive)
	stub.resetErr = errors.New("database operation failed")
	svc := newTestService(stub)

	const pw = "SuperSecret9"
	_, err := svc.ResetPassword(context.Background(), adminActor(), target, model.ResetPasswordInput{
		Mode: model.ResetPasswordModeManual, TemporaryPassword: pw,
	})
	if err == nil {
		t.Fatal("expected a repository error")
	}
	if strings.Contains(err.Error(), pw) {
		t.Fatal("repository error leaked the temporary password")
	}
}

func TestResetInactiveAccountAllowed(t *testing.T) {
	target := uuid.New()
	stub := resetTargetStub(target, authmodel.RoleSalesExecutive, authmodel.UserInactive)
	svc := newTestService(stub)

	result, err := svc.ResetPassword(context.Background(), adminActor(), target, model.ResetPasswordInput{Mode: model.ResetPasswordModeAuto})
	if err != nil {
		t.Fatalf("inactive reset: %v", err)
	}
	if result.TemporaryPassword == "" {
		t.Fatal("expected a temporary password for an INACTIVE account")
	}
}

func TestResetSelfByAdministratorAllowed(t *testing.T) {
	target := uuid.New()
	stub := resetTargetStub(target, authmodel.RoleAdministrator, authmodel.UserActive)
	svc := newTestService(stub)

	_, err := svc.ResetPassword(context.Background(), Actor{UserID: target, Role: authmodel.RoleAdministrator}, target, model.ResetPasswordInput{Mode: model.ResetPasswordModeAuto})
	if err != nil {
		t.Fatalf("self reset: %v", err)
	}
	if stub.lastReset == nil || stub.lastReset.actorID != target || stub.lastReset.targetID != target {
		t.Fatal("self reset should record actor and target as the same user")
	}
}

func TestResetResultShape(t *testing.T) {
	target := uuid.New()
	stub := resetTargetStub(target, authmodel.RoleSalesExecutive, authmodel.UserActive)
	stub.revokedCount = 2
	svc := newTestService(stub)

	result, err := svc.ResetPassword(context.Background(), adminActor(), target, model.ResetPasswordInput{Mode: model.ResetPasswordModeAuto})
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if result.UserID != target {
		t.Fatalf("userId=%v, want %v", result.UserID, target)
	}
	if result.TemporaryPassword == "" {
		t.Fatal("temporaryPassword must be non-empty")
	}
	if !result.MustChangePassword {
		t.Fatal("mustChangePassword must be true")
	}
	if result.SessionsRevoked != 2 {
		t.Fatalf("sessionsRevoked=%d, want 2", result.SessionsRevoked)
	}
}
