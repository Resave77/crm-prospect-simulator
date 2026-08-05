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
	ErrConflict      = errors.New("prospect is still referenced by existing records")
)

type Repository interface {
	ListAssigned(context.Context, uuid.UUID) ([]model.Prospect, error)
	ListWon(context.Context) ([]model.Prospect, error)
	ListAll(context.Context) ([]model.Prospect, error)
	ListSalesExecutives(context.Context) ([]model.SalesExecutive, error)
	ListMentionUsers(context.Context) ([]model.SalesExecutive, error)
	FindReview(context.Context, uuid.UUID) (model.Review, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, model.Status, model.Status, string) (model.Prospect, error)
	Create(context.Context, model.SaveProspectInput, uuid.UUID) (model.Prospect, error)
	CheckIn(context.Context, uuid.UUID, uuid.UUID, model.CheckInInput) (model.Visit, error)
	CheckOut(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, model.CheckOutInput) (model.Visit, error)
	ListVisitMonitoring(context.Context, model.VisitMonitoringFilter) ([]model.VisitMonitoringItem, error)
	ListMyVisits(context.Context, uuid.UUID, model.VisitMonitoringFilter) ([]model.VisitMonitoringItem, error)
	ListProspectVisits(context.Context, uuid.UUID) ([]model.Visit, error)
	DeleteVisit(context.Context, uuid.UUID, uuid.UUID) (model.Visit, error)
	DeleteProspect(context.Context, uuid.UUID) ([]string, error)
	RequestDeletion(context.Context, uuid.UUID, uuid.UUID) error
	ApproveDeletion(context.Context, uuid.UUID) ([]string, error)
	RejectDeletion(context.Context, uuid.UUID) error
	ListComments(context.Context, uuid.UUID) ([]model.ProspectComment, error)
	CreateComment(context.Context, uuid.UUID, uuid.UUID, string, []model.CommentAttachment) (model.ProspectComment, error)
	DeleteComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]model.CommentAttachment, error)
	FindCommentAttachment(context.Context, uuid.UUID, uuid.UUID) (model.CommentAttachment, error)
	FindProspectOwner(context.Context, uuid.UUID) (uuid.UUID, error)
	ListPhotoTags(context.Context, uuid.UUID) ([]model.ProspectPhotoTag, error)
	UpsertPhotoTag(context.Context, uuid.UUID, int, model.PhotoCategory, uuid.UUID) (model.ProspectPhotoTag, error)
	ProspectAccessibleTo(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}
