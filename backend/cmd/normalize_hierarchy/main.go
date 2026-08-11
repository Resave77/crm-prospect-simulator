package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	targetDatabase = "crm_prospect_simulator_dev"
	effectiveFrom  = "2026-08-01"

	roleSuperAdminID = "00000000-0000-0000-0000-000000000100"
	roleLevel1ID     = "00000000-0000-0000-0000-000000000101"
	roleLevel2ID     = "00000000-0000-0000-0000-000000000102"
	roleLevel3ID     = "00000000-0000-0000-0000-000000000104"

	assignmentAdminRootID    = "840abe29-2f74-4422-89b7-d4834c1cf61c"
	assignmentLevel1RootID   = "945c73a2-cedb-42bf-88c2-9e78388967e5"
	assignmentLevel2ID       = "6d0b6630-f256-41bd-8a37-acbff3ab542f"
	assignmentLevel3ID       = "989db891-e70e-4f2a-a414-6b5f24a3a6c7"
	assignmentSalesLevel3ID  = "a57c54e4-c2f8-4d10-aa4a-a1fbd6e3324e"
	assignmentSales2Level3ID = "18209b51-ddb5-4990-b8f3-e7dc68998d51"
	assignmentSales3Level3ID = "f93b718f-d2ad-5189-b75a-b6224252750e"
)

type roleTarget struct {
	ID          uuid.UUID
	Name        string
	Level       int
	LandingPage string
	Description string
}

type assignmentTarget struct {
	ID          uuid.UUID
	Email       string
	RoleID      uuid.UUID
	ParentEmail string
}

type userRow struct {
	ID          uuid.UUID
	Email       string
	FullName    string
	Role        string
	Status      string
	SalesRoleID *uuid.UUID
}

type roleRow struct {
	ID             uuid.UUID
	Name           string
	NormalizedName string
	Level          int
	LandingPage    string
	IsActive       bool
}

type assignmentRow struct {
	ID            uuid.UUID
	UserEmail     string
	UserName      string
	UserRole      string
	RoleID        uuid.UUID
	RoleName      string
	RoleLevel     int
	ParentEmail   string
	ParentName    string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

var businessTables = []string{
	"prospects",
	"prospect_visits",
	"prospect_status_history",
	"parent_companies",
	"customer_sites",
	"prospect_comments",
	"prospect_photo_tags",
}

func uid(s string) uuid.UUID {
	return uuid.MustParse(s)
}

func roleTargets() []roleTarget {
	return []roleTarget{
		{ID: uid(roleSuperAdminID), Name: "SUPER_ADMIN", Level: 1, LandingPage: "/admin/dashboard", Description: "Top-level organizational role for the primary Super Admin"},
		{ID: uid(roleLevel1ID), Name: "Sales Level 1", Level: 2, LandingPage: "/sales/dashboard", Description: "Top-level sales role under Super Admin"},
		{ID: uid(roleLevel2ID), Name: "Sales Level 2", Level: 3, LandingPage: "/sales/dashboard", Description: "Second-level sales role"},
		{ID: uid(roleLevel3ID), Name: "Sales Level 3", Level: 4, LandingPage: "/sales/dashboard", Description: "Operational sales role"},
	}
}

func assignmentTargets() []assignmentTarget {
	return []assignmentTarget{
		{ID: uid(assignmentAdminRootID), Email: "admin@yummy.test", RoleID: uid(roleSuperAdminID)},
		{ID: uid(assignmentLevel1RootID), Email: "level1@yummy.test", RoleID: uid(roleLevel1ID), ParentEmail: "admin@yummy.test"},
		{ID: uid(assignmentLevel2ID), Email: "level2@yummy.test", RoleID: uid(roleLevel2ID), ParentEmail: "level1@yummy.test"},
		{ID: uid(assignmentLevel3ID), Email: "level3@yummy.test", RoleID: uid(roleLevel3ID), ParentEmail: "level2@yummy.test"},
		{ID: uid(assignmentSalesLevel3ID), Email: "sales@yummy.test", RoleID: uid(roleLevel3ID), ParentEmail: "level2@yummy.test"},
		{ID: uid(assignmentSales2Level3ID), Email: "sales2@yummy.test", RoleID: uid(roleLevel3ID), ParentEmail: "level2@yummy.test"},
		{ID: uid(assignmentSales3Level3ID), Email: "sales3@yummy.test", RoleID: uid(roleLevel3ID), ParentEmail: "level2@yummy.test"},
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "normalize_hierarchy:", err)
		os.Exit(1)
	}
}

func run() error {
	execute := flag.Bool("execute", false, "apply changes; default is dry-run/read-only rollback")
	flag.Parse()

	ctx := context.Background()
	_ = godotenv.Load(".env", "../.env")
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if err := guardDatabaseURL(databaseURL); err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

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

	if err := guardCurrentDatabase(ctx, tx); err != nil {
		return err
	}

	beforeBusiness, err := businessCounts(ctx, tx)
	if err != nil {
		return err
	}
	beforeRoles, err := listTargetRoles(ctx, tx)
	if err != nil {
		return err
	}
	beforeHierarchy, err := listActiveHierarchy(ctx, tx)
	if err != nil {
		return err
	}

	users, err := preflight(ctx, tx)
	if err != nil {
		return err
	}
	if err := applyNormalization(ctx, tx, users); err != nil {
		return err
	}
	if err := validateHierarchy(ctx, tx); err != nil {
		return err
	}
	afterBusiness, err := businessCounts(ctx, tx)
	if err != nil {
		return err
	}
	if err := compareBusinessCounts(beforeBusiness, afterBusiness); err != nil {
		return err
	}
	afterRoles, err := listTargetRoles(ctx, tx)
	if err != nil {
		return err
	}
	afterHierarchy, err := listActiveHierarchy(ctx, tx)
	if err != nil {
		return err
	}

	printReport(*execute, beforeBusiness, afterBusiness, beforeRoles, afterRoles, beforeHierarchy, afterHierarchy)

	if !*execute {
		fmt.Println("DRY-RUN complete. Transaction rolled back; no database changes were committed.")
		return nil
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	fmt.Println("EXECUTE complete. Transaction committed.")
	return nil
}

func guardDatabaseURL(databaseURL string) error {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName != targetDatabase {
		return fmt.Errorf("refusing to run: database must be %q, got %q", targetDatabase, dbName)
	}
	return nil
}

func guardCurrentDatabase(ctx context.Context, tx pgx.Tx) error {
	var current string
	if err := tx.QueryRow(ctx, `SELECT current_database()`).Scan(&current); err != nil {
		return fmt.Errorf("read current database: %w", err)
	}
	if current != targetDatabase {
		return fmt.Errorf("refusing to run: connected database must be %q, got %q", targetDatabase, current)
	}
	return nil
}

func preflight(ctx context.Context, tx pgx.Tx) (map[string]userRow, error) {
	required := map[string]string{
		"admin@yummy.test":  "SUPER_ADMIN",
		"level1@yummy.test": "SALES_MANAGER",
		"level2@yummy.test": "SALES_MANAGER",
		"level3@yummy.test": "SALES_EXECUTIVE",
		"sales@yummy.test":  "SALES_EXECUTIVE",
		"sales2@yummy.test": "SALES_EXECUTIVE",
		"sales3@yummy.test": "SALES_EXECUTIVE",
	}
	users := make(map[string]userRow, len(required))
	for email, expectedRole := range required {
		user, err := findUser(ctx, tx, email)
		if err != nil {
			return nil, fmt.Errorf("preflight user %s: %w", email, err)
		}
		if user.Status != "ACTIVE" {
			return nil, fmt.Errorf("preflight user %s: expected ACTIVE, got %s", email, user.Status)
		}
		if user.Role != expectedRole {
			return nil, fmt.Errorf("preflight user %s: expected role %s, got %s", email, expectedRole, user.Role)
		}
		users[email] = user
	}
	return users, nil
}

func findUser(ctx context.Context, tx pgx.Tx, email string) (userRow, error) {
	var row userRow
	var salesRoleID *uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT id, email, full_name, role::text, status::text, sales_role_id
		FROM users
		WHERE lower(btrim(email)) = lower(btrim($1)) AND deleted_at IS NULL
		ORDER BY created_at
		LIMIT 1`, email).Scan(&row.ID, &row.Email, &row.FullName, &row.Role, &row.Status, &salesRoleID)
	if err != nil {
		return userRow{}, err
	}
	row.SalesRoleID = salesRoleID
	return row, nil
}

func applyNormalization(ctx context.Context, tx pgx.Tx, users map[string]userRow) error {
	for _, role := range roleTargets() {
		if err := retireConflictingActiveRoleNames(ctx, tx, role); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO sales_roles (id, name, normalized_name, level, description, landing_page, is_active, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, true, $7, now())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				normalized_name = EXCLUDED.normalized_name,
				level = EXCLUDED.level,
				description = EXCLUDED.description,
				landing_page = EXCLUDED.landing_page,
				is_active = true,
				updated_by = EXCLUDED.updated_by,
				updated_at = now()`,
			role.ID, role.Name, normalizeName(role.Name), role.Level, role.Description, role.LandingPage, users["admin@yummy.test"].ID); err != nil {
			return fmt.Errorf("upsert role %s: %w", role.Name, err)
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET sales_role_id = NULL, manager_id = NULL, updated_at = now()
		WHERE role = 'SUPER_ADMIN' AND deleted_at IS NULL AND (sales_role_id IS NOT NULL OR manager_id IS NOT NULL)`); err != nil {
		return fmt.Errorf("clear SUPER_ADMIN direct sales role links: %w", err)
	}

	for _, target := range assignmentTargets() {
		user := users[target.Email]
		var parentID *uuid.UUID
		if target.ParentEmail != "" {
			parent := users[target.ParentEmail]
			parentID = &parent.ID
		}
		if err := upsertActiveAssignment(ctx, tx, target.ID, user.ID, target.RoleID, parentID, users["admin@yummy.test"].ID); err != nil {
			return fmt.Errorf("assignment %s: %w", target.Email, err)
		}
	}

	adminID := users["admin@yummy.test"].ID
	if _, err := tx.Exec(ctx, `
		UPDATE sales_structure_assignments a
		SET effective_from = LEAST(a.effective_from, $2::date - 1),
			effective_to = $2::date - 1,
			updated_at = now()
		FROM sales_roles r
		WHERE a.sales_role_id = r.id
		  AND r.level = 1
		  AND a.effective_from <= $2::date
		  AND (a.effective_to IS NULL OR a.effective_to >= $2::date)
		  AND a.user_id <> $1`, adminID, effectiveFrom); err != nil {
		return fmt.Errorf("close non-admin level-1 roots: %w", err)
	}
	return nil
}

func retireConflictingActiveRoleNames(ctx context.Context, tx pgx.Tx, role roleTarget) error {
	normalized := normalizeName(role.Name)
	_, err := tx.Exec(ctx, `
		UPDATE sales_roles
		SET is_active = false,
			normalized_name = normalized_name || ' retired ' || left(id::text, 8),
			updated_at = now()
		WHERE id <> $1
		  AND is_active = true
		  AND lower(btrim(normalized_name)) = lower(btrim($2))`, role.ID, normalized)
	if err != nil {
		return fmt.Errorf("retire conflicting role name %s: %w", role.Name, err)
	}
	return nil
}

func upsertActiveAssignment(ctx context.Context, tx pgx.Tx, id, userID, roleID uuid.UUID, parentID *uuid.UUID, actorID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
		UPDATE sales_structure_assignments
		SET effective_from = LEAST(effective_from, $2::date - 1),
			effective_to = $2::date - 1,
			updated_at = now()
		WHERE user_id = $1
		  AND effective_from <= $2::date
		  AND (effective_to IS NULL OR effective_to >= $2::date)
		  AND id <> $3`, userID, effectiveFrom, id); err != nil {
		return fmt.Errorf("close overlapping assignments: %w", err)
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO sales_structure_assignments (id, user_id, sales_role_id, parent_user_id, effective_from, effective_to, created_by, updated_at)
		VALUES ($1, $2, $3, $4, $5::date, NULL, $6, now())
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			sales_role_id = EXCLUDED.sales_role_id,
			parent_user_id = EXCLUDED.parent_user_id,
			effective_from = EXCLUDED.effective_from,
			effective_to = NULL,
			created_by = EXCLUDED.created_by,
			updated_at = now()`,
		id, userID, roleID, parentID, effectiveFrom, actorID)
	return err
}

func validateHierarchy(ctx context.Context, tx pgx.Tx) error {
	var rootCount int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM sales_structure_assignments a
		JOIN sales_roles r ON r.id = a.sales_role_id
		JOIN users u ON u.id = a.user_id
		WHERE a.effective_from <= $1::date
		  AND (a.effective_to IS NULL OR a.effective_to >= $1::date)
		  AND r.level = 1
		  AND u.deleted_at IS NULL`, effectiveFrom).Scan(&rootCount); err != nil {
		return fmt.Errorf("validate root count: %w", err)
	}
	if rootCount != 1 {
		return fmt.Errorf("validate root count: expected 1 active level-1 root, got %d", rootCount)
	}

	for _, target := range assignmentTargets() {
		var ok bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM sales_structure_assignments a
				JOIN users u ON u.id = a.user_id
				WHERE a.id = $1
				  AND lower(btrim(u.email)) = lower(btrim($2))
				  AND a.sales_role_id = $3
				  AND a.effective_from <= $4::date
				  AND (a.effective_to IS NULL OR a.effective_to >= $4::date)
			)`, target.ID, target.Email, target.RoleID, effectiveFrom).Scan(&ok)
		if err != nil {
			return fmt.Errorf("validate assignment %s: %w", target.Email, err)
		}
		if !ok {
			return fmt.Errorf("validate assignment %s: target active assignment missing", target.Email)
		}
	}
	return nil
}

func businessCounts(ctx context.Context, tx pgx.Tx) (map[string]int64, error) {
	counts := make(map[string]int64, len(businessTables))
	for _, table := range businessTables {
		var count int64
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM `+table).Scan(&count); err != nil {
			return nil, fmt.Errorf("count %s: %w", table, err)
		}
		counts[table] = count
	}
	return counts, nil
}

func compareBusinessCounts(before, after map[string]int64) error {
	for _, table := range businessTables {
		if before[table] != after[table] {
			return fmt.Errorf("business row count changed for %s: before=%d after=%d", table, before[table], after[table])
		}
	}
	return nil
}

func listTargetRoles(ctx context.Context, tx pgx.Tx) ([]roleRow, error) {
	ids := make([]uuid.UUID, 0, len(roleTargets()))
	for _, role := range roleTargets() {
		ids = append(ids, role.ID)
	}
	rows, err := tx.Query(ctx, `
		SELECT id, name, normalized_name, level, COALESCE(landing_page, ''), is_active
		FROM sales_roles
		WHERE id = ANY($1)
		ORDER BY level, name`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []roleRow
	for rows.Next() {
		var row roleRow
		if err := rows.Scan(&row.ID, &row.Name, &row.NormalizedName, &row.Level, &row.LandingPage, &row.IsActive); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func listActiveHierarchy(ctx context.Context, tx pgx.Tx) ([]assignmentRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT a.id, u.email, u.full_name, u.role::text, r.id, r.name, r.level,
		       COALESCE(p.email, ''), COALESCE(p.full_name, ''),
		       a.effective_from, a.effective_to
		FROM sales_structure_assignments a
		JOIN users u ON u.id = a.user_id
		JOIN sales_roles r ON r.id = a.sales_role_id
		LEFT JOIN users p ON p.id = a.parent_user_id
		WHERE a.effective_from <= $1::date
		  AND (a.effective_to IS NULL OR a.effective_to >= $1::date)
		  AND u.deleted_at IS NULL
		ORDER BY r.level, u.email`, effectiveFrom)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []assignmentRow
	for rows.Next() {
		var row assignmentRow
		if err := rows.Scan(&row.ID, &row.UserEmail, &row.UserName, &row.UserRole, &row.RoleID, &row.RoleName, &row.RoleLevel, &row.ParentEmail, &row.ParentName, &row.EffectiveFrom, &row.EffectiveTo); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func printReport(execute bool, beforeBusiness, afterBusiness map[string]int64, beforeRoles, afterRoles []roleRow, beforeHierarchy, afterHierarchy []assignmentRow) {
	mode := "DRY-RUN"
	if execute {
		mode = "EXECUTE"
	}
	fmt.Printf("Mode: %s\n\n", mode)
	fmt.Println("Business row counts:")
	for _, table := range businessTables {
		fmt.Printf("  %-25s before=%d after=%d\n", table, beforeBusiness[table], afterBusiness[table])
	}
	fmt.Println()
	printRoles("Before roles", beforeRoles)
	printRoles("After roles", afterRoles)
	printHierarchy("Before active hierarchy", beforeHierarchy)
	printHierarchy("After active hierarchy", afterHierarchy)
}

func printRoles(title string, roles []roleRow) {
	fmt.Println(title + ":")
	if len(roles) == 0 {
		fmt.Println("  (none)")
		return
	}
	for _, role := range roles {
		fmt.Printf("  %s | %-13s | normalized=%-13s | level=%d | active=%t | landing=%s\n",
			role.ID, role.Name, role.NormalizedName, role.Level, role.IsActive, role.LandingPage)
	}
	fmt.Println()
}

func printHierarchy(title string, rows []assignmentRow) {
	fmt.Println(title + ":")
	if len(rows) == 0 {
		fmt.Println("  (none)")
		return
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].RoleLevel == rows[j].RoleLevel {
			return rows[i].UserEmail < rows[j].UserEmail
		}
		return rows[i].RoleLevel < rows[j].RoleLevel
	})
	for _, row := range rows {
		parent := "NULL"
		if row.ParentEmail != "" {
			parent = row.ParentEmail
		}
		fmt.Printf("  L%d %-13s user=%-18s parent=%-18s assignment=%s from=%s\n",
			row.RoleLevel, row.RoleName, row.UserEmail, parent, row.ID, row.EffectiveFrom.Format("2006-01-02"))
	}
	fmt.Println()
}

func normalizeName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}
