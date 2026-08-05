package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/auth/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ClientContext struct {
	UserAgent string
	IPAddress string
}

type AuthResult struct {
	AccessToken      string           `json:"accessToken"`
	AccessExpiresAt  time.Time        `json:"accessExpiresAt"`
	RefreshToken     string           `json:"-"`
	RefreshExpiresAt time.Time        `json:"-"`
	User             model.PublicUser `json:"user"`
}

type ChangePasswordResult struct {
	PasswordChanged          bool `json:"passwordChanged"`
	MustChangePassword       bool `json:"mustChangePassword"`
	SessionsRevoked          int  `json:"sessionsRevoked"`
	ReauthenticationRequired bool `json:"reauthenticationRequired"`
}

type Principal struct {
	UserID             uuid.UUID
	Role               model.Role
	SessionID          uuid.UUID
	TokenVersion       int
	MustChangePassword bool
}

type AuthService struct {
	users      repository.UserRepository
	sessions   repository.SessionRepository
	tokens     *TokenManager
	refreshTTL time.Duration
	now        func() time.Time
}

func NewAuthService(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	tokens *TokenManager,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:      users,
		sessions:   sessions,
		tokens:     tokens,
		refreshTTL: refreshTTL,
		now:        time.Now,
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
	client ClientContext,
) (AuthResult, error) {
	user, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	if user.Status != model.UserActive {
		return AuthResult{}, ErrUserInactive
	}

	return s.createLogin(ctx, user, client, uuid.Nil)
}

func (s *AuthService) Refresh(
	ctx context.Context,
	rawRefresh string,
	client ClientContext,
) (AuthResult, error) {
	sessionID, secret, err := parseRefreshCredential(rawRefresh)
	if err != nil {
		return AuthResult{}, ErrInvalidToken
	}

	session, err := s.sessions.FindSessionByID(ctx, sessionID)
	if err != nil || !refreshSecretMatches(secret, session.TokenHash) {
		return AuthResult{}, ErrInvalidToken
	}

	now := s.now().UTC()
	if session.RevokedAt != nil {
		if session.RevokeReason == "ROTATED" {
			_ = s.sessions.RevokeAllForUser(ctx, session.UserID, "ROTATION_REUSE", now)
		}
		return AuthResult{}, ErrInvalidToken
	}

	if !session.Active(now) {
		return AuthResult{}, ErrSessionExpired
	}

	user, err := s.users.FindUserByID(ctx, session.UserID)
	if err != nil {
		return AuthResult{}, ErrInvalidToken
	}

	if user.Status != model.UserActive {
		return AuthResult{}, ErrUserInactive
	}

	return s.createLogin(ctx, user, client, session.ID)
}

func (s *AuthService) createLogin(
	ctx context.Context,
	user model.User,
	client ClientContext,
	replaceSessionID uuid.UUID,
) (AuthResult, error) {
	// TEMP DEMO: Mandatory first-login password change is temporarily disabled for the
	// demo/progress deployment. This modifies only the in-memory auth response
	// and access-token source; the database column and ChangePassword feature
	// remain intact for later reactivation.
	user.MustChangePassword = false

	now := s.now().UTC()
	sessionID, refreshToken, refreshHash, err := newRefreshCredential()
	if err != nil {
		return AuthResult{}, err
	}

	refreshExpiresAt := now.Add(s.refreshTTL)
	session := model.RefreshSession{
		ID:        sessionID,
		UserID:    user.ID,
		TokenHash: refreshHash,
		UserAgent: client.UserAgent,
		IPAddress: client.IPAddress,
		ExpiresAt: refreshExpiresAt,
	}

	if replaceSessionID == uuid.Nil {
		err = s.sessions.Create(ctx, session)
	} else {
		err = s.sessions.Rotate(ctx, replaceSessionID, session, now)
	}
	if err != nil {
		return AuthResult{}, fmt.Errorf("persist refresh session: %w", err)
	}

	accessToken, accessExpiresAt, err := s.tokens.IssueAccess(user, sessionID, now)
	if err != nil {
		_ = s.sessions.Revoke(ctx, sessionID, "ACCESS_TOKEN_FAILURE", now)
		return AuthResult{}, err
	}

	if replaceSessionID == uuid.Nil {
		_ = s.users.RecordLogin(ctx, user.ID, now)
	}

	return AuthResult{
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
		User:             user.Public(),
	}, nil
}

func (s *AuthService) AuthenticateAccess(
	ctx context.Context,
	rawToken string,
) (Principal, error) {
	claims, err := s.tokens.ParseAccess(rawToken)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}

	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil ||
		user.Status != model.UserActive ||
		user.Role != claims.Role ||
		user.TokenVersion != claims.TokenVersion {
		return Principal{}, ErrInvalidToken
	}

	return Principal{
		UserID:       user.ID,
		Role:         user.Role,
		SessionID:    claims.SessionID,
		TokenVersion: user.TokenVersion,

		// TEMP DEMO: Mandatory first-login password enforcement is disabled.
		MustChangePassword: false,
	}, nil
}

func (s *AuthService) Me(
	ctx context.Context,
	principal Principal,
) (model.PublicUser, error) {
	user, err := s.users.FindUserByID(ctx, principal.UserID)
	if err != nil {
		return model.PublicUser{}, ErrInvalidToken
	}

	// TEMP DEMO: Keep frontend auth state aligned with disabled mandatory first-login enforcement.
	user.MustChangePassword = false
	return user.Public(), nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefresh string) error {
	sessionID, _, err := parseRefreshCredential(rawRefresh)
	if err != nil {
		return nil
	}

	return s.sessions.Revoke(ctx, sessionID, "LOGOUT", s.now().UTC())
}

func (s *AuthService) LogoutAll(ctx context.Context, principal Principal) error {
	return s.sessions.RevokeAllForUser(
		ctx,
		principal.UserID,
		"LOGOUT_ALL",
		s.now().UTC(),
	)
}

func (s *AuthService) ChangePassword(
	ctx context.Context,
	principal Principal,
	currentPassword string,
	newPassword string,
	confirmPassword string,
) (ChangePasswordResult, error) {
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		return ChangePasswordResult{}, ErrMissingFields
	}

	if newPassword != confirmPassword {
		return ChangePasswordResult{}, ErrPasswordMismatch
	}

	if !strongPassword(newPassword) {
		return ChangePasswordResult{}, ErrPasswordTooWeak
	}

	user, err := s.users.FindUserByID(ctx, principal.UserID)
	if err != nil || user.Status != model.UserActive {
		return ChangePasswordResult{}, ErrInvalidToken
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)) != nil {
		return ChangePasswordResult{}, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword)) == nil {
		return ChangePasswordResult{}, ErrPasswordSame
	}

	newHash, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return ChangePasswordResult{}, fmt.Errorf("hash new password: %w", err)
	}

	revoked, err := s.users.ChangePassword(
		ctx,
		user.ID,
		string(newHash),
		"PASSWORD_CHANGED",
		s.now().UTC(),
	)
	if err != nil {
		return ChangePasswordResult{}, err
	}

	return ChangePasswordResult{
		PasswordChanged:          true,
		MustChangePassword:       false,
		SessionsRevoked:          revoked,
		ReauthenticationRequired: true,
	}, nil
}

func strongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var upper, lower, digit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		}
	}

	return upper && lower && digit
}

func IsClientAuthError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrSessionExpired) ||
		errors.Is(err, ErrUserInactive)
}
