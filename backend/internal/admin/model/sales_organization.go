package model

import (
	"encoding/json"
	"fmt"
	"time"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
)

const DateLayout = "2006-01-02"

type SalesRole struct {
	ID              uuid.UUID    `json:"id"`
	Name            string       `json:"name"`
	Level           int          `json:"level"`
	Description     string       `json:"description"`
	IsActive        bool         `json:"isActive"`
	LandingPage     *string      `json:"landingPage"`
	PermissionCount int          `json:"permissionCount,omitempty"`
	Permissions     []Permission `json:"permissions,omitempty"`
	CreatedBy       *uuid.UUID   `json:"createdBy"`
	UpdatedBy       *uuid.UUID   `json:"updatedBy"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
}

type CreateSalesRoleInput struct {
	Name           string   `json:"name"`
	Level          int      `json:"level"`
	Description    string   `json:"description"`
	LandingPage    *string  `json:"landingPage"`
	PermissionKeys []string `json:"permissionKeys"`
}

type UpdateSalesRoleInput struct {
	Name           *string  `json:"name"`
	Level          *int     `json:"level"`
	Description    *string  `json:"description"`
	LandingPage    *string  `json:"landingPage"`
	PermissionKeys []string `json:"permissionKeys"`
}

type UpdateSalesRoleStatusInput struct {
	IsActive bool `json:"isActive"`
}

type PermissionNodeType string

const (
	PermissionNodeGroup  PermissionNodeType = "GROUP"
	PermissionNodeMenu   PermissionNodeType = "MENU"
	PermissionNodeAction PermissionNodeType = "ACTION"
)

type Permission struct {
	ID          uuid.UUID          `json:"id"`
	Key         string             `json:"key"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	GroupKey    string             `json:"groupKey"`
	ParentKey   *string            `json:"parentKey"`
	NodeType    PermissionNodeType `json:"nodeType"`
	RoutePath   *string            `json:"routePath"`
	IsActive    bool               `json:"isActive"`
	SortOrder   int                `json:"sortOrder"`
	CreatedAt   time.Time          `json:"createdAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt,omitempty"`
}

type SalesStructureAssignment struct {
	ID            uuid.UUID  `json:"assignmentId"`
	UserID        uuid.UUID  `json:"userId"`
	SalesRoleID   uuid.UUID  `json:"salesRoleId"`
	ParentUserID  *uuid.UUID `json:"parentUserId"`
	EffectiveFrom time.Time  `json:"effectiveFrom"`
	EffectiveTo   *time.Time `json:"effectiveTo"`
}

type SalesStructureDate struct {
	time.Time
}

func (d *SalesStructureDate) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("date must be YYYY-MM-DD")
	}
	parsed, err := time.Parse(DateLayout, raw)
	if err != nil {
		return fmt.Errorf("date must be YYYY-MM-DD")
	}
	d.Time = parsed
	return nil
}

func (d SalesStructureDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format(DateLayout))
}

type CreateAssignmentInput struct {
	UserID        uuid.UUID          `json:"userId"`
	SalesRoleID   uuid.UUID          `json:"salesRoleId"`
	ParentUserID  *uuid.UUID         `json:"parentUserId"`
	EffectiveFrom SalesStructureDate `json:"effectiveFrom"`
}

type MoveAssignmentInput struct {
	SalesRoleID   uuid.UUID          `json:"salesRoleId"`
	ParentUserID  *uuid.UUID         `json:"parentUserId"`
	EffectiveFrom SalesStructureDate `json:"effectiveFrom"`
}

type EndAssignmentInput struct {
	EffectiveTo SalesStructureDate `json:"effectiveTo"`
}

type SalesStructureItem struct {
	AssignmentID  uuid.UUID      `json:"assignmentId"`
	UserID        uuid.UUID      `json:"userId"`
	SalesName     string         `json:"salesName"`
	SystemRole    authmodel.Role `json:"systemRole"`
	SalesRole     SalesRoleRef   `json:"salesRole"`
	ParentUserID  *uuid.UUID     `json:"parentUserId"`
	ParentName    *string        `json:"parentName"`
	EffectiveFrom string         `json:"effectiveFrom"`
	EffectiveTo   *string        `json:"effectiveTo"`
}

type SalesRoleRef struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Level int       `json:"level"`
}

type AssignmentHistoryItem struct {
	AssignmentID  uuid.UUID    `json:"assignmentId"`
	SalesRole     SalesRoleRef `json:"salesRole"`
	ParentUserID  *uuid.UUID   `json:"parentUserId"`
	ParentName    *string      `json:"parentName"`
	EffectiveFrom string       `json:"effectiveFrom"`
	EffectiveTo   *string      `json:"effectiveTo"`
	Status        string       `json:"status"`
}
