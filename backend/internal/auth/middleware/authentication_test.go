package middleware

import (
	"testing"

	"crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/auth/service"
)

func TestPermissionPolicyMatchesSystemAndSalesRoles(t *testing.T) {
	const permission = "view_ai_summary"
	if !hasPermission(service.Principal{Role: model.RoleSuperAdmin}, permission) {
		t.Fatal("SUPER_ADMIN must remain unrestricted without a sales role")
	}
	if hasPermission(service.Principal{Role: model.RoleAdministrator}, permission) {
		t.Fatal("ADMINISTRATOR without an explicit sales role permission must be denied")
	}
	withPermission := service.Principal{Role: model.RoleSalesExecutive, SalesRole: &model.SalesRoleSummary{PermissionKeys: []string{permission}}}
	if !hasPermission(withPermission, permission) {
		t.Fatal("assigned sales permission must be honored")
	}
	withoutPermission := service.Principal{Role: model.RoleSalesExecutive, SalesRole: &model.SalesRoleSummary{PermissionKeys: []string{}}}
	if hasPermission(withoutPermission, permission) {
		t.Fatal("missing sales permission must be denied")
	}
}
