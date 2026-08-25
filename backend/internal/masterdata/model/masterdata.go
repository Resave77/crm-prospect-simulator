package model

import "time"

const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
)

type Segment struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	CategoryCount int        `json:"categoryCount"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`
}

type Category struct {
	ID          string     `json:"id"`
	SegmentID   string     `json:"segmentId"`
	SegmentName string     `json:"segmentName"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	PlaceAPI    string     `json:"placeApi"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type SegmentInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type CategoryInput struct {
	SegmentID   string `json:"segmentId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PlaceAPI    string `json:"placeApi"`
	Status      string `json:"status"`
}
