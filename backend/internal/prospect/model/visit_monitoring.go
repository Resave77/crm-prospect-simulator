package model

import (
	"time"

	"github.com/google/uuid"
)

type VisitMonitoringItem struct {
	ID                   uuid.UUID  `json:"id"`
	ProspectID           uuid.UUID  `json:"prospectId"`
	CustomerID           *uuid.UUID `json:"customerId,omitempty"`
	EntityType           string     `json:"entityType"`
	CustomerName         string     `json:"customerName"`
	CustomerCategory     string     `json:"customerCategory"`
	IndustryGroup        string     `json:"industryGroup"`
	FormattedAddress     string     `json:"formattedAddress"`
	PhoneNumber          string     `json:"phoneNumber"`
	ProspectLatitude     *float64   `json:"prospectLatitude"`
	ProspectLongitude    *float64   `json:"prospectLongitude"`
	SalesExecutiveID     uuid.UUID  `json:"salesExecutiveId"`
	SalesExecutiveName   string     `json:"salesExecutiveName"`
	CheckInAt            time.Time  `json:"checkInAt"`
	CheckOutAt           *time.Time `json:"checkOutAt,omitempty"`
	CheckInLatitude      float64    `json:"checkInLatitude"`
	CheckInLongitude     float64    `json:"checkInLongitude"`
	CheckOutLatitude     *float64   `json:"checkOutLatitude"`
	CheckOutLongitude    *float64   `json:"checkOutLongitude"`
	DistanceMeters       float64    `json:"distanceMeters"`
	DurationSeconds      *float64   `json:"durationSeconds,omitempty"`
	RadiusStatus         string     `json:"radiusStatus"`
	ProspectStatus       string     `json:"prospectStatus"`
	SelfieReference      string     `json:"selfieReference"`
	VisitNotes           string     `json:"visitNotes"`
	FollowUpNotes        string     `json:"followUpNotes"`
	VisitResult          string     `json:"visitResult"`
	VisitOutcome         string     `json:"visitOutcome"`
	VisitCount           int        `json:"visitCount"`
}

type VisitMonitoringFilter struct {
	DateFrom          string `query:"dateFrom"`
	DateTo            string `query:"dateTo"`
	SalesExecutiveID  string `query:"salesExecutiveId"`
	CustomerName      string `query:"customerName"`
	RadiusStatus      string `query:"radiusStatus"`
}
