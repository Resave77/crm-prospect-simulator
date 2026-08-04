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
	DeletionRequested        bool       `json:"deletionRequested"`
	ConvertedAt              *time.Time `json:"convertedAt,omitempty"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                time.Time  `json:"updatedAt"`
}

var ActiveStatuses = []Status{StatusNewLead, StatusContacted, StatusInterested, StatusQualified, StatusProposalSent, StatusNegotiation}

type SalesExecutive struct {
	ID                  uuid.UUID `json:"id"`
	FullName            string    `json:"fullName"`
	ActiveProspectCount int       `json:"activeProspectCount"`
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
	VisitResult        string     `json:"visitResult"`
	VisitOutcome       string     `json:"visitOutcome"`
}

type CheckInInput struct {
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	SelfieReference string  `json:"selfieReference"`
	VisitNotes      string  `json:"visitNotes"`
}

type CheckOutInput struct {
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	FollowUpNotes string  `json:"followUpNotes"`
	VisitResult   string  `json:"visitResult"`
	VisitOutcome  string  `json:"visitOutcome"`
}

type PlacePhoto struct {
	Name        string `json:"name"`
	PhotoURL    string `json:"photoUrl"`
	WidthPx     int    `json:"widthPx"`
	HeightPx    int    `json:"heightPx"`
	Attribution string `json:"attribution"`
}

type PlaceOpeningHours struct {
	OpenNow  bool     `json:"openNow"`
	Weekdays []string `json:"weekdays"`
}

type PlaceReview struct {
	AuthorName   string  `json:"authorName"`
	AuthorPhoto  string  `json:"authorPhoto"`
	Rating       float64 `json:"rating"`
	Text         string  `json:"text"`
	Time         string  `json:"time"`
	LanguageCode string  `json:"languageCode"`
}

type PlaceDetails struct {
	GooglePlaceID        string              `json:"googlePlaceId"`
	PlaceName            string              `json:"placeName"`
	FormattedAddress     string              `json:"formattedAddress"`
	Latitude             float64             `json:"latitude"`
	Longitude            float64             `json:"longitude"`
	PlaceCategory        string              `json:"placeCategory"`
	PlaceTypes           []string            `json:"placeTypes"`
	PhoneNumber          string              `json:"phoneNumber"`
	InternationalPhone   string              `json:"internationalPhone"`
	WebsiteURL           string              `json:"websiteUrl"`
	GoogleMapsURL        string              `json:"googleMapsUrl"`
	Rating               float64             `json:"rating"`
	UserRatingCount      int                 `json:"userRatingCount"`
	BusinessStatus       string              `json:"businessStatus"`
	PriceLevel           string              `json:"priceLevel"`
	EditorialSummary     string              `json:"editorialSummary"`
	UTCOffsetMinutes     int                 `json:"utcOffsetMinutes"`
	Photos               []PlacePhoto        `json:"photos"`
	OpeningHours         *PlaceOpeningHours  `json:"openingHours"`
	Reviews              []PlaceReview       `json:"reviews"`
	Delivery             bool                `json:"delivery"`
	DineIn               bool                `json:"dineIn"`
	Takeout              bool                `json:"takeout"`
	CurbsidePickup       bool                `json:"curbsidePickup"`
	ParkingOptions       *PlaceParking       `json:"parkingOptions"`
	PaymentOptions       *PlacePayments      `json:"paymentOptions"`
	AccessibilityOptions *PlaceAccessibility `json:"accessibilityOptions"`
}

type PlaceParking struct {
	PaidStreetParking bool `json:"paidStreetParking"`
	PaidParkingLot    bool `json:"paidParkingLot"`
	FreeStreetParking bool `json:"freeStreetParking"`
	FreeParkingLot    bool `json:"freeParkingLot"`
	ValetParking      bool `json:"valetParking"`
	GarageParking     bool `json:"garageParking"`
}

type PlacePayments struct {
	CashOnly       bool `json:"cashOnly"`
	CreditCardOnly bool `json:"creditCardOnly"`
	DebitCardOnly  bool `json:"debitCardOnly"`
	NfcOnly        bool `json:"nfcOnly"`
}

type PlaceAccessibility struct {
	WheelchairAccessibleEntrance bool `json:"wheelchairAccessibleEntrance"`
	WheelchairAccessibleParking  bool `json:"wheelchairAccessibleParking"`
	WheelchairAccessibleRestroom bool `json:"wheelchairAccessibleRestroom"`
	WheelchairAccessibleSeating  bool `json:"wheelchairAccessibleSeating"`
}

type ProspectComment struct {
	ID          uuid.UUID           `json:"id"`
	ProspectID  uuid.UUID           `json:"prospectId"`
	UserID      uuid.UUID           `json:"userId"`
	UserName    string              `json:"userName"`
	Content     string              `json:"content"`
	Attachments []CommentAttachment `json:"attachments"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

type CommentAttachment struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	Path        string    `json:"-"`
}

type PhotoCategory string

const (
	PhotoCategoryMenu  PhotoCategory = "MENU"
	PhotoCategoryPlace PhotoCategory = "PLACE"
)

type ProspectPhotoTag struct {
	ID         uuid.UUID     `json:"id"`
	ProspectID uuid.UUID     `json:"prospectId"`
	PhotoName  string        `json:"photoName"`
	Category   PhotoCategory `json:"category"`
	UpdatedBy  *uuid.UUID    `json:"updatedBy"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
}
