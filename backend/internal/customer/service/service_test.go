package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	customermodel "crm-prospect-simulator/backend/internal/customer/model"
	customerrepository "crm-prospect-simulator/backend/internal/customer/repository"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	prospectrepository "crm-prospect-simulator/backend/internal/prospect/repository"
	prospectservice "crm-prospect-simulator/backend/internal/prospect/service"
	"github.com/google/uuid"
)

type fakeProspectSource struct {
	prospect prospectmodel.Prospect
}

func (f *fakeProspectSource) ListAssigned(context.Context, uuid.UUID) ([]prospectmodel.Prospect, error) {
	return []prospectmodel.Prospect{f.prospect}, nil
}
func (f *fakeProspectSource) ListWon(context.Context) ([]prospectmodel.Prospect, error) {
	return []prospectmodel.Prospect{f.prospect}, nil
}
func (f *fakeProspectSource) ListAll(context.Context) ([]prospectmodel.Prospect, error) {
	return []prospectmodel.Prospect{f.prospect}, nil
}
func (f *fakeProspectSource) ListSalesExecutives(context.Context) ([]prospectmodel.SalesExecutive, error) {
	return nil, nil
}
func (f *fakeProspectSource) FindReview(context.Context, uuid.UUID) (prospectmodel.Review, error) {
	return prospectmodel.Review{Prospect: f.prospect}, nil
}
func (f *fakeProspectSource) Transition(context.Context, uuid.UUID, uuid.UUID, prospectmodel.Status, prospectmodel.Status, string) (prospectmodel.Prospect, error) {
	return prospectmodel.Prospect{}, prospectrepository.ErrInvalidStatus
}
func (f *fakeProspectSource) Create(context.Context, prospectmodel.SaveProspectInput, uuid.UUID) (prospectmodel.Prospect, error) {
	return f.prospect, nil
}
func (f *fakeProspectSource) CheckIn(context.Context, uuid.UUID, uuid.UUID, prospectmodel.CheckInInput) (prospectmodel.Visit, error) {
	return prospectmodel.Visit{}, nil
}
func (f *fakeProspectSource) CheckOut(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, prospectmodel.CheckOutInput) (prospectmodel.Visit, error) {
	return prospectmodel.Visit{}, nil
}

type fakeCustomerRepository struct {
	prospectStatus prospectmodel.Status
	googlePlaceID  string
	duplicatePlace bool
	customers      []customermodel.CustomerSite
	parents        []customermodel.ParentCompany
}

func (f *fakeCustomerRepository) SearchParentCompanies(context.Context, string) ([]customermodel.ParentCompany, error) {
	return f.parents, nil
}
func (f *fakeCustomerRepository) ListActiveSalesExecutives(context.Context) ([]customermodel.UserOption, error) {
	return []customermodel.UserOption{{ID: uuid.New(), FullName: "Active Sales"}}, nil
}
func (f *fakeCustomerRepository) Convert(_ context.Context, prospectID, adminID uuid.UUID, input customermodel.ConversionInput) (customermodel.CustomerSite, error) {
	if f.prospectStatus != prospectmodel.StatusWon && f.prospectStatus != prospectmodel.StatusConverted {
		return customermodel.CustomerSite{}, customerrepository.ErrProspectNotWon
	}
	if f.prospectStatus == prospectmodel.StatusConverted {
		return customermodel.CustomerSite{}, customerrepository.ErrAlreadyConverted
	}
	if f.duplicatePlace {
		return customermodel.CustomerSite{}, customerrepository.ErrDuplicatePlace
	}
	parentID := uuid.New()
	parentCode := fmt.Sprintf("PC-%06d", len(f.parents)+1)
	parentName := input.ParentCompanyName
	if input.ParentMethod == customermodel.ParentMethodExisting {
		if input.ExistingParentCompanyID == nil {
			return customermodel.CustomerSite{}, customerrepository.ErrParentUnavailable
		}
		parentID = *input.ExistingParentCompanyID
		for _, parent := range f.parents {
			if parent.ID == parentID {
				parentCode, parentName = parent.ParentCode, parent.Name
			}
		}
	} else {
		f.parents = append(f.parents, customermodel.ParentCompany{ID: parentID, ParentCode: parentCode, Name: parentName})
	}
	item := customermodel.CustomerSite{
		ID: uuid.New(), CustomerCode: fmt.Sprintf("%s-S%03d", parentCode, len(f.customers)+1),
		ParentCompanyID: parentID, ParentCode: parentCode, ParentCompanyName: parentName,
		SourceProspectID: prospectID, SourceGooglePlaceID: f.googlePlaceID,
		Name: input.CustomerName, Segment: input.CustomerSegment, Category: input.CustomerCategory,
		SalesExecutiveID: input.SalesExecutiveID, ConvertedAt: time.Now(), ConvertedByAdminID: adminID,
	}
	f.customers = append(f.customers, item)
	f.prospectStatus = prospectmodel.StatusConverted
	return item, nil
}
func (f *fakeCustomerRepository) ListCustomers(context.Context) ([]customermodel.CustomerSite, error) {
	return f.customers, nil
}
func (f *fakeCustomerRepository) ListCustomersForSales(_ context.Context, salesID uuid.UUID) ([]customermodel.CustomerSite, error) {
	items := make([]customermodel.CustomerSite, 0)
	for _, item := range f.customers {
		if item.SalesExecutiveID == salesID {
			items = append(items, item)
		}
	}
	return items, nil
}
func (f *fakeCustomerRepository) FindCustomerForSales(_ context.Context, id, salesID uuid.UUID) (customermodel.CustomerDetail, error) {
	for _, item := range f.customers {
		if item.ID == id && item.SalesExecutiveID == salesID {
			return customermodel.CustomerDetail{Customer: item}, nil
		}
	}
	return customermodel.CustomerDetail{}, customerrepository.ErrNotFound
}

func newCustomerService(repo *fakeCustomerRepository) *Service {
	prospects := prospectservice.New(&fakeProspectSource{prospect: prospectmodel.Prospect{ID: uuid.New(), Status: repo.prospectStatus}})
	return New(repo, prospects)
}

func validInput(salesID uuid.UUID) customermodel.ConversionInput {
	return customermodel.ConversionInput{
		CustomerName: "Toko Kopi Tuku - Cipete", CustomerSegment: "Food Service", CustomerCategory: "Cafe",
		ParentMethod: customermodel.ParentMethodManual, ParentCompanyName: "PT Kopi Tuku Indonesia",
		SiteAddress:    customermodel.Address{Mode: "MANUAL", Province: "DKI Jakarta", District: "Jakarta Selatan", SubDistrict: "Cilandak", Village: "Cipete Selatan", PreviewAddress: "Jl. Cipete Raya"},
		CompanyAddress: customermodel.Address{}, SalesExecutiveID: salesID,
	}
}

func TestLostProspectCannotBeConverted(t *testing.T) {
	repo := &fakeCustomerRepository{prospectStatus: prospectmodel.StatusLost}
	_, err := newCustomerService(repo).Convert(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, uuid.New(), validInput(uuid.New()))
	if !errors.Is(err, customerrepository.ErrProspectNotWon) {
		t.Fatalf("expected not won, got %v", err)
	}
}

func TestNonAdministratorCannotConvert(t *testing.T) {
	repo := &fakeCustomerRepository{prospectStatus: prospectmodel.StatusWon}
	_, err := newCustomerService(repo).Convert(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}, uuid.New(), validInput(uuid.New()))
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestConversionCreatesCustomerSetsConvertedAndPublishesToLists(t *testing.T) {
	salesID := uuid.New()
	repo := &fakeCustomerRepository{prospectStatus: prospectmodel.StatusWon, googlePlaceID: "sim-place-1"}
	service := newCustomerService(repo)
	admin := Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}
	created, err := service.Convert(context.Background(), admin, uuid.New(), validInput(salesID))
	if err != nil {
		t.Fatal(err)
	}
	if created.ParentCompanyID == uuid.Nil || created.CustomerCode == "" || repo.prospectStatus != prospectmodel.StatusConverted {
		t.Fatalf("conversion result incomplete: %+v status=%s", created, repo.prospectStatus)
	}
	adminItems, _ := service.AdminCustomers(context.Background(), admin)
	salesItems, _ := service.MyCustomers(context.Background(), Actor{UserID: salesID, Role: authmodel.RoleSalesExecutive})
	if len(adminItems) != 1 || len(salesItems) != 1 || adminItems[0].ID != salesItems[0].ID {
		t.Fatalf("converted customer was not published to both lists")
	}
	_, err = service.Convert(context.Background(), admin, created.SourceProspectID, validInput(salesID))
	if !errors.Is(err, customerrepository.ErrAlreadyConverted) {
		t.Fatalf("expected second conversion to fail, got %v", err)
	}
}

func TestConversionLinksExistingParentCompany(t *testing.T) {
	parent := customermodel.ParentCompany{ID: uuid.New(), ParentCode: "PC-000900", Name: "Existing Parent"}
	repo := &fakeCustomerRepository{prospectStatus: prospectmodel.StatusWon, parents: []customermodel.ParentCompany{parent}}
	input := validInput(uuid.New())
	input.ParentMethod = customermodel.ParentMethodExisting
	input.ExistingParentCompanyID = &parent.ID
	created, err := newCustomerService(repo).Convert(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, uuid.New(), input)
	if err != nil || created.ParentCompanyID != parent.ID || len(repo.parents) != 1 {
		t.Fatalf("existing parent was not linked: result=%+v err=%v", created, err)
	}
}

func TestDuplicateGooglePlaceIsBlocked(t *testing.T) {
	repo := &fakeCustomerRepository{prospectStatus: prospectmodel.StatusWon, duplicatePlace: true}
	_, err := newCustomerService(repo).Convert(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, uuid.New(), validInput(uuid.New()))
	if !errors.Is(err, customerrepository.ErrDuplicatePlace) {
		t.Fatalf("expected duplicate place, got %v", err)
	}
}

func TestAssignmentOverlapIsRejected(t *testing.T) {
	repo := &fakeCustomerRepository{prospectStatus: prospectmodel.StatusWon}
	input := validInput(uuid.New())
	owner := uuid.NewString()
	input.SalesAssignments = []customermodel.PeriodAssignment{
		{OwnerID: owner, StartMonth: 1, StartYear: 2026, End: "2026-12"},
		{OwnerID: owner, StartMonth: 6, StartYear: 2026, End: "UNTIL_NOW"},
	}
	_, err := newCustomerService(repo).Convert(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, uuid.New(), input)
	if !errors.Is(err, ErrAssignmentOverlap) {
		t.Fatalf("expected assignment overlap, got %v", err)
	}
}

func TestSalesAssignmentMustReferenceSalesUserID(t *testing.T) {
	repo := &fakeCustomerRepository{prospectStatus: prospectmodel.StatusWon}
	input := validInput(uuid.New())
	input.SalesAssignments = []customermodel.PeriodAssignment{{OwnerID: "not-a-user-id", StartMonth: 1, StartYear: 2026, End: "UNTIL_NOW"}}
	_, err := newCustomerService(repo).Convert(context.Background(), Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}, uuid.New(), input)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected sales assignment validation, got %v", err)
	}
}
