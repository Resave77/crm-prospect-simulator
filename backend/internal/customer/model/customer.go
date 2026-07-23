package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	ParentMethodManual    = "MANUAL_ENTRY"
	ParentMethodMatchSite = "MATCH_CUSTOMER_NAME"
	ParentMethodExisting  = "EXISTING_COMPANY"
)

type Address struct {
	Mode           string   `json:"mode"`
	Province       string   `json:"province"`
	District       string   `json:"district"`
	SubDistrict    string   `json:"subDistrict"`
	Village        string   `json:"village"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	PreviewAddress string   `json:"previewAddress"`
}

type Contact struct {
	Name     string `json:"name"`
	Position string `json:"position"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type PeriodAssignment struct {
	OwnerID    string `json:"ownerId"`
	OwnerName  string `json:"ownerName"`
	StartMonth int    `json:"startMonth"`
	StartYear  int    `json:"startYear"`
	End        string `json:"end"`
}

type ParentCompany struct {
	ID             uuid.UUID          `json:"id"`
	ParentCode     string             `json:"parentCode"`
	Name           string             `json:"name"`
	Address        Address            `json:"address"`
	Contacts       []Contact          `json:"contacts"`
	NPWPName       string             `json:"npwpName"`
	NPWPAddress    string             `json:"npwpAddress"`
	NPWPNumber     string             `json:"npwpNumber"`
	TermOfPayment  string             `json:"termOfPayment"`
	KAMAssignments []PeriodAssignment `json:"kamAssignments"`
}

type CustomerSite struct {
	ID                     uuid.UUID          `json:"id"`
	CustomerCode           string             `json:"customerCode"`
	ParentCompanyID        uuid.UUID          `json:"parentCompanyId"`
	ParentCode             string             `json:"parentCode"`
	ParentCompanyName      string             `json:"parentCompanyName"`
	SourceProspectID       uuid.UUID          `json:"sourceProspectId"`
	SourceGooglePlaceID    string             `json:"sourceGooglePlaceId"`
	Name                   string             `json:"name"`
	Segment                string             `json:"segment"`
	Category               string             `json:"category"`
	Region                 string             `json:"region"`
	Address                Address            `json:"address"`
	Contacts               []Contact          `json:"contacts"`
	PPN                    string             `json:"ppn"`
	IDTKUNumber            string             `json:"idTkuNumber"`
	NIK                    string             `json:"nik"`
	ShipmentCost           string             `json:"shipmentCost"`
	InvoiceType            string             `json:"invoiceType"`
	BankAccount            string             `json:"bankAccount"`
	BillToSource           string             `json:"billToSource"`
	ShipToSource           string             `json:"shipToSource"`
	BillingAddressPreview  string             `json:"billingAddressPreview"`
	ShippingAddressPreview string             `json:"shippingAddressPreview"`
	SalesExecutiveID       uuid.UUID          `json:"salesExecutiveId"`
	SalesExecutiveName     string             `json:"salesExecutiveName"`
	SalesAssignments       []PeriodAssignment `json:"salesAssignments"`
	ConvertedAt            time.Time          `json:"convertedAt"`
	UpdatedAt              time.Time          `json:"updatedAt"`
	ConvertedByAdminID     uuid.UUID          `json:"convertedByAdminId"`
}

type UserOption struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"fullName"`
}

type CustomerDetail struct {
	Customer           CustomerSite  `json:"customer"`
	ParentCompany      ParentCompany `json:"parentCompany"`
	SourceProspectName string        `json:"sourceProspectName"`
}

type ConversionInput struct {
	CustomerName            string             `json:"customerName"`
	CustomerSegment         string             `json:"customerSegment"`
	CustomerCategory        string             `json:"customerCategory"`
	ParentMethod            string             `json:"parentMethod"`
	ExistingParentCompanyID *uuid.UUID         `json:"existingParentCompanyId"`
	ParentCompanyName       string             `json:"parentCompanyName"`
	SameAsSiteAddress       bool               `json:"sameAsSiteAddress"`
	SiteAddress             Address            `json:"siteAddress"`
	CompanyAddress          Address            `json:"companyAddress"`
	SiteContacts            []Contact          `json:"siteContacts"`
	CompanyContacts         []Contact          `json:"companyContacts"`
	PPN                     string             `json:"ppn"`
	IDTKUNumber             string             `json:"idTkuNumber"`
	NIK                     string             `json:"nik"`
	CompanyNPWPName         string             `json:"companyNpwpName"`
	CompanyNPWPAddress      string             `json:"companyNpwpAddress"`
	CompanyNPWPNumber       string             `json:"companyNpwpNumber"`
	ShipmentCost            string             `json:"shipmentCost"`
	InvoiceType             string             `json:"invoiceType"`
	BankAccount             string             `json:"bankAccount"`
	TermOfPayment           string             `json:"termOfPayment"`
	BillToSource            string             `json:"billToSource"`
	ShipToSource            string             `json:"shipToSource"`
	BillingAddressPreview   string             `json:"billingAddressPreview"`
	ShippingAddressPreview  string             `json:"shippingAddressPreview"`
	SalesExecutiveID        uuid.UUID          `json:"salesExecutiveId"`
	SalesAssignments        []PeriodAssignment `json:"salesAssignments"`
	KAMAssignments          []PeriodAssignment `json:"kamAssignments"`
}

type CustomerListParams struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
	Segment string `json:"segment"`
	Category string `json:"category"`
	Sales   string `json:"sales"`
	Region  string `json:"region"`
	Sort    string `json:"sort"`
}

type CustomerListResult struct {
	Items []CustomerSite `json:"items"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
	Pages int            `json:"pages"`
}

type ListFilterOptions struct {
	Segments  []string       `json:"segments"`
	Categories []string      `json:"categories"`
	Regions   []string       `json:"regions"`
	SalesExec []UserOption   `json:"salesExecutives"`
}

type MasterOptions struct {
	Segments           []string  `json:"segments"`
	Categories         []string  `json:"categories"`
	ShipmentCosts      []string  `json:"shipmentCosts"`
	InvoiceTypes       []string  `json:"invoiceTypes"`
	TermsOfPayment     []string  `json:"termsOfPayment"`
	KAMs               []string  `json:"kams"`
	AddressSuggestions []Address `json:"addressSuggestions"`
}
