package model

import (
	"encoding/json"
	"fmt"
	"time"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
)

type ListFilter struct {
	Page      int
	Limit     int
	Search    string
	Role      string
	Status    string
	ManagerID string
}

type UserListItem struct {
	ID                 uuid.UUID                  `json:"id"`
	EmployeeID         string                     `json:"employeeId"`
	FullName           string                     `json:"fullName"`
	Email              string                     `json:"email"`
	Phone              string                     `json:"phone"`
	Role               authmodel.Role             `json:"role"`
	Status             authmodel.UserStatus       `json:"status"`
	ManagerID          *uuid.UUID                 `json:"managerId"`
	ManagerName        string                     `json:"managerName"`
	OrganizationalRole *OrganizationalRoleSummary `json:"organizationalRole"`
	MustChangePassword bool                       `json:"mustChangePassword"`
	CreatedAt          time.Time                  `json:"createdAt"`
	UpdatedAt          time.Time                  `json:"updatedAt"`
}

type UserListResult struct {
	Items []UserListItem `json:"data"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
	Pages int            `json:"pages"`
}

type UserDetail struct {
	ID                 uuid.UUID                  `json:"id"`
	EmployeeID         string                     `json:"employeeId"`
	FullName           string                     `json:"fullName"`
	Email              string                     `json:"email"`
	Phone              string                     `json:"phone"`
	Role               authmodel.Role             `json:"role"`
	Status             authmodel.UserStatus       `json:"status"`
	ManagerID          *uuid.UUID                 `json:"managerId"`
	ManagerName        string                     `json:"managerName"`
	OrganizationalRole *OrganizationalRoleSummary `json:"organizationalRole"`
	MustChangePassword bool                       `json:"mustChangePassword"`
	CreatedBy          *uuid.UUID                 `json:"createdBy"`
	UpdatedBy          *uuid.UUID                 `json:"updatedBy"`
	CreatedAt          time.Time                  `json:"createdAt"`
	UpdatedAt          time.Time                  `json:"updatedAt"`
}

type OrganizationalRoleSummary struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Level           int       `json:"level"`
	LandingPage     *string   `json:"landingPage"`
	PermissionCount int       `json:"permissionCount"`
	IsActive        bool      `json:"isActive"`
	Description     string    `json:"description,omitempty"`
}

type ManagerOption struct {
	ID         uuid.UUID `json:"id"`
	EmployeeID string    `json:"employeeId"`
	FullName   string    `json:"name"`
	Email      string    `json:"email"`
}

type AccountType string

const (
	AccountTypeSuperAdmin   AccountType = "SUPER_ADMIN"
	AccountTypeSalesAccount AccountType = "SALES_ACCOUNT"
)

type CreateUserInput struct {
	EmployeeID        string         `json:"employeeId"`
	FullName          string         `json:"name"`
	Email             string         `json:"email"`
	Phone             string         `json:"phone"`
	AccountType       AccountType    `json:"accountType"`
	Role              authmodel.Role `json:"role"`
	SalesRoleID       *uuid.UUID     `json:"salesRoleId"`
	ManagerID         *uuid.UUID     `json:"managerId"`
	TemporaryPassword string         `json:"temporaryPassword"`
}

// OptionalUUID tracks whether a JSON field was present so a PATCH request
// can distinguish an omitted managerId from an explicit null (clear).
type OptionalUUID struct {
	Present bool
	Value   *uuid.UUID
}

func (o *OptionalUUID) UnmarshalJSON(data []byte) error {
	o.Present = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("value must be a UUID or null")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return fmt.Errorf("value must be a valid UUID or null")
	}
	o.Value = &id
	return nil
}

func (o OptionalUUID) MarshalJSON() ([]byte, error) {
	if o.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(o.Value)
}

type UpdateUserInput struct {
	EmployeeID  *string         `json:"employeeId"`
	FullName    *string         `json:"name"`
	Email       *string         `json:"email"`
	Phone       *string         `json:"phone"`
	AccountType *AccountType    `json:"accountType"`
	Role        *authmodel.Role `json:"role"`
	SalesRoleID OptionalUUID    `json:"salesRoleId"`
	ManagerID   OptionalUUID    `json:"managerId"`
}

type UpdateStatusInput struct {
	Status authmodel.UserStatus `json:"status"`
}

type ResetPasswordMode string

const (
	ResetPasswordModeAuto   ResetPasswordMode = "AUTO"
	ResetPasswordModeManual ResetPasswordMode = "MANUAL"
)

type ResetPasswordInput struct {
	Mode              ResetPasswordMode `json:"mode"`
	TemporaryPassword string            `json:"temporaryPassword,omitempty"`
}

type ResetPasswordResult struct {
	UserID             uuid.UUID `json:"userId"`
	TemporaryPassword  string    `json:"temporaryPassword"`
	MustChangePassword bool      `json:"mustChangePassword"`
	SessionsRevoked    int64     `json:"sessionsRevoked"`
}
