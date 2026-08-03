package repository

import (
	"context"
	"errors"
	"time"

	"crm-prospect-simulator/backend/internal/admin/model"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("user not found")
	ErrConflict       = errors.New("record conflict")
	ErrSelfDeactivate = errors.New("cannot deactivate yourself")
	ErrRoleInUse      = errors.New("sales role in use")
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
	ListSalesRoles(ctx context.Context) ([]model.SalesRole, error)
	FindSalesRole(ctx context.Context, id uuid.UUID) (model.SalesRole, error)
	CreateSalesRole(ctx context.Context, id uuid.UUID, input model.CreateSalesRoleInput, actorID uuid.UUID) error
	UpdateSalesRole(ctx context.Context, id uuid.UUID, input model.UpdateSalesRoleInput, actorID uuid.UUID) error
	UpdateSalesRoleStatus(ctx context.Context, id uuid.UUID, isActive bool, actorID uuid.UUID) error
	DeleteSalesRole(ctx context.Context, id uuid.UUID) error
	SalesRoleNameExists(ctx context.Context, normalizedName string, excludeID *uuid.UUID) (bool, error)
	SalesRoleHasAssignments(ctx context.Context, id uuid.UUID) (bool, error)
	CreateSalesAssignment(ctx context.Context, id uuid.UUID, input model.CreateAssignmentInput, actorID uuid.UUID) error
	MoveSalesAssignment(ctx context.Context, currentID uuid.UUID, newID uuid.UUID, input model.MoveAssignmentInput, actorID uuid.UUID) error
	FindSalesAssignment(ctx context.Context, id uuid.UUID) (model.SalesStructureAssignment, error)
	FindEffectiveSalesAssignment(ctx context.Context, userID uuid.UUID, effectiveDate time.Time) (model.SalesStructureAssignment, model.SalesRole, error)
	SalesAssignmentOverlaps(ctx context.Context, userID uuid.UUID, from time.Time, to *time.Time, excludeID *uuid.UUID) (bool, error)
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
	ListSalesStructure(ctx context.Context, effectiveDate time.Time) ([]model.SalesStructureItem, error)
	ListSalesAssignmentHistory(ctx context.Context, userID uuid.UUID) ([]model.AssignmentHistoryItem, error)
}
