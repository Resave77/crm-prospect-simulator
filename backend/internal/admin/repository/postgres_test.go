package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"crm-prospect-simulator/backend/internal/admin/model"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func strPtr(s string) *string { return &s }

type mockTx struct {
	pgx.Tx
	execFunc func(sql string, args ...any) (int64, error)
	calls    []string
}

func (m *mockTx) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	m.calls = append(m.calls, sql)
	rows := int64(1)
	if m.execFunc != nil {
		var err error
		rows, err = m.execFunc(sql, args...)
		if err != nil {
			return pgconn.CommandTag{}, err
		}
	}
	return pgconn.NewCommandTag("UPDATE " + itoa(int(rows))), nil
}

func TestUpdateSetsExplicitNullClearsManager(t *testing.T) {
	role := authmodel.RoleSalesManager
	input := model.UpdateUserInput{
		Role:      &role,
		ManagerID: model.OptionalUUID{Present: true, Value: nil},
	}
	sets, _ := updateSets(input, uuid.New())
	if !strings.Contains(sets, `manager_id = NULL`) {
		t.Fatalf("expected manager_id = NULL in SET, got: %s", sets)
	}
	if strings.Contains(sets, `manager_id = $`) {
		t.Fatalf("explicit null must not bind a manager argument, got: %s", sets)
	}
}

func TestUpdateSetsUUIDBindsManagerArgument(t *testing.T) {
	managerID := uuid.New()
	input := model.UpdateUserInput{
		ManagerID: model.OptionalUUID{Present: true, Value: &managerID},
	}
	sets, args := updateSets(input, uuid.New())
	if !strings.Contains(sets, `manager_id = $2`) {
		t.Fatalf("expected manager_id = $2 in SET, got: %s", sets)
	}
	found := false
	for _, a := range args {
		if v, ok := a.(uuid.UUID); ok && v == managerID {
			found = true
		}
	}
	if !found {
		t.Fatalf("manager argument not bound: %v", args)
	}
}

func TestUpdateSetsOmittedManagerHasNoManagerClause(t *testing.T) {
	input := model.UpdateUserInput{FullName: strPtr("Updated")}
	sets, _ := updateSets(input, uuid.New())
	if strings.Contains(sets, "manager_id") {
		t.Fatalf("omitted managerId must not change manager_id, got: %s", sets)
	}
}

func TestUpdateSetsEmptyInputReturnsEmpty(t *testing.T) {
	sets, args := updateSets(model.UpdateUserInput{}, uuid.New())
	if sets != "" || len(args) != 0 {
		t.Fatalf("empty input should produce no updates, got sets=%q args=%v", sets, args)
	}
}

func TestResetPasswordUpdateStatement(t *testing.T) {
	if strings.Contains(resetPasswordUpdate, "COALESCE") {
		t.Fatalf("password_hash must not use COALESCE, got: %s", resetPasswordUpdate)
	}
	for _, want := range []string{
		`password_hash = $1`,
		`must_change_password = TRUE`,
		`token_version = token_version + 1`,
		`updated_by = $2`,
		`updated_at`,
		`WHERE id = $3`,
	} {
		if !strings.Contains(resetPasswordUpdate, want) {
			t.Fatalf("resetPasswordUpdate missing %q, got: %s", want, resetPasswordUpdate)
		}
	}
	for _, forbidden := range []string{"role", "status", "manager_id", "employee_id"} {
		if strings.Contains(resetPasswordUpdate, forbidden) {
			t.Fatalf("resetPasswordUpdate must not touch %q, got: %s", forbidden, resetPasswordUpdate)
		}
	}
}

func TestRevokeActiveSessionsStatement(t *testing.T) {
	for _, want := range []string{
		`user_id = $1`,
		`revoked_at IS NULL`,
		`revoke_reason = 'ADMIN_PASSWORD_RESET'`,
	} {
		if !strings.Contains(revokeActiveSessions, want) {
			t.Fatalf("revokeActiveSessions missing %q, got: %s", want, revokeActiveSessions)
		}
	}
}

func TestResetPasswordTxUpdatesUserAndRevokesSessions(t *testing.T) {
	tx := &mockTx{execFunc: func(sql string, _ ...any) (int64, error) {
		if sql == revokeActiveSessions {
			return 0, nil
		}
		return 1, nil
	}}
	revoked, err := resetPasswordTx(context.Background(), tx, "hash", uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("resetPasswordTx: %v", err)
	}
	if revoked != 0 {
		t.Fatalf("revoked=%d, want 0", revoked)
	}
	if len(tx.calls) != 2 {
		t.Fatalf("expected 2 statements, got %d: %v", len(tx.calls), tx.calls)
	}
	if tx.calls[0] != resetPasswordUpdate {
		t.Fatalf("first statement must update the user, got: %s", tx.calls[0])
	}
	if tx.calls[1] != revokeActiveSessions {
		t.Fatalf("second statement must revoke sessions, got: %s", tx.calls[1])
	}
}

func TestResetPasswordTxCountsRevokedSessions(t *testing.T) {
	tx := &mockTx{execFunc: func(sql string, _ ...any) (int64, error) {
		if sql == revokeActiveSessions {
			return 3, nil
		}
		return 1, nil
	}}
	revoked, err := resetPasswordTx(context.Background(), tx, "hash", uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("resetPasswordTx: %v", err)
	}
	if revoked != 3 {
		t.Fatalf("revoked=%d, want 3", revoked)
	}
}

func TestResetPasswordTxZeroSessionsIsSuccess(t *testing.T) {
	tx := &mockTx{execFunc: func(sql string, _ ...any) (int64, error) {
		if sql == revokeActiveSessions {
			return 0, nil
		}
		return 1, nil
	}}
	revoked, err := resetPasswordTx(context.Background(), tx, "hash", uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("resetPasswordTx with no active sessions: %v", err)
	}
	if revoked != 0 {
		t.Fatalf("revoked=%d, want 0", revoked)
	}
}

func TestResetPasswordTxMissingUserReturnsNotFound(t *testing.T) {
	tx := &mockTx{execFunc: func(sql string, _ ...any) (int64, error) {
		if sql == resetPasswordUpdate {
			return 0, nil
		}
		return 1, nil
	}}
	_, err := resetPasswordTx(context.Background(), tx, "hash", uuid.New(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
	if len(tx.calls) != 1 {
		t.Fatalf("session revocation must not run when the user update affects no rows, got calls: %v", tx.calls)
	}
}

func TestResetPasswordTxPropagatesUpdateError(t *testing.T) {
	tx := &mockTx{execFunc: func(sql string, _ ...any) (int64, error) {
		if sql == resetPasswordUpdate {
			return 0, errors.New("constraint violation")
		}
		return 1, nil
	}}
	_, err := resetPasswordTx(context.Background(), tx, "hash", uuid.New(), uuid.New())
	if err == nil || !strings.Contains(err.Error(), "constraint violation") {
		t.Fatalf("err=%v, want wrapped constraint violation", err)
	}
}

func TestResetPasswordTxPropagatesRevokeError(t *testing.T) {
	tx := &mockTx{execFunc: func(sql string, _ ...any) (int64, error) {
		if sql == revokeActiveSessions {
			return 0, errors.New("session revoke failed")
		}
		return 1, nil
	}}
	_, err := resetPasswordTx(context.Background(), tx, "hash", uuid.New(), uuid.New())
	if err == nil || !strings.Contains(err.Error(), "session revoke failed") {
		t.Fatalf("err=%v, want wrapped revoke error", err)
	}
}
