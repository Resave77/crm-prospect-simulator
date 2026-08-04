package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"crm-prospect-simulator/backend/internal/admin/model"
	"github.com/google/uuid"
)

var (
	ErrPermissionNotFound          = errors.New("permission not found")
	ErrInvalidPermissionKey        = errors.New("invalid permission key")
	ErrInvalidLandingPage          = errors.New("invalid landing page")
	ErrLandingPagePermissionNeeded = errors.New("landing page permission required")
	ErrRolePermissionUpdateFailed  = errors.New("role permission update failed")
)

func (s *Service) ListPermissions(ctx context.Context, actor Actor, search string) ([]model.Permission, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repo.ListPermissions(ctx, strings.TrimSpace(search))
}

func (s *Service) normalizePermissionKeys(ctx context.Context, keys []string) ([]string, map[string]model.Permission, error) {
	seen := map[string]bool{}
	cleaned := []string{}
	for _, raw := range keys {
		key := strings.TrimSpace(raw)
		if key == "" || key != strings.ToLower(key) || strings.Contains(key, " ") {
			return nil, nil, fmt.Errorf("%w: %s", ErrInvalidPermissionKey, raw)
		}
		if !seen[key] {
			seen[key] = true
			cleaned = append(cleaned, key)
		}
	}
	if len(cleaned) == 0 {
		return []string{}, map[string]model.Permission{}, nil
	}
	perms, err := s.repo.FindPermissionsByKeys(ctx, cleaned)
	if err != nil {
		return nil, nil, err
	}
	byKey := map[string]model.Permission{}
	for _, permission := range perms {
		byKey[permission.Key] = permission
	}
	for _, key := range cleaned {
		if _, ok := byKey[key]; !ok {
			return nil, nil, fmt.Errorf("%w: %s", ErrPermissionNotFound, key)
		}
	}
	for _, key := range cleaned {
		permission := byKey[key]
		for permission.ParentKey != nil {
			parentKey := *permission.ParentKey
			if !seen[parentKey] {
				parent, err := s.repo.FindPermissionByKey(ctx, parentKey)
				if err != nil {
					return nil, nil, err
				}
				seen[parentKey] = true
				byKey[parentKey] = parent
			}
			parent := byKey[parentKey]
			permission = parent
		}
	}
	normalized := []string{}
	for _, permission := range permissionCatalog {
		if seen[permission.Key] {
			normalized = append(normalized, permission.Key)
		}
	}
	for key := range seen {
		found := false
		for _, normalizedKey := range normalized {
			if normalizedKey == key {
				found = true
				break
			}
		}
		if !found {
			normalized = append(normalized, key)
		}
	}
	return normalized, byKey, nil
}

func (s *Service) validateLandingPage(landingPage *string, permissionKeys map[string]bool) error {
	if landingPage == nil || strings.TrimSpace(*landingPage) == "" {
		return nil
	}
	page := strings.TrimSpace(*landingPage)
	required, ok := landingPagePermissions[page]
	if !ok {
		return fmt.Errorf("%w: %s", ErrInvalidLandingPage, page)
	}
	if !permissionKeys[required] {
		return fmt.Errorf("%w: %s requires %s", ErrLandingPagePermissionNeeded, page, required)
	}
	*landingPage = page
	return nil
}

func keysSet(keys []string) map[string]bool {
	set := map[string]bool{}
	for _, key := range keys {
		set[key] = true
	}
	return set
}

func permissionsSet(perms []model.Permission) map[string]bool {
	set := map[string]bool{}
	for _, permission := range perms {
		set[permission.Key] = true
	}
	return set
}

func permissionKeysFromPermissions(perms []model.Permission) []string {
	keys := make([]string, 0, len(perms))
	for _, permission := range perms {
		keys = append(keys, permission.Key)
	}
	return keys
}

func (s *Service) validateRolePermissionsForCreate(ctx context.Context, input *model.CreateSalesRoleInput) ([]string, error) {
	keys, _, err := s.normalizePermissionKeys(ctx, input.PermissionKeys)
	if err != nil {
		return nil, err
	}
	if err := s.validateLandingPage(input.LandingPage, keysSet(keys)); err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *Service) validateRolePermissionsForUpdate(ctx context.Context, current model.SalesRole, input *model.UpdateSalesRoleInput) ([]string, bool, error) {
	if input.PermissionKeys != nil {
		keys, _, err := s.normalizePermissionKeys(ctx, input.PermissionKeys)
		if err != nil {
			return nil, true, err
		}
		landingPage := current.LandingPage
		if input.LandingPage != nil {
			landingPage = input.LandingPage
		}
		if err := s.validateLandingPage(landingPage, keysSet(keys)); err != nil {
			return nil, true, err
		}
		return keys, true, nil
	}
	if input.LandingPage != nil {
		currentPermissions, err := s.repo.ListRolePermissions(ctx, current.ID)
		if err != nil {
			return nil, false, err
		}
		if err := s.validateLandingPage(input.LandingPage, permissionsSet(currentPermissions)); err != nil {
			return nil, false, err
		}
	}
	return nil, false, nil
}

func (s *Service) roleWithPermissions(ctx context.Context, role model.SalesRole) (model.SalesRole, error) {
	permissions, err := s.repo.ListRolePermissions(ctx, role.ID)
	if err != nil {
		return model.SalesRole{}, err
	}
	role.Permissions = permissions
	role.PermissionCount = len(permissions)
	return role, nil
}

func defaultPermissionKeys(level int) []string {
	switch level {
	case 1:
		keys := make([]string, 0, len(permissionCatalog))
		for _, permission := range permissionCatalog {
			keys = append(keys, permission.Key)
		}
		return keys
	case 2:
		return []string{"menu_sales_dashboard", "view_sales_dashboard", "menu_sales_structure", "view_sales_structure", "menu_my_prospects", "view_my_prospects", "menu_my_customers", "view_my_customers", "menu_sales_history", "view_sales_history", "menu_profile", "view_own_profile", "change_own_password"}
	case 3:
		return []string{"menu_sales_dashboard", "view_sales_dashboard", "menu_my_prospects", "view_my_prospects", "view_my_prospect_detail", "menu_my_customers", "view_my_customers", "view_my_customer_detail", "menu_sales_history", "view_sales_history", "menu_profile", "view_own_profile", "change_own_password"}
	case 4:
		return []string{"menu_sales_dashboard", "view_sales_dashboard", "menu_my_prospects", "view_my_prospects", "view_my_prospect_detail", "check_in_prospect", "check_out_prospect", "update_visit_result", "menu_my_customers", "view_my_customers", "view_my_customer_detail", "check_in_customer", "check_out_customer", "menu_sales_history", "view_sales_history", "menu_profile", "view_own_profile", "change_own_password"}
	default:
		return []string{}
	}
}

func defaultLandingPage(level int) *string {
	switch level {
	case 1:
		return strPtr("/admin/dashboard")
	case 2, 3, 4:
		return strPtr("/sales/dashboard")
	default:
		return nil
	}
}

func defaultSalesRolePermissionSeed() map[uuid.UUID][]string {
	return map[uuid.UUID][]string{
		uuid.MustParse("00000000-0000-0000-0000-000000000101"): defaultPermissionKeys(1),
		uuid.MustParse("00000000-0000-0000-0000-000000000102"): defaultPermissionKeys(2),
		uuid.MustParse("00000000-0000-0000-0000-000000000103"): defaultPermissionKeys(3),
		uuid.MustParse("00000000-0000-0000-0000-000000000104"): defaultPermissionKeys(4),
	}
}
