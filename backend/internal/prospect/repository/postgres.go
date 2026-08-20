package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"crm-prospect-simulator/backend/internal/prospect/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func decodeAttachments(raw []byte) ([]model.CommentAttachment, error) {
	items := make([]model.CommentAttachment, 0)
	if len(raw) == 0 {
		return items, nil
	}
	return items, json.Unmarshal(raw, &items)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const prospectSelect = `
	SELECT p.id, p.google_place_id, p.place_name, p.formatted_address,
	       p.latitude, p.longitude, p.place_category, p.place_types,
	       p.industry_group,
	       COALESCE(p.phone_number, ''), COALESCE(p.website_url, ''), COALESCE(p.google_maps_url, ''), p.assigned_sales_executive_id,
	       u.full_name, p.visit_notes, p.follow_up_notes, p.status::text,
	       p.deletion_requested, p.converted_at, p.created_at, p.updated_at
	FROM prospects p
	JOIN users u ON u.id = p.assigned_sales_executive_id`

func (r *PostgresRepository) ListAssigned(ctx context.Context, salesExecutiveID uuid.UUID) ([]model.Prospect, error) {
	rows, err := r.pool.Query(ctx, prospectSelect+`
		WHERE p.assigned_sales_executive_id = $1
		ORDER BY p.updated_at DESC`, salesExecutiveID)
	if err != nil {
		return nil, fmt.Errorf("list assigned prospects: %w", err)
	}
	defer rows.Close()
	return scanProspects(rows)
}

func (r *PostgresRepository) TeamDashboard(ctx context.Context, actorID uuid.UUID) (model.TeamDashboard, error) {
	activeAssignment := func(alias string) string {
		return alias + `.effective_from <= CURRENT_DATE AND (` + alias + `.effective_to IS NULL OR ` + alias + `.effective_to >= CURRENT_DATE)`
	}
	var item model.TeamDashboard
	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.full_name, sr.name, sr.level, a.effective_from
		FROM sales_structure_assignments a
		JOIN users u ON u.id = a.user_id
		JOIN sales_roles sr ON sr.id = a.sales_role_id
		WHERE a.user_id = $1 AND `+activeAssignment("a")+` AND u.deleted_at IS NULL
		ORDER BY a.effective_from DESC
		LIMIT 1`, actorID).Scan(&item.Lead.ID, &item.Lead.FullName, &item.Lead.RoleName, &item.Lead.RoleLevel, &item.Lead.EffectiveFrom)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.TeamDashboard{Lead: model.TeamLeadInfo{ID: actorID}, PipelineCounts: map[model.Status]int{}}, nil
	}
	if err != nil {
		return model.TeamDashboard{}, fmt.Errorf("read team lead assignment: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE descendants AS (
			SELECT a.user_id, a.parent_user_id, sr.name AS role_name, sr.level AS role_level, 1 AS depth
			FROM sales_structure_assignments a
			JOIN sales_roles sr ON sr.id = a.sales_role_id
			JOIN users u ON u.id = a.user_id
			WHERE a.parent_user_id = $1 AND `+activeAssignment("a")+` AND u.deleted_at IS NULL
			UNION ALL
			SELECT child.user_id, child.parent_user_id, child_role.name, child_role.level, d.depth + 1
			FROM sales_structure_assignments child
			JOIN sales_roles child_role ON child_role.id = child.sales_role_id
			JOIN users child_user ON child_user.id = child.user_id
			JOIN descendants d ON d.user_id = child.parent_user_id
			WHERE `+activeAssignment("child")+` AND child_user.deleted_at IS NULL
		)
		SELECT d.user_id, u.full_name, d.role_name, d.role_level, d.parent_user_id,
		       COUNT(DISTINCT p.id) FILTER (WHERE p.status IN ('NEW_LEAD','CONTACTED','INTERESTED','QUALIFIED','PROPOSAL_SENT','NEGOTIATION'))::int,
		       COUNT(DISTINCT cs.id)::int,
		       COUNT(DISTINCT v.id) FILTER (WHERE v.check_in_at::date = CURRENT_DATE)::int,
		       COUNT(DISTINCT v.id) FILTER (WHERE v.check_in_at::date = CURRENT_DATE AND v.check_out_at IS NOT NULL)::int,
		       COUNT(DISTINCT v.id) FILTER (WHERE v.check_in_at::date = CURRENT_DATE AND v.check_out_at IS NULL)::int
		FROM descendants d
		JOIN users u ON u.id = d.user_id
		LEFT JOIN prospects p ON p.assigned_sales_executive_id = d.user_id
		LEFT JOIN customer_sites cs ON cs.sales_executive_id = d.user_id
		LEFT JOIN prospect_visits v ON v.sales_executive_id = d.user_id
		GROUP BY d.user_id, u.full_name, d.role_name, d.role_level, d.parent_user_id, d.depth
		ORDER BY d.depth, d.role_level, u.full_name`, actorID)
	if err != nil {
		return model.TeamDashboard{}, fmt.Errorf("read team descendants: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var member model.TeamMemberSummary
		if err := rows.Scan(&member.UserID, &member.FullName, &member.RoleName, &member.RoleLevel, &member.ParentUserID, &member.ActiveProspects, &member.Customers, &member.VisitsToday, &member.CompletedVisits, &member.PendingVisits); err != nil {
			return model.TeamDashboard{}, err
		}
		item.Members = append(item.Members, member)
		item.TotalDescendantCount++
		if member.ParentUserID == actorID {
			item.DirectMemberCount++
		}
		item.ActiveProspects += member.ActiveProspects
		item.Customers += member.Customers
		item.VisitsToday += member.VisitsToday
		item.CompletedVisits += member.CompletedVisits
		item.PendingVisits += member.PendingVisits
	}
	if err := rows.Err(); err != nil {
		return model.TeamDashboard{}, err
	}
	item.HasTeam = item.TotalDescendantCount > 0

	item.PipelineCounts = map[model.Status]int{}
	statusRows, err := r.pool.Query(ctx, `
		WITH RECURSIVE descendants AS (
			SELECT a.user_id
			FROM sales_structure_assignments a
			JOIN users u ON u.id = a.user_id
			WHERE a.parent_user_id = $1 AND `+activeAssignment("a")+` AND u.deleted_at IS NULL
			UNION ALL
			SELECT child.user_id
			FROM sales_structure_assignments child
			JOIN users child_user ON child_user.id = child.user_id
			JOIN descendants d ON d.user_id = child.parent_user_id
			WHERE `+activeAssignment("child")+` AND child_user.deleted_at IS NULL
		)
		SELECT p.status::text, COUNT(*)::int
		FROM prospects p
		JOIN descendants d ON d.user_id = p.assigned_sales_executive_id
		GROUP BY p.status`, actorID)
	if err != nil {
		return model.TeamDashboard{}, fmt.Errorf("read team pipeline counts: %w", err)
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var status model.Status
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return model.TeamDashboard{}, err
		}
		item.PipelineCounts[status] = count
	}
	return item, statusRows.Err()
}

func (r *PostgresRepository) ListAll(ctx context.Context) ([]model.Prospect, error) {
	rows, err := r.pool.Query(ctx, prospectSelect+` ORDER BY p.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list prospects: %w", err)
	}
	defer rows.Close()
	return scanProspects(rows)
}

func (r *PostgresRepository) ListSalesExecutives(ctx context.Context) ([]model.SalesExecutive, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.full_name, COUNT(p.id)::int AS active_prospect_count
		FROM users u
		LEFT JOIN prospects p ON p.assigned_sales_executive_id = u.id AND p.status IN ('NEW_LEAD','CONTACTED','INTERESTED','QUALIFIED','PROPOSAL_SENT','NEGOTIATION')
		WHERE u.role = 'SALES_EXECUTIVE' AND u.status = 'ACTIVE'
		GROUP BY u.id, u.full_name
		ORDER BY u.full_name`)
	if err != nil {
		return nil, fmt.Errorf("list sales executives: %w", err)
	}
	defer rows.Close()
	items := make([]model.SalesExecutive, 0)
	for rows.Next() {
		var item model.SalesExecutive
		if err := rows.Scan(&item.ID, &item.FullName, &item.ActiveProspectCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListMentionUsers(ctx context.Context) ([]model.SalesExecutive, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, full_name, 0 FROM users WHERE status='ACTIVE' ORDER BY full_name`)
	if err != nil {
		return nil, fmt.Errorf("list mention users: %w", err)
	}
	defer rows.Close()
	items := make([]model.SalesExecutive, 0)
	for rows.Next() {
		var item model.SalesExecutive
		if err := rows.Scan(&item.ID, &item.FullName, &item.ActiveProspectCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListWon(ctx context.Context) ([]model.Prospect, error) {
	rows, err := r.pool.Query(ctx, prospectSelect+`
		WHERE p.status = 'WON' ORDER BY p.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list won prospects: %w", err)
	}
	defer rows.Close()
	return scanProspects(rows)
}

func (r *PostgresRepository) FindReview(ctx context.Context, id uuid.UUID) (model.Review, error) {
	prospect, err := scanProspect(r.pool.QueryRow(ctx, prospectSelect+` WHERE p.id = $1`, id))
	if err != nil {
		return model.Review{}, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT h.id, h.from_status::text, h.to_status::text, h.changed_by_user_id,
		       u.full_name, h.notes, h.created_at
		FROM prospect_status_history h
		JOIN users u ON u.id = h.changed_by_user_id
		WHERE h.prospect_id = $1 ORDER BY h.created_at`, id)
	if err != nil {
		return model.Review{}, fmt.Errorf("list prospect history: %w", err)
	}
	defer rows.Close()
	history := make([]model.StatusHistory, 0)
	for rows.Next() {
		var item model.StatusHistory
		var from *string
		if err := rows.Scan(&item.ID, &from, &item.ToStatus, &item.ChangedByUserID, &item.ChangedByName, &item.Notes, &item.CreatedAt); err != nil {
			return model.Review{}, fmt.Errorf("scan prospect history: %w", err)
		}
		if from != nil {
			status := model.Status(*from)
			item.FromStatus = &status
		}
		history = append(history, item)
	}
	if err := rows.Err(); err != nil {
		return model.Review{}, err
	}
	visitRows, err := r.pool.Query(ctx, `
		SELECT v.id, v.prospect_id, v.sales_executive_id, u.full_name,
		       v.check_in_at, v.check_in_latitude, v.check_in_longitude,
		       v.check_out_at, v.check_out_latitude, v.check_out_longitude,
		       v.selfie_reference, v.visit_notes, v.follow_up_notes,
		       v.visit_result, v.visit_outcome
		FROM prospect_visits v JOIN users u ON u.id = v.sales_executive_id
		WHERE v.prospect_id = $1 ORDER BY v.check_in_at DESC`, id)
	if err != nil {
		return model.Review{}, fmt.Errorf("list prospect visits: %w", err)
	}
	defer visitRows.Close()
	visits := make([]model.Visit, 0)
	for visitRows.Next() {
		item, scanErr := scanVisit(visitRows)
		if scanErr != nil {
			return model.Review{}, scanErr
		}
		visits = append(visits, item)
	}
	return model.Review{Prospect: prospect, History: history, Visits: visits}, visitRows.Err()
}

func (r *PostgresRepository) Transition(ctx context.Context, id, salesExecutiveID uuid.UUID, expected, status model.Status, notes string) (model.Prospect, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Prospect{}, fmt.Errorf("begin prospect decision: %w", err)
	}
	defer tx.Rollback(ctx)

	var current model.Status
	var owner uuid.UUID
	err = tx.QueryRow(ctx, `SELECT status::text, assigned_sales_executive_id FROM prospects WHERE id = $1 FOR UPDATE`, id).Scan(&current, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Prospect{}, ErrNotFound
	}
	if err != nil {
		return model.Prospect{}, fmt.Errorf("lock prospect decision: %w", err)
	}
	if owner != salesExecutiveID {
		return model.Prospect{}, ErrNotOwner
	}
	if current != expected {
		return model.Prospect{}, ErrInvalidStatus
	}
	if _, err = tx.Exec(ctx, `UPDATE prospects SET status = $2, updated_at = now() WHERE id = $1`, id, status); err != nil {
		return model.Prospect{}, fmt.Errorf("update prospect decision: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO prospect_status_history
			(id, prospect_id, from_status, to_status, changed_by_user_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6)`, uuid.New(), id, current, status, salesExecutiveID, notes); err != nil {
		return model.Prospect{}, fmt.Errorf("record prospect decision: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return model.Prospect{}, fmt.Errorf("commit prospect decision: %w", err)
	}
	return scanProspect(r.pool.QueryRow(ctx, prospectSelect+` WHERE p.id = $1`, id))
}

func (r *PostgresRepository) Create(ctx context.Context, input model.SaveProspectInput, administratorID uuid.UUID) (model.Prospect, error) {
	placeTypes, err := json.Marshal(input.Place.PlaceTypes)
	if err != nil {
		return model.Prospect{}, fmt.Errorf("encode place types: %w", err)
	}
	id := uuid.New()
	_, err = r.pool.Exec(ctx, `
		WITH inserted AS (
			INSERT INTO prospects (id, google_place_id, place_name, formatted_address, latitude, longitude,
				place_category, industry_group, place_types, phone_number, website_url, google_maps_url, assigned_sales_executive_id, status, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,'NEW_LEAD',now()) RETURNING id
		)
		INSERT INTO prospect_status_history (id, prospect_id, from_status, to_status, changed_by_user_id, notes)
		SELECT $14, id, NULL, 'NEW_LEAD', $15, 'Saved from Prospect Finder and assigned' FROM inserted`,
		id, input.Place.GooglePlaceID, input.Place.PlaceName, input.Place.FormattedAddress,
		input.Place.Latitude, input.Place.Longitude, input.Place.PlaceCategory, input.IndustryGroup,
		placeTypes, input.Place.PhoneNumber, input.Place.WebsiteURL, input.Place.GoogleMapsURL,
		input.AssignedSalesExecutiveID, uuid.New(), administratorID)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			return model.Prospect{}, ErrDuplicate
		}
		return model.Prospect{}, fmt.Errorf("create prospect: %w", err)
	}
	return scanProspect(r.pool.QueryRow(ctx, prospectSelect+` WHERE p.id = $1`, id))
}

func (r *PostgresRepository) CheckIn(ctx context.Context, prospectID, salesExecutiveID uuid.UUID, input model.CheckInInput) (model.Visit, error) {
	selfie := input.SelfieReference
	id := uuid.New()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Visit{}, fmt.Errorf("begin check in visit: %w", err)
	}
	defer tx.Rollback(ctx)

	// Validate Visit B and ownership before resolving any previous active visit.
	var eligible bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM prospects p
			WHERE p.id=$1 AND p.assigned_sales_executive_id=$2 AND p.status NOT IN ('LOST')
			OR EXISTS (
				SELECT 1 FROM customer_sites cs
				WHERE cs.source_prospect_id=p.id AND cs.sales_executive_id=$2
			)
		)`, prospectID, salesExecutiveID).Scan(&eligible); err != nil {
		return model.Visit{}, fmt.Errorf("validate check in visit: %w", err)
	}
	if !eligible {
		return model.Visit{}, ErrNotOwner
	}

	// A Sales user may have only one active visit. Resolve an older visit only
	// after Visit B has passed ownership/status validation. The current Visit B
	// coordinates are the only reliable location available at this event.
	if _, err := tx.Exec(ctx, `
		UPDATE prospect_visits
		SET check_out_at=now(), check_out_latitude=$2, check_out_longitude=$3,
			visit_outcome=CASE WHEN visit_outcome='' THEN 'AUTO_NEXT_VISIT' ELSE visit_outcome || ' | AUTO_NEXT_VISIT' END,
			updated_at=now()
		WHERE sales_executive_id=$1 AND check_out_at IS NULL AND prospect_id<>$4`,
		salesExecutiveID, input.Latitude, input.Longitude, prospectID); err != nil {
		return model.Visit{}, fmt.Errorf("auto check out previous visit: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO prospect_visits (id, prospect_id, sales_executive_id, check_in_at, check_in_latitude, check_in_longitude, selfie_reference, visit_notes, updated_at)
		SELECT $1, p.id, $3, now(), $4, $5, $6, $7, now() FROM prospects p
		WHERE p.id = $2
		  AND (
			p.assigned_sales_executive_id = $3
			OR EXISTS (
				SELECT 1 FROM customer_sites cs
				WHERE cs.source_prospect_id = p.id AND cs.sales_executive_id = $3
			)
		  )
		  AND p.status NOT IN ('LOST')`,
		id, prospectID, salesExecutiveID, input.Latitude, input.Longitude, selfie, input.VisitNotes)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			return model.Visit{}, ErrVisitOpen
		}
		return model.Visit{}, fmt.Errorf("check in prospect visit: %w", err)
	}
	if input.VisitNotes != "" {
		if _, err := tx.Exec(ctx, `UPDATE prospects SET visit_notes=$2, updated_at=now() WHERE id=$1`, prospectID, input.VisitNotes); err != nil {
			return model.Visit{}, fmt.Errorf("save visit notes: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Visit{}, fmt.Errorf("commit check in visit: %w", err)
	}
	return r.findVisit(ctx, id)
}

const checkoutToleranceMeters = 500.0

func (r *PostgresRepository) CheckOut(ctx context.Context, prospectID, visitID, salesExecutiveID uuid.UUID, input model.CheckOutInput) (model.Visit, error) {
	command, err := r.pool.Exec(ctx, `
		UPDATE prospect_visits v SET check_out_at=now(), check_out_latitude=$4, check_out_longitude=$5, follow_up_notes=$6, visit_result=$7, visit_outcome=$8, updated_at=now()
		FROM prospects p WHERE v.id=$2 AND v.prospect_id=$1 AND p.id=v.prospect_id
		AND v.sales_executive_id=$3
		AND (
			p.assigned_sales_executive_id=$3
			OR EXISTS (
				SELECT 1 FROM customer_sites cs
				WHERE cs.source_prospect_id = p.id AND cs.sales_executive_id = $3
			)
		)
		AND v.check_out_at IS NULL
		-- Checkout tolerance is measured from the actual check-in position.
		-- This allows the user to finish the visit without returning to the
		-- prospect/customer point, while still limiting location drift.
		AND ($9 OR (2 * 6371000 * ASIN(SQRT(POWER(SIN(RADIANS($4 - v.check_in_latitude) / 2), 2) + COS(RADIANS(v.check_in_latitude)) * COS(RADIANS($4)) * POWER(SIN(RADIANS($5 - v.check_in_longitude) / 2), 2))))) <= `+fmt.Sprintf("%.0f", checkoutToleranceMeters)+`))`,
		prospectID, visitID, salesExecutiveID, input.Latitude, input.Longitude, input.FollowUpNotes, input.VisitResult, input.VisitOutcome, input.AutoCheckOut)
	if err != nil {
		return model.Visit{}, fmt.Errorf("check out prospect visit: %w", err)
	}
	if command.RowsAffected() != 1 {
		var exists bool
		_ = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM prospect_visits WHERE id=$1 AND prospect_id=$2)`, visitID, prospectID).Scan(&exists)
		if exists {
			var openWithinTolerance bool
			_ = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM prospect_visits v WHERE v.id=$1 AND v.check_out_at IS NULL AND (2 * 6371000 * ASIN(SQRT(POWER(SIN(RADIANS($2 - v.check_in_latitude) / 2), 2) + COS(RADIANS(v.check_in_latitude)) * COS(RADIANS($2)) * POWER(SIN(RADIANS($3 - v.check_in_longitude) / 2), 2))))) <= `+fmt.Sprintf("%.0f", checkoutToleranceMeters)+`)`, visitID, input.Latitude, input.Longitude).Scan(&openWithinTolerance)
			if !openWithinTolerance {
				return model.Visit{}, ErrVisitOutsideTolerance
			}
			return model.Visit{}, ErrVisitClosed
		}
		return model.Visit{}, ErrNotOwner
	}
	if input.FollowUpNotes != "" {
		_, _ = r.pool.Exec(ctx, `UPDATE prospects SET follow_up_notes=$2, updated_at=now() WHERE id=$1`, prospectID, input.FollowUpNotes)
	}
	return r.findVisit(ctx, visitID)
}

func (r *PostgresRepository) ListVisitMonitoring(ctx context.Context, filter model.VisitMonitoringFilter) ([]model.VisitMonitoringItem, error) {
	return r.listVisitsWithFilter(ctx, "", filter)
}

func (r *PostgresRepository) ListMyVisits(ctx context.Context, salesExecutiveID uuid.UUID, filter model.VisitMonitoringFilter) ([]model.VisitMonitoringItem, error) {
	return r.listVisitsWithFilter(ctx, salesExecutiveID.String(), filter)
}

func (r *PostgresRepository) listVisitsWithFilter(ctx context.Context, salesExecutiveID string, filter model.VisitMonitoringFilter) ([]model.VisitMonitoringItem, error) {
	const baseQuery = `
		SELECT v.id, v.prospect_id, COALESCE(cs.name, p.place_name), COALESCE(cs.category, p.place_category),
		       COALESCE(p.industry_group, ''),
		       COALESCE(cs.preview_address, p.formatted_address, ''),
		       COALESCE(p.phone_number, ''),
		       COALESCE(cs.latitude, p.latitude), COALESCE(cs.longitude, p.longitude),
		       v.sales_executive_id, u.full_name,
		       v.check_in_at, v.check_out_at,
		       v.check_in_latitude, v.check_in_longitude,
		       v.check_out_latitude, v.check_out_longitude,
		       CASE WHEN COALESCE(cs.latitude, p.latitude) IS NOT NULL AND COALESCE(cs.longitude, p.longitude) IS NOT NULL THEN
		           (2 * 6371000 * ASIN(SQRT(
		             POWER(SIN(RADIANS(v.check_in_latitude - COALESCE(cs.latitude, p.latitude)) / 2), 2) +
		             COS(RADIANS(COALESCE(cs.latitude, p.latitude))) * COS(RADIANS(v.check_in_latitude)) *
		             POWER(SIN(RADIANS(v.check_in_longitude - COALESCE(cs.longitude, p.longitude)) / 2), 2)
		           )))
		       ELSE 0 END AS distance_meters,
		       CASE WHEN v.check_out_at IS NOT NULL THEN
		           EXTRACT(EPOCH FROM (v.check_out_at - v.check_in_at))
		       ELSE NULL END AS duration_seconds,
		       CASE WHEN COALESCE(cs.latitude, p.latitude) IS NOT NULL AND COALESCE(cs.longitude, p.longitude) IS NOT NULL THEN
		           CASE WHEN (2 * 6371000 * ASIN(SQRT(
		             POWER(SIN(RADIANS(v.check_in_latitude - COALESCE(cs.latitude, p.latitude)) / 2), 2) +
		             COS(RADIANS(COALESCE(cs.latitude, p.latitude))) * COS(RADIANS(v.check_in_latitude)) *
		             POWER(SIN(RADIANS(v.check_in_longitude - COALESCE(cs.longitude, p.longitude)) / 2), 2)
		           ))) <= 100 THEN 'INSIDE' ELSE 'OUTSIDE' END
		       ELSE 'UNKNOWN' END AS radius_status,
		       p.status::text AS prospect_status,
		       v.selfie_reference, v.visit_notes, v.follow_up_notes,
		       (SELECT COUNT(*)::int FROM prospect_visits pv WHERE pv.prospect_id = v.prospect_id) AS visit_count,
		       v.visit_result, v.visit_outcome,
		       cs.id, CASE WHEN cs.id IS NULL THEN 'prospect' ELSE 'customer' END
		FROM prospect_visits v
		JOIN users u ON u.id = v.sales_executive_id
		JOIN prospects p ON p.id = v.prospect_id
		LEFT JOIN customer_sites cs ON cs.source_prospect_id = p.id
		  AND cs.sales_executive_id = v.sales_executive_id`

	whereClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if salesExecutiveID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.sales_executive_id = $%d", argIdx))
		args = append(args, salesExecutiveID)
		argIdx++
	}

	if filter.DateFrom != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.check_in_at::date >= $%d", argIdx))
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.check_in_at::date <= $%d", argIdx))
		args = append(args, filter.DateTo)
		argIdx++
	}
	if filter.SalesExecutiveID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.sales_executive_id = $%d", argIdx))
		args = append(args, filter.SalesExecutiveID)
		argIdx++
	}
	if filter.CustomerName != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(cs.name, p.place_name) ILIKE $%d", argIdx))
		args = append(args, "%"+filter.CustomerName+"%")
		argIdx++
	}

	query := baseQuery
	if len(whereClauses) > 0 {
		query += " WHERE "
		for i, clause := range whereClauses {
			if i > 0 {
				query += " AND "
			}
			query += clause
		}
	}
	query += " ORDER BY v.check_in_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list visit monitoring: %w", err)
	}
	defer rows.Close()

	items := make([]model.VisitMonitoringItem, 0)
	for rows.Next() {
		var item model.VisitMonitoringItem
		if err := rows.Scan(
			&item.ID, &item.ProspectID, &item.CustomerName, &item.CustomerCategory,
			&item.IndustryGroup, &item.FormattedAddress, &item.PhoneNumber,
			&item.ProspectLatitude, &item.ProspectLongitude,
			&item.SalesExecutiveID, &item.SalesExecutiveName,
			&item.CheckInAt, &item.CheckOutAt,
			&item.CheckInLatitude, &item.CheckInLongitude,
			&item.CheckOutLatitude, &item.CheckOutLongitude,
			&item.DistanceMeters, &item.DurationSeconds, &item.RadiusStatus,
			&item.ProspectStatus, &item.SelfieReference, &item.VisitNotes, &item.FollowUpNotes,
			&item.VisitCount, &item.VisitResult, &item.VisitOutcome,
			&item.CustomerID, &item.EntityType,
		); err != nil {
			return nil, fmt.Errorf("scan visit monitoring: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if filter.RadiusStatus != "" && filter.RadiusStatus != "ALL" {
		filtered := make([]model.VisitMonitoringItem, 0)
		for _, item := range items {
			if item.RadiusStatus == filter.RadiusStatus {
				filtered = append(filtered, item)
			}
		}
		return filtered, nil
	}

	return items, nil
}

func (r *PostgresRepository) ListProspectVisits(ctx context.Context, prospectID uuid.UUID) ([]model.Visit, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT v.id, v.prospect_id, v.sales_executive_id, u.full_name,
		       v.check_in_at, v.check_in_latitude, v.check_in_longitude,
		       v.check_out_at, v.check_out_latitude, v.check_out_longitude,
		       v.selfie_reference, v.visit_notes, v.follow_up_notes,
		       v.visit_result, v.visit_outcome
		FROM prospect_visits v JOIN users u ON u.id = v.sales_executive_id
		WHERE v.prospect_id = $1 ORDER BY v.check_in_at DESC`, prospectID)
	if err != nil {
		return nil, fmt.Errorf("list prospect visits: %w", err)
	}
	defer rows.Close()
	items := make([]model.Visit, 0)
	for rows.Next() {
		item, scanErr := scanVisit(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) DeleteVisit(ctx context.Context, visitID uuid.UUID, salesExecutiveID uuid.UUID) (model.Visit, error) {
	var (
		visit model.Visit
		err   error
	)
	if salesExecutiveID == uuid.Nil {
		visit, err = r.findVisit(ctx, visitID)
	} else {
		visit, err = r.findVisitOwnedBy(ctx, visitID, salesExecutiveID)
	}
	if err != nil {
		return model.Visit{}, err
	}
	if _, err := r.pool.Exec(ctx, `DELETE FROM prospect_visits WHERE id = $1`, visitID); err != nil {
		return model.Visit{}, fmt.Errorf("delete visit: %w", err)
	}
	return visit, nil
}

func (r *PostgresRepository) DeleteProspect(ctx context.Context, id uuid.UUID) ([]string, error) {
	var converted bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM customer_sites WHERE source_prospect_id = $1)`, id).Scan(&converted); err != nil {
		return nil, fmt.Errorf("check prospect conversion: %w", err)
	}
	if converted {
		return nil, ErrConflict
	}
	rows, err := r.pool.Query(ctx, `SELECT selfie_reference FROM prospect_visits WHERE prospect_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("list prospect visit selfies: %w", err)
	}
	defer rows.Close()
	selfies := make([]string, 0)
	for rows.Next() {
		var ref string
		if err := rows.Scan(&ref); err != nil {
			return nil, fmt.Errorf("scan prospect visit selfie: %w", err)
		}
		selfies = append(selfies, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list prospect visit selfies: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM prospect_comments WHERE prospect_id = $1`, id); err != nil {
		return nil, fmt.Errorf("delete prospect comments: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM prospect_visits WHERE prospect_id = $1`, id); err != nil {
		return nil, fmt.Errorf("delete prospect visits: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM prospect_status_history WHERE prospect_id = $1`, id); err != nil {
		return nil, fmt.Errorf("delete prospect status history: %w", err)
	}
	tag, err := tx.Exec(ctx, `DELETE FROM prospects WHERE id = $1`, id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("delete prospect: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return selfies, nil
}

func isForeignKeyViolation(err error) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == "23503"
}

func (r *PostgresRepository) RequestDeletion(ctx context.Context, prospectID, salesExecID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE prospects SET deletion_requested = TRUE, updated_at = now()
		WHERE id = $1 AND assigned_sales_executive_id = $2`, prospectID, salesExecID)
	if err != nil {
		return fmt.Errorf("request deletion: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ApproveDeletion(ctx context.Context, prospectID uuid.UUID) ([]string, error) {
	return r.DeleteProspect(ctx, prospectID)
}

func (r *PostgresRepository) RejectDeletion(ctx context.Context, prospectID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE prospects SET deletion_requested = FALSE, updated_at = now()
		WHERE id = $1`, prospectID)
	if err != nil {
		return fmt.Errorf("reject deletion: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ListComments(ctx context.Context, prospectID uuid.UUID) ([]model.ProspectComment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.prospect_id, c.user_id, u.full_name, c.content, c.attachments, c.created_at, c.updated_at
		FROM prospect_comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.prospect_id = $1
		ORDER BY c.created_at ASC`, prospectID)
	if err != nil {
		return nil, fmt.Errorf("list prospect comments: %w", err)
	}
	defer rows.Close()
	items := make([]model.ProspectComment, 0)
	for rows.Next() {
		var item model.ProspectComment
		var raw []byte
		if err := rows.Scan(&item.ID, &item.ProspectID, &item.UserID, &item.UserName, &item.Content, &raw, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan prospect comment: %w", err)
		}
		if item.Attachments, err = decodeAttachments(raw); err != nil {
			return nil, fmt.Errorf("decode comment attachments: %w", err)
		}
		for i := range item.Attachments {
			item.Attachments[i].Path = ""
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateAIChat(ctx context.Context, item model.ProspectAIChat) (model.ProspectAIChat, error) {
	item.ID = uuid.New()
	err := r.pool.QueryRow(ctx, `INSERT INTO prospect_ai_chats (id, prospect_id, user_id, message, answer, skill, insight, why, recommended_action) VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''),NULLIF($9,'')) RETURNING created_at`, item.ID, item.ProspectID, item.UserID, item.Message, item.Answer, item.Skill, item.Insight, item.Why, item.RecommendedAction).Scan(&item.CreatedAt)
	if err != nil {
		return model.ProspectAIChat{}, fmt.Errorf("create prospect AI chat: %w", err)
	}
	return item, nil
}

func (r *PostgresRepository) ListAIChats(ctx context.Context, prospectID uuid.UUID, limit int) ([]model.ProspectAIChat, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `SELECT c.id, c.prospect_id, c.user_id, c.message, c.answer, c.skill, COALESCE(c.insight,''), COALESCE(c.why,''), COALESCE(c.recommended_action,''), u.full_name, u.role::text, c.created_at FROM prospect_ai_chats c JOIN users u ON u.id=c.user_id WHERE c.prospect_id=$1 ORDER BY c.created_at ASC LIMIT $2`, prospectID, limit)
	if err != nil {
		return nil, fmt.Errorf("list prospect AI chats: %w", err)
	}
	defer rows.Close()
	items := []model.ProspectAIChat{}
	for rows.Next() {
		var item model.ProspectAIChat
		if err := rows.Scan(&item.ID, &item.ProspectID, &item.UserID, &item.Message, &item.Answer, &item.Skill, &item.Insight, &item.Why, &item.RecommendedAction, &item.AuthorName, &item.AuthorRole, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListRecentAIChats(ctx context.Context, prospectID uuid.UUID, limit int) ([]model.ProspectAIChat, error) {
	if limit <= 0 {
		limit = 8
	}
	rows, err := r.pool.Query(ctx, `SELECT c.id, c.prospect_id, c.user_id, c.message, c.answer, c.skill, COALESCE(c.insight,''), COALESCE(c.why,''), COALESCE(c.recommended_action,''), u.full_name, u.role::text, c.created_at FROM prospect_ai_chats c JOIN users u ON u.id=c.user_id WHERE c.prospect_id=$1 ORDER BY c.created_at DESC LIMIT $2`, prospectID, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent prospect AI chats: %w", err)
	}
	defer rows.Close()
	items := []model.ProspectAIChat{}
	for rows.Next() {
		var item model.ProspectAIChat
		if err := rows.Scan(&item.ID, &item.ProspectID, &item.UserID, &item.Message, &item.Answer, &item.Skill, &item.Insight, &item.Why, &item.RecommendedAction, &item.AuthorName, &item.AuthorRole, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateComment(ctx context.Context, prospectID, userID uuid.UUID, content string, attachments []model.CommentAttachment) (model.ProspectComment, error) {
	id := uuid.New()
	raw, err := json.Marshal(attachments)
	if err != nil {
		return model.ProspectComment{}, fmt.Errorf("encode comment attachments: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO prospect_comments (id, prospect_id, user_id, content, attachments, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())`, id, prospectID, userID, content, raw)
	if err != nil {
		return model.ProspectComment{}, fmt.Errorf("create prospect comment: %w", err)
	}
	var item model.ProspectComment
	err = r.pool.QueryRow(ctx, `
		SELECT c.id, c.prospect_id, c.user_id, u.full_name, c.content, c.attachments, c.created_at, c.updated_at
		FROM prospect_comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1`, id).Scan(&item.ID, &item.ProspectID, &item.UserID, &item.UserName, &item.Content, &raw, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return model.ProspectComment{}, fmt.Errorf("read created comment: %w", err)
	}
	item.Attachments, err = decodeAttachments(raw)
	for i := range item.Attachments {
		item.Attachments[i].Path = ""
	}
	return item, err
}

func (r *PostgresRepository) DeleteComment(ctx context.Context, prospectID, commentID, userID uuid.UUID) ([]model.CommentAttachment, error) {
	query := `DELETE FROM prospect_comments WHERE id=$1 AND prospect_id=$2`
	args := []any{commentID, prospectID}
	if userID != uuid.Nil {
		query += ` AND user_id=$3`
		args = append(args, userID)
	}
	query += ` RETURNING attachments`

	var raw []byte
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("delete prospect comment: %w", err)
	}
	items, err := decodeAttachments(raw)
	if err != nil {
		return nil, fmt.Errorf("decode deleted comment attachments: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) FindCommentAttachment(ctx context.Context, prospectID, attachmentID uuid.UUID) (model.CommentAttachment, error) {
	rows, err := r.pool.Query(ctx, `SELECT attachments FROM prospect_comments WHERE prospect_id=$1 ORDER BY created_at DESC`, prospectID)
	if err != nil {
		return model.CommentAttachment{}, fmt.Errorf("find comment attachment: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return model.CommentAttachment{}, fmt.Errorf("scan comment attachment: %w", err)
		}
		items, err := decodeAttachments(raw)
		if err != nil {
			return model.CommentAttachment{}, err
		}
		for _, item := range items {
			if item.ID == attachmentID {
				return item, nil
			}
		}
	}
	if err := rows.Err(); err != nil {
		return model.CommentAttachment{}, fmt.Errorf("read comment attachments: %w", err)
	}
	return model.CommentAttachment{}, ErrNotFound
}

func (r *PostgresRepository) FindProspectOwner(ctx context.Context, prospectID uuid.UUID) (uuid.UUID, error) {
	var ownerID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT assigned_sales_executive_id FROM prospects WHERE id = $1`, prospectID).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("find prospect owner: %w", err)
	}
	return ownerID, nil
}

func (r *PostgresRepository) ListPhotoTags(ctx context.Context, prospectID uuid.UUID) ([]model.ProspectPhotoTag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, prospect_id, photo_name, photo_index, category, updated_by, created_at, updated_at
		FROM prospect_photo_tags
		WHERE prospect_id = $1
		ORDER BY created_at ASC`, prospectID)
	if err != nil {
		return nil, fmt.Errorf("list prospect photo tags: %w", err)
	}
	defer rows.Close()
	items := make([]model.ProspectPhotoTag, 0)
	for rows.Next() {
		item, err := scanPhotoTag(rows)
		if err != nil {
			return nil, fmt.Errorf("scan prospect photo tag: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) UpsertPhotoTag(ctx context.Context, prospectID uuid.UUID, photoName string, photoIndex *int, category model.PhotoCategory, userID uuid.UUID) (model.ProspectPhotoTag, error) {
	id := uuid.New()
	var item model.ProspectPhotoTag
	const returning = `RETURNING id, prospect_id, photo_name, photo_index, category, updated_by, created_at, updated_at`
	var err error
	if photoIndex != nil {
		item, err = scanPhotoTag(r.pool.QueryRow(ctx, `UPDATE prospect_photo_tags SET photo_name=$3, category=$4, updated_by=$5, updated_at=now() WHERE prospect_id=$1 AND photo_index=$2 `+returning, prospectID, *photoIndex, photoName, category, userID))
	}
	if err == pgx.ErrNoRows || photoIndex == nil {
		item, err = scanPhotoTag(r.pool.QueryRow(ctx, `
			INSERT INTO prospect_photo_tags (id, prospect_id, photo_name, photo_index, category, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, now())
			ON CONFLICT (prospect_id, photo_name)
			DO UPDATE SET photo_index = EXCLUDED.photo_index, category = EXCLUDED.category, updated_by = EXCLUDED.updated_by, updated_at = now()
			`+returning, id, prospectID, photoName, photoIndex, category, userID))
	}
	if err != nil {
		return model.ProspectPhotoTag{}, fmt.Errorf("upsert prospect photo tag: %w", err)
	}
	return item, nil
}

func scanPhotoTag(row rowScanner) (model.ProspectPhotoTag, error) {
	var item model.ProspectPhotoTag
	err := row.Scan(&item.ID, &item.ProspectID, &item.PhotoName, &item.PhotoIndex, &item.Category, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *PostgresRepository) ProspectAccessibleTo(ctx context.Context, prospectID, userID uuid.UUID) (bool, error) {
	var accessible bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM prospects WHERE id = $1 AND assigned_sales_executive_id = $2
			UNION
			SELECT 1 FROM customer_sites WHERE source_prospect_id = $1 AND sales_executive_id = $2
		)`, prospectID, userID).Scan(&accessible)
	if err != nil {
		return false, fmt.Errorf("check prospect accessibility: %w", err)
	}
	return accessible, nil
}

func (r *PostgresRepository) ExistingCustomerPlaceIDs(ctx context.Context, placeIDs []string) (map[string]bool, error) {
	result := make(map[string]bool, len(placeIDs))
	if len(placeIDs) == 0 {
		return result, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT source_google_place_id FROM customer_sites
		WHERE source_google_place_id IS NOT NULL AND source_google_place_id = ANY($1)`, placeIDs)
	if err != nil {
		return nil, fmt.Errorf("list existing customer place ids: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan existing customer place id: %w", err)
		}
		result[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list existing customer place ids: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) ListCustomerMarkers(ctx context.Context) ([]model.CustomerMarker, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, customer_code, COALESCE(source_google_place_id, ''), name, COALESCE(preview_address, ''), latitude, longitude
		FROM customer_sites
		WHERE latitude IS NOT NULL AND longitude IS NOT NULL
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list customer markers: %w", err)
	}
	defer rows.Close()
	items := make([]model.CustomerMarker, 0)
	for rows.Next() {
		var item model.CustomerMarker
		if err := rows.Scan(&item.CustomerID, &item.CustomerCode, &item.GooglePlaceID, &item.PlaceName, &item.FormattedAddress, &item.Latitude, &item.Longitude); err != nil {
			return nil, fmt.Errorf("scan customer marker: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list customer markers: %w", err)
	}
	return items, nil
}

const visitSelect = `SELECT v.id,v.prospect_id,v.sales_executive_id,u.full_name,v.check_in_at,v.check_in_latitude,v.check_in_longitude,v.check_out_at,v.check_out_latitude,v.check_out_longitude,v.selfie_reference,v.visit_notes,v.follow_up_notes,v.visit_result,v.visit_outcome FROM prospect_visits v JOIN users u ON u.id=v.sales_executive_id`

func (r *PostgresRepository) findVisit(ctx context.Context, id uuid.UUID) (model.Visit, error) {
	return scanVisit(r.pool.QueryRow(ctx, visitSelect+` WHERE v.id=$1`, id))
}

func (r *PostgresRepository) findVisitOwnedBy(ctx context.Context, id, salesExecutiveID uuid.UUID) (model.Visit, error) {
	return scanVisit(r.pool.QueryRow(ctx, visitSelect+` WHERE v.id=$1 AND v.sales_executive_id=$2`, id, salesExecutiveID))
}

func scanVisit(row rowScanner) (model.Visit, error) {
	var item model.Visit
	err := row.Scan(&item.ID, &item.ProspectID, &item.SalesExecutiveID, &item.SalesExecutiveName, &item.CheckInAt, &item.CheckInLatitude, &item.CheckInLongitude, &item.CheckOutAt, &item.CheckOutLatitude, &item.CheckOutLongitude, &item.SelfieReference, &item.VisitNotes, &item.FollowUpNotes, &item.VisitResult, &item.VisitOutcome)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Visit{}, ErrNotFound
	}
	if err != nil {
		return model.Visit{}, fmt.Errorf("scan prospect visit: %w", err)
	}
	return item, nil
}

func scanProspects(rows pgx.Rows) ([]model.Prospect, error) {
	items := make([]model.Prospect, 0)
	for rows.Next() {
		item, err := scanProspect(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type rowScanner interface {
	Scan(...any) error
}

func scanProspect(row rowScanner) (model.Prospect, error) {
	var item model.Prospect
	var placeTypes []byte
	err := row.Scan(&item.ID, &item.GooglePlaceID, &item.PlaceName, &item.FormattedAddress,
		&item.Latitude, &item.Longitude, &item.PlaceCategory, &placeTypes, &item.IndustryGroup,
		&item.PhoneNumber, &item.WebsiteURL, &item.GoogleMapsURL, &item.AssignedSalesExecutiveID, &item.AssignedSalesExecutive,
		&item.VisitNotes, &item.FollowUpNotes, &item.Status, &item.DeletionRequested, &item.ConvertedAt,
		&item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Prospect{}, ErrNotFound
	}
	if err != nil {
		return model.Prospect{}, fmt.Errorf("scan prospect: %w", err)
	}
	if err := json.Unmarshal(placeTypes, &item.PlaceTypes); err != nil {
		return model.Prospect{}, fmt.Errorf("decode prospect place types: %w", err)
	}
	return item, nil
}
