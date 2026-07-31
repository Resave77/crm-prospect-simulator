package repository

import (
	"context"
	"errors"

	"crm-prospect-simulator/backend/internal/admin/model"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("user not found")
	ErrConflict       = errors.New("record conflict")
	ErrSelfDeactivate = errors.New("cannot deactivate yourself")
)

type Repository interface {
	ListUsers(ctx context.Context, filter model.ListFilter) (model.UserListResult, error)
	FindUserDetail(ctx context.Context, id uuid.UUID) (model.UserDetail, error)
	CreateUser(ctx context.Context, id uuid.UUID, input model.CreateUserInput, passwordHash string, actorID uuid.UUID) error
	UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput, actorID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status authmodel.UserStatus, actorID uuid.UUID) error
	ListActiveManagers(ctx context.Context) ([]model.ManagerOption, error)
	ExistsByEmail(ctx context.Context, email string, excludeID *uuid.UUID) (bool, error)
	ExistsByEmployeeID(ctx context.Context, employeeID string, excludeID *uuid.UUID) (bool, error)
	FindManagerByID(ctx context.Context, id uuid.UUID) (authmodel.User, error)
	CountActiveAdministrators(ctx context.Context) (int, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (authmodel.User, error)
	ResetPassword(ctx context.Context, targetUserID uuid.UUID, actorUserID uuid.UUID, passwordHash string) (int64, error)
}
