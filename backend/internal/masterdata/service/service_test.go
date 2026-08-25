package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	masterdatamodel "crm-prospect-simulator/backend/internal/masterdata/model"
	"crm-prospect-simulator/backend/internal/masterdata/repository"
	"github.com/google/uuid"
)

type fakeMasterDataRepository struct {
	segments   []masterdatamodel.Segment
	categories []masterdatamodel.Category
}

func (f *fakeMasterDataRepository) ListSegments(_ context.Context, search, status string) ([]masterdatamodel.Segment, error) {
	items := make([]masterdatamodel.Segment, 0)
	for _, segment := range f.segments {
		if search != "" && !strings.Contains(strings.ToLower(segment.Name), strings.ToLower(search)) {
			continue
		}
		if status != "" && segment.Status != status {
			continue
		}
		items = append(items, segment)
	}
	return items, nil
}

func (f *fakeMasterDataRepository) GetSegment(_ context.Context, id uuid.UUID) (masterdatamodel.Segment, error) {
	for _, segment := range f.segments {
		if segment.ID == id.String() {
			return segment, nil
		}
	}
	return masterdatamodel.Segment{}, repository.ErrNotFound
}

func (f *fakeMasterDataRepository) CreateSegment(_ context.Context, input masterdatamodel.SegmentInput) (masterdatamodel.Segment, error) {
	segment := masterdatamodel.Segment{ID: uuid.New().String(), Name: input.Name, Description: input.Description, Status: input.Status}
	f.segments = append(f.segments, segment)
	return segment, nil
}

func (f *fakeMasterDataRepository) UpdateSegment(ctx context.Context, id uuid.UUID, input masterdatamodel.SegmentInput) (masterdatamodel.Segment, error) {
	if _, err := f.GetSegment(ctx, id); err != nil {
		return masterdatamodel.Segment{}, err
	}
	segment := masterdatamodel.Segment{ID: id.String(), Name: input.Name, Description: input.Description, Status: input.Status}
	return segment, nil
}

func (f *fakeMasterDataRepository) DeleteSegment(context.Context, uuid.UUID) error { return nil }

func (f *fakeMasterDataRepository) ListTrashedSegments(context.Context) ([]masterdatamodel.Segment, error) {
	return nil, nil
}

func (f *fakeMasterDataRepository) RestoreSegment(context.Context, uuid.UUID) error { return nil }

func (f *fakeMasterDataRepository) ListCategories(context.Context, string, string, string) ([]masterdatamodel.Category, error) {
	return f.categories, nil
}

func (f *fakeMasterDataRepository) ListCategoriesBySegment(_ context.Context, segmentID uuid.UUID) ([]masterdatamodel.Category, error) {
	items := make([]masterdatamodel.Category, 0)
	for _, category := range f.categories {
		if category.SegmentID == segmentID.String() {
			items = append(items, category)
		}
	}
	return items, nil
}

func (f *fakeMasterDataRepository) GetCategory(context.Context, uuid.UUID) (masterdatamodel.Category, error) {
	return masterdatamodel.Category{}, repository.ErrNotFound
}

func (f *fakeMasterDataRepository) CreateCategory(_ context.Context, input masterdatamodel.CategoryInput) (masterdatamodel.Category, error) {
	category := masterdatamodel.Category{ID: uuid.New().String(), SegmentID: input.SegmentID, Name: input.Name, Status: input.Status}
	f.categories = append(f.categories, category)
	return category, nil
}

func (f *fakeMasterDataRepository) UpdateCategory(context.Context, uuid.UUID, masterdatamodel.CategoryInput) (masterdatamodel.Category, error) {
	return masterdatamodel.Category{}, nil
}

func (f *fakeMasterDataRepository) DeleteCategory(context.Context, uuid.UUID) error { return nil }

func (f *fakeMasterDataRepository) ListTrashedCategories(context.Context) ([]masterdatamodel.Category, error) {
	return nil, nil
}

func (f *fakeMasterDataRepository) RestoreCategory(context.Context, uuid.UUID) error { return nil }

func adminActor() Actor {
	return Actor{UserID: uuid.New(), Role: authmodel.RoleAdministrator}
}

func salesActor() Actor {
	return Actor{UserID: uuid.New(), Role: authmodel.RoleSalesExecutive}
}

func TestCreateSegmentRejectsEmptyName(t *testing.T) {
	svc := New(&fakeMasterDataRepository{})
	_, err := svc.CreateSegment(context.Background(), adminActor(), masterdatamodel.SegmentInput{Name: "   ", Status: "ACTIVE"})
	var validationErr *ValidationError
	if err == nil || !errors.As(err, &validationErr) {
		t.Fatalf("err=%v, want validation error", err)
	}
}

func TestCreateSegmentNormalizesAndCreates(t *testing.T) {
	repo := &fakeMasterDataRepository{}
	svc := New(repo)
	segment, err := svc.CreateSegment(context.Background(), adminActor(), masterdatamodel.SegmentInput{Name: "  Food & Beverage  ", Status: "active"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if segment.Name != "Food & Beverage" {
		t.Fatalf("name=%q, want trimmed name", segment.Name)
	}
	if segment.Status != "ACTIVE" {
		t.Fatalf("status=%q, want ACTIVE", segment.Status)
	}
}

func TestCreateCategoryRequiresExistingSegment(t *testing.T) {
	repo := &fakeMasterDataRepository{}
	svc := New(repo)
	_, err := svc.CreateCategory(context.Background(), adminActor(), masterdatamodel.CategoryInput{
		SegmentID: uuid.New().String(),
		Name:      "Restaurant",
		Status:    "ACTIVE",
	})
	var validationErr *ValidationError
	if err == nil || !errors.As(err, &validationErr) {
		t.Fatalf("err=%v, want validation error for unknown segment", err)
	}
}

func TestCreateCategorySucceedsWithValidSegment(t *testing.T) {
	segmentID := uuid.New()
	repo := &fakeMasterDataRepository{segments: []masterdatamodel.Segment{{ID: segmentID.String(), Name: "Food & Beverage", Status: "ACTIVE"}}}
	svc := New(repo)
	category, err := svc.CreateCategory(context.Background(), adminActor(), masterdatamodel.CategoryInput{
		SegmentID: segmentID.String(),
		Name:      "Restaurant",
		Status:    "ACTIVE",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if category.Name != "Restaurant" || category.SegmentID != segmentID.String() {
		t.Fatalf("category=%+v, unexpected", category)
	}
}

func TestSalesRoleIsForbidden(t *testing.T) {
	svc := New(&fakeMasterDataRepository{})
	if _, err := svc.ListSegments(context.Background(), salesActor(), "", ""); err != ErrForbidden {
		t.Fatalf("err=%v, want ErrForbidden", err)
	}
	if _, err := svc.CreateSegment(context.Background(), salesActor(), masterdatamodel.SegmentInput{Name: "X", Status: "ACTIVE"}); err != ErrForbidden {
		t.Fatalf("err=%v, want ErrForbidden", err)
	}
	if err := svc.DeleteCategory(context.Background(), salesActor(), uuid.New()); err != ErrForbidden {
		t.Fatalf("err=%v, want ErrForbidden", err)
	}
}

func TestTrashOperationsRequireAdminRole(t *testing.T) {
	svc := New(&fakeMasterDataRepository{})
	if _, err := svc.ListTrashedSegments(context.Background(), salesActor()); err != ErrForbidden {
		t.Fatalf("err=%v, want ErrForbidden", err)
	}
	if _, err := svc.ListTrashedCategories(context.Background(), salesActor()); err != ErrForbidden {
		t.Fatalf("err=%v, want ErrForbidden", err)
	}
	if err := svc.RestoreSegment(context.Background(), salesActor(), uuid.New()); err != ErrForbidden {
		t.Fatalf("err=%v, want ErrForbidden", err)
	}
	if err := svc.RestoreCategory(context.Background(), salesActor(), uuid.New()); err != ErrForbidden {
		t.Fatalf("err=%v, want ErrForbidden", err)
	}
}
