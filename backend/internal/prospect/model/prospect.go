package model

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusNewLead      Status = "NEW_LEAD"
	StatusContacted    Status = "CONTACTED"
	StatusInterested   Status = "INTERESTED"
	StatusQualified    Status = "QUALIFIED"
	StatusProposalSent Status = "PROPOSAL_SENT"
	StatusNegotiation  Status = "NEGOTIATION"
	StatusWon          Status = "WON"
	StatusLost         Status = "LOST"
	StatusConverted    Status = "CONVERTED"
)

type Prospect struct {
	ID                       uuid.UUID  `json:"id"`
	GooglePlaceID            string     `json:"googlePlaceId"`
	PlaceName                string     `json:"placeName"`
	FormattedAddress         string     `json:"formattedAddress"`
	Latitude                 *float64   `json:"latitude"`
	Longitude                *float64   `json:"longitude"`
	PlaceCategory            string     `json:"placeCategory"`
	IndustryGroup            string     `json:"industryGroup"`
	PlaceTypes               []string   `json:"placeTypes"`
	PhoneNumber              string     `json:"phoneNumber"`
	WebsiteURL               string     `json:"websiteUrl"`
	GoogleMapsURL            string     `json:"googleMapsUrl"`
	AssignedSalesExecutiveID uuid.UUID  `json:"assignedSalesExecutiveId"`
	AssignedSalesExecutive   string     `json:"assignedSalesExecutive"`
	VisitNotes               string     `json:"visitNotes"`
	FollowUpNotes            string     `json:"followUpNotes"`
	Status                   Status     `json:"status"`
	ConvertedAt              *time.Time `json:"convertedAt,omitempty"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                time.Time  `json:"updatedAt"`
}

var ActiveStatuses = []Status{StatusNewLead, StatusContacted, StatusInterested, StatusQualified, StatusProposalSent, StatusNegotiation}

type SalesExecutive struct {
	ID                 uuid.UUID `json:"id"`
	FullName           string    `json:"fullName"`
	ActiveProspectCount int      `json:"activeProspectCount"`
}

type PlaceResult struct {
	GooglePlaceID    string   `json:"googlePlaceId"`
	PlaceName        string   `json:"name"`
	FormattedAddress string   `json:"address"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	PlaceCategory    string   `json:"category"`
	PlaceTypes       []string `json:"placeTypes"`
	PhoneNumber      string   `json:"phone"`
	Distance         float64  `json:"distance"`
	Rating           float64  `json:"rating"`
	UserRatingCount  int      `json:"userRatingCount"`
	BusinessStatus   string   `json:"businessStatus"`
	WebsiteURL       string   `json:"website"`
	GoogleMapsURL    string   `json:"googleMapsUrl"`
	MarkerCategory   string   `json:"markerCategory"`
	MarkerColor      string   `json:"markerColor"`
	MarkerIcon       string   `json:"markerIcon"`
}

type PlaceSearchInput struct {
	Keyword    string   `json:"keyword"`
	Categories []string `json:"categories"`
	Latitude   float64  `json:"latitude"`
	Longitude  float64  `json:"longitude"`
	Radius     float64  `json:"radius"`
}

type SaveProspectInput struct {
	Place                    PlaceResult `json:"place"`
	IndustryGroup            string      `json:"industryGroup"`
	AssignedSalesExecutiveID uuid.UUID   `json:"assignedSalesExecutiveId"`
}

type StatusHistory struct {
	ID              uuid.UUID `json:"id"`
	FromStatus      *Status   `json:"fromStatus"`
	ToStatus        Status    `json:"toStatus"`
	ChangedByUserID uuid.UUID `json:"changedByUserId"`
	ChangedByName   string    `json:"changedByName"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Review struct {
	Prospect Prospect        `json:"prospect"`
	History  []StatusHistory `json:"history"`
	Visits   []Visit         `json:"visits"`
}

type Visit struct {
	ID                 uuid.UUID  `json:"id"`
	ProspectID         uuid.UUID  `json:"prospectId"`
	SalesExecutiveID   uuid.UUID  `json:"salesExecutiveId"`
	SalesExecutiveName string     `json:"salesExecutiveName"`
	CheckInAt          time.Time  `json:"checkInAt"`
	CheckInLatitude    float64    `json:"checkInLatitude"`
	CheckInLongitude   float64    `json:"checkInLongitude"`
	CheckOutAt         *time.Time `json:"checkOutAt,omitempty"`
	CheckOutLatitude   *float64   `json:"checkOutLatitude"`
	CheckOutLongitude  *float64   `json:"checkOutLongitude"`
	SelfieReference    string     `json:"selfieReference"`
	VisitNotes         string     `json:"visitNotes"`
	FollowUpNotes      string     `json:"followUpNotes"`
}

type CheckInInput struct {
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	SelfiePlaceholder bool    `json:"selfiePlaceholder"`
	VisitNotes        string  `json:"visitNotes"`
}

type CheckOutInput struct {
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	FollowUpNotes string  `json:"followUpNotes"`
}
