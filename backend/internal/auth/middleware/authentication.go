package middleware

import (
	"strings"

	"crm-prospect-simulator/backend/internal/auth/model"

	"crm-prospect-simulator/backend/internal/auth/service"

	"crm-prospect-simulator/backend/internal/shared/response"

	"github.com/gofiber/fiber/v2"
)

const principalKey = "auth.principal"

type Middleware struct {
	auth *service.AuthService
}

func New(auth *service.AuthService) *Middleware {

	return &Middleware{auth: auth}

}

func (m *Middleware) Authenticate(c *fiber.Ctx) error {

	header := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))

	parts := strings.Fields(header)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {

		return response.Error(c, fiber.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication is required.")

	}

	principal, err := m.auth.AuthenticateAccess(c.UserContext(), parts[1])

	if err != nil {

		return response.Error(c, fiber.StatusUnauthorized, "ACCESS_TOKEN_INVALID", "The access token is invalid or expired.")

	}

	c.Locals(principalKey, principal)

	return c.Next()

}

func (m *Middleware) RequireRole(roles ...model.Role) fiber.Handler {

	allowed := make(map[model.Role]bool, len(roles))

	for _, role := range roles {

		allowed[role] = true

	}

	return func(c *fiber.Ctx) error {

		principal, ok := Principal(c)

		if !ok || !roleAllowed(allowed, principal.Role) {

			return response.Error(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "You do not have permission to perform this action.")

		}

		return c.Next()

	}

}

func (m *Middleware) RequirePermission(permissionKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		principal, ok := Principal(c)
		if !ok || !hasPermission(principal, permissionKey) {
			return response.Error(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "You do not have permission to perform this action.")
		}

		return c.Next()
	}
}

func roleAllowed(allowed map[model.Role]bool, actual model.Role) bool {

	if allowed[actual] {

		return true

	}

	return actual == model.RoleSuperAdmin && allowed[model.RoleAdministrator]

}

func hasPermission(principal service.Principal, permissionKey string) bool {
	if permissionKey == "" {
		return true
	}
	// SUPER_ADMIN is the system-wide authority and is not assigned a sales role.
	// Keep this aligned with frontend navigation.hasPermission.
	if principal.Role == model.RoleSuperAdmin {
		return true
	}
	if principal.Role == model.RoleSalesExecutive && principal.SalesRole == nil {
		return true
	}
	if principal.SalesRole == nil {
		return false
	}
	for _, key := range principal.SalesRole.PermissionKeys {
		if key == permissionKey {
			return true
		}
	}
	return false
}

func (m *Middleware) RequirePasswordChanged(c *fiber.Ctx) error {
	principal, ok := Principal(c)
	if !ok {

		return response.Error(c, fiber.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication is required.")

	}
	if principal.MustChangePassword {
		return response.Error(c, fiber.StatusForbidden, "PASSWORD_CHANGE_REQUIRED", "You must change your password before continuing.")
	}

	return c.Next()

}

func Principal(c *fiber.Ctx) (service.Principal, bool) {

	principal, ok := c.Locals(principalKey).(service.Principal)

	return principal, ok

}
