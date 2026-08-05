package server

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"crm-prospect-simulator/backend/config"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
)

func TestHealthRoutesReturnJSON(t *testing.T) {
	app := New(config.Config{AllowedOrigins: "http://localhost:5173"}, nil, nil, nil, nil)
	for _, path := range []string{"/api/health", "/api/v1/health"} {
		response, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		body, _ := io.ReadAll(response.Body)
		if response.StatusCode != 200 || !strings.Contains(response.Header.Get("Content-Type"), "application/json") {
			t.Fatalf("%s returned status=%d content-type=%q", path, response.StatusCode, response.Header.Get("Content-Type"))
		}
		if strings.Contains(strings.ToLower(string(body)), "<html") {
			t.Fatalf("%s returned HTML", path)
		}
	}
}

func TestUnknownAPIRouteReturnsJSON404(t *testing.T) {
	app := New(config.Config{AllowedOrigins: "http://localhost:5173"}, nil, nil, nil, nil)
	response, err := app.Test(httptest.NewRequest("GET", "/api/unknown", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 404 || !strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		t.Fatalf("status=%d content-type=%q", response.StatusCode, response.Header.Get("Content-Type"))
	}
}

func TestChangePasswordNeedsAuthentication(t *testing.T) {
	app, _ := buildTestApp(authmodel.User{ID: uuid.New(), Email: "admin@test.test", Role: authmodel.RoleAdministrator, TokenVersion: 1})
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", strings.NewReader(`{"currentPassword":"password","newPassword":"NewPass456","confirmPassword":"NewPass456"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("status=%d, want 401", resp.StatusCode)
	}
}

func TestChangePasswordAllowedForEveryRole(t *testing.T) {
	for _, role := range []authmodel.Role{authmodel.RoleAdministrator, authmodel.RoleSalesManager, authmodel.RoleSalesExecutive} {
		app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "user@test.test", Role: role, Status: authmodel.UserActive, TokenVersion: 1})
		req := httptest.NewRequest("POST", "/api/v1/auth/change-password", strings.NewReader(`{"currentPassword":"password","newPassword":"NewPass456","confirmPassword":"NewPass456"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("role=%s status=%d body=%s, want 200", role, resp.StatusCode, string(body))
		}
		var body struct {
			Data struct {
				PasswordChanged          bool `json:"passwordChanged"`
				MustChangePassword       bool `json:"mustChangePassword"`
				ReauthenticationRequired bool `json:"reauthenticationRequired"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if !body.Data.PasswordChanged || body.Data.MustChangePassword || !body.Data.ReauthenticationRequired {
			t.Fatalf("role=%s unexpected body=%+v", role, body)
		}
	}
}
func TestProtectedRouteAllowsNormalAuthenticatedUser(t *testing.T) {
	app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "admin@test.test", Role: authmodel.RoleAdministrator, Status: authmodel.UserActive, TokenVersion: 1})
	req := httptest.NewRequest("GET", "/api/v1/dashboard/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d, want 200, body=%s", resp.StatusCode, body)
	}
}

func TestAdminRouteAllowsSuperAdmin(t *testing.T) {
	app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "super@test.test", Role: authmodel.RoleSuperAdmin, Status: authmodel.UserActive, TokenVersion: 1})
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d, want 200 for SUPER_ADMIN on admin route, body=%s", resp.StatusCode, body)
	}
}

func TestProtectedRoutePassesThroughWhenPasswordChangeRequired(t *testing.T) {
	app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "admin@test.test", Role: authmodel.RoleAdministrator, Status: authmodel.UserActive, TokenVersion: 1, MustChangePassword: true})
	req := httptest.NewRequest("GET", "/api/v1/dashboard/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || strings.Contains(string(body), "PASSWORD_CHANGE_REQUIRED") {
		t.Fatalf("status=%d body=%s, want 200 without PASSWORD_CHANGE_REQUIRED", resp.StatusCode, body)
	}
}

func TestForcedChangeUserCanCallChangePassword(t *testing.T) {
	app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "sales@test.test", Role: authmodel.RoleSalesExecutive, Status: authmodel.UserActive, TokenVersion: 1, MustChangePassword: true})
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", strings.NewReader(`{"currentPassword":"password","newPassword":"NewPass456","confirmPassword":"NewPass456"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d, want handler success 200, body=%s", resp.StatusCode, body)
	}
}

func TestForcedChangeUserCanCallLogoutAndMe(t *testing.T) {
	app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "sales@test.test", Role: authmodel.RoleSalesExecutive, Status: authmodel.UserActive, TokenVersion: 1, MustChangePassword: true})

	logoutReq := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	logoutResp, err := app.Test(logoutReq)
	if err != nil {
		t.Fatal(err)
	}
	if logoutResp.StatusCode != 204 {
		t.Fatalf("logout status=%d, want 204", logoutResp.StatusCode)
	}

	meReq := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meResp, err := app.Test(meReq)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(meResp.Body)
	if meResp.StatusCode != 200 || !strings.Contains(string(body), "mustChangePassword") {
		t.Fatalf("me status=%d body=%s, want 200 with mustChangePassword", meResp.StatusCode, body)
	}
}

func TestForcedChangeRolesUseNormalRoleMiddleware(t *testing.T) {
	cases := []struct {
		name       string
		role       authmodel.Role
		path       string
		wantStatus int
	}{
		{"administrator", authmodel.RoleAdministrator, "/api/v1/admin/users", 200},
		{"sales-manager", authmodel.RoleSalesManager, "/api/v1/dashboard/admin", 403},
		{"sales-executive", authmodel.RoleSalesExecutive, "/api/v1/dashboard/sales", 200},
	}
	for _, tc := range cases {
		app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: tc.name + "@test.test", Role: tc.role, Status: authmodel.UserActive, TokenVersion: 1, MustChangePassword: true})
		req := httptest.NewRequest("GET", tc.path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != tc.wantStatus || strings.Contains(string(body), "PASSWORD_CHANGE_REQUIRED") {
			t.Fatalf("%s status=%d body=%s, want %d without PASSWORD_CHANGE_REQUIRED", tc.name, resp.StatusCode, body, tc.wantStatus)
		}
	}
}

func TestRoleMiddlewareUnchangedWhenPasswordChangeNotRequired(t *testing.T) {
	app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "manager@test.test", Role: authmodel.RoleSalesManager, Status: authmodel.UserActive, TokenVersion: 1})
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d, want role forbidden 403, body=%s", resp.StatusCode, body)
	}
}

func TestForcedChangeSalesDashboardPassesThrough(t *testing.T) {
	app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "sales@test.test", Role: authmodel.RoleSalesExecutive, Status: authmodel.UserActive, TokenVersion: 1, MustChangePassword: true})
	req := httptest.NewRequest("GET", "/api/v1/dashboard/sales", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || strings.Contains(string(body), "PASSWORD_CHANGE_REQUIRED") {
		t.Fatalf("status=%d body=%s, want 200 without PASSWORD_CHANGE_REQUIRED", resp.StatusCode, body)
	}
}
func TestForcedChangeUserCanCallLogoutAll(t *testing.T) {
	app, token := buildTestApp(authmodel.User{ID: uuid.New(), Email: "sales@test.test", Role: authmodel.RoleSalesExecutive, Status: authmodel.UserActive, TokenVersion: 1, MustChangePassword: true})
	req := httptest.NewRequest("POST", "/api/v1/auth/logout-all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d, want 204, body=%s", resp.StatusCode, body)
	}
}
