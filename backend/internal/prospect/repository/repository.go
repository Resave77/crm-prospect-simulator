package repository

import (
	"context"
	"errors"

	"crm-prospect-simulator/backend/internal/prospect/model"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("prospect not found")
	ErrInvalidStatus = errors.New("prospect status does not allow this operation")
	ErrNotOwner      = errors.New("prospect is not assigned to this sales executive")
	ErrDuplicate     = errors.New("Google Place is already saved as a prospect")
	ErrVisitOpen     = errors.New("prospect already has an open visit")
	ErrVisitClosed   = errors.New("visit is already checked out")
)

type Repository interface {
	ListAssigned(context.Context, uuid.UUID) ([]model.Prospect, error)
	ListWon(context.Context) ([]model.Prospect, error)
	ListAll(context.Context) ([]model.Prospect, error)
	ListSalesExecutives(context.Context) ([]model.SalesExecutive, error)
	FindReview(context.Context, uuid.UUID) (model.Review, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, model.Status, model.Status, string) (model.Prospect, error)
	Create(context.Context, model.SaveProspectInput, uuid.UUID) (model.Prospect, error)
	CheckIn(context.Context, uuid.UUID, uuid.UUID, model.CheckInInput) (model.Visit, error)
	CheckOut(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, model.CheckOutInput) (model.Visit, error)
}
