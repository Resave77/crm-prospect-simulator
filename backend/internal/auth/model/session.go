package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	TokenHash           string
	UserAgent           string
	IPAddress           string
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	RevokeReason        string
	ReplacedBySessionID *uuid.UUID
	CreatedAt           time.Time
}

func (s RefreshSession) Active(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}
