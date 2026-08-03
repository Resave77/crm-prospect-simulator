package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"crm-prospect-simulator/backend/internal/admin/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepository) ListSalesRoles(ctx context.Context) ([]model.SalesRole, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, level, COALESCE(description,''), is_active, created_by, updated_by, created_at, updated_at FROM sales_roles ORDER BY level, name`)
	if err != nil {
		return nil, fmt.Errorf("list sales roles: %w", err)
	}
	defer rows.Close()
	items := []model.SalesRole{}
	for rows.Next() {
		item, err := scanSalesRole(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) FindSalesRole(ctx context.Context, id uuid.UUID) (model.SalesRole, error) {
	return scanSalesRole(r.pool.QueryRow(ctx, `SELECT id, name, level, COALESCE(description,''), is_active, created_by, updated_by, created_at, updated_at FROM sales_roles WHERE id=$1`, id))
}

func (r *PostgresRepository) CreateSalesRole(ctx context.Context, id uuid.UUID, input model.CreateSalesRoleInput, actorID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO sales_roles (id,name,normalized_name,level,description,created_by,updated_by,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$6,now())`, id, strings.TrimSpace(input.Name), normalizeName(input.Name), input.Level, strings.TrimSpace(input.Description), actorID)
	return mapError(err)
}

func (r *PostgresRepository) UpdateSalesRole(ctx context.Context, id uuid.UUID, input model.UpdateSalesRoleInput, actorID uuid.UUID) error {
	sets := []string{"updated_by = $2", "updated_at = now()"}
	args := []any{id, actorID}
	idx := 3
	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, strings.TrimSpace(*input.Name))
		idx++
		sets = append(sets, fmt.Sprintf("normalized_name = $%d", idx))
		args = append(args, normalizeName(*input.Name))
		idx++
	}
	if input.Level != nil {
		sets = append(sets, fmt.Sprintf("level = $%d", idx))
		args = append(args, *input.Level)
		idx++
	}
	if input.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", idx))
		args = append(args, strings.TrimSpace(*input.Description))
		idx++
	}
	cmd, err := r.pool.Exec(ctx, `UPDATE sales_roles SET `+strings.Join(sets, ", ")+` WHERE id = $1`, args...)
	if err != nil {
		return mapError(err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) UpdateSalesRoleStatus(ctx context.Context, id uuid.UUID, isActive bool, actorID uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `UPDATE sales_roles SET is_active=$2, updated_by=$3, updated_at=now() WHERE id=$1`, id, isActive, actorID)
	if err != nil {
		return mapError(err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) SalesRoleNameExists(ctx context.Context, normalizedName string, excludeID *uuid.UUID) (bool, error) {
	var exists bool
	var err error
	if excludeID == nil {
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sales_roles WHERE normalized_name=$1 AND is_active=true)`, normalizedName).Scan(&exists)
	} else {
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sales_roles WHERE normalized_name=$1 AND is_active=true AND id<>$2)`, normalizedName, *excludeID).Scan(&exists)
	}
	return exists, err
}

func (r *PostgresRepository) SalesRoleHasAssignments(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sales_structure_assignments WHERE sales_role_id=$1)`, id).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, id).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) CreateSalesAssignment(ctx context.Context, id uuid.UUID, input model.CreateAssignmentInput, actorID uuid.UUID) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO sales_structure_assignments (id,user_id,sales_role_id,parent_user_id,effective_from,created_by,updated_at) VALUES ($1,$2,$3,$4,$5,$6,now())`, id, input.UserID, input.SalesRoleID, input.ParentUserID, input.EffectiveFrom.Time, actorID)
	if err != nil {
		return mapError(err)
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) MoveSalesAssignment(ctx context.Context, currentID uuid.UUID, newID uuid.UUID, input model.MoveAssignmentInput, actorID uuid.UUID) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID uuid.UUID
	var from time.Time
	err = tx.QueryRow(ctx, `SELECT user_id, effective_from FROM sales_structure_assignments WHERE id=$1 FOR UPDATE`, currentID).Scan(&userID, &from)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	closeDate := input.EffectiveFrom.Time.AddDate(0, 0, -1)
	cmd, err := tx.Exec(ctx, `UPDATE sales_structure_assignments SET effective_to=$2, updated_at=now() WHERE id=$1`, currentID, closeDate)
	if err != nil {
		return mapError(err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	_, err = tx.Exec(ctx, `INSERT INTO sales_structure_assignments (id,user_id,sales_role_id,parent_user_id,effective_from,created_by,updated_at) VALUES ($1,$2,$3,$4,$5,$6,now())`, newID, userID, input.SalesRoleID, input.ParentUserID, input.EffectiveFrom.Time, actorID)
	if err != nil {
		return mapError(err)
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) FindSalesAssignment(ctx context.Context, id uuid.UUID) (model.SalesStructureAssignment, error) {
	return scanAssignment(r.pool.QueryRow(ctx, `SELECT id,user_id,sales_role_id,parent_user_id,effective_from,effective_to FROM sales_structure_assignments WHERE id=$1`, id))
}

func (r *PostgresRepository) FindEffectiveSalesAssignment(ctx context.Context, userID uuid.UUID, effectiveDate time.Time) (model.SalesStructureAssignment, model.SalesRole, error) {
	row := r.pool.QueryRow(ctx, `SELECT a.id,a.user_id,a.sales_role_id,a.parent_user_id,a.effective_from,a.effective_to,r.id,r.name,r.level,COALESCE(r.description,''),r.is_active,r.created_by,r.updated_by,r.created_at,r.updated_at FROM sales_structure_assignments a JOIN sales_roles r ON r.id=a.sales_role_id WHERE a.user_id=$1 AND a.effective_from <= $2 AND (a.effective_to IS NULL OR a.effective_to >= $2) ORDER BY a.effective_from DESC LIMIT 1`, userID, effectiveDate)
	var a model.SalesStructureAssignment
	var role model.SalesRole
	err := row.Scan(&a.ID, &a.UserID, &a.SalesRoleID, &a.ParentUserID, &a.EffectiveFrom, &a.EffectiveTo, &role.ID, &role.Name, &role.Level, &role.Description, &role.IsActive, &role.CreatedBy, &role.UpdatedBy, &role.CreatedAt, &role.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, role, ErrNotFound
	}
	return a, role, err
}

func (r *PostgresRepository) SalesAssignmentOverlaps(ctx context.Context, userID uuid.UUID, from time.Time, to *time.Time, excludeID *uuid.UUID) (bool, error) {
	end := to
	var exists bool
	args := []any{userID, from, end}
	q := `SELECT EXISTS(SELECT 1 FROM sales_structure_assignments WHERE user_id=$1 AND effective_from <= COALESCE($3::date, '9999-12-31'::date) AND COALESCE(effective_to, '9999-12-31'::date) >= $2`
	if excludeID != nil {
		q += ` AND id<>$4`
		args = append(args, *excludeID)
	}
	q += `)`
	err := r.pool.QueryRow(ctx, q, args...).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) ListSalesStructure(ctx context.Context, effectiveDate time.Time) ([]model.SalesStructureItem, error) {
	rows, err := r.pool.Query(ctx, `SELECT a.id,u.id,u.full_name,u.role::text,r.id,r.name,r.level,a.parent_user_id,p.full_name,a.effective_from,a.effective_to FROM sales_structure_assignments a JOIN users u ON u.id=a.user_id JOIN sales_roles r ON r.id=a.sales_role_id LEFT JOIN users p ON p.id=a.parent_user_id WHERE a.effective_from <= $1 AND (a.effective_to IS NULL OR a.effective_to >= $1) ORDER BY r.level,u.full_name`, effectiveDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.SalesStructureItem{}
	for rows.Next() {
		item, err := scanStructureItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListSalesAssignmentHistory(ctx context.Context, userID uuid.UUID) ([]model.AssignmentHistoryItem, error) {
	rows, err := r.pool.Query(ctx, `SELECT a.id,r.id,r.name,r.level,a.parent_user_id,p.full_name,a.effective_from,a.effective_to FROM sales_structure_assignments a JOIN sales_roles r ON r.id=a.sales_role_id LEFT JOIN users p ON p.id=a.parent_user_id WHERE a.user_id=$1 ORDER BY a.effective_from DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.AssignmentHistoryItem{}
	for rows.Next() {
		item, err := scanHistoryItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanSalesRole(row pgx.Row) (model.SalesRole, error) {
	var item model.SalesRole
	err := row.Scan(&item.ID, &item.Name, &item.Level, &item.Description, &item.IsActive, &item.CreatedBy, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, ErrNotFound
	}
	return item, err
}
func scanAssignment(row pgx.Row) (model.SalesStructureAssignment, error) {
	var item model.SalesStructureAssignment
	err := row.Scan(&item.ID, &item.UserID, &item.SalesRoleID, &item.ParentUserID, &item.EffectiveFrom, &item.EffectiveTo)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, ErrNotFound
	}
	return item, err
}

func scanStructureItem(row pgx.Row) (model.SalesStructureItem, error) {
	var item model.SalesStructureItem
	var parentName pgtype.Text
	var from time.Time
	var to *time.Time
	err := row.Scan(&item.AssignmentID, &item.UserID, &item.SalesName, &item.SystemRole, &item.SalesRole.ID, &item.SalesRole.Name, &item.SalesRole.Level, &item.ParentUserID, &parentName, &from, &to)
	if err != nil {
		return item, err
	}
	if parentName.Valid {
		item.ParentName = &parentName.String
	}
	item.EffectiveFrom = from.Format(model.DateLayout)
	if to != nil {
		s := to.Format(model.DateLayout)
		item.EffectiveTo = &s
	}
	return item, nil
}
func scanHistoryItem(row pgx.Row) (model.AssignmentHistoryItem, error) {
	var item model.AssignmentHistoryItem
	var parentName pgtype.Text
	var from time.Time
	var to *time.Time
	err := row.Scan(&item.AssignmentID, &item.SalesRole.ID, &item.SalesRole.Name, &item.SalesRole.Level, &item.ParentUserID, &parentName, &from, &to)
	if err != nil {
		return item, err
	}
	if parentName.Valid {
		item.ParentName = &parentName.String
	}
	item.EffectiveFrom = from.Format(model.DateLayout)
	if to != nil {
		s := to.Format(model.DateLayout)
		item.EffectiveTo = &s
	}
	return item, nil
}
func normalizeName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}
