package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"crm-prospect-simulator/backend/internal/auth/model"
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

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, employee_id, phone,
		       role::text, status::text, token_version, last_login_at,
		       must_change_password, manager_id, created_by, updated_by,
		       created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`, email))
}

func (r *PostgresRepository) FindUserByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, employee_id, phone,
		       role::text, status::text, token_version, last_login_at,
		       must_change_password, manager_id, created_by, updated_by,
		       created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`, id))
}

func (r *PostgresRepository) scanUser(row pgx.Row) (model.User, error) {
	var user model.User
	var employeeID pgtype.Text
	var phone pgtype.Text
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
		&employeeID, &phone,
		&user.Role, &user.Status, &user.TokenVersion, &user.LastLoginAt,
		&user.MustChangePassword, &user.ManagerID, &user.CreatedBy, &user.UpdatedBy,
		&user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("scan user: %w", err)
	}
	if employeeID.Valid {
		user.EmployeeID = employeeID.String
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	return user, nil
}

func (r *PostgresRepository) RecordLogin(ctx context.Context, userID uuid.UUID, at time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_login_at = $2, updated_at = $2 WHERE id = $1`, userID, at)
	if err != nil {
		return fmt.Errorf("record login: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpsertSeed(ctx context.Context, user model.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, employee_id, phone, role, status, token_version, must_change_password, manager_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1, false, $9, now())
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			employee_id = COALESCE(EXCLUDED.employee_id, users.employee_id),
			phone = COALESCE(EXCLUDED.phone, users.phone),
			role = EXCLUDED.role,
			status = EXCLUDED.status,
			manager_id = EXCLUDED.manager_id,
			updated_at = now()`,
		user.ID, user.Email, user.PasswordHash, user.FullName, user.EmployeeID, user.Phone, user.Role, user.Status, user.ManagerID)
	if err != nil {
		return fmt.Errorf("seed user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Create(ctx context.Context, session model.RefreshSession) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_sessions
			(id, user_id, token_hash, user_agent, ip_address, expires_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())`, session.ID, session.UserID,
		session.TokenHash, session.UserAgent, session.IPAddress, session.ExpiresAt)
	return databaseError("create refresh session", err)
}

func (r *PostgresRepository) FindSessionByID(ctx context.Context, id uuid.UUID) (model.RefreshSession, error) {
	var session model.RefreshSession
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, user_agent, ip_address, expires_at,
		       revoked_at, COALESCE(revoke_reason, ''), replaced_by_session_id, created_at
		FROM refresh_sessions WHERE id = $1`, id).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.UserAgent,
		&session.IPAddress, &session.ExpiresAt, &session.RevokedAt,
		&session.RevokeReason, &session.ReplacedBySessionID, &session.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.RefreshSession{}, ErrNotFound
	}
	if err != nil {
		return model.RefreshSession{}, fmt.Errorf("find refresh session: %w", err)
	}
	return session, nil
}

func (r *PostgresRepository) Rotate(ctx context.Context, oldID uuid.UUID, replacement model.RefreshSession, at time.Time) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin session rotation: %w", err)
	}
	defer tx.Rollback(ctx)

	command, err := tx.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = $2, revoke_reason = 'ROTATED', replaced_by_session_id = $3
		WHERE id = $1 AND revoked_at IS NULL AND expires_at > $2`, oldID, at, replacement.ID)
	if err != nil {
		return fmt.Errorf("revoke rotated session: %w", err)
	}
	if command.RowsAffected() != 1 {
		return ErrConflict
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO refresh_sessions
			(id, user_id, token_hash, user_agent, ip_address, expires_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, replacement.ID, replacement.UserID,
		replacement.TokenHash, replacement.UserAgent, replacement.IPAddress, replacement.ExpiresAt, at)
	if err != nil {
		return databaseError("insert rotated session", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit session rotation: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Revoke(ctx context.Context, id uuid.UUID, reason string, at time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_sessions SET revoked_at = COALESCE(revoked_at, $2), revoke_reason = $3
		WHERE id = $1`, id, at, reason)
	return databaseError("revoke refresh session", err)
}

func (r *PostgresRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID, reason string, at time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_sessions SET revoked_at = $2, revoke_reason = $3
		WHERE user_id = $1 AND revoked_at IS NULL`, userID, at, reason)
	return databaseError("revoke user sessions", err)
}

func (r *PostgresRepository) ChangePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string, revokeReason string, at time.Time) (int, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin change password txn: %w", err)
	}
	defer tx.Rollback(ctx)

	cmdUser, err := tx.Exec(ctx, `
		UPDATE users
		SET password_hash = $2,
		    must_change_password = false,
		    token_version = token_version + 1,
		    updated_at = $3
		WHERE id = $1`, userID, newPasswordHash, at)
	if err != nil {
		return 0, fmt.Errorf("update user password: %w", err)
	}
	if cmdUser.RowsAffected() != 1 {
		return 0, ErrNotFound
	}

	cmdSessions, err := tx.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = $2, revoke_reason = $3
		WHERE user_id = $1 AND revoked_at IS NULL`, userID, at, revokeReason)
	if err != nil {
		return 0, fmt.Errorf("revoke user sessions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit change password: %w", err)
	}

	return int(cmdSessions.RowsAffected()), nil
}

func databaseError(operation string, err error) error {
	if err == nil {
		return nil
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" {
		return ErrConflict
	}
	return fmt.Errorf("%s: %w", operation, err)
}
