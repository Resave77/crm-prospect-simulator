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
		SELECT id, email, password_hash, full_name, role::text, status::text,
		       token_version, last_login_at, created_at, updated_at
		FROM users WHERE email = $1`, email))
}

func (r *PostgresRepository) FindUserByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role::text, status::text,
		       token_version, last_login_at, created_at, updated_at
		FROM users WHERE id = $1`, id))
}

func (r *PostgresRepository) scanUser(row pgx.Row) (model.User, error) {
	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role,
		&user.Status, &user.TokenVersion, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("scan user: %w", err)
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
		INSERT INTO users (id, email, password_hash, full_name, role, status, token_version)
		VALUES ($1, $2, $3, $4, $5, $6, 1)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = EXCLUDED.role,
			status = EXCLUDED.status,
			updated_at = now()`,
		user.ID, user.Email, user.PasswordHash, user.FullName, user.Role, user.Status)
	if err != nil {
		return fmt.Errorf("seed user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Create(ctx context.Context, session model.RefreshSession) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_sessions
			(id, user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`, session.ID, session.UserID,
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
			(id, user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`, replacement.ID, replacement.UserID,
		replacement.TokenHash, replacement.UserAgent, replacement.IPAddress, replacement.ExpiresAt)
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
