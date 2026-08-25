package repository

import (
	"context"
	"errors"

	"crm-prospect-simulator/backend/internal/masterdata/model"
	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("master data record not found")
	ErrDuplicateName     = errors.New("a record with the same name already exists")
	ErrSegmentInUse      = errors.New("segment still has categories and cannot be deleted")
	ErrParentInTrash     = errors.New("parent segment is still in trash")
)

type Repository interface {
	ListSegments(ctx context.Context, search, status string) ([]model.Segment, error)
	GetSegment(ctx context.Context, id uuid.UUID) (model.Segment, error)
	CreateSegment(ctx context.Context, input model.SegmentInput) (model.Segment, error)
	UpdateSegment(ctx context.Context, id uuid.UUID, input model.SegmentInput) (model.Segment, error)
	DeleteSegment(ctx context.Context, id uuid.UUID) error
	ListTrashedSegments(ctx context.Context) ([]model.Segment, error)
	RestoreSegment(ctx context.Context, id uuid.UUID) error

	ListCategories(ctx context.Context, search, segmentID, status string) ([]model.Category, error)
	ListCategoriesBySegment(ctx context.Context, segmentID uuid.UUID) ([]model.Category, error)
	GetCategory(ctx context.Context, id uuid.UUID) (model.Category, error)
	CreateCategory(ctx context.Context, input model.CategoryInput) (model.Category, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, input model.CategoryInput) (model.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	ListTrashedCategories(ctx context.Context) ([]model.Category, error)
	RestoreCategory(ctx context.Context, id uuid.UUID) error
}
