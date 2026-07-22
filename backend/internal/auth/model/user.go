package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdministrator  Role = "ADMINISTRATOR"
	RoleSalesExecutive Role = "SALES_EXECUTIVE"
)

func (r Role) Valid() bool {
	return r == RoleAdministrator || r == RoleSalesExecutive
}

type UserStatus string

const (
	UserActive   UserStatus = "ACTIVE"
	UserInactive UserStatus = "INACTIVE"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	Role         Role
	Status       UserStatus
	TokenVersion int
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PublicUser struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"fullName"`
	Role     Role      `json:"role"`
}

func (u User) Public() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role}
}
