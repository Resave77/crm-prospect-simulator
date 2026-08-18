package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/auth/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userRepositoryStub struct {
	user              model.User
	changePasswordErr error
	changedHash       string
	changedUserID     uuid.UUID
	revoked           int
}

func (r *userRepositoryStub) FindByEmail(context.Context, string) (model.User, error) {
	return r.user, nil
}
func (r *userRepositoryStub) FindUserByID(context.Context, uuid.UUID) (model.User, error) {
	return r.user, nil
}
func (r *userRepositoryStub) RecordLogin(context.Context, uuid.UUID, time.Time) error { return nil }
func (r *userRepositoryStub) UpsertSeed(context.Context, model.User) error            { return nil }
func (r *userRepositoryStub) ChangePassword(_ context.Context, userID uuid.UUID, hash string, _ string, _ time.Time) (int, error) {
	if r.changePasswordErr != nil {
		return 0, r.changePasswordErr
	}
	r.changedUserID = userID
	r.changedHash = hash
	return r.revoked, nil
}

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

func TestLoginHandlesNullableUserProfileFields(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := model.User{
		ID: uuid.New(), Email: "manager@yummy.test", PasswordHash: string(hash),
		FullName: "Sales Manager", Role: model.RoleSalesManager, Status: model.UserActive, TokenVersion: 1,
		EmployeeID: "", Phone: "", ManagerID: nil, CreatedBy: nil, UpdatedBy: nil, LastLoginAt: nil,
	}
	users := &userRepositoryStub{user: user}
	sessions := &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}}
	tokens := NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute)
	auth := NewAuthService(users, sessions, tokens, time.Hour)

	result, err := auth.Login(context.Background(), user.Email, "password123", ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if result.User.Role != model.RoleSalesManager {
		t.Fatalf("role=%s, want %s", result.User.Role, model.RoleSalesManager)
	}
	if result.User.EmployeeID != "" || result.User.Phone != "" || result.User.ManagerID != nil || result.User.MustChangePassword {
		t.Fatalf("unexpected public nullable fields: %+v", result.User)
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

func TestChangePasswordSuccess(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
	user := model.User{
		ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash),
		Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1, MustChangePassword: true,
	}
	users := &userRepositoryStub{user: user, revoked: 2}
	sessions := &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}}
	auth := NewAuthService(users, sessions, NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	res, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: user.Role}, "OldPass123", "NewPass456", "NewPass456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.PasswordChanged || res.MustChangePassword || res.ReauthenticationRequired || res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if res.SessionsRevoked != 2 {
		t.Fatalf("sessionsRevoked=%d, want 2", res.SessionsRevoked)
	}
}

func TestChangePasswordIncorrectCurrent(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
	user := model.User{
		ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash),
		Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1,
	}
	users := &userRepositoryStub{user: user}
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	_, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: user.Role}, "WrongPass123", "NewPass456", "NewPass456")
	if err != ErrInvalidCredentials {
		t.Fatalf("err=%v, want ErrInvalidCredentials", err)
	}
}

func TestChangePasswordMinimumLength(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
	user := model.User{
		ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash),
		Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1,
	}
	users := &userRepositoryStub{user: user}
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	for _, weak := range []string{"short", "12345", "abcde"} {
		_, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: user.Role}, "OldPass123", weak, weak)
		if err != ErrPasswordTooWeak {
			t.Fatalf("err=%v, want ErrPasswordTooWeak for %q", err, weak)
		}
	}
}

func TestChangePasswordAcceptsSixCharacterPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
	user := model.User{ID: uuid.New(), PasswordHash: string(hash), Role: model.RoleSalesExecutive, Status: model.UserActive}
	users := &userRepositoryStub{user: user}
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	if _, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: user.Role}, "OldPass123", "abcdef", "abcdef"); err != nil {
		t.Fatalf("six-character password rejected: %v", err)
	}
}

func TestChangePasswordMismatch(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
	user := model.User{
		ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash),
		Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1,
	}
	users := &userRepositoryStub{user: user}
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	_, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: user.Role}, "OldPass123", "NewPass456", "Different789")
	if err == nil {
		t.Fatalf("expected error for password mismatch")
	}
}

func TestChangePasswordSame(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
	user := model.User{
		ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash),
		Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1,
	}
	users := &userRepositoryStub{user: user}
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	_, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: user.Role}, "OldPass123", "OldPass123", "OldPass123")
	if err == nil {
		t.Fatalf("expected error when new password equals current password")
	}
}

func TestChangePasswordMissingFields(t *testing.T) {
	auth := NewAuthService(&userRepositoryStub{}, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	cases := [][3]string{
		{"", "NewPass456", "NewPass456"},
		{"OldPass123", "", "NewPass456"},
		{"OldPass123", "NewPass456", ""},
	}
	for _, tc := range cases {
		_, err := auth.ChangePassword(context.Background(), Principal{UserID: uuid.New()}, tc[0], tc[1], tc[2])
		if err != ErrMissingFields {
			t.Fatalf("err=%v, want ErrMissingFields", err)
		}
	}
}

func TestChangePasswordStoresBcryptHashForNewPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
	user := model.User{ID: uuid.New(), PasswordHash: string(hash), Role: model.RoleSalesExecutive, Status: model.UserActive}
	users := &userRepositoryStub{user: user}
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	_, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: user.Role}, "OldPass123", "NewPass456", "NewPass456")
	if err != nil {
		t.Fatal(err)
	}
	if users.changedHash == "NewPass456" {
		t.Fatal("stored raw password, want bcrypt hash")
	}
	if bcrypt.CompareHashAndPassword([]byte(users.changedHash), []byte("NewPass456")) != nil {
		t.Fatal("stored hash does not match new password")
	}
}

func TestChangePasswordAllowsEveryRole(t *testing.T) {
	for _, role := range []model.Role{model.RoleAdministrator, model.RoleSalesManager, model.RoleSalesExecutive} {
		hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
		user := model.User{ID: uuid.New(), PasswordHash: string(hash), Role: role, Status: model.UserActive}
		users := &userRepositoryStub{user: user}
		auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
			NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

		if _, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: role}, "OldPass123", "NewPass456", "NewPass456"); err != nil {
			t.Fatalf("role %s err=%v, want nil", role, err)
		}
	}
}

func TestChangePasswordRollbackBehaviorSurfacesRepositoryFailure(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.MinCost)
	user := model.User{ID: uuid.New(), PasswordHash: string(hash), Role: model.RoleAdministrator, Status: model.UserActive}
	persistErr := errors.New("rollback")
	users := &userRepositoryStub{user: user, changePasswordErr: persistErr}
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}},
		NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute), time.Hour)

	result, err := auth.ChangePassword(context.Background(), Principal{UserID: user.ID, Role: user.Role}, "OldPass123", "NewPass456", "NewPass456")
	if !errors.Is(err, persistErr) {
		t.Fatalf("err=%v, want repository failure", err)
	}
	if result.PasswordChanged {
		t.Fatalf("passwordChanged=%v, want false", result.PasswordChanged)
	}
}
func TestAuthenticateAccessIncludesMustChangePasswordFromUserRow(t *testing.T) {
	user := model.User{ID: uuid.New(), Email: "admin@yummy.test", Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1, MustChangePassword: true}
	users := &userRepositoryStub{user: user}
	tokens := NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute)
	auth := NewAuthService(users, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}}, tokens, time.Hour)
	access, _, err := tokens.IssueAccess(user, uuid.New(), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	principal, err := auth.AuthenticateAccess(context.Background(), access)
	if err != nil {
		t.Fatal(err)
	}
	if !principal.MustChangePassword {
		t.Fatal("principal.MustChangePassword=false, want true from user row")
	}
}

func TestLoginAndRefreshReturnMustChangePassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := model.User{ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash), FullName: "Administrator", Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1, MustChangePassword: true}
	users := &userRepositoryStub{user: user}
	sessions := &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}}
	tokens := NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute)
	auth := NewAuthService(users, sessions, tokens, time.Hour)

	login, err := auth.Login(context.Background(), user.Email, "password123", ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !login.User.MustChangePassword {
		t.Fatal("login mustChangePassword=false, want true from user row")
	}
	refreshed, err := auth.Refresh(context.Background(), login.RefreshToken, ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !refreshed.User.MustChangePassword {
		t.Fatal("refresh mustChangePassword=false, want true from user row")
	}
}

func TestLoginAndRefreshReturnSuperAdminRole(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := model.User{ID: uuid.New(), Email: "super@yummy.test", PasswordHash: string(hash), FullName: "Super Admin", Role: model.RoleSuperAdmin, Status: model.UserActive, TokenVersion: 1}
	users := &userRepositoryStub{user: user}
	sessions := &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}}
	tokens := NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute)
	auth := NewAuthService(users, sessions, tokens, time.Hour)

	login, err := auth.Login(context.Background(), user.Email, "password123", ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if login.User.Role != model.RoleSuperAdmin {
		t.Fatalf("login role=%s, want SUPER_ADMIN", login.User.Role)
	}
	refreshed, err := auth.Refresh(context.Background(), login.RefreshToken, ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.User.Role != model.RoleSuperAdmin {
		t.Fatalf("refresh role=%s, want SUPER_ADMIN", refreshed.User.Role)
	}
}

func TestAuthenticateAccessRejectsInactiveUserBeforePasswordChangeGuard(t *testing.T) {
	user := model.User{ID: uuid.New(), Email: "inactive@yummy.test", Role: model.RoleAdministrator, Status: model.UserActive, TokenVersion: 1, MustChangePassword: true}
	tokens := NewTokenManager("01234567890123456789012345678901", "test", "test-api", time.Minute)
	access, _, err := tokens.IssueAccess(user, uuid.New(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	user.Status = model.UserInactive
	auth := NewAuthService(&userRepositoryStub{user: user}, &sessionRepositoryStub{sessions: map[uuid.UUID]model.RefreshSession{}}, tokens, time.Hour)

	if _, err := auth.AuthenticateAccess(context.Background(), access); err != ErrInvalidToken {
		t.Fatalf("err=%v, want ErrInvalidToken", err)
	}
}
