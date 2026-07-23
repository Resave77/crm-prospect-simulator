package repository

import (
	"context"
	"errors"

	"crm-prospect-simulator/backend/internal/customer/model"
	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("customer record not found")
	ErrProspectNotWon    = errors.New("prospect is not won")
	ErrAlreadyConverted  = errors.New("prospect was already converted")
	ErrDuplicatePlace    = errors.New("google place is already linked to a customer")
	ErrCodeConflict      = errors.New("generated code conflicts with an existing record")
	ErrParentUnavailable = errors.New("parent company cannot be created or linked")
	ErrSalesUnavailable  = errors.New("sales executive is not active")
)

type Repository interface {
	SearchParentCompanies(context.Context, string) ([]model.ParentCompany, error)
	ListActiveSalesExecutives(context.Context) ([]model.UserOption, error)
	Convert(context.Context, uuid.UUID, uuid.UUID, model.ConversionInput) (model.CustomerSite, error)
	ListCustomers(context.Context) ([]model.CustomerSite, error)
	ListCustomersForSales(context.Context, uuid.UUID) ([]model.CustomerSite, error)
	FindCustomerForSales(context.Context, uuid.UUID, uuid.UUID) (model.CustomerDetail, error)
}
