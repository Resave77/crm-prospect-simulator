package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleSuperAdmin     Role = "SUPER_ADMIN"
	RoleAdministrator  Role = "ADMINISTRATOR"
	RoleSalesManager   Role = "SALES_MANAGER"
	RoleSalesExecutive Role = "SALES_EXECUTIVE"
)

func (r Role) Valid() bool {
	return r == RoleSuperAdmin || r == RoleAdministrator || r == RoleSalesManager || r == RoleSalesExecutive
}

func (r Role) IsAdminRole() bool {
	return r == RoleSuperAdmin || r == RoleAdministrator
}

type UserStatus string

const (
	UserActive   UserStatus = "ACTIVE"
	UserInactive UserStatus = "INACTIVE"
)

type User struct {
	ID                 uuid.UUID
	Email              string
	PasswordHash       string
	FullName           string
	EmployeeID         string
	Phone              string
	Role               Role
	Status             UserStatus
	TokenVersion       int
	LastLoginAt        *time.Time
	MustChangePassword bool
	ManagerID          *uuid.UUID
	SalesRole          *SalesRoleSummary
	CreatedBy          *uuid.UUID
	UpdatedBy          *uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SalesRoleSummary struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Level          int       `json:"level"`
	LandingPage    *string   `json:"landingPage"`
	PermissionKeys []string  `json:"permissionKeys,omitempty"`
}

type PublicUser struct {
	ID                 uuid.UUID         `json:"id"`
	Email              string            `json:"email"`
	FullName           string            `json:"fullName"`
	EmployeeID         string            `json:"employeeId"`
	Phone              string            `json:"phone"`
	Role               Role              `json:"role"`
	MustChangePassword bool              `json:"mustChangePassword"`
	ManagerID          *uuid.UUID        `json:"managerId"`
	SalesRole          *SalesRoleSummary `json:"salesRole"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		ID:                 u.ID,
		Email:              u.Email,
		FullName:           u.FullName,
		EmployeeID:         u.EmployeeID,
		Phone:              u.Phone,
		Role:               u.Role,
		MustChangePassword: u.MustChangePassword,
		ManagerID:          u.ManagerID,
		SalesRole:          u.SalesRole,
	}
}
