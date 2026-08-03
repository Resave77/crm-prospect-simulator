package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"crm-prospect-simulator/backend/internal/admin/model"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const listColumns = `u.id, u.email, u.full_name, u.employee_id, u.phone,
	u.role::text, u.status::text, u.must_change_password,
	u.manager_id, COALESCE(m.full_name, ''),
	COALESCE(sr.name, CASE WHEN u.role::text = 'SUPER_ADMIN' THEN 'Super Admin' ELSE '' END), sr.level,
	u.created_at, u.updated_at`

const detailColumns = `u.id, u.email, u.full_name, u.employee_id, u.phone,
	u.role::text, u.status::text, u.must_change_password,
	u.manager_id, COALESCE(m.full_name, ''),
	u.created_by, u.updated_by,
	u.created_at, u.updated_at`

const userJoin = `FROM users u LEFT JOIN users m ON m.id = u.manager_id
	LEFT JOIN LATERAL (
		SELECT r.name, r.level
		FROM sales_structure_assignments a
		JOIN sales_roles r ON r.id = a.sales_role_id
		WHERE a.user_id = u.id
		  AND a.effective_from <= CURRENT_DATE
		  AND (a.effective_to IS NULL OR a.effective_to >= CURRENT_DATE)
		ORDER BY a.effective_from DESC
		LIMIT 1
	) sr ON true`

func (r *PostgresRepository) ListUsers(ctx context.Context, filter model.ListFilter) (model.UserListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10
	}

	where, args := buildListWhere(filter)
	idx := len(args) + 1

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+userJoin+where, args...).Scan(&total)
	if err != nil {
		return model.UserListResult{}, fmt.Errorf("count users: %w", err)
	}

	pages := total / filter.Limit
	if total%filter.Limit > 0 {
		pages++
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := `SELECT ` + listColumns + ` ` + userJoin + where +
		` ORDER BY u.created_at DESC LIMIT $` + itoa(idx) + ` OFFSET $` + itoa(idx+1)
	dataArgs := append(args, filter.Limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return model.UserListResult{}, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	items := make([]model.UserListItem, 0)
	for rows.Next() {
		item, err := scanUserListItem(rows)
		if err != nil {
			return model.UserListResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.UserListResult{}, fmt.Errorf("rows iteration: %w", err)
	}

	return model.UserListResult{
		Items: items,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
		Pages: pages,
	}, nil
}

func buildListWhere(filter model.ListFilter) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)
	idx := 1

	if filter.Search != "" {
		pattern := "%" + strings.TrimSpace(filter.Search) + "%"
		conditions = append(conditions, `(`+
			`u.employee_id ILIKE $`+itoa(idx)+` OR `+
			`u.full_name ILIKE $`+itoa(idx)+` OR `+
			`u.email ILIKE $`+itoa(idx)+` OR `+
			`u.phone ILIKE $`+itoa(idx)+
			`)`)
		args = append(args, pattern)
		idx++
	}
	if filter.Role != "" {
		conditions = append(conditions, `u.role::text = $`+itoa(idx))
		args = append(args, filter.Role)
		idx++
	}
	if filter.Status != "" {
		conditions = append(conditions, `u.status::text = $`+itoa(idx))
		args = append(args, filter.Status)
		idx++
	}
	if filter.ManagerID != "" {
		conditions = append(conditions, `u.manager_id = $`+itoa(idx))
		args = append(args, filter.ManagerID)
		idx++
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func (r *PostgresRepository) FindUserDetail(ctx context.Context, id uuid.UUID) (model.UserDetail, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+detailColumns+` `+userJoin+` WHERE u.id = $1`, id)
	return scanUserDetail(row)
}

func (r *PostgresRepository) CreateUser(ctx context.Context, id uuid.UUID, input model.CreateUserInput, passwordHash string, actorID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, employee_id, phone, role, status,
		                   must_change_password, manager_id, created_by, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'ACTIVE', true, $8, $9, $9, now())`,
		id, strings.ToLower(strings.TrimSpace(input.Email)), passwordHash,
		strings.TrimSpace(input.FullName), strings.TrimSpace(input.EmployeeID),
		strings.TrimSpace(input.Phone), input.Role, input.ManagerID, actorID)
	if err != nil {
		return mapError(err)
	}
	return nil
}

func updateSets(input model.UpdateUserInput, actorID uuid.UUID) (string, []any) {
	sets := make([]string, 0)
	args := make([]any, 0)
	idx := 2

	add := func(column string, value any) {
		sets = append(sets, column+` = $`+itoa(idx))
		args = append(args, value)
		idx++
	}

	if input.FullName != nil {
		add(`full_name`, strings.TrimSpace(*input.FullName))
	}
	if input.Email != nil {
		add(`email`, strings.ToLower(strings.TrimSpace(*input.Email)))
	}
	if input.Phone != nil {
		add(`phone`, strings.TrimSpace(*input.Phone))
	}
	if input.EmployeeID != nil {
		add(`employee_id`, strings.TrimSpace(*input.EmployeeID))
	}
	if input.Role != nil {
		sets = append(sets, `role = $`+itoa(idx)+`::UserRole`)
		args = append(args, string(*input.Role))
		idx++
	}
	if input.ManagerID.Present {
		if input.ManagerID.Value == nil {
			sets = append(sets, `manager_id = NULL`)
		} else {
			add(`manager_id`, *input.ManagerID.Value)
		}
	}

	if len(sets) == 0 {
		return "", nil
	}

	sets = append(sets, `updated_by = $`+itoa(idx))
	args = append(args, actorID)
	sets = append(sets, `updated_at = now()`)

	return strings.Join(sets, ", "), args
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput, actorID uuid.UUID) error {
	sets, args := updateSets(input, actorID)
	if sets == "" {
		return nil
	}

	query := `UPDATE users SET ` + sets + ` WHERE id = $1`
	command, err := r.pool.Exec(ctx, query, append([]any{id}, args...)...)
	if err != nil {
		return mapError(err)
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status authmodel.UserStatus, actorID uuid.UUID) error {
	command, err := r.pool.Exec(ctx, `
		UPDATE users SET status = $2, updated_by = $3, updated_at = now() WHERE id = $1`,
		id, status, actorID)
	if err != nil {
		return mapError(err)
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ListActiveManagers(ctx context.Context) ([]model.ManagerOption, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, employee_id, full_name, email FROM users
		WHERE role = 'SALES_MANAGER' AND status = 'ACTIVE'
		ORDER BY full_name`)
	if err != nil {
		return nil, fmt.Errorf("list active managers: %w", err)
	}
	defer rows.Close()

	items := make([]model.ManagerOption, 0)
	for rows.Next() {
		var item model.ManagerOption
		var empID pgtype.Text
		if err := rows.Scan(&item.ID, &empID, &item.FullName, &item.Email); err != nil {
			return nil, fmt.Errorf("scan manager option: %w", err)
		}
		if empID.Valid {
			item.EmployeeID = empID.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ExistsByEmail(ctx context.Context, email string, excludeID *uuid.UUID) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var exists bool
	var err error
	if excludeID != nil {
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`, email, *excludeID).Scan(&exists)
	} else {
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	}
	if err != nil {
		return false, fmt.Errorf("check email: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) ExistsByEmployeeID(ctx context.Context, employeeID string, excludeID *uuid.UUID) (bool, error) {
	employeeID = strings.TrimSpace(employeeID)
	var exists bool
	var err error
	if excludeID != nil {
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE employee_id = $1 AND id != $2)`, employeeID, *excludeID).Scan(&exists)
	} else {
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE employee_id = $1)`, employeeID).Scan(&exists)
	}
	if err != nil {
		return false, fmt.Errorf("check employee_id: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) FindManagerByID(ctx context.Context, id uuid.UUID) (authmodel.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, employee_id, phone,
		       role::text, status::text, token_version, last_login_at,
		       must_change_password, manager_id, created_by, updated_by,
		       created_at, updated_at
		FROM users WHERE id = $1`, id))
}

func (r *PostgresRepository) FindUserByID(ctx context.Context, id uuid.UUID) (authmodel.User, error) {
	user, err := r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, employee_id, phone,
		       role::text, status::text, token_version, last_login_at,
		       must_change_password, manager_id, created_by, updated_by,
		       created_at, updated_at
		FROM users WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return authmodel.User{}, ErrNotFound
	}
	return user, err
}

func (r *PostgresRepository) CountActiveAdministrators(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role IN ('SUPER_ADMIN', 'ADMINISTRATOR') AND status = 'ACTIVE'`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active administrators: %w", err)
	}
	return count, nil
}

const resetPasswordUpdate = `UPDATE users SET
	password_hash = $1,
	must_change_password = TRUE,
	token_version = token_version + 1,
	updated_by = $2,
	updated_at = now()
WHERE id = $3`

const revokeActiveSessions = `UPDATE refresh_sessions SET
	revoked_at = now(),
	revoke_reason = 'ADMIN_PASSWORD_RESET'
WHERE user_id = $1 AND revoked_at IS NULL`

// resetPasswordTx applies the password reset update and refresh-session
// revocation inside the caller's transaction.
func resetPasswordTx(ctx context.Context, tx pgx.Tx, passwordHash string, actorUserID uuid.UUID, targetUserID uuid.UUID) (int64, error) {
	command, err := tx.Exec(ctx, resetPasswordUpdate, passwordHash, actorUserID, targetUserID)
	if err != nil {
		return 0, mapError(err)
	}
	if command.RowsAffected() == 0 {
		return 0, ErrNotFound
	}

	revoked, err := tx.Exec(ctx, revokeActiveSessions, targetUserID)
	if err != nil {
		return 0, mapError(err)
	}
	return revoked.RowsAffected(), nil
}

// ResetPassword replaces the target user's password hash, forces a password
// change, bumps token_version, and revokes all active refresh sessions in a
// single transaction. Returns the number of sessions revoked.
func (r *PostgresRepository) ResetPassword(ctx context.Context, targetUserID uuid.UUID, actorUserID uuid.UUID, passwordHash string) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin password reset: %w", err)
	}
	defer tx.Rollback(ctx)

	revoked, err := resetPasswordTx(ctx, tx, passwordHash, actorUserID, targetUserID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit password reset: %w", err)
	}
	return revoked, nil
}

func scanUserListItem(row pgx.Row) (model.UserListItem, error) {
	var item model.UserListItem
	var employeeID pgtype.Text
	var phone pgtype.Text
	var orgRoleLevel pgtype.Int2
	var managerID *uuid.UUID
	err := row.Scan(&item.ID, &item.Email, &item.FullName,
		&employeeID, &phone,
		&item.Role, &item.Status, &item.MustChangePassword,
		&managerID, &item.ManagerName,
		&item.OrganizationalRole, &orgRoleLevel,
		&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return model.UserListItem{}, fmt.Errorf("scan user list item: %w", err)
	}
	if employeeID.Valid {
		item.EmployeeID = employeeID.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	if orgRoleLevel.Valid {
		level := int(orgRoleLevel.Int16)
		item.OrganizationalRoleLevel = &level
	}
	item.ManagerID = managerID
	return item, nil
}

func scanUserDetail(row pgx.Row) (model.UserDetail, error) {
	var item model.UserDetail
	var employeeID pgtype.Text
	var phone pgtype.Text
	var managerID *uuid.UUID
	err := row.Scan(&item.ID, &item.Email, &item.FullName,
		&employeeID, &phone,
		&item.Role, &item.Status, &item.MustChangePassword,
		&managerID, &item.ManagerName,
		&item.CreatedBy, &item.UpdatedBy,
		&item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.UserDetail{}, ErrNotFound
	}
	if err != nil {
		return model.UserDetail{}, fmt.Errorf("scan user detail: %w", err)
	}
	if employeeID.Valid {
		item.EmployeeID = employeeID.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	item.ManagerID = managerID
	return item, nil
}

func (r *PostgresRepository) scanUser(row pgx.Row) (authmodel.User, error) {
	var user authmodel.User
	var employeeID pgtype.Text
	var phone pgtype.Text
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
		&employeeID, &phone,
		&user.Role, &user.Status, &user.TokenVersion, &user.LastLoginAt,
		&user.MustChangePassword, &user.ManagerID, &user.CreatedBy, &user.UpdatedBy,
		&user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return authmodel.User{}, ErrNotFound
	}
	if err != nil {
		return authmodel.User{}, fmt.Errorf("scan user: %w", err)
	}
	if employeeID.Valid {
		user.EmployeeID = employeeID.String
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	return user, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" {
		return ErrConflict
	}
	return fmt.Errorf("database operation: %w", err)
}
