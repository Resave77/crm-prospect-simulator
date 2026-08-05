package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"crm-prospect-simulator/backend/config"
	"crm-prospect-simulator/backend/platform/database"
)

const (
	userAdminID  = "c67edd75-23fe-4a91-b3b6-4c1181cc9d8d"
	userDeleteID = "097e9036-0920-4f4a-a5fb-a01012e5fc70"
	userLevel1ID = "1391e74f-f2a4-405f-b924-49bf0617bb5d"
	userLevel2ID = "e540034d-e1ca-444d-990e-6a55a734e852"
	userLevel3ID = "113de61e-99f4-41d0-bf9c-db76333783be"
	userSalesID  = "e83fc480-f228-4f80-bca6-55f9b42736ca"
	userSales2ID = "f0fc9903-cc99-43e5-b08f-7cbc1cb9402f"
	userSales3ID = "4ef6f759-c0e9-45ea-996e-8f07c38ab201"
	userSandyID  = "3f01df1b-950d-43f1-bce2-d8cc658d2101"

	roleLevel1ID = "00000000-0000-0000-0000-000000000101"
	roleLevel2ID = "00000000-0000-0000-0000-000000000102"
	roleLevel3ID = "00000000-0000-0000-0000-000000000103"
	roleLevel4ID = "00000000-0000-0000-0000-000000000104"

	assignmentLevel1RootID   = "945c73a2-cedb-42bf-88c2-9e78388967e5"
	assignmentLevel2ID       = "6d0b6630-f256-41bd-8a37-acbff3ab542f"
	assignmentLevel3ID       = "989db891-e70e-4f2a-a414-6b5f24a3a6c7"
	assignmentSalesLevel3ID  = "a57c54e4-c2f8-4d10-aa4a-a1fbd6e3324e"
	assignmentSales2Level3ID = "18209b51-ddb5-4990-b8f3-e7dc68998d51"
	assignmentSales3Level3ID = "f93b718f-d2ad-5189-b75a-b6224252750e"

	passwordHashAdmin  = "$2a$10$lwUpcBt35Icbgyur1LEnnuz20ZNdL0wKZjOq8.9TwSKgsBsPG8d5K"
	passwordHashDelete = "$2a$10$4cQnfxHhD9uS5jyl2gTa3e8sIYXOJoNlAp3hBsKRtiAyblCEBFyi."
	passwordHashLevel1 = "$2a$10$9Bp/SGHaxPBOxRrwUhO4tu9vBttkf66piBG3T9gBhX8X7vYer5ppK"
	passwordHashLevel2 = "$2a$10$GruyFaUHg7gC5F1Qop3v0uipXxQASfvTXgNoWtIE/s7jgkB1AIEPe"
	passwordHashLevel3 = "$2a$10$OPbMhtFWpd3wn3/UwpAIQ.917hlz4OgVxxNEFUm8PJozQ9uyTfNMO"
	passwordHashSales  = "$2a$10$Z8BxSU0x17PaG22C54KaXu5NEbl6foX4/WhgiscMIq6ybRbXYWql2"
	passwordHashSandy  = "$2a$10$hoRg4AHVKo0Bhn/vPm78QuA.KIfBGzFV/u1L76he2brBQpdgufAKy"
)

var levelOnePermissions = []string{
	"approve_prospect_deletion",
	"change_own_password",
	"check_in_customer",
	"check_in_prospect",
	"check_out_customer",
	"check_out_prospect",
	"convert_prospect",
	"create_account",
	"create_customer",
	"create_prospect",
	"create_role",
	"create_sales_assignment",
	"delete_customer",
	"delete_prospect",
	"delete_role",
	"delete_visit",
	"manage_prospect_comments",
	"manage_role_permissions",
	"menu_accounts",
	"menu_admin_dashboard",
	"menu_customers",
	"menu_my_customers",
	"menu_my_prospects",
	"menu_profile",
	"menu_prospect_finder",
	"menu_prospect_list",
	"menu_prospect_pipeline",
	"menu_reports",
	"menu_roles",
	"menu_sales_dashboard",
	"menu_sales_history",
	"menu_sales_pipeline",
	"menu_sales_structure",
	"menu_visit_monitoring",
	"move_sales_assignment",
	"reject_prospect_deletion",
	"request_prospect_deletion",
	"reset_account_password",
	"update_account",
	"update_account_status",
	"update_company",
	"update_customer",
	"update_prospect_pipeline",
	"update_role",
	"update_role_status",
	"update_visit_result",
	"view_account_detail",
	"view_accounts",
	"view_admin_dashboard",
	"view_company_detail",
	"view_customer_detail",
	"view_customers",
	"view_my_customer_detail",
	"view_my_customers",
	"view_my_prospect_detail",
	"view_my_prospects",
	"view_own_profile",
	"view_own_visits",
	"view_prospect_detail",
	"view_prospect_finder",
	"view_prospect_list",
	"view_prospect_pipeline",
	"view_reports",
	"view_role_detail",
	"view_roles",
	"view_sales_assignment_history",
	"view_sales_dashboard",
	"view_sales_history",
	"view_sales_structure",
	"view_visit_evidence",
	"view_visit_monitoring",
}

var levelTwoPermissions = []string{
	"change_own_password",
	"menu_my_customers",
	"menu_my_prospects",
	"menu_profile",
	"menu_sales_dashboard",
	"menu_sales_history",
	"menu_sales_structure",
	"view_my_customers",
	"view_my_prospects",
	"view_own_profile",
	"view_sales_dashboard",
	"view_sales_history",
	"view_sales_structure",
}

var levelThreePermissions = []string{
	"change_own_password",
	"menu_my_customers",
	"menu_my_prospects",
	"menu_profile",
	"menu_sales_dashboard",
	"menu_sales_history",
	"view_my_customer_detail",
	"view_my_customers",
	"view_my_prospect_detail",
	"view_my_prospects",
	"view_own_profile",
	"view_sales_dashboard",
	"view_sales_history",
}

var levelFourPermissions = []string{
	"change_own_password",
	"check_in_customer",
	"check_in_prospect",
	"check_out_customer",
	"check_out_prospect",
	"menu_my_customers",
	"menu_my_prospects",
	"menu_profile",
	"menu_sales_dashboard",
	"menu_sales_history",
	"update_visit_result",
	"view_my_customer_detail",
	"view_my_customers",
	"view_my_prospect_detail",
	"view_my_prospects",
	"view_own_profile",
	"view_sales_dashboard",
	"view_sales_history",
}

type baselineUser struct {
	ID                 uuid.UUID
	Email              string
	PasswordHash       string
	FullName           string
	EmployeeID         string
	Phone              string
	SystemRole         string
	Status             string
	MustChangePassword bool
	TokenVersion       int
	ManagerID          *uuid.UUID
	SalesRoleID        *uuid.UUID
}

type baselineRole struct {
	ID             uuid.UUID
	Name           string
	Level          int
	Description    string
	LandingPage    string
	PermissionKeys []string
	UpdatedBy      *uuid.UUID
}

type baselineAssignment struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	RoleID        uuid.UUID
	ParentID      *uuid.UUID
	EffectiveFrom string
	EffectiveTo   *string
	CreatedBy     *uuid.UUID
}

type resolvedUser struct {
	user    baselineUser
	actual  uuid.UUID
	existed bool
}

type resolvedRole struct {
	role    baselineRole
	actual  uuid.UUID
	existed bool
}

type counters struct {
	Created int
	Updated int
}

func uid(s string) uuid.UUID {
	return uuid.MustParse(s)
}

func optUUID(s string) *uuid.UUID {
	u := uid(s)
	return &u
}

func optString(s string) *string {
	return &s
}

func normalizeRoleName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func buildUsers() []baselineUser {
	return []baselineUser{
		{
			ID: uid(userAdminID), Email: "admin@yummy.test", PasswordHash: passwordHashAdmin,
			FullName: "Yummy Super Admin", EmployeeID: "ADM-0001", SystemRole: "SUPER_ADMIN",
			Status: "ACTIVE", TokenVersion: 2,
		},
		{
			ID: uid(userDeleteID), Email: "delete@yummy.test", PasswordHash: passwordHashDelete,
			FullName: "tes delete", EmployeeID: "EMP-2026-DEL-9732", SystemRole: "SUPER_ADMIN",
			Status: "ACTIVE", TokenVersion: 1,
		},
		{
			ID: uid(userLevel1ID), Email: "level1@yummy.test", PasswordHash: passwordHashLevel1,
			FullName: "Dudung Mujaer", EmployeeID: "EMP-2026-LEV-3075", SystemRole: "SALES_MANAGER",
			Status: "ACTIVE", TokenVersion: 1, SalesRoleID: optUUID(roleLevel1ID),
		},
		{
			ID: uid(userLevel2ID), Email: "level2@yummy.test", PasswordHash: passwordHashLevel2,
			FullName: "Dobleh Dongdong", EmployeeID: "EMP-2026-LEV-6765", SystemRole: "SALES_MANAGER",
			Status: "ACTIVE", TokenVersion: 1, SalesRoleID: optUUID(roleLevel2ID),
		},
		{
			ID: uid(userLevel3ID), Email: "level3@yummy.test", PasswordHash: passwordHashLevel3,
			FullName: "Jamal Kabur", EmployeeID: "EMP-2026-LEV-5264", SystemRole: "SALES_EXECUTIVE",
			Status: "ACTIVE", TokenVersion: 1, SalesRoleID: optUUID(roleLevel4ID),
		},
		{
			ID: uid(userSalesID), Email: "sales@yummy.test", PasswordHash: passwordHashSales,
			FullName: "Nurdin Pratama", EmployeeID: "SE-0001", SystemRole: "SALES_EXECUTIVE",
			Status: "ACTIVE", TokenVersion: 5, SalesRoleID: optUUID(roleLevel4ID),
		},
		{
			ID: uid(userSales2ID), Email: "sales2@yummy.test", PasswordHash: passwordHashAdmin,
			FullName: "Alicia Ramadhan", EmployeeID: "SE-0002", SystemRole: "SALES_EXECUTIVE",
			Status: "ACTIVE", TokenVersion: 1, SalesRoleID: optUUID(roleLevel4ID),
		},
		{
			ID: uid(userSales3ID), Email: "sales3@yummy.test", PasswordHash: passwordHashAdmin,
			FullName: "Rizky Ananda", EmployeeID: "SE-0003", SystemRole: "SALES_EXECUTIVE",
			Status: "ACTIVE", TokenVersion: 1,
		},
		{
			ID: uid(userSandyID), Email: "sandy@yummy.test", PasswordHash: passwordHashSandy,
			FullName: "Sandy Ramadhan", EmployeeID: "EMP-2026-SAN-8100", Phone: "083222114455", SystemRole: "SUPER_ADMIN",
			Status: "ACTIVE", TokenVersion: 3,
		},
	}
}

func buildRoles() []baselineRole {
	return []baselineRole{
		{
			ID: uid(roleLevel1ID), Name: "Sales Level 1", Level: 1,
			Description: "Default editable sales organization role", LandingPage: "/admin/dashboard",
			PermissionKeys: levelOnePermissions, UpdatedBy: optUUID(userAdminID),
		},
		{
			ID: uid(roleLevel2ID), Name: "Sales Level 2", Level: 2,
			Description: "Default editable sales organization role", LandingPage: "/sales/dashboard",
			PermissionKeys: levelTwoPermissions, UpdatedBy: optUUID(userAdminID),
		},
		{
			ID: uid(roleLevel3ID), Name: "Sales Regional Supervisor", Level: 3,
			Description: "Supervises Level 4 sales in one team", LandingPage: "/sales/dashboard",
			PermissionKeys: levelThreePermissions, UpdatedBy: optUUID(userAdminID),
		},
		{
			ID: uid(roleLevel4ID), Name: "Sales Level 4", Level: 4,
			Description: "Default editable sales organization role", LandingPage: "/sales/dashboard",
			PermissionKeys: levelFourPermissions, UpdatedBy: optUUID(userAdminID),
		},
	}
}

func buildAssignments() []baselineAssignment {
	return []baselineAssignment{
		{
			ID: uid(assignmentLevel1RootID), UserID: uid(userLevel1ID), RoleID: uid(roleLevel1ID),
			EffectiveFrom: "2026-08-01", EffectiveTo: optString("2026-08-31"),
			CreatedBy: optUUID(userAdminID),
		},
		{
			ID: uid(assignmentLevel2ID), UserID: uid(userLevel2ID), RoleID: uid(roleLevel2ID),
			ParentID: optUUID(userLevel1ID), EffectiveFrom: "2026-08-01", EffectiveTo: optString("2026-08-31"),
			CreatedBy: optUUID(userAdminID),
		},
		{
			ID: uid(assignmentLevel3ID), UserID: uid(userLevel3ID), RoleID: uid(roleLevel3ID),
			ParentID: optUUID(userLevel2ID), EffectiveFrom: "2026-08-01", EffectiveTo: optString("2026-08-31"),
			CreatedBy: optUUID(userAdminID),
		},
		{
			ID: uid(assignmentSalesLevel3ID), UserID: uid(userSalesID), RoleID: uid(roleLevel4ID),
			ParentID: optUUID(userLevel3ID), EffectiveFrom: "2026-08-01", EffectiveTo: optString("2026-08-31"),
			CreatedBy: optUUID(userAdminID),
		},
		{
			ID: uid(assignmentSales2Level3ID), UserID: uid(userSales2ID), RoleID: uid(roleLevel4ID),
			ParentID: optUUID(userLevel3ID), EffectiveFrom: "2026-08-01", EffectiveTo: optString("2026-08-31"),
			CreatedBy: optUUID(userAdminID),
		},
		{
			ID: uid(assignmentSales3Level3ID), UserID: uid(userSales3ID), RoleID: uid(roleLevel4ID),
			ParentID: optUUID(userLevel3ID), EffectiveFrom: "2026-08-01", EffectiveTo: optString("2026-08-31"),
			CreatedBy: optUUID(userAdminID),
		},
	}
}

func resolveUserIDByEmail(ctx context.Context, tx pgx.Tx, email string) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM users WHERE lower(btrim(email)) = lower(btrim($1)) ORDER BY created_at LIMIT 1`, email).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	return id, true, nil
}

func resolveRoleIDByName(ctx context.Context, tx pgx.Tx, normalized string) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM sales_roles WHERE lower(btrim(normalized_name)) = lower(btrim($1)) ORDER BY created_at LIMIT 1`, normalized).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	return id, true, nil
}

func resolveAssignmentID(ctx context.Context, tx pgx.Tx, baselineID uuid.UUID, userID, roleID uuid.UUID, parentID *uuid.UUID, from, to string) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM sales_structure_assignments WHERE id = $1`, baselineID).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, err
	}
	var parent any
	if parentID != nil {
		parent = *parentID
	}
	err = tx.QueryRow(ctx, `
		SELECT id FROM sales_structure_assignments
		WHERE user_id = $1 AND sales_role_id = $2 AND effective_from = $3::date
		  AND effective_to IS NOT DISTINCT FROM $4::date
		  AND parent_user_id IS NOT DISTINCT FROM $5
		ORDER BY created_at LIMIT 1`, userID, roleID, from, to, parent).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, err
	}
	return baselineID, false, nil
}

func seedUsers(ctx context.Context, tx pgx.Tx) (map[uuid.UUID]uuid.UUID, counters, error) {
	resolved := make([]resolvedUser, 0, len(buildUsers()))
	userIDMap := make(map[uuid.UUID]uuid.UUID, len(buildUsers()))

	for _, u := range buildUsers() {
		actualID, existed, err := resolveUserIDByEmail(ctx, tx, u.Email)
		if err != nil {
			return nil, counters{}, fmt.Errorf("resolve user %s: %w", u.Email, err)
		}
		if !existed {
			actualID = u.ID
		}
		userIDMap[u.ID] = actualID
		resolved = append(resolved, resolvedUser{user: u, actual: actualID, existed: existed})
	}

	var counts counters
	for _, r := range resolved {
		var managerID any
		if r.user.ManagerID != nil {
			managerID = userIDMap[*r.user.ManagerID]
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, employee_id, phone, role, status, must_change_password, token_version, manager_id, deleted_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7::"UserRole", $8::"UserStatus", $9, $10, $11, NULL, now())
			ON CONFLICT (id) DO UPDATE SET
				email = EXCLUDED.email,
				password_hash = EXCLUDED.password_hash,
				full_name = EXCLUDED.full_name,
				employee_id = EXCLUDED.employee_id,
				phone = EXCLUDED.phone,
				role = EXCLUDED.role,
				status = EXCLUDED.status,
				must_change_password = EXCLUDED.must_change_password,
				token_version = EXCLUDED.token_version,
				manager_id = EXCLUDED.manager_id,
				deleted_at = NULL,
				updated_at = now()`,
			r.actual, r.user.Email, r.user.PasswordHash, r.user.FullName, r.user.EmployeeID, r.user.Phone,
			r.user.SystemRole, r.user.Status, r.user.MustChangePassword, r.user.TokenVersion, managerID)
		if err != nil {
			return nil, counts, fmt.Errorf("upsert user %s: %w", r.user.Email, err)
		}
		if r.existed {
			counts.Updated++
		} else {
			counts.Created++
		}
	}
	return userIDMap, counts, nil
}

func seedRoles(ctx context.Context, tx pgx.Tx, userIDMap map[uuid.UUID]uuid.UUID) (map[uuid.UUID]uuid.UUID, counters, error) {
	resolved := make([]resolvedRole, 0, len(buildRoles()))
	roleIDMap := make(map[uuid.UUID]uuid.UUID, len(buildRoles()))

	for _, r := range buildRoles() {
		actualID, existed, err := resolveRoleIDByName(ctx, tx, normalizeRoleName(r.Name))
		if err != nil {
			return nil, counters{}, fmt.Errorf("resolve role %s: %w", r.Name, err)
		}
		if !existed {
			actualID = r.ID
		}
		roleIDMap[r.ID] = actualID
		resolved = append(resolved, resolvedRole{role: r, actual: actualID, existed: existed})
	}

	var counts counters
	for _, r := range resolved {
		var updatedBy any
		if r.role.UpdatedBy != nil {
			updatedBy = userIDMap[*r.role.UpdatedBy]
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO sales_roles (id, name, normalized_name, level, description, landing_page, is_active, created_by, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, true, NULL, $7, now())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				normalized_name = EXCLUDED.normalized_name,
				level = EXCLUDED.level,
				description = EXCLUDED.description,
				landing_page = EXCLUDED.landing_page,
				is_active = true,
				created_by = EXCLUDED.created_by,
				updated_by = EXCLUDED.updated_by,
				updated_at = now()`,
			r.actual, r.role.Name, normalizeRoleName(r.role.Name), r.role.Level, r.role.Description, r.role.LandingPage, updatedBy)
		if err != nil {
			return nil, counts, fmt.Errorf("upsert role %s: %w", r.role.Name, err)
		}
		if r.existed {
			counts.Updated++
		} else {
			counts.Created++
		}
	}
	return roleIDMap, counts, nil
}

func seedRolePermissions(ctx context.Context, tx pgx.Tx, roleIDMap map[uuid.UUID]uuid.UUID) (int, error) {
	total := 0
	for _, r := range buildRoles() {
		tag, err := tx.Exec(ctx, `
			INSERT INTO role_permissions (sales_role_id, permission_id)
			SELECT $1, id FROM permissions WHERE key = ANY($2) AND is_active = true
			ON CONFLICT DO NOTHING`, roleIDMap[r.ID], r.PermissionKeys)
		if err != nil {
			return total, fmt.Errorf("seed permissions for role %s: %w", r.Name, err)
		}
		total += int(tag.RowsAffected())
	}
	return total, nil
}

func seedUserSalesRoles(ctx context.Context, tx pgx.Tx, userIDMap, roleIDMap map[uuid.UUID]uuid.UUID) (int, error) {
	linked := 0
	for _, u := range buildUsers() {
		if u.SalesRoleID == nil {
			continue
		}
		tag, err := tx.Exec(ctx, `UPDATE users SET sales_role_id = $2, updated_at = now() WHERE id = $1`, userIDMap[u.ID], roleIDMap[*u.SalesRoleID])
		if err != nil {
			return linked, fmt.Errorf("link sales role for %s: %w", u.Email, err)
		}
		linked += int(tag.RowsAffected())
	}
	return linked, nil
}

func seedAssignments(ctx context.Context, tx pgx.Tx, userIDMap, roleIDMap map[uuid.UUID]uuid.UUID) (counters, error) {
	var counts counters
	for _, a := range buildAssignments() {
		resolvedUserID := userIDMap[a.UserID]
		resolvedRoleID := roleIDMap[a.RoleID]
		var resolvedParentID *uuid.UUID
		if a.ParentID != nil {
			resolvedParentID = ptr(userIDMap[*a.ParentID])
		}
		var resolvedCreatedBy any
		if a.CreatedBy != nil {
			resolvedCreatedBy = userIDMap[*a.CreatedBy]
		}
		var effectiveTo any
		if a.EffectiveTo != nil {
			effectiveTo = *a.EffectiveTo
		}

		targetID, existed, err := resolveAssignmentID(ctx, tx, a.ID, resolvedUserID, resolvedRoleID, resolvedParentID, a.EffectiveFrom, *a.EffectiveTo)
		if err != nil {
			return counts, err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO sales_structure_assignments (id, user_id, sales_role_id, parent_user_id, effective_from, effective_to, created_by, updated_at)
			VALUES ($1, $2, $3, $4, $5::date, $6::date, $7, now())
			ON CONFLICT (id) DO UPDATE SET
				user_id = EXCLUDED.user_id,
				sales_role_id = EXCLUDED.sales_role_id,
				parent_user_id = EXCLUDED.parent_user_id,
				effective_from = EXCLUDED.effective_from,
				effective_to = EXCLUDED.effective_to,
				created_by = EXCLUDED.created_by,
				updated_at = now()`,
			targetID, resolvedUserID, resolvedRoleID, resolvedParentID, a.EffectiveFrom, effectiveTo, resolvedCreatedBy)
		if err != nil {
			return counts, err
		}
		if existed {
			counts.Updated++
		} else {
			counts.Created++
		}
	}
	return counts, nil
}

func ptr(u uuid.UUID) *uuid.UUID {
	return &u
}

func isLocalDatabase(databaseURL string) bool {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1"
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	if !isLocalDatabase(cfg.DatabaseURL) {
		return fmt.Errorf("refusing to seed baseline on non-local database %q", cfg.DatabaseURL)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	userIDMap, users, err := seedUsers(ctx, tx)
	if err != nil {
		return err
	}
	roleIDMap, roles, err := seedRoles(ctx, tx, userIDMap)
	if err != nil {
		return err
	}
	permissionGrants, err := seedRolePermissions(ctx, tx, roleIDMap)
	if err != nil {
		return err
	}
	userRoleLinks, err := seedUserSalesRoles(ctx, tx, userIDMap, roleIDMap)
	if err != nil {
		return err
	}
	assignments, err := seedAssignments(ctx, tx, userIDMap, roleIDMap)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	fmt.Printf("Baseline seeded. users created=%d updated=%d; roles created=%d updated=%d; permission grants added=%d; user sales-role links set=%d; assignments created=%d updated=%d\n",
		users.Created, users.Updated, roles.Created, roles.Updated, permissionGrants, userRoleLinks, assignments.Created, assignments.Updated)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed_baseline:", err)
		os.Exit(1)
	}
}
