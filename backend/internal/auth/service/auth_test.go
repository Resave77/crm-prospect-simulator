package service

import (
	"context"
	"testing"
	"time"

	"crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/auth/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userRepositoryStub struct {
	user model.User
}

func (r *userRepositoryStub) FindByEmail(context.Context, string) (model.User, error) {
	return r.user, nil
}
func (r *userRepositoryStub) FindUserByID(context.Context, uuid.UUID) (model.User, error) {
	return r.user, nil
}
func (r *userRepositoryStub) RecordLogin(context.Context, uuid.UUID, time.Time) error { return nil }
func (r *userRepositoryStub) UpsertSeed(context.Context, model.User) error            { return nil }

type sessionRepositoryStub struct {
	sessions map[uuid.UUID]model.RefreshSession
}

func (r *sessionRepositoryStub) Create(_ context.Context, session model.RefreshSession) error {
	r.sessions[session.ID] = session
	return nil
}
func (r *sessionRepositoryStub) FindSessionByID(_ context.Context, id uuid.UUID) (model.RefreshSession, error) {
	session, ok := r.sessions[id]
	if !ok {
		return model.RefreshSession{}, repository.ErrNotFound
	}
	return session, nil
}
func (r *sessionRepositoryStub) Rotate(_ context.Context, oldID uuid.UUID, replacement model.RefreshSession, at time.Time) error {
	current := r.sessions[oldID]
	current.RevokedAt = &at
	r.sessions[oldID] = current
	r.sessions[replacement.ID] = replacement
	return nil
}
func (r *sessionRepositoryStub) Revoke(_ context.Context, id uuid.UUID, reason string, at time.Time) error {
	current := r.sessions[id]
	current.RevokedAt = &at
	current.RevokeReason = reason
	r.sessions[id] = current
	return nil
}
func (r *sessionRepositoryStub) RevokeAllForUser(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}

func TestLoginUsesBcryptAndIssuesSession(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := model.User{
		ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash),
		FullName: "Administrator", Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1,
	}
	users := &userRepositoryStub{user: user}
	sessions := &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}}
	tokens := NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute)
	auth := NewAuthService(users, sessions, tokens, time.Hour)

	result, err := auth.Login(context.Background(), user.Email, "password123", ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken == "" || result.RefreshToken == "" || result.User.Role != model.RoleAdministrator {
		t.Fatalf("incomplete authentication result: %+v", result)
	}
	if len(sessions.sessions) != 1 {
		t.Fatalf("session count=%d, want 1", len(sessions.sessions))
	}
}

func TestLoginRejectsIncorrectBcryptPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	users := &userRepositoryStub{user: model.User{
		ID: uuid.New(), Email: "sales@yummy.test", PasswordHash: string(hash),
		Role: model.RoleSalesExecutive, Status: model.UserActive,
	}}
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	if _, err := auth.Login(context.Background(), users.user.Email, "wrong-password", ClientContext{}); err != ErrInvalidCredentials {
		t.Fatalf("err=%v, want ErrInvalidCredentials", err)
	}
}
