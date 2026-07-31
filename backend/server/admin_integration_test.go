package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"crm-prospect-simulator/backend/config"
	adminmodel "crm-prospect-simulator/backend/internal/admin/model"
	adminrepo "crm-prospect-simulator/backend/internal/admin/repository"
	adminservice "crm-prospect-simulator/backend/internal/admin/service"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	authrepo "crm-prospect-simulator/backend/internal/auth/repository"
	authservice 	"crm-prospect-simulator/backend/internal/auth/service"
	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type stubUserRepo struct {
	user authmodel.User
}

func (r *stubUserRepo) FindByEmail(_ context.Context, _ string) (authmodel.User, error) {
	return r.user, nil
}
func (r *stubUserRepo) FindUserByID(_ context.Context, _ uuid.UUID) (authmodel.User, error) {
	return r.user, nil
}
func (r *stubUserRepo) RecordLogin(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return nil
}
func (r *stubUserRepo) UpsertSeed(_ context.Context, _ authmodel.User) error {
	return nil
}

type stubSessionRepo struct{}

func (r *stubSessionRepo) Create(_ context.Context, _ authmodel.RefreshSession) error {
	return nil
}
func (r *stubSessionRepo) FindSessionByID(_ context.Context, _ uuid.UUID) (authmodel.RefreshSession, error) {
	return authmodel.RefreshSession{}, authrepo.ErrNotFound
}
func (r *stubSessionRepo) Rotate(_ context.Context, _ uuid.UUID, _ authmodel.RefreshSession, _ time.Time) error {
	return nil
}
func (r *stubSessionRepo) Revoke(_ context.Context, _ uuid.UUID, _ string, _ time.Time) error {
	return nil
}
func (r *stubSessionRepo) RevokeAllForUser(_ context.Context, _ uuid.UUID, _ string, _ time.Time) error {
	return nil
}

type adminRepoStub struct{}

func (r *adminRepoStub) ListUsers(_ context.Context, _ adminmodel.ListFilter) (adminmodel.UserListResult, error) {
	return adminmodel.UserListResult{Items: []adminmodel.UserListItem{}, Total: 0, Page: 1, Limit: 10, Pages: 0}, nil
}
func (r *adminRepoStub) FindUserDetail(_ context.Context, _ uuid.UUID) (adminmodel.UserDetail, error) {
	return adminmodel.UserDetail{}, adminrepo.ErrNotFound
}
func (r *adminRepoStub) CreateUser(_ context.Context, _ uuid.UUID, _ adminmodel.CreateUserInput, _ string, _ uuid.UUID) error {
	return nil
}
func (r *adminRepoStub) UpdateUser(_ context.Context, _ uuid.UUID, _ adminmodel.UpdateUserInput, _ uuid.UUID) error {
	return nil
}
func (r *adminRepoStub) UpdateStatus(_ context.Context, _ uuid.UUID, _ authmodel.UserStatus, _ uuid.UUID) error {
	return adminrepo.ErrNotFound
}
func (r *adminRepoStub) ListActiveManagers(_ context.Context) ([]adminmodel.ManagerOption, error) {
	return nil, nil
}
func (r *adminRepoStub) ExistsByEmail(_ context.Context, _ string, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *adminRepoStub) ExistsByEmployeeID(_ context.Context, _ string, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *adminRepoStub) FindManagerByID(_ context.Context, _ uuid.UUID) (authmodel.User, error) {
	return authmodel.User{}, adminrepo.ErrNotFound
}
func (r *adminRepoStub) CountActiveAdministrators(_ context.Context) (int, error) {
	return 1, nil
}
func (r *adminRepoStub) FindUserByID(_ context.Context, _ uuid.UUID) (authmodel.User, error) {
	return authmodel.User{}, adminrepo.ErrNotFound
}
func (r *adminRepoStub) ResetPassword(_ context.Context, _, _ uuid.UUID, _ string) (int64, error) {
	return 0, nil
}

func buildTestApp(user authmodel.User) (*fiber.App, string) {
	return buildTestAppWithAdmin(user, adminservice.New(&adminRepoStub{}))
}

func buildTestAppWithAdmin(user authmodel.User, adminSvc *adminservice.Service) (*fiber.App, string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	user.PasswordHash = string(hash)
	if user.Status == "" {
		user.Status = authmodel.UserActive
	}

	users := &stubUserRepo{user: user}
	sessions := &stubSessionRepo{}
	jwtSecret := "01234567890123456789012345678901"
	tokens := authservice.NewTokenManager(jwtSecret, "test", "test-api", time.Minute)
	authSvc := authservice.NewAuthService(users, sessions, tokens, time.Hour)
	app := New(config.Config{AllowedOrigins: "http://localhost:5173"}, authSvc, nil, nil, adminSvc)
	access, _, _ := tokens.IssueAccess(user, uuid.New(), time.Now())
	return app, access
}

func TestAdminListUsersNeedsAuthentication(t *testing.T) {
	app, _ := buildTestApp(authmodel.User{ID: uuid.New(), Email: "admin@test.test", Role: authmodel.RoleAdministrator, TokenVersion: 1})
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("unauthenticated status=%d, want 401", resp.StatusCode)
	}
}

func TestAdminListUsersAdminGets200(t *testing.T) {
	app, token := buildTestApp(authmodel.User{
		ID: uuid.New(), Email: "admin@test.test", FullName: "Admin",
		Role: authmodel.RoleAdministrator, Status: authmodel.UserActive, TokenVersion: 1,
	})
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("admin status=%d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "passwordHash") || strings.Contains(string(body), "tokenVersion") {
		t.Fatal("response contains passwordHash or tokenVersion")
	}
}

func TestManagerGets403OnAdminUsers(t *testing.T) {
	app, token := buildTestApp(authmodel.User{
		ID: uuid.New(), Email: "manager@test.test", FullName: "Manager",
		Role: authmodel.RoleSalesManager, Status: authmodel.UserActive, TokenVersion: 1,
	})
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("manager status=%d, want 403", resp.StatusCode)
	}
}

func TestSalesExecutiveGets403OnAdminUsers(t *testing.T) {
	app, token := buildTestApp(authmodel.User{
		ID: uuid.New(), Email: "sales@test.test", FullName: "Sales",
		Role: authmodel.RoleSalesExecutive, Status: authmodel.UserActive, TokenVersion: 1,
	})
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("sales status=%d, want 403", resp.StatusCode)
	}
}

func TestAdminGetManagersReturnsJSONArray(t *testing.T) {
	app, token := buildTestApp(authmodel.User{
		ID: uuid.New(), Email: "admin@test.test", FullName: "Admin",
		Role: authmodel.RoleAdministrator, Status: authmodel.UserActive, TokenVersion: 1,
	})
	req := httptest.NewRequest("GET", "/api/v1/admin/users/options/managers", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("managers status=%d, want 200", resp.StatusCode)
	}
	var respBody map[string]any
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respBody)
	if _, ok := respBody["data"]; !ok {
		t.Fatal("managers response missing data key")
	}
}

type patchAdminRepo struct {
	current  authmodel.User
	detail   adminmodel.UserDetail
	captured *adminmodel.UpdateUserInput
}

func (r *patchAdminRepo) ListUsers(_ context.Context, _ adminmodel.ListFilter) (adminmodel.UserListResult, error) {
	return adminmodel.UserListResult{Items: []adminmodel.UserListItem{}, Total: 0, Page: 1, Limit: 10, Pages: 0}, nil
}
func (r *patchAdminRepo) FindUserDetail(_ context.Context, _ uuid.UUID) (adminmodel.UserDetail, error) {
	return r.detail, nil
}
func (r *patchAdminRepo) CreateUser(_ context.Context, _ uuid.UUID, _ adminmodel.CreateUserInput, _ string, _ uuid.UUID) error {
	return nil
}
func (r *patchAdminRepo) UpdateUser(_ context.Context, _ uuid.UUID, input adminmodel.UpdateUserInput, _ uuid.UUID) error {
	r.captured = &input
	return nil
}
func (r *patchAdminRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ authmodel.UserStatus, _ uuid.UUID) error {
	return adminrepo.ErrNotFound
}
func (r *patchAdminRepo) ListActiveManagers(_ context.Context) ([]adminmodel.ManagerOption, error) {
	return nil, nil
}
func (r *patchAdminRepo) ExistsByEmail(_ context.Context, _ string, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *patchAdminRepo) ExistsByEmployeeID(_ context.Context, _ string, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (r *patchAdminRepo) FindManagerByID(_ context.Context, _ uuid.UUID) (authmodel.User, error) {
	return authmodel.User{}, adminrepo.ErrNotFound
}
func (r *patchAdminRepo) CountActiveAdministrators(_ context.Context) (int, error) {
	return 1, nil
}
func (r *patchAdminRepo) FindUserByID(_ context.Context, _ uuid.UUID) (authmodel.User, error) {
	return r.current, nil
}
func (r *patchAdminRepo) ResetPassword(_ context.Context, _, _ uuid.UUID, _ string) (int64, error) {
	return 0, nil
}

type resetAdminRepo struct {
	adminRepoStub
	revoked  int64
	resetErr error
}

func (r *resetAdminRepo) ResetPassword(_ context.Context, _, _ uuid.UUID, _ string) (int64, error) {
	return r.revoked, r.resetErr
}

func TestAdminUpdatePromotesSEToSalesManagerAndClearsManager(t *testing.T) {
	managerA := uuid.New()
	current := authmodel.User{
		ID: uuid.New(), Email: "se@test.test", FullName: "SE",
		Role: authmodel.RoleSalesExecutive, Status: authmodel.UserActive, ManagerID: &managerA, TokenVersion: 1,
	}
	repo := &patchAdminRepo{
		current: current,
		detail:  adminmodel.UserDetail{ID: current.ID, FullName: "SE", Role: authmodel.RoleSalesManager},
	}
	admin := authmodel.User{ID: uuid.New(), Email: "admin@test.test", FullName: "Admin", Role: authmodel.RoleAdministrator, Status: authmodel.UserActive, TokenVersion: 1}
	app, token := buildTestAppWithAdmin(admin, adminservice.New(repo))

	req := httptest.NewRequest("PATCH", "/api/v1/admin/users/"+current.ID.String(), strings.NewReader(`{"role":"SALES_MANAGER","managerId":null}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d, want 200, body=%s", resp.StatusCode, body)
	}
	if repo.captured == nil {
		t.Fatal("repo UpdateUser not called")
	}
	if !repo.captured.ManagerID.Present || repo.captured.ManagerID.Value != nil {
		t.Fatalf("expected manager cleared (present with nil value), got %+v", repo.captured.ManagerID)
	}
	if repo.captured.Role == nil || string(*repo.captured.Role) != "SALES_MANAGER" {
		t.Fatalf("expected role SALES_MANAGER, got %v", repo.captured.Role)
	}
}

func TestAdminUpdateInvalidManagerUUIDIsNot500(t *testing.T) {
	admin := authmodel.User{ID: uuid.New(), Email: "admin@test.test", FullName: "Admin", Role: authmodel.RoleAdministrator, Status: authmodel.UserActive, TokenVersion: 1}
	repo := &patchAdminRepo{}
	app, token := buildTestAppWithAdmin(admin, adminservice.New(repo))

	req := httptest.NewRequest("PATCH", "/api/v1/admin/users/"+uuid.New().String(), strings.NewReader(`{"managerId":"not-a-uuid"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode < 400 || resp.StatusCode >= 500 {
		t.Fatalf("invalid manager UUID status=%d, want 4xx (not 500)", resp.StatusCode)
	}
}

func adminResetRequest(target string, body string) *http.Request {
	req := httptest.NewRequest("POST", "/api/v1/admin/users/"+target+"/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func adminUser() authmodel.User {
	return authmodel.User{ID: uuid.New(), Email: "admin@test.test", FullName: "Admin", Role: authmodel.RoleAdministrator, Status: authmodel.UserActive, TokenVersion: 1}
}

func TestResetPasswordNeedsAuthentication(t *testing.T) {
	app, _ := buildTestApp(adminUser())
	resp, err := app.Test(adminResetRequest(uuid.New().String(), `{"mode":"AUTO"}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("unauthenticated status=%d, want 401", resp.StatusCode)
	}
}

func TestResetPasswordForbiddenForNonAdmins(t *testing.T) {
	for _, role := range []authmodel.Role{authmodel.RoleSalesManager, authmodel.RoleSalesExecutive} {
		app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "u@test.test", FullName: "U", Role: role, Status: authmodel.UserActive, TokenVersion: 1})
		req := adminResetRequest(uuid.New().String(), `{"mode":"AUTO"}`)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 403 {
			t.Fatalf("%s reset status=%d, want 403", role, resp.StatusCode)
		}
	}
}

func TestResetPasswordInvalidUUID400(t *testing.T) {
	app, token := buildTestApp(adminUser())
	req := adminResetRequest("not-a-uuid", `{"mode":"AUTO"}`)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("invalid UUID status=%d, want 400", resp.StatusCode)
	}
}

func TestResetPasswordMalformedJSON400(t *testing.T) {
	app, token := buildTestApp(adminUser())
	req := adminResetRequest(uuid.New().String(), `{"mode":`)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("malformed body status=%d, want 400", resp.StatusCode)
	}
}

func TestResetPasswordInvalidMode422(t *testing.T) {
	app, token := buildTestApp(adminUser())
	req := adminResetRequest(uuid.New().String(), `{"mode":"RANDOM"}`)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 422 {
		t.Fatalf("invalid mode status=%d, want 422", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "INVALID_RESET_MODE") {
		t.Fatalf("expected INVALID_RESET_MODE error code, body=%s", body)
	}
}

func TestResetPasswordManualMissingPassword422(t *testing.T) {
	app, token := buildTestApp(adminUser())
	req := adminResetRequest(uuid.New().String(), `{"mode":"MANUAL"}`)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 422 {
		t.Fatalf("missing password status=%d, want 422", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "TEMPORARY_PASSWORD_REQUIRED") {
		t.Fatalf("expected TEMPORARY_PASSWORD_REQUIRED error code, body=%s", body)
	}
}

func TestResetPasswordWeakPassword422(t *testing.T) {
	app, token := buildTestApp(adminUser())
	req := adminResetRequest(uuid.New().String(), `{"mode":"MANUAL","temporaryPassword":"password"}`)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 422 {
		t.Fatalf("weak password status=%d, want 422", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "WEAK_TEMPORARY_PASSWORD") {
		t.Fatalf("expected WEAK_TEMPORARY_PASSWORD error code, body=%s", body)
	}
}

func TestResetPasswordUnknownTarget404(t *testing.T) {
	repo := &resetAdminRepo{resetErr: adminrepo.ErrNotFound}
	app, token := buildTestAppWithAdmin(adminUser(), adminservice.New(repo))
	req := adminResetRequest(uuid.New().String(), `{"mode":"AUTO"}`)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("unknown target status=%d, want 404", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "USER_NOT_FOUND") {
		t.Fatalf("expected USER_NOT_FOUND error code, body=%s", body)
	}
}

func TestResetPasswordAutoSuccess(t *testing.T) {
	repo := &resetAdminRepo{revoked: 2}
	app, token := buildTestAppWithAdmin(adminUser(), adminservice.New(repo))
	target := uuid.New()

	req := adminResetRequest(target.String(), `{"mode":"AUTO"}`)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("auto reset status=%d, want 200, body=%s", resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)
	raw := string(body)
	for _, leaked := range []string{"passwordHash", "tokenVersion", "refreshToken", "refresh_token", "accessToken"} {
		if strings.Contains(raw, leaked) {
			t.Fatalf("response leaked %q: %s", leaked, raw)
		}
	}

	var respBody map[string]any
	if err := json.Unmarshal(body, &respBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data, ok := respBody["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %v", respBody)
	}
	if data["userId"] != target.String() {
		t.Fatalf("userId=%v, want %s", data["userId"], target)
	}
	if pw, _ := data["temporaryPassword"].(string); pw == "" {
		t.Fatal("temporaryPassword must be non-empty")
	}
	if mustChange, _ := data["mustChangePassword"].(bool); !mustChange {
		t.Fatal("mustChangePassword must be true")
	}
	if revoked, _ := data["sessionsRevoked"].(float64); revoked != 2 {
		t.Fatalf("sessionsRevoked=%v, want 2", data["sessionsRevoked"])
	}
}

func TestResetPasswordManualSuccess(t *testing.T) {
	repo := &resetAdminRepo{revoked: 1}
	app, token := buildTestAppWithAdmin(adminUser(), adminservice.New(repo))
	target := uuid.New()

	req := adminResetRequest(target.String(), `{"mode":"MANUAL","temporaryPassword":"TempPass123!"}`)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("manual reset status=%d, want 200, body=%s", resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)
	var respBody map[string]any
	if err := json.Unmarshal(body, &respBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data, ok := respBody["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %v", respBody)
	}
	if pw, _ := data["temporaryPassword"].(string); pw != "TempPass123!" {
		t.Fatalf("temporaryPassword=%q, want the exact submitted password", pw)
	}
	if revoked, _ := data["sessionsRevoked"].(float64); revoked != 1 {
		t.Fatalf("sessionsRevoked=%v, want 1", data["sessionsRevoked"])
	}
}

type tokenVersionRepo struct {
	stubUserRepo
	userRef *authmodel.User
}

func (r *tokenVersionRepo) FindUserByID(_ context.Context, _ uuid.UUID) (authmodel.User, error) {
	return *r.userRef, nil
}

func TestBumpedTokenVersionInvalidatesOldAccessToken(t *testing.T) {
	user := adminUser()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	user.PasswordHash = string(hash)

	users := &tokenVersionRepo{userRef: &user}
	sessions := &stubSessionRepo{}
	jwtSecret := "01234567890123456789012345678901"
	tokens := authservice.NewTokenManager(jwtSecret, "test", "test-api", time.Minute)
	authSvc := authservice.NewAuthService(users, sessions, tokens, time.Hour)
	adminSvc := adminservice.New(&adminRepoStub{})
	app := New(config.Config{AllowedOrigins: "http://localhost:5173"}, authSvc, nil, nil, adminSvc)
	access, _, _ := tokens.IssueAccess(user, uuid.New(), time.Now())

	user.TokenVersion = 2

	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("access token issued before version bump status=%d, want 401", resp.StatusCode)
	}
}
