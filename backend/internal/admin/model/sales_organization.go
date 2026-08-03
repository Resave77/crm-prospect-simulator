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
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Level       int        `json:"level"`
	Description string     `json:"description"`
	IsActive    bool       `json:"isActive"`
	CreatedBy   *uuid.UUID `json:"createdBy"`
	UpdatedBy   *uuid.UUID `json:"updatedBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type CreateSalesRoleInput struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type UpdateSalesRoleInput struct {
	Name        *string `json:"name"`
	Level       *int    `json:"level"`
	Description *string `json:"description"`
}

type UpdateSalesRoleStatusInput struct {
	IsActive bool `json:"isActive"`
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
}
