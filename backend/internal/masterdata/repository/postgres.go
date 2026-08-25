package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"crm-prospect-simulator/backend/internal/masterdata/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const segmentColumns = `
	SELECT s.id, s.name, s.description, s.status,
	       (SELECT count(*) FROM categories c WHERE c.segment_id = s.id AND c.deleted_at IS NULL) AS category_count,
	       s.created_at, s.updated_at, s.deleted_at
	FROM segments s`

const categoryColumns = `
	SELECT c.id, c.segment_id, s.name,
	       c.name, c.description, c.place_api, c.status, c.created_at, c.updated_at, c.deleted_at
	FROM categories c
	JOIN segments s ON s.id = c.segment_id`

func mapConstraint(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrDuplicateName
		case "23503":
			return ErrSegmentInUse
		}
	}
	return err
}

func (r *PostgresRepository) ListSegments(ctx context.Context, search, status string) ([]model.Segment, error) {
	query := segmentColumns + ` WHERE s.deleted_at IS NULL
		AND ($1 = '' OR s.name ILIKE '%' || $1 || '%')
		AND ($2 = '' OR s.status = $2)
		ORDER BY s.name`
	rows, err := r.pool.Query(ctx, query, strings.TrimSpace(search), strings.TrimSpace(status))
	if err != nil {
		return nil, fmt.Errorf("list segments: %w", err)
	}
	defer rows.Close()
	items := make([]model.Segment, 0)
	for rows.Next() {
		item, err := scanSegment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetSegment(ctx context.Context, id uuid.UUID) (model.Segment, error) {
	row := r.pool.QueryRow(ctx, segmentColumns+` WHERE s.deleted_at IS NULL AND s.id = $1`, id)
	item, err := scanSegment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Segment{}, ErrNotFound
	}
	if err != nil {
		return model.Segment{}, fmt.Errorf("get segment: %w", err)
	}
	return item, nil
}

func (r *PostgresRepository) CreateSegment(ctx context.Context, input model.SegmentInput) (model.Segment, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO segments (id, name, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, status, 0::bigint, created_at, updated_at, deleted_at`,
		uuid.New(), input.Name, input.Description, input.Status)
	item, err := scanSegment(row)
	if err != nil {
		return model.Segment{}, mapConstraint(fmt.Errorf("create segment: %w", err))
	}
	return item, nil
}

func (r *PostgresRepository) UpdateSegment(ctx context.Context, id uuid.UUID, input model.SegmentInput) (model.Segment, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE segments SET name = $2, description = $3, status = $4, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, description, status,
		          (SELECT count(*) FROM categories c WHERE c.segment_id = segments.id AND c.deleted_at IS NULL),
		          created_at, updated_at, deleted_at`,
		id, input.Name, input.Description, input.Status)
	item, err := scanSegment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Segment{}, ErrNotFound
	}
	if err != nil {
		return model.Segment{}, mapConstraint(fmt.Errorf("update segment: %w", err))
	}
	return item, nil
}

func (r *PostgresRepository) DeleteSegment(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE segments SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM categories c WHERE c.segment_id = segments.id AND c.deleted_at IS NULL
		  )`, id)
	if err != nil {
		return fmt.Errorf("delete segment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		var deletedAt *time.Time
		err := r.pool.QueryRow(ctx, `SELECT deleted_at FROM segments WHERE id = $1`, id).Scan(&deletedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("delete segment lookup: %w", err)
		}
		if deletedAt != nil {
			return ErrNotFound
		}
		return ErrSegmentInUse
	}
	return nil
}

func (r *PostgresRepository) ListTrashedSegments(ctx context.Context) ([]model.Segment, error) {
	rows, err := r.pool.Query(ctx, segmentColumns+` WHERE s.deleted_at IS NOT NULL ORDER BY s.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list trashed segments: %w", err)
	}
	defer rows.Close()
	items := make([]model.Segment, 0)
	for rows.Next() {
		item, err := scanSegment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) RestoreSegment(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE segments SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return mapConstraint(fmt.Errorf("restore segment: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ListCategories(ctx context.Context, search, segmentID, status string) ([]model.Category, error) {
	query := categoryColumns + ` WHERE c.deleted_at IS NULL
		AND ($1 = '' OR c.name ILIKE '%' || $1 || '%')
		AND ($2 = '' OR c.segment_id = $2::uuid)
		AND ($3 = '' OR c.status = $3)
		ORDER BY s.name, c.name`
	rows, err := r.pool.Query(ctx, query, strings.TrimSpace(search), strings.TrimSpace(segmentID), strings.TrimSpace(status))
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	items, err := collectCategories(rows)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) ListCategoriesBySegment(ctx context.Context, segmentID uuid.UUID) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx, categoryColumns+` WHERE c.deleted_at IS NULL AND c.segment_id = $1 ORDER BY c.name`, segmentID)
	if err != nil {
		return nil, fmt.Errorf("list categories by segment: %w", err)
	}
	defer rows.Close()
	items, err := collectCategories(rows)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) GetCategory(ctx context.Context, id uuid.UUID) (model.Category, error) {
	row := r.pool.QueryRow(ctx, categoryColumns+` WHERE c.deleted_at IS NULL AND c.id = $1`, id)
	item, err := scanCategory(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Category{}, ErrNotFound
	}
	if err != nil {
		return model.Category{}, fmt.Errorf("get category: %w", err)
	}
	return item, nil
}

func (r *PostgresRepository) CreateCategory(ctx context.Context, input model.CategoryInput) (model.Category, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO categories (id, segment_id, name, description, place_api, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, segment_id, '', name, description, place_api, status, created_at, updated_at, deleted_at`,
		uuid.New(), input.SegmentID, input.Name, input.Description, input.PlaceAPI, input.Status)
	item, err := scanCategory(row)
	if err != nil {
		return model.Category{}, mapConstraint(fmt.Errorf("create category: %w", err))
	}
	return r.reloadCategory(ctx, item.ID)
}

func (r *PostgresRepository) UpdateCategory(ctx context.Context, id uuid.UUID, input model.CategoryInput) (model.Category, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE categories SET segment_id = $2, name = $3, description = $4, place_api = $5, status = $6, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, segment_id, '', name, description, place_api, status, created_at, updated_at, deleted_at`,
		id, input.SegmentID, input.Name, input.Description, input.PlaceAPI, input.Status)
	item, err := scanCategory(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Category{}, ErrNotFound
	}
	if err != nil {
		return model.Category{}, mapConstraint(fmt.Errorf("update category: %w", err))
	}
	return r.reloadCategory(ctx, item.ID)
}

func (r *PostgresRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE categories SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return mapConstraint(fmt.Errorf("delete category: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ListTrashedCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx, categoryColumns+` WHERE c.deleted_at IS NOT NULL ORDER BY c.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list trashed categories: %w", err)
	}
	defer rows.Close()
	items, err := collectCategories(rows)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) RestoreCategory(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE categories SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL
		  AND EXISTS (
		    SELECT 1 FROM segments s WHERE s.id = categories.segment_id AND s.deleted_at IS NULL
		  )`, id)
	if err != nil {
		return mapConstraint(fmt.Errorf("restore category: %w", err))
	}
	if tag.RowsAffected() == 0 {
		var deletedAt *time.Time
		err := r.pool.QueryRow(ctx, `SELECT deleted_at FROM categories WHERE id = $1`, id).Scan(&deletedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("restore category lookup: %w", err)
		}
		if deletedAt == nil {
			return ErrNotFound
		}
		return ErrParentInTrash
	}
	return nil
}

func (r *PostgresRepository) reloadCategory(ctx context.Context, id string) (model.Category, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return model.Category{}, ErrNotFound
	}
	return r.GetCategory(ctx, parsed)
}

type rowScanner interface{ Scan(dest ...any) error }

func scanSegment(row rowScanner) (model.Segment, error) {
	var item model.Segment
	if err := row.Scan(&item.ID, &item.Name, &item.Description, &item.Status, &item.CategoryCount, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Segment{}, ErrNotFound
		}
		return model.Segment{}, fmt.Errorf("scan segment: %w", err)
	}
	return item, nil
}

func scanCategory(row rowScanner) (model.Category, error) {
	var item model.Category
	if err := row.Scan(&item.ID, &item.SegmentID, &item.SegmentName, &item.Name, &item.Description, &item.PlaceAPI, &item.Status, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, ErrNotFound
		}
		return model.Category{}, fmt.Errorf("scan category: %w", err)
	}
	return item, nil
}

func collectCategories(rows pgx.Rows) ([]model.Category, error) {
	items := make([]model.Category, 0)
	for rows.Next() {
		item, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
