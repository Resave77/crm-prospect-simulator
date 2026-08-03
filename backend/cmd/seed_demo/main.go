package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"crm-prospect-simulator/backend/config"
	"crm-prospect-simulator/backend/internal/admin/demo"
	"crm-prospect-simulator/backend/platform/database"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type counters struct {
	Created int
	Updated int
	Skipped int
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "seed demo:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ds := demo.BuildDataset()
	if err := demo.Validate(ds); err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if !isLocalDatabase(cfg.DatabaseURL) {
		return fmt.Errorf("database target is not recognized as local/disposable; refusing to seed")
	}
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(demo.DemoPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	roleCounts, err := seedRoles(ctx, tx, ds)
	if err != nil {
		return err
	}
	userCounts, err := seedUsers(ctx, tx, ds, string(hash))
	if err != nil {
		return err
	}
	assignmentCounts, err := seedAssignments(ctx, tx, ds)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	fmt.Printf("Demo sales organization seeded. roles created=%d updated=%d skipped=%d; users created=%d updated=%d skipped=%d; assignments created=%d updated=%d skipped=%d\n",
		roleCounts.Created, roleCounts.Updated, roleCounts.Skipped,
		userCounts.Created, userCounts.Updated, userCounts.Skipped,
		assignmentCounts.Created, assignmentCounts.Updated, assignmentCounts.Skipped)
	return nil
}

func seedRoles(ctx context.Context, tx pgx.Tx, ds demo.Dataset) (counters, error) {
	var counts counters
	for _, role := range ds.Roles {
		normalized := demo.NormalizeRoleName(role.Name)
		// The sales organization foundation migration seeds default roles under
		// placeholder ids. A demo role may share its active normalized_name with
		// one of those placeholders, so release (deactivate) the conflicting
		// placeholder row before upserting the canonical demo role by id. Only
		// release rows that are not yet referenced by any assignment.
		if _, err := tx.Exec(ctx, `
			UPDATE sales_roles
			SET is_active = false, updated_at = now()
			WHERE normalized_name = $1
				AND id <> $2
				AND is_active = true
				AND NOT EXISTS (
					SELECT 1 FROM sales_structure_assignments a
					WHERE a.sales_role_id = sales_roles.id
				)`, normalized, role.ID); err != nil {
			return counts, fmt.Errorf("release placeholder role %s: %w", role.Name, err)
		}
		tag, err := tx.Exec(ctx, `
			INSERT INTO sales_roles (id, name, normalized_name, level, description, is_active, updated_at)
			VALUES ($1, $2, $3, $4, $5, true, now())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				normalized_name = EXCLUDED.normalized_name,
				level = EXCLUDED.level,
				description = EXCLUDED.description,
				is_active = true,
				updated_at = now()
			WHERE sales_roles.name IS DISTINCT FROM EXCLUDED.name
				OR sales_roles.normalized_name IS DISTINCT FROM EXCLUDED.normalized_name
				OR sales_roles.level IS DISTINCT FROM EXCLUDED.level
				OR sales_roles.description IS DISTINCT FROM EXCLUDED.description
				OR sales_roles.is_active IS DISTINCT FROM true`,
			role.ID, role.Name, demo.NormalizeRoleName(role.Name), role.Level, role.Description)
		if err != nil {
			return counts, fmt.Errorf("upsert role %s: %w", role.Name, err)
		}
		if tag.RowsAffected() == 0 {
			counts.Skipped++
		} else {
			var existed bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sales_roles WHERE id = $1 AND created_at < now() - interval '1 second')`, role.ID).Scan(&existed); err != nil {
				return counts, err
			}
			if existed {
				counts.Updated++
			} else {
				counts.Created++
			}
		}
	}
	return counts, nil
}

func seedUsers(ctx context.Context, tx pgx.Tx, ds demo.Dataset, passwordHash string) (counters, error) {
	var counts counters
	userByKey := map[string]demo.User{}
	for _, user := range ds.Users {
		userByKey[user.Key] = user
	}
	for _, user := range ds.Users {
		var managerID any
		if user.ManagerKey != "" {
			manager := userByKey[user.ManagerKey]
			managerID = manager.ID
		}
		var existed bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, user.ID).Scan(&existed); err != nil {
			return counts, err
		}
		tag, err := tx.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, employee_id, phone, role, status, must_change_password, manager_id, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7::"UserRole", 'ACTIVE', false, $8, now())
			ON CONFLICT (id) DO UPDATE SET
				email = EXCLUDED.email,
				full_name = EXCLUDED.full_name,
				employee_id = EXCLUDED.employee_id,
				phone = EXCLUDED.phone,
				role = EXCLUDED.role,
				status = 'ACTIVE',
				must_change_password = false,
				manager_id = EXCLUDED.manager_id,
				updated_at = now()
			WHERE users.email LIKE '%@demo.yummy.local'
				AND (users.email IS DISTINCT FROM EXCLUDED.email
					OR users.full_name IS DISTINCT FROM EXCLUDED.full_name
					OR users.employee_id IS DISTINCT FROM EXCLUDED.employee_id
					OR users.phone IS DISTINCT FROM EXCLUDED.phone
					OR users.role IS DISTINCT FROM EXCLUDED.role
					OR users.status IS DISTINCT FROM 'ACTIVE'
					OR users.must_change_password IS DISTINCT FROM false
					OR users.manager_id IS DISTINCT FROM EXCLUDED.manager_id)`,
			user.ID, strings.ToLower(user.Email), passwordHash, user.FullName, user.EmployeeID, user.Phone, user.SystemRole, managerID)
		if err != nil {
			return counts, fmt.Errorf("upsert user %s: %w", user.EmployeeID, err)
		}
		if tag.RowsAffected() == 0 {
			counts.Skipped++
		} else if existed {
			counts.Updated++
		} else {
			counts.Created++
		}
	}
	return counts, nil
}

func seedAssignments(ctx context.Context, tx pgx.Tx, ds demo.Dataset) (counters, error) {
	var counts counters
	users := map[string]demo.User{}
	roles := map[string]demo.Role{}
	for _, user := range ds.Users {
		users[user.Key] = user
	}
	for _, role := range ds.Roles {
		roles[role.Key] = role
	}
	for _, assignment := range ds.Assignments {
		user := users[assignment.UserKey]
		role := roles[assignment.RoleKey]
		var parentID any
		if assignment.ParentKey != "" {
			parent := users[assignment.ParentKey]
			parentID = parent.ID
		}
		var existed bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sales_structure_assignments WHERE id = $1)`, assignment.ID).Scan(&existed); err != nil {
			return counts, err
		}
		tag, err := tx.Exec(ctx, `
			INSERT INTO sales_structure_assignments (id, user_id, sales_role_id, parent_user_id, effective_from, effective_to, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, now())
			ON CONFLICT (id) DO UPDATE SET
				user_id = EXCLUDED.user_id,
				sales_role_id = EXCLUDED.sales_role_id,
				parent_user_id = EXCLUDED.parent_user_id,
				effective_from = EXCLUDED.effective_from,
				effective_to = EXCLUDED.effective_to,
				updated_at = now()
			WHERE sales_structure_assignments.user_id IS DISTINCT FROM EXCLUDED.user_id
				OR sales_structure_assignments.sales_role_id IS DISTINCT FROM EXCLUDED.sales_role_id
				OR sales_structure_assignments.parent_user_id IS DISTINCT FROM EXCLUDED.parent_user_id
				OR sales_structure_assignments.effective_from IS DISTINCT FROM EXCLUDED.effective_from
				OR sales_structure_assignments.effective_to IS DISTINCT FROM EXCLUDED.effective_to`,
			assignment.ID, user.ID, role.ID, parentID, assignment.From, assignment.To)
		if err != nil {
			return counts, fmt.Errorf("upsert assignment %s: %w", assignment.Key, err)
		}
		if tag.RowsAffected() == 0 {
			counts.Skipped++
		} else if existed {
			counts.Updated++
		} else {
			counts.Created++
		}
	}
	return counts, nil
}

func isLocalDatabase(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	name := strings.ToLower(strings.TrimPrefix(parsed.Path, "/"))
	if host != "localhost" && host != "127.0.0.1" && host != "::1" {
		return false
	}
	return !strings.Contains(name, "prod") && !strings.Contains(name, "production")
}
