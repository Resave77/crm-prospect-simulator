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
	Timezone           string
	City               *string
	Province           *string
	District           *string
	JobTitle           *string
	PositionGrade      *string
	SubDepartment      *string
	JoinDate           *time.Time
	Gender             *string
	DateOfBirth        *time.Time
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
	Timezone           string             `json:"timezone,omitempty"`
	City               *string            `json:"city,omitempty"`
	Province           *string            `json:"province,omitempty"`
	District           *string            `json:"district,omitempty"`
	JobTitle           *string            `json:"jobTitle,omitempty"`
	PositionGrade      *string            `json:"positionGrade,omitempty"`
	SubDepartment      *string            `json:"subDepartment,omitempty"`
	JoinDate           *time.Time         `json:"joinDate,omitempty"`
	Gender             *string            `json:"gender,omitempty"`
	DateOfBirth        *time.Time         `json:"dateOfBirth,omitempty"`
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
		Timezone: u.Timezone, City: u.City, Province: u.Province, District: u.District,
		JobTitle: u.JobTitle, PositionGrade: u.PositionGrade, SubDepartment: u.SubDepartment,
		JoinDate: u.JoinDate, Gender: u.Gender, DateOfBirth: u.DateOfBirth,
	}
}
