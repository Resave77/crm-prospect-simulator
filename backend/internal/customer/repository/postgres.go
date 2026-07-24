package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"crm-prospect-simulator/backend/internal/customer/model"
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

const parentSelect = `
	SELECT pc.id, pc.parent_code, pc.name, pc.address_mode, pc.province,
	       pc.district, pc.sub_district, pc.village, pc.latitude, pc.longitude,
	       pc.preview_address, pc.company_contacts, pc.npwp_name,
	       pc.npwp_address, pc.npwp_number, pc.term_of_payment, pc.kam_assignments
	FROM parent_companies pc`

const customerSelect = `
	SELECT cs.id, cs.customer_code, cs.parent_company_id, pc.parent_code, pc.name,
	       cs.source_prospect_id, cs.source_google_place_id, cs.name, cs.segment,
	       cs.category, cs.address_mode, cs.province, cs.district, cs.sub_district,
	       cs.village, cs.latitude, cs.longitude, cs.preview_address, cs.site_contacts,
	       cs.ppn, cs.id_tku_number, cs.nik, cs.shipment_cost, cs.invoice_type,
	       cs.bank_account, cs.bill_to_source, cs.ship_to_source,
	       cs.billing_address_preview, cs.shipping_address_preview,
	       cs.sales_executive_id, u.full_name, cs.sales_assignments,
	       cs.converted_at, cs.updated_at, cs.converted_by_admin_id
	FROM customer_sites cs
	JOIN parent_companies pc ON pc.id = cs.parent_company_id
	JOIN users u ON u.id = cs.sales_executive_id`

func (r *PostgresRepository) SearchParentCompanies(ctx context.Context, search string) ([]model.ParentCompany, error) {
	pattern := "%" + strings.TrimSpace(search) + "%"
	rows, err := r.pool.Query(ctx, parentSelect+`
		WHERE ($1 = '%%' OR pc.name ILIKE $1 OR pc.parent_code ILIKE $1)
		ORDER BY pc.name LIMIT 20`, pattern)
	if err != nil {
		return nil, fmt.Errorf("search parent companies: %w", err)
	}
	defer rows.Close()
	items := make([]model.ParentCompany, 0)
	for rows.Next() {
		item, err := scanParent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListActiveSalesExecutives(ctx context.Context) ([]model.UserOption, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, full_name FROM users
		WHERE role = 'SALES_EXECUTIVE' AND status = 'ACTIVE' ORDER BY full_name`)
	if err != nil {
		return nil, fmt.Errorf("list active sales executives: %w", err)
	}
	defer rows.Close()
	items := make([]model.UserOption, 0)
	for rows.Next() {
		var item model.UserOption
		if err := rows.Scan(&item.ID, &item.FullName); err != nil {
			return nil, fmt.Errorf("scan sales executive: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Convert(ctx context.Context, prospectID, administratorID uuid.UUID, input model.ConversionInput) (model.CustomerSite, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("begin prospect conversion: %w", err)
	}
	defer tx.Rollback(ctx)

	var prospectStatus, googlePlaceID string
	err = tx.QueryRow(ctx, `
		SELECT status::text, google_place_id FROM prospects WHERE id = $1 FOR UPDATE`, prospectID).
		Scan(&prospectStatus, &googlePlaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.CustomerSite{}, ErrNotFound
	}
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("lock prospect for conversion: %w", err)
	}
	if prospectStatus == "CONVERTED" {
		return model.CustomerSite{}, ErrAlreadyConverted
	}
	if prospectStatus != "WON" {
		return model.CustomerSite{}, ErrProspectNotWon
	}

	var duplicate bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM customer_sites WHERE source_prospect_id = $1 OR source_google_place_id = $2)`,
		prospectID, googlePlaceID).Scan(&duplicate); err != nil {
		return model.CustomerSite{}, fmt.Errorf("check customer duplicate: %w", err)
	}
	if duplicate {
		return model.CustomerSite{}, ErrDuplicatePlace
	}

	parent, err := r.resolveParentCompany(ctx, tx, input)
	if err != nil {
		return model.CustomerSite{}, err
	}

	var salesName string
	err = tx.QueryRow(ctx, `
		SELECT full_name FROM users WHERE id = $1 AND role = 'SALES_EXECUTIVE' AND status = 'ACTIVE'`, input.SalesExecutiveID).
		Scan(&salesName)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.CustomerSite{}, ErrSalesUnavailable
	}
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("validate sales executive: %w", err)
	}
	for _, assignment := range input.SalesAssignments {
		if assignment.OwnerID == "" {
			continue
		}
		assignmentOwnerID, parseErr := uuid.Parse(assignment.OwnerID)
		if parseErr != nil {
			return model.CustomerSite{}, ErrSalesUnavailable
		}
		var active bool
		if queryErr := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'SALES_EXECUTIVE' AND status = 'ACTIVE')`, assignmentOwnerID).Scan(&active); queryErr != nil {
			return model.CustomerSite{}, fmt.Errorf("validate additional sales assignment: %w", queryErr)
		}
		if !active {
			return model.CustomerSite{}, ErrSalesUnavailable
		}
	}

	var customerSequence int64
	if err := tx.QueryRow(ctx, `SELECT nextval('customer_site_code_seq')`).Scan(&customerSequence); err != nil {
		return model.CustomerSite{}, fmt.Errorf("generate customer code: %w", err)
	}
	customerCode := simulationCustomerCode(parent.ParentCode, customerSequence)

	siteContacts, err := json.Marshal(input.SiteContacts)
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("encode site contacts: %w", err)
	}
	salesAssignments, err := json.Marshal(input.SalesAssignments)
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("encode sales assignments: %w", err)
	}
	convertedAt := time.Now().UTC()
	customerID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO customer_sites (
			id, customer_code, parent_company_id, source_prospect_id, source_google_place_id,
			name, segment, category, address_mode, province, district, sub_district, village,
			latitude, longitude, preview_address, site_contacts, ppn, id_tku_number, nik,
			shipment_cost, invoice_type, bank_account, bill_to_source, ship_to_source,
			billing_address_preview, shipping_address_preview, sales_executive_id,
			sales_assignments, converted_at, converted_by_admin_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31)`,
		customerID, customerCode, parent.ID, prospectID, googlePlaceID,
		input.CustomerName, input.CustomerSegment, input.CustomerCategory,
		input.SiteAddress.Mode, input.SiteAddress.Province, input.SiteAddress.District,
		input.SiteAddress.SubDistrict, input.SiteAddress.Village, input.SiteAddress.Latitude,
		input.SiteAddress.Longitude, input.SiteAddress.PreviewAddress, siteContacts,
		input.PPN, input.IDTKUNumber, input.NIK, input.ShipmentCost, input.InvoiceType,
		input.BankAccount, input.BillToSource, input.ShipToSource,
		input.BillingAddressPreview, input.ShippingAddressPreview,
		input.SalesExecutiveID, salesAssignments, convertedAt, administratorID)
	if err != nil {
		return model.CustomerSite{}, mapDatabaseError(err)
	}
	command, err := tx.Exec(ctx, `
		UPDATE prospects SET status = 'CONVERTED', converted_at = $2, updated_at = $2
		WHERE id = $1 AND status = 'WON'`, prospectID, convertedAt)
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("mark prospect converted: %w", err)
	}
	if command.RowsAffected() != 1 {
		return model.CustomerSite{}, ErrAlreadyConverted
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO prospect_status_history
			(id, prospect_id, from_status, to_status, changed_by_user_id, notes)
		VALUES ($1, $2, 'WON', 'CONVERTED', $3, $4)`,
		uuid.New(), prospectID, administratorID, "Converted to Customer Existing "+customerCode)
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("record conversion history: %w", err)
	}
	result, err := scanCustomer(tx.QueryRow(ctx, customerSelect+` WHERE cs.id = $1`, customerID))
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("read converted customer: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.CustomerSite{}, mapDatabaseError(err)
	}
	return result, nil
}

func (r *PostgresRepository) AutoConvert(ctx context.Context, prospectID uuid.UUID) (model.CustomerSite, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("begin auto conversion: %w", err)
	}
	defer tx.Rollback(ctx)

	var prospectStatus, googlePlaceID, placeName, formattedAddress, placeCategory string
	var latitude, longitude *float64
	var assignedSalesExecID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT status::text, google_place_id, place_name, formatted_address,
		       latitude, longitude, COALESCE(place_category, ''), assigned_sales_executive_id
		FROM prospects WHERE id = $1 FOR UPDATE`, prospectID).
		Scan(&prospectStatus, &googlePlaceID, &placeName, &formattedAddress, &latitude, &longitude, &placeCategory, &assignedSalesExecID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.CustomerSite{}, ErrNotFound
	}
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("lock prospect for auto conversion: %w", err)
	}
	if prospectStatus == "CONVERTED" {
		return model.CustomerSite{}, ErrAlreadyConverted
	}
	if prospectStatus != "WON" {
		return model.CustomerSite{}, ErrProspectNotWon
	}

	var duplicate bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM customer_sites WHERE source_prospect_id = $1 OR source_google_place_id = $2)`,
		prospectID, googlePlaceID).Scan(&duplicate); err != nil {
		return model.CustomerSite{}, fmt.Errorf("check customer duplicate: %w", err)
	}
	if duplicate {
		return model.CustomerSite{}, ErrDuplicatePlace
	}

	var salesName string
	if err := tx.QueryRow(ctx, `
		SELECT full_name FROM users WHERE id = $1 AND role = 'SALES_EXECUTIVE' AND status = 'ACTIVE'`, assignedSalesExecID).Scan(&salesName); err != nil {
		return model.CustomerSite{}, ErrSalesUnavailable
	}

	var parentSequence int64
	if err := tx.QueryRow(ctx, `SELECT nextval('parent_company_code_seq')`).Scan(&parentSequence); err != nil {
		return model.CustomerSite{}, fmt.Errorf("generate parent code: %w", err)
	}
	parentCode := simulationParentCode(parentSequence)
	parentID := uuid.New()
	emptyContacts, _ := json.Marshal([]model.Contact{})
	emptyKams, _ := json.Marshal([]model.PeriodAssignment{})
	_, err = tx.Exec(ctx, `
		INSERT INTO parent_companies (
			id, parent_code, name, address_mode, province, district, sub_district, village,
			latitude, longitude, preview_address, company_contacts, npwp_name,
			npwp_address, npwp_number, term_of_payment, kam_assignments)
		VALUES ($1,$2,$3,'AUTO_CONVERTED','','','','',$4,$5,$6,$7,'','','','',$8)`,
		parentID, parentCode, placeName, latitude, longitude, formattedAddress, emptyContacts, emptyKams)
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("create auto parent company: %w", err)
	}

	var customerSequence int64
	if err := tx.QueryRow(ctx, `SELECT nextval('customer_site_code_seq')`).Scan(&customerSequence); err != nil {
		return model.CustomerSite{}, fmt.Errorf("generate customer code: %w", err)
	}
	customerCode := simulationCustomerCode(parentCode, customerSequence)

	customerID := uuid.New()
	convertedAt := time.Now().UTC()
	siteContacts, _ := json.Marshal([]model.Contact{})
	salesAssignments, _ := json.Marshal([]model.PeriodAssignment{})
	if placeCategory == "" {
		placeCategory = "Uncategorized"
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO customer_sites (
			id, customer_code, parent_company_id, source_prospect_id, source_google_place_id,
			name, segment, category, address_mode, province, district, sub_district, village,
			latitude, longitude, preview_address, site_contacts, ppn, id_tku_number, nik,
			shipment_cost, invoice_type, bank_account, bill_to_source, ship_to_source,
			billing_address_preview, shipping_address_preview, sales_executive_id,
			sales_assignments, converted_at, converted_by_admin_id)
		VALUES ($1,$2,$3,$4,$5,$6,'General Trade',$7,'AUTO_CONVERTED','','','','',$8,$9,$10,$11,'','','','','','','',$12,$13,$14,$15)`,
		customerID, customerCode, parentID, prospectID, googlePlaceID,
		placeName, placeCategory, latitude, longitude, formattedAddress, siteContacts,
		assignedSalesExecID, salesAssignments, convertedAt, uuid.Nil)
	if err != nil {
		return model.CustomerSite{}, mapDatabaseError(err)
	}
	if _, err = tx.Exec(ctx, `
		UPDATE prospects SET status = 'CONVERTED', converted_at = $2, updated_at = $2
		WHERE id = $1 AND status = 'WON'`, prospectID, convertedAt); err != nil {
		return model.CustomerSite{}, fmt.Errorf("mark prospect converted: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO prospect_status_history
			(id, prospect_id, from_status, to_status, changed_by_user_id, notes)
		VALUES ($1, $2, 'WON', 'CONVERTED', $3, $4)`,
		uuid.New(), prospectID, assignedSalesExecID, "Auto-converted on WON transition "+customerCode); err != nil {
		return model.CustomerSite{}, fmt.Errorf("record auto conversion history: %w", err)
	}
	result, err := scanCustomer(tx.QueryRow(ctx, customerSelect+` WHERE cs.id = $1`, customerID))
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("read auto converted customer: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.CustomerSite{}, mapDatabaseError(err)
	}
	return result, nil
}

func (r *PostgresRepository) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM customer_sites WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete customer site: %w", err)
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) resolveParentCompany(ctx context.Context, tx pgx.Tx, input model.ConversionInput) (model.ParentCompany, error) {
	if input.ParentMethod == model.ParentMethodExisting {
		if input.ExistingParentCompanyID == nil {
			return model.ParentCompany{}, ErrParentUnavailable
		}
		parent, err := scanParent(tx.QueryRow(ctx, parentSelect+` WHERE pc.id = $1 FOR SHARE`, *input.ExistingParentCompanyID))
		if errors.Is(err, ErrNotFound) {
			return model.ParentCompany{}, ErrParentUnavailable
		}
		return parent, err
	}

	var parentSequence int64
	if err := tx.QueryRow(ctx, `SELECT nextval('parent_company_code_seq')`).Scan(&parentSequence); err != nil {
		return model.ParentCompany{}, fmt.Errorf("generate parent code: %w", err)
	}
	parentCode := simulationParentCode(parentSequence)
	contacts, err := json.Marshal(input.CompanyContacts)
	if err != nil {
		return model.ParentCompany{}, fmt.Errorf("encode company contacts: %w", err)
	}
	kams, err := json.Marshal(input.KAMAssignments)
	if err != nil {
		return model.ParentCompany{}, fmt.Errorf("encode KAM assignments: %w", err)
	}
	parentID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO parent_companies (
			id, parent_code, name, address_mode, province, district, sub_district, village,
			latitude, longitude, preview_address, company_contacts, npwp_name,
			npwp_address, npwp_number, term_of_payment, kam_assignments)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		parentID, parentCode, input.ParentCompanyName, input.CompanyAddress.Mode,
		input.CompanyAddress.Province, input.CompanyAddress.District, input.CompanyAddress.SubDistrict,
		input.CompanyAddress.Village, input.CompanyAddress.Latitude, input.CompanyAddress.Longitude,
		input.CompanyAddress.PreviewAddress, contacts, input.CompanyNPWPName,
		input.CompanyNPWPAddress, input.CompanyNPWPNumber, input.TermOfPayment, kams)
	if err != nil {
		return model.ParentCompany{}, mapDatabaseError(err)
	}
	return model.ParentCompany{ID: parentID, ParentCode: parentCode, Name: input.ParentCompanyName}, nil
}

func (r *PostgresRepository) ListCustomers(ctx context.Context) ([]model.CustomerSite, error) {
	return r.listCustomers(ctx, customerSelect+` ORDER BY cs.converted_at DESC`)
}

func (r *PostgresRepository) ListCustomersForSales(ctx context.Context, salesExecutiveID uuid.UUID) ([]model.CustomerSite, error) {
	return r.listCustomers(ctx, customerSelect+` WHERE cs.sales_executive_id = $1 ORDER BY cs.converted_at DESC`, salesExecutiveID)
}

func (r *PostgresRepository) FindCustomerForSales(ctx context.Context, id, salesExecutiveID uuid.UUID) (model.CustomerDetail, error) {
	customer, err := scanCustomer(r.pool.QueryRow(ctx, customerSelect+` WHERE cs.id=$1 AND cs.sales_executive_id=$2`, id, salesExecutiveID))
	if err != nil {
		return model.CustomerDetail{}, err
	}
	parent, err := scanParent(r.pool.QueryRow(ctx, parentSelect+` WHERE pc.id=$1`, customer.ParentCompanyID))
	if err != nil {
		return model.CustomerDetail{}, err
	}
	var sourceName string
	if err := r.pool.QueryRow(ctx, `SELECT place_name FROM prospects WHERE id=$1`, customer.SourceProspectID).Scan(&sourceName); err != nil {
		return model.CustomerDetail{}, fmt.Errorf("read source prospect: %w", err)
	}
	return model.CustomerDetail{Customer: customer, ParentCompany: parent, SourceProspectName: sourceName}, nil
}

func (r *PostgresRepository) FindCustomer(ctx context.Context, id uuid.UUID) (model.CustomerDetail, error) {
	customer, err := scanCustomer(r.pool.QueryRow(ctx, customerSelect+` WHERE cs.id=$1`, id))
	if err != nil {
		return model.CustomerDetail{}, err
	}
	parent, err := scanParent(r.pool.QueryRow(ctx, parentSelect+` WHERE pc.id=$1`, customer.ParentCompanyID))
	if err != nil {
		return model.CustomerDetail{}, err
	}
	var sourceName string
	if err := r.pool.QueryRow(ctx, `SELECT place_name FROM prospects WHERE id=$1`, customer.SourceProspectID).Scan(&sourceName); err != nil {
		return model.CustomerDetail{}, fmt.Errorf("read source prospect: %w", err)
	}
	return model.CustomerDetail{Customer: customer, ParentCompany: parent, SourceProspectName: sourceName}, nil
}

func (r *PostgresRepository) listCustomers(ctx context.Context, query string, args ...any) ([]model.CustomerSite, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list customer sites: %w", err)
	}
	defer rows.Close()
	items := make([]model.CustomerSite, 0)
	for rows.Next() {
		item, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

const customerListBase = `
	FROM customer_sites cs
	JOIN parent_companies pc ON pc.id = cs.parent_company_id
	JOIN users u ON u.id = cs.sales_executive_id`

const customerListSelect = `
	SELECT cs.id, cs.customer_code, cs.parent_company_id, pc.parent_code, pc.name,
	       cs.source_prospect_id, cs.source_google_place_id, cs.name, cs.segment,
	       cs.category, cs.address_mode, cs.province, cs.district, cs.sub_district,
	       cs.village, cs.latitude, cs.longitude, cs.preview_address, cs.site_contacts,
	       cs.ppn, cs.id_tku_number, cs.nik, cs.shipment_cost, cs.invoice_type,
	       cs.bank_account, cs.bill_to_source, cs.ship_to_source,
	       cs.billing_address_preview, cs.shipping_address_preview,
	       cs.sales_executive_id, u.full_name, cs.sales_assignments,
	       cs.converted_at, cs.updated_at, cs.converted_by_admin_id
` + customerListBase

func (r *PostgresRepository) ListCustomersPaged(ctx context.Context, params model.CustomerListParams) (model.CustomerListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}

	where, args := buildCustomerWhere(params)
	sortClause := buildCustomerSort(params.Sort)

	countQuery := `SELECT COUNT(*) ` + customerListBase + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return model.CustomerListResult{}, fmt.Errorf("count customer sites: %w", err)
	}

	pages := total / params.Limit
	if total%params.Limit > 0 {
		pages++
	}

	offset := (params.Page - 1) * params.Limit
	dataQuery := customerListSelect + where + sortClause + ` LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)
	dataArgs := append(args, params.Limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return model.CustomerListResult{}, fmt.Errorf("query customer sites: %w", err)
	}
	defer rows.Close()

	items := make([]model.CustomerSite, 0)
	for rows.Next() {
		item, err := scanCustomer(rows)
		if err != nil {
			return model.CustomerListResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.CustomerListResult{}, err
	}

	return model.CustomerListResult{
		Items: items,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
		Pages: pages,
	}, nil
}

func (r *PostgresRepository) ListFilterOptions(ctx context.Context) (model.ListFilterOptions, error) {
	segments, err := r.distinctColumn(ctx, `SELECT DISTINCT segment FROM customer_sites WHERE segment != '' ORDER BY segment`)
	if err != nil {
		return model.ListFilterOptions{}, err
	}
	categories, err := r.distinctColumn(ctx, `SELECT DISTINCT category FROM customer_sites WHERE category != '' ORDER BY category`)
	if err != nil {
		return model.ListFilterOptions{}, err
	}
	regions, err := r.distinctColumn(ctx, `SELECT DISTINCT province FROM customer_sites WHERE province != '' ORDER BY province`)
	if err != nil {
		return model.ListFilterOptions{}, err
	}
	sales, err := r.ListActiveSalesExecutives(ctx)
	if err != nil {
		return model.ListFilterOptions{}, err
	}
	return model.ListFilterOptions{
		Segments:   segments,
		Categories: categories,
		Regions:    regions,
		SalesExec:  sales,
	}, nil
}

func (r *PostgresRepository) distinctColumn(ctx context.Context, query string) ([]string, error) {
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query distinct column: %w", err)
	}
	defer rows.Close()
	items := make([]string, 0)
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		items = append(items, val)
	}
	return items, rows.Err()
}

func buildCustomerWhere(params model.CustomerListParams) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)
	idx := 1

	if params.Keyword != "" {
		pattern := "%" + strings.TrimSpace(params.Keyword) + "%"
		conditions = append(conditions, `(
			cs.name ILIKE $`+itoa(idx)+` OR
			cs.customer_code ILIKE $`+itoa(idx)+` OR
			pc.name ILIKE $`+itoa(idx)+` OR
			pc.parent_code ILIKE $`+itoa(idx)+` OR
			cs.site_contacts::text ILIKE $`+itoa(idx)+` OR
			cs.preview_address ILIKE $`+itoa(idx)+`
		)`)
		args = append(args, pattern)
		idx++
	}
	if params.Segment != "" {
		conditions = append(conditions, `cs.segment = $`+itoa(idx))
		args = append(args, params.Segment)
		idx++
	}
	if params.Category != "" {
		conditions = append(conditions, `cs.category = $`+itoa(idx))
		args = append(args, params.Category)
		idx++
	}
	if params.Sales != "" {
		conditions = append(conditions, `u.full_name ILIKE $`+itoa(idx))
		args = append(args, "%"+params.Sales+"%")
		idx++
	}
	if params.Region != "" {
		conditions = append(conditions, `cs.province = $`+itoa(idx))
		args = append(args, params.Region)
		idx++
	}

	if len(conditions) == 0 {
		return "", args
	}
	where := " WHERE " + strings.Join(conditions, " AND ")
	return where, args
}

func buildCustomerSort(sort string) string {
	switch sort {
	case "oldest":
		return ` ORDER BY cs.converted_at ASC`
	case "name":
		return ` ORDER BY cs.name ASC`
	case "code":
		return ` ORDER BY cs.customer_code ASC`
	case "converted":
		return ` ORDER BY cs.converted_at DESC`
	case "updated":
		return ` ORDER BY cs.updated_at DESC`
	default:
		return ` ORDER BY cs.converted_at DESC`
	}
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

type rowScanner interface {
	Scan(...any) error
}

func scanParent(row rowScanner) (model.ParentCompany, error) {
	var item model.ParentCompany
	var contacts, kams []byte
	err := row.Scan(&item.ID, &item.ParentCode, &item.Name, &item.Address.Mode,
		&item.Address.Province, &item.Address.District, &item.Address.SubDistrict,
		&item.Address.Village, &item.Address.Latitude, &item.Address.Longitude,
		&item.Address.PreviewAddress, &contacts, &item.NPWPName, &item.NPWPAddress,
		&item.NPWPNumber, &item.TermOfPayment, &kams)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ParentCompany{}, ErrNotFound
	}
	if err != nil {
		return model.ParentCompany{}, fmt.Errorf("scan parent company: %w", err)
	}
	if err := decodeJSON(contacts, &item.Contacts); err != nil {
		return model.ParentCompany{}, err
	}
	if err := decodeJSON(kams, &item.KAMAssignments); err != nil {
		return model.ParentCompany{}, err
	}
	return item, nil
}

func scanCustomer(row rowScanner) (model.CustomerSite, error) {
	var item model.CustomerSite
	var contacts, assignments []byte
	err := row.Scan(&item.ID, &item.CustomerCode, &item.ParentCompanyID, &item.ParentCode,
		&item.ParentCompanyName, &item.SourceProspectID, &item.SourceGooglePlaceID,
		&item.Name, &item.Segment, &item.Category, &item.Address.Mode,
		&item.Address.Province, &item.Address.District, &item.Address.SubDistrict,
		&item.Address.Village, &item.Address.Latitude, &item.Address.Longitude,
		&item.Address.PreviewAddress, &contacts, &item.PPN, &item.IDTKUNumber,
		&item.NIK, &item.ShipmentCost, &item.InvoiceType, &item.BankAccount,
		&item.BillToSource, &item.ShipToSource, &item.BillingAddressPreview,
		&item.ShippingAddressPreview, &item.SalesExecutiveID, &item.SalesExecutiveName,
		&assignments, &item.ConvertedAt, &item.UpdatedAt, &item.ConvertedByAdminID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.CustomerSite{}, ErrNotFound
	}
	if err != nil {
		return model.CustomerSite{}, fmt.Errorf("scan customer site: %w", err)
	}
	item.Region = item.Address.Province
	if err := decodeJSON(contacts, &item.Contacts); err != nil {
		return model.CustomerSite{}, err
	}
	if err := decodeJSON(assignments, &item.SalesAssignments); err != nil {
		return model.CustomerSite{}, err
	}
	return item, nil
}

func decodeJSON(raw []byte, destination any) error {
	if err := json.Unmarshal(raw, destination); err != nil {
		return fmt.Errorf("decode simulation JSON: %w", err)
	}
	return nil
}

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" {
		if strings.Contains(pgError.ConstraintName, "source_google_place_id") {
			return ErrDuplicatePlace
		}
		if strings.Contains(pgError.ConstraintName, "source_prospect_id") {
			return ErrAlreadyConverted
		}
		return ErrCodeConflict
	}
	return fmt.Errorf("persist conversion: %w", err)
}

// These formats exist only for the local mentor simulation. Final ownership,
// prefixes, and sequencing remain an ERP/business decision.
func simulationParentCode(sequence int64) string {
	return fmt.Sprintf("PC-%06d", sequence)
}

func simulationCustomerCode(parentCode string, sequence int64) string {
	return fmt.Sprintf("%s-S%03d", parentCode, sequence)
}
