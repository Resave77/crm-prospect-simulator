package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"crm-prospect-simulator/backend/internal/auth/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	Role         model.Role `json:"role"`
	SessionID    uuid.UUID  `json:"sid"`
	TokenVersion int        `json:"ver"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret    []byte
	issuer    string
	audience  string
	accessTTL time.Duration
}

func NewTokenManager(secret, issuer, audience string, accessTTL time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), issuer: issuer, audience: audience, accessTTL: accessTTL}
}

func (m *TokenManager) IssueAccess(user model.User, sessionID uuid.UUID, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(m.accessTTL)
	claims := AccessClaims{
		Role: user.Role, SessionID: sessionID, TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: m.issuer, Subject: user.ID.String(), Audience: jwt.ClaimStrings{m.audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt), NotBefore: jwt.NewNumericDate(now),
			IssuedAt: jwt.NewNumericDate(now), ID: uuid.NewString(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return token, expiresAt, nil
}

func (m *TokenManager) ParseAccess(raw string) (AccessClaims, error) {
	claims := AccessClaims{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience), jwt.WithLeeway(30*time.Second))
	if err != nil || !token.Valid || !claims.Role.Valid() || claims.Subject == "" || claims.SessionID == uuid.Nil {
		return AccessClaims{}, ErrInvalidToken
	}
	return claims, nil
}

func newRefreshCredential() (uuid.UUID, string, string, error) {
	id := uuid.New()
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return uuid.Nil, "", "", fmt.Errorf("generate refresh credential: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(secret)
	return id, id.String() + "." + encoded, hashRefreshSecret(encoded), nil
}

func parseRefreshCredential(raw string) (uuid.UUID, string, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || parts[1] == "" {
		return uuid.Nil, "", ErrInvalidToken
	}
	id, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	return id, parts[1], nil
}

func hashRefreshSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func refreshSecretMatches(secret, expectedHash string) bool {
	actual := hashRefreshSecret(secret)
	return len(actual) == len(expectedHash) && subtle.ConstantTimeCompare([]byte(actual), []byte(expectedHash)) == 1
}
