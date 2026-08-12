package service

import (
	"context"
	"testing"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	customermodel "crm-prospect-simulator/backend/internal/customer/model"
	"github.com/google/uuid"
)

type fakeCustomerRepository struct {
	teamCustomersByActor map[uuid.UUID]customermodel.TeamCustomers
	personalByActor      map[uuid.UUID][]customermodel.CustomerSite
}

func (f *fakeCustomerRepository) SearchParentCompanies(context.Context, string) ([]customermodel.ParentCompany, error) {
	return nil, nil
}
func (f *fakeCustomerRepository) ListActiveSalesExecutives(context.Context) ([]customermodel.UserOption, error) {
	return nil, nil
}
func (f *fakeCustomerRepository) Convert(context.Context, uuid.UUID, uuid.UUID, customermodel.ConversionInput) (customermodel.CustomerSite, error) {
	return customermodel.CustomerSite{}, nil
}
func (f *fakeCustomerRepository) AutoConvert(context.Context, uuid.UUID) (customermodel.CustomerSite, error) {
	return customermodel.CustomerSite{}, nil
}
func (f *fakeCustomerRepository) DeleteCustomer(context.Context, uuid.UUID) error { return nil }
func (f *fakeCustomerRepository) ListCustomers(context.Context) ([]customermodel.CustomerSite, error) {
	return nil, nil
}
func (f *fakeCustomerRepository) ListCustomersPaged(context.Context, customermodel.CustomerListParams) (customermodel.CustomerListResult, error) {
	return customermodel.CustomerListResult{}, nil
}
func (f *fakeCustomerRepository) ListFilterOptions(context.Context) (customermodel.ListFilterOptions, error) {
	return customermodel.ListFilterOptions{}, nil
}
func (f *fakeCustomerRepository) ListCustomersForSales(_ context.Context, actorID uuid.UUID) ([]customermodel.CustomerSite, error) {
	return f.personalByActor[actorID], nil
}
func (f *fakeCustomerRepository) ListTeamCustomers(_ context.Context, actorID uuid.UUID) (customermodel.TeamCustomers, error) {
	return f.teamCustomersByActor[actorID], nil
}
func (f *fakeCustomerRepository) FindCustomerForSales(context.Context, uuid.UUID, uuid.UUID) (customermodel.CustomerDetail, error) {
	return customermodel.CustomerDetail{}, nil
}
func (f *fakeCustomerRepository) FindCustomer(context.Context, uuid.UUID) (customermodel.CustomerDetail, error) {
	return customermodel.CustomerDetail{}, nil
}
func (f *fakeCustomerRepository) UpdateParentCompany(context.Context, uuid.UUID, customermodel.UpdateParentCompanyInput) (customermodel.ParentCompany, error) {
	return customermodel.ParentCompany{}, nil
}
func (f *fakeCustomerRepository) FindParentCompanyByCode(context.Context, string) (customermodel.ParentCompany, error) {
	return customermodel.ParentCompany{}, nil
}

func TestTeamCustomersIncludesDescendantsOnly(t *testing.T) {
	managerID := uuid.New()
	childID := uuid.New()
	nestedID := uuid.New()
	siblingID := uuid.New()
	unrelatedID := uuid.New()
	endedID := uuid.New()
	descendantCustomer := customermodel.CustomerSite{ID: uuid.New(), Name: "Child Customer", SalesExecutiveID: childID}
	nestedCustomer := customermodel.CustomerSite{ID: uuid.New(), Name: "Nested Customer", SalesExecutiveID: nestedID}
	repo := &fakeCustomerRepository{teamCustomersByActor: map[uuid.UUID]customermodel.TeamCustomers{
		managerID: {
			HasTeam:              true,
			DirectMemberCount:    1,
			TotalDescendantCount: 2,
			Customers:            []customermodel.CustomerSite{descendantCustomer, nestedCustomer},
		},
		siblingID:   {HasTeam: false, Customers: []customermodel.CustomerSite{}},
		unrelatedID: {HasTeam: false, Customers: []customermodel.CustomerSite{}},
		endedID:     {HasTeam: false, Customers: []customermodel.CustomerSite{}},
	}}
	svc := New(repo, nil)

	got, err := svc.TeamCustomers(context.Background(), Actor{
		UserID:         managerID,
		Role:           authmodel.RoleSalesManager,
		PermissionKeys: []string{"view_team_dashboard"},
	})
	if err != nil {
		t.Fatalf("team customers: %v", err)
	}
	if !got.HasTeam || got.TotalDescendantCount != 2 || len(got.Customers) != 2 {
		t.Fatalf("unexpected scoped result: %+v", got)
	}
	for _, forbiddenID := range []uuid.UUID{siblingID, unrelatedID, endedID} {
		for _, customer := range got.Customers {
			if customer.SalesExecutiveID == forbiddenID {
				t.Fatalf("included out-of-scope customer for %s", forbiddenID)
			}
		}
	}
}

func TestTeamCustomersRequiresTeamCapabilityButAllowsNoDescendants(t *testing.T) {
	actorID := uuid.New()
	svc := New(&fakeCustomerRepository{teamCustomersByActor: map[uuid.UUID]customermodel.TeamCustomers{
		actorID: {HasTeam: false, Customers: []customermodel.CustomerSite{}},
	}}, nil)

	if _, err := svc.TeamCustomers(context.Background(), Actor{UserID: actorID, Role: authmodel.RoleSalesManager}); err == nil {
		t.Fatal("expected forbidden without view_team_dashboard")
	}
	got, err := svc.TeamCustomers(context.Background(), Actor{
		UserID:         actorID,
		Role:           authmodel.RoleSalesManager,
		PermissionKeys: []string{"view_team_dashboard"},
	})
	if err != nil {
		t.Fatalf("team customers with permission: %v", err)
	}
	if got.HasTeam || len(got.Customers) != 0 {
		t.Fatalf("no-descendant actor should get empty normal result: %+v", got)
	}
}

func TestLeafPersonalMyCustomersUnchanged(t *testing.T) {
	leafID := uuid.New()
	personal := []customermodel.CustomerSite{{ID: uuid.New(), Name: "Own Customer", SalesExecutiveID: leafID}}
	svc := New(&fakeCustomerRepository{personalByActor: map[uuid.UUID][]customermodel.CustomerSite{leafID: personal}}, nil)

	got, err := svc.MyCustomers(context.Background(), Actor{
		UserID:         leafID,
		Role:           authmodel.RoleSalesExecutive,
		PermissionKeys: []string{"view_my_customers"},
	})
	if err != nil {
		t.Fatalf("my customers: %v", err)
	}
	if len(got) != 1 || got[0].ID != personal[0].ID {
		t.Fatalf("personal customers changed: %+v", got)
	}
}
