package repository

import (
	"context"
	"errors"
	"time"

	"crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record conflict")
)

type UserRepository interface {
	FindByEmail(context.Context, string) (model.User, error)
	FindUserByID(context.Context, uuid.UUID) (model.User, error)
	RecordLogin(context.Context, uuid.UUID, time.Time) error
	UpsertSeed(context.Context, model.User) error
}

type SessionRepository interface {
	Create(context.Context, model.RefreshSession) error
	FindSessionByID(context.Context, uuid.UUID) (model.RefreshSession, error)
	Rotate(context.Context, uuid.UUID, model.RefreshSession, time.Time) error
	Revoke(context.Context, uuid.UUID, string, time.Time) error
	RevokeAllForUser(context.Context, uuid.UUID, string, time.Time) error
}
