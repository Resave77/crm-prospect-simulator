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
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "cleanup demo:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if !isLocalDatabase(cfg.DatabaseURL) {
		return fmt.Errorf("database target is not recognized as local/disposable; refusing to cleanup")
	}
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	assignments, err := deleteDemoAssignments(ctx, tx)
	if err != nil {
		return err
	}
	sessions, err := deleteDemoSessions(ctx, tx)
	if err != nil {
		return err
	}
	users, err := deleteDemoUsers(ctx, tx)
	if err != nil {
		return err
	}
	roles, skippedRoles, err := deleteDemoRoles(ctx, tx)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	fmt.Printf("Demo sales organization cleanup complete. assignments deleted=%d; sessions deleted=%d; users deleted=%d; roles deleted=%d skipped_roles_in_use=%d\n",
		assignments, sessions, users, roles, skippedRoles)
	return nil
}

func deleteDemoAssignments(ctx context.Context, tx pgx.Tx) (int64, error) {
	tag, err := tx.Exec(ctx, `
		DELETE FROM sales_structure_assignments a
		USING users u
		WHERE a.user_id = u.id
			AND (u.email LIKE '%@demo.yummy.local' OR u.employee_id LIKE 'DEMO-%')`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func deleteDemoSessions(ctx context.Context, tx pgx.Tx) (int64, error) {
	tag, err := tx.Exec(ctx, `
		DELETE FROM refresh_sessions s
		USING users u
		WHERE s.user_id = u.id
			AND (u.email LIKE '%@demo.yummy.local' OR u.employee_id LIKE 'DEMO-%')`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func deleteDemoUsers(ctx context.Context, tx pgx.Tx) (int64, error) {
	tag, err := tx.Exec(ctx, `
		DELETE FROM users
		WHERE email LIKE '%@demo.yummy.local'
			OR employee_id LIKE 'DEMO-%'`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func deleteDemoRoles(ctx context.Context, tx pgx.Tx) (int64, int, error) {
	ds := demo.BuildDataset()
	deleted := int64(0)
	skipped := 0
	for _, role := range ds.Roles {
		var referenced bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sales_structure_assignments WHERE sales_role_id = $1)`, role.ID).Scan(&referenced); err != nil {
			return deleted, skipped, err
		}
		if referenced {
			skipped++
			continue
		}
		tag, err := tx.Exec(ctx, `DELETE FROM sales_roles WHERE id = $1`, role.ID)
		if err != nil {
			return deleted, skipped, err
		}
		deleted += tag.RowsAffected()
	}
	return deleted, skipped, nil
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
