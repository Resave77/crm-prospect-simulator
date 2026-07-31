package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdministrator  Role = "ADMINISTRATOR"
	RoleSalesManager   Role = "SALES_MANAGER"
	RoleSalesExecutive Role = "SALES_EXECUTIVE"
)

func (r Role) Valid() bool {
	return r == RoleAdministrator || r == RoleSalesManager || r == RoleSalesExecutive
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
	CreatedBy          *uuid.UUID
	UpdatedBy          *uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type PublicUser struct {
	ID                 uuid.UUID  `json:"id"`
	Email              string     `json:"email"`
	FullName           string     `json:"fullName"`
	EmployeeID         string     `json:"employeeId"`
	Phone              string     `json:"phone"`
	Role               Role       `json:"role"`
	MustChangePassword bool       `json:"mustChangePassword"`
	ManagerID          *uuid.UUID `json:"managerId"`
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
	}
}
