package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sort"
	"strconv"
	"strings"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	customermodel "crm-prospect-simulator/backend/internal/customer/model"
	"crm-prospect-simulator/backend/internal/customer/repository"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	prospectservice "crm-prospect-simulator/backend/internal/prospect/service"
	"github.com/google/uuid"
)

var (
	ErrForbidden         = errors.New("customer operation forbidden")
	ErrValidation        = errors.New("conversion form is invalid")
	ErrAssignmentOverlap = errors.New("assignment periods for the same owner overlap")
)

type Actor struct {
	UserID uuid.UUID
	Role   authmodel.Role
}

type ConversionForm struct {
	Prospect            prospectmodel.Review          `json:"prospect"`
	ParentCompanies     []customermodel.ParentCompany `json:"parentCompanies"`
	SalesExecutives     []customermodel.UserOption    `json:"salesExecutives"`
	ParentCodePreview   string                        `json:"parentCodePreview"`
	CustomerCodePreview string                        `json:"customerCodePreview"`
	SellerIdentity      string                        `json:"sellerIdentity"`
	Options             customermodel.MasterOptions   `json:"options"`
}

type Service struct {
	repository repository.Repository
	prospects  *prospectservice.Service
}

func New(repo repository.Repository, prospects *prospectservice.Service) *Service {
	return &Service{repository: repo, prospects: prospects}
}

func (s *Service) ConversionForm(ctx context.Context, actor Actor, prospectID uuid.UUID) (ConversionForm, error) {
	if actor.Role != authmodel.RoleAdministrator {
		return ConversionForm{}, ErrForbidden
	}
	review, err := s.prospects.Review(ctx, prospectservice.Actor{UserID: actor.UserID, Role: actor.Role}, prospectID)
	if err != nil {
		return ConversionForm{}, err
	}
	if review.Prospect.Status != prospectmodel.StatusWon {
		return ConversionForm{}, repository.ErrProspectNotWon
	}
	companies, err := s.repository.SearchParentCompanies(ctx, "")
	if err != nil {
		return ConversionForm{}, err
	}
	sales, err := s.repository.ListActiveSalesExecutives(ctx)
	if err != nil {
		return ConversionForm{}, err
	}
	return ConversionForm{
		Prospect: review, ParentCompanies: companies, SalesExecutives: sales,
		ParentCodePreview: "Generated on save", CustomerCodePreview: "Generated on save",
		SellerIdentity: "PT Yummy Corp — Simulation",
		Options:        simulationOptions(),
	}, nil
}

func (s *Service) SearchParentCompanies(ctx context.Context, actor Actor, search string) ([]customermodel.ParentCompany, error) {
	if actor.Role != authmodel.RoleAdministrator {
		return nil, ErrForbidden
	}
	return s.repository.SearchParentCompanies(ctx, strings.TrimSpace(search))
}

func (s *Service) Convert(ctx context.Context, actor Actor, prospectID uuid.UUID, input customermodel.ConversionInput) (customermodel.CustomerSite, error) {
	if actor.Role != authmodel.RoleAdministrator {
		return customermodel.CustomerSite{}, ErrForbidden
	}
	normalize(&input)
	if err := validate(input); err != nil {
		return customermodel.CustomerSite{}, err
	}
	return s.repository.Convert(ctx, prospectID, actor.UserID, input)
}

func (s *Service) AdminCustomers(ctx context.Context, actor Actor) ([]customermodel.CustomerSite, error) {
	if actor.Role != authmodel.RoleAdministrator {
		return nil, ErrForbidden
	}
	return s.repository.ListCustomers(ctx)
}

func (s *Service) AdminCustomersList(ctx context.Context, actor Actor, params customermodel.CustomerListParams) (customermodel.CustomerListResult, error) {
	if actor.Role != authmodel.RoleAdministrator {
		return customermodel.CustomerListResult{}, ErrForbidden
	}
	return s.repository.ListCustomersPaged(ctx, params)
}

func (s *Service) CustomerFilterOptions(ctx context.Context, actor Actor) (customermodel.ListFilterOptions, error) {
	if actor.Role != authmodel.RoleAdministrator {
		return customermodel.ListFilterOptions{}, ErrForbidden
	}
	return s.repository.ListFilterOptions(ctx)
}

func (s *Service) MyCustomers(ctx context.Context, actor Actor) ([]customermodel.CustomerSite, error) {
	if actor.Role != authmodel.RoleSalesExecutive {
		return nil, ErrForbidden
	}
	return s.repository.ListCustomersForSales(ctx, actor.UserID)
}

func (s *Service) MyCustomer(ctx context.Context, actor Actor, id uuid.UUID) (customermodel.CustomerDetail, error) {
	if actor.Role != authmodel.RoleSalesExecutive {
		return customermodel.CustomerDetail{}, ErrForbidden
	}
	return s.repository.FindCustomerForSales(ctx, id, actor.UserID)
}

func (s *Service) AdminCustomer(ctx context.Context, actor Actor, id uuid.UUID) (customermodel.CustomerDetail, error) {
	if actor.Role != authmodel.RoleAdministrator {
		return customermodel.CustomerDetail{}, ErrForbidden
	}
	return s.repository.FindCustomer(ctx, id)
}

func normalize(input *customermodel.ConversionInput) {
	input.CustomerName = strings.TrimSpace(input.CustomerName)
	input.CustomerSegment = strings.TrimSpace(input.CustomerSegment)
	input.CustomerCategory = strings.TrimSpace(input.CustomerCategory)
	input.ParentCompanyName = strings.TrimSpace(input.ParentCompanyName)
	if input.ParentMethod == customermodel.ParentMethodMatchSite {
		input.ParentCompanyName = input.CustomerName
	}
	if input.SameAsSiteAddress && input.ParentMethod != customermodel.ParentMethodExisting {
		input.CompanyAddress = input.SiteAddress
	}
}

func validate(input customermodel.ConversionInput) error {
	if input.CustomerName == "" || input.CustomerSegment == "" || input.CustomerCategory == "" {
		return fmt.Errorf("%w: customer name, segment, and category are required", ErrValidation)
	}
	if input.ParentMethod != customermodel.ParentMethodManual && input.ParentMethod != customermodel.ParentMethodMatchSite && input.ParentMethod != customermodel.ParentMethodExisting {
		return fmt.Errorf("%w: parent company method is required", ErrValidation)
	}
	if input.ParentMethod == customermodel.ParentMethodExisting {
		if input.ExistingParentCompanyID == nil || *input.ExistingParentCompanyID == uuid.Nil {
			return fmt.Errorf("%w: an existing parent company must be selected", ErrValidation)
		}
	} else if input.ParentCompanyName == "" {
		return fmt.Errorf("%w: parent company name is required", ErrValidation)
	}
	if input.SiteAddress.Mode == "" || input.SiteAddress.Province == "" || input.SiteAddress.District == "" || input.SiteAddress.SubDistrict == "" || input.SiteAddress.Village == "" || strings.TrimSpace(input.SiteAddress.PreviewAddress) == "" {
		return fmt.Errorf("%w: site address mode, region, and preview address are required", ErrValidation)
	}
	if err := validateCoordinates(input.SiteAddress); err != nil {
		return err
	}
	if input.ParentMethod != customermodel.ParentMethodExisting {
		if err := validateCoordinates(input.CompanyAddress); err != nil {
			return err
		}
		if err := validateOptionalAddress(input.CompanyAddress); err != nil {
			return err
		}
	}
	if input.SalesExecutiveID == uuid.Nil {
		return fmt.Errorf("%w: an active Sales Executive is required", ErrValidation)
	}
	if err := validateContacts(input.SiteContacts); err != nil {
		return err
	}
	if err := validateContacts(input.CompanyContacts); err != nil {
		return err
	}
	for _, assignment := range input.SalesAssignments {
		if assignment.OwnerID == "" && assignment.OwnerName == "" && assignment.StartMonth == 0 && assignment.StartYear == 0 && assignment.End == "" {
			continue
		}
		if _, err := uuid.Parse(assignment.OwnerID); err != nil {
			return fmt.Errorf("%w: Sales Assignment owner must reference a Sales Executive user", ErrValidation)
		}
	}
	if err := validateAssignments(input.SalesAssignments); err != nil {
		return err
	}
	return validateAssignments(input.KAMAssignments)
}

func validateCoordinates(address customermodel.Address) error {
	if address.Latitude != nil && (*address.Latitude < -90 || *address.Latitude > 90) {
		return fmt.Errorf("%w: latitude is outside its valid range", ErrValidation)
	}
	if address.Longitude != nil && (*address.Longitude < -180 || *address.Longitude > 180) {
		return fmt.Errorf("%w: longitude is outside its valid range", ErrValidation)
	}
	return nil
}

func validateOptionalAddress(address customermodel.Address) error {
	started := address.Mode != "" || address.Province != "" || address.District != "" || address.SubDistrict != "" ||
		address.Village != "" || address.Latitude != nil || address.Longitude != nil || strings.TrimSpace(address.PreviewAddress) != ""
	if !started {
		return nil
	}
	if address.Mode == "" || address.Province == "" || address.District == "" || address.SubDistrict == "" || address.Village == "" || strings.TrimSpace(address.PreviewAddress) == "" {
		return fmt.Errorf("%w: a started company address requires its mode, region, and preview address", ErrValidation)
	}
	return nil
}

func validateContacts(contacts []customermodel.Contact) error {
	for _, contact := range contacts {
		started := strings.TrimSpace(contact.Name+contact.Position+contact.Phone+contact.Email) != ""
		if !started {
			continue
		}
		if strings.TrimSpace(contact.Name) == "" && strings.TrimSpace(contact.Phone) == "" {
			return fmt.Errorf("%w: a started contact requires a contact name or phone number", ErrValidation)
		}
		if contact.Email != "" {
			if _, err := mail.ParseAddress(contact.Email); err != nil {
				return fmt.Errorf("%w: contact email is invalid", ErrValidation)
			}
		}
	}
	return nil
}

func validateAssignments(assignments []customermodel.PeriodAssignment) error {
	type period struct{ start, end int }
	byOwner := make(map[string][]period)
	for _, assignment := range assignments {
		started := assignment.OwnerID != "" || assignment.OwnerName != "" || assignment.StartMonth != 0 || assignment.StartYear != 0 || assignment.End != ""
		if !started {
			continue
		}
		owner := strings.TrimSpace(assignment.OwnerID)
		if owner == "" {
			owner = strings.TrimSpace(assignment.OwnerName)
		}
		if owner == "" || assignment.StartMonth < 1 || assignment.StartMonth > 12 || assignment.StartYear < 2000 {
			return fmt.Errorf("%w: a started assignment requires owner, start month, and start year", ErrValidation)
		}
		start := assignment.StartYear*12 + assignment.StartMonth
		end := int(^uint(0) >> 1)
		if assignment.End != "" && assignment.End != "UNTIL_NOW" {
			parts := strings.Split(assignment.End, "-")
			if len(parts) != 2 {
				return fmt.Errorf("%w: assignment end must be YYYY-MM or UNTIL_NOW", ErrValidation)
			}
			year, yearErr := strconv.Atoi(parts[0])
			month, monthErr := strconv.Atoi(parts[1])
			if yearErr != nil || monthErr != nil || month < 1 || month > 12 {
				return fmt.Errorf("%w: assignment end is invalid", ErrValidation)
			}
			end = year*12 + month
			if end < start {
				return fmt.Errorf("%w: assignment end precedes its start", ErrValidation)
			}
		}
		byOwner[strings.ToLower(owner)] = append(byOwner[strings.ToLower(owner)], period{start: start, end: end})
	}
	for _, periods := range byOwner {
		sort.Slice(periods, func(i, j int) bool { return periods[i].start < periods[j].start })
		for i := 1; i < len(periods); i++ {
			if periods[i].start <= periods[i-1].end {
				return ErrAssignmentOverlap
			}
		}
	}
	return nil
}

func simulationOptions() customermodel.MasterOptions {
	return customermodel.MasterOptions{
		Segments:       []string{"General Trade", "Modern Trade", "Food Service", "Key Account"},
		Categories:     []string{"Cafe", "Restaurant", "Retail Store", "Hotel", "Office"},
		ShipmentCosts:  []string{"Standard", "Free Shipping", "Special Rate"},
		InvoiceTypes:   []string{"Tax Invoice", "Commercial Invoice", "Receipt"},
		TermsOfPayment: []string{"Cash", "NET 14", "NET 30", "NET 45"},
		KAMs:           []string{"Andini Putri", "Bagus Pratama", "Citra Lestari"},
		AddressSuggestions: []customermodel.Address{
			{Mode: "GMAPS_SIMULATION", Province: "DKI Jakarta", District: "Jakarta Selatan", SubDistrict: "Cilandak", Village: "Cipete Selatan", Latitude: pointer(-6.2771), Longitude: pointer(106.8016), PreviewAddress: "Jl. Cipete Raya, Cipete Selatan, Jakarta Selatan"},
			{Mode: "GMAPS_SIMULATION", Province: "DKI Jakarta", District: "Jakarta Pusat", SubDistrict: "Menteng", Village: "Gondangdia", Latitude: pointer(-6.1872), Longitude: pointer(106.8307), PreviewAddress: "Jl. RP. Soeroso, Gondangdia, Jakarta Pusat"},
		},
	}
}

func pointer(value float64) *float64 { return &value }
