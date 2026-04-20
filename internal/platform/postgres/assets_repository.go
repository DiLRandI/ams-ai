package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"ams-ai/internal/domain"

	"github.com/jackc/pgx/v5"
)

type AssetRepository struct {
	db                *DB
	reminderWindowDay int
}

func NewAssetRepository(db *DB, reminderWindowDays int) *AssetRepository {
	if reminderWindowDays <= 0 {
		reminderWindowDays = 30
	}
	return &AssetRepository{db: db, reminderWindowDay: reminderWindowDays}
}

func (r *AssetRepository) CreateAsset(ctx context.Context, a domain.Asset) (domain.Asset, error) {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return domain.Asset{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id int64
	if err := tx.QueryRow(ctx, `SELECT nextval('assets_id_seq')`).Scan(&id); err != nil {
		return domain.Asset{}, err
	}
	a.ID = id
	a.Code = fmt.Sprintf("AMS-%06d", id)
	row := tx.QueryRow(ctx, `
		INSERT INTO assets (
			id, code, type, category_id, name, brand, model, serial_number,
			purchase_date, purchase_price, status, condition, location, assigned_to,
			assigned_user_id, notes, warranty_start_date, warranty_expiry_date,
			warranty_notes, created_by, updated_by
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21
		)
		RETURNING `+assetSelectColumns(r.reminderWindowDay)+`
	`, a.ID, a.Code, a.Type, a.CategoryID, a.Name, a.Brand, a.Model, a.SerialNumber,
		a.PurchaseDate, a.PurchasePrice, a.Status, a.Condition, a.Location, a.AssignedTo,
		a.AssignedUserID, a.Notes, a.WarrantyStartDate, a.WarrantyExpiryDate, a.WarrantyNotes,
		a.CreatedBy, a.UpdatedBy)
	created, err := scanAsset(row)
	if err != nil {
		return domain.Asset{}, mapPgErr(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Asset{}, err
	}
	return created, nil
}

func (r *AssetRepository) UpdateAsset(ctx context.Context, a domain.Asset) (domain.Asset, error) {
	row := r.db.pool.QueryRow(ctx, `
		UPDATE assets SET
			type = $2,
			category_id = $3,
			name = $4,
			brand = $5,
			model = $6,
			serial_number = $7,
			purchase_date = $8,
			purchase_price = $9,
			status = $10,
			condition = $11,
			location = $12,
			assigned_to = $13,
			assigned_user_id = $14,
			notes = $15,
			warranty_start_date = $16,
			warranty_expiry_date = $17,
			warranty_notes = $18,
			updated_by = $19,
			updated_at = now()
		WHERE id = $1
		RETURNING `+assetSelectColumns(r.reminderWindowDay)+`
	`, a.ID, a.Type, a.CategoryID, a.Name, a.Brand, a.Model, a.SerialNumber,
		a.PurchaseDate, a.PurchasePrice, a.Status, a.Condition, a.Location, a.AssignedTo,
		a.AssignedUserID, a.Notes, a.WarrantyStartDate, a.WarrantyExpiryDate, a.WarrantyNotes, a.UpdatedBy)
	updated, err := scanAsset(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Asset{}, domain.ErrNotFound
	}
	return updated, mapPgErr(err)
}

func (r *AssetRepository) ArchiveAsset(ctx context.Context, id, userID int64) error {
	cmd, err := r.db.pool.Exec(ctx, `
		UPDATE assets
		SET archived_at = now(), updated_by = $2, updated_at = now()
		WHERE id = $1
	`, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AssetRepository) RestoreAsset(ctx context.Context, id, userID int64) error {
	cmd, err := r.db.pool.Exec(ctx, `
		UPDATE assets
		SET archived_at = NULL, updated_by = $2, updated_at = now()
		WHERE id = $1
	`, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AssetRepository) GetAsset(ctx context.Context, id int64) (domain.Asset, error) {
	row := r.db.pool.QueryRow(ctx, `
		SELECT `+assetSelectColumnsWithJoins()+`
		WHERE a.id = $2
	`, r.reminderWindowDay, id)
	a, err := scanAsset(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Asset{}, domain.ErrNotFound
	}
	return a, err
}

func (r *AssetRepository) ListAssets(ctx context.Context, f domain.AssetFilters) ([]domain.Asset, error) {
	args := []any{f.ReminderWindowDay}
	where := []string{"1 = 1"}
	if !f.IncludeArchived {
		where = append(where, "a.archived_at IS NULL")
	}
	if f.CurrentUserRole != domain.RoleAdmin {
		args = append(args, f.CurrentUserID)
		where = append(where, fmt.Sprintf("(a.created_by = $%d OR a.assigned_user_id = $%d)", len(args), len(args)))
	}
	if f.Query != "" {
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(f.Query))+"%")
		idx := len(args)
		where = append(where, fmt.Sprintf(`(
			lower(a.name) LIKE $%d OR lower(a.brand) LIKE $%d OR lower(a.model) LIKE $%d OR
			lower(a.serial_number) LIKE $%d OR lower(a.notes) LIKE $%d OR
			EXISTS (SELECT 1 FROM vehicle_profiles vp WHERE vp.asset_id = a.id AND lower(vp.registration_number) LIKE $%d)
		)`, idx, idx, idx, idx, idx, idx))
	}
	if f.CategoryID > 0 {
		args = append(args, f.CategoryID)
		where = append(where, fmt.Sprintf("a.category_id = $%d", len(args)))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("a.status = $%d", len(args)))
	}
	if f.Location != "" {
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(f.Location))+"%")
		where = append(where, fmt.Sprintf("lower(a.location) LIKE $%d", len(args)))
	}
	if f.AssignedUserID > 0 {
		args = append(args, f.AssignedUserID)
		where = append(where, fmt.Sprintf("a.assigned_user_id = $%d", len(args)))
	}
	if f.WarrantyState != "" {
		where = append(where, fmt.Sprintf(`CASE
			WHEN a.warranty_expiry_date IS NULL THEN 'not_set'
			WHEN a.warranty_expiry_date < current_date THEN 'expired'
			WHEN a.warranty_expiry_date <= current_date + ($1::int * interval '1 day') THEN 'expiring_soon'
			ELSE 'active'
		END = '%s'`, strings.ReplaceAll(f.WarrantyState, "'", "''")))
	}
	if f.HasDocuments != nil {
		if *f.HasDocuments {
			where = append(where, "doc_counts.document_count > 0")
		} else {
			where = append(where, "COALESCE(doc_counts.document_count, 0) = 0")
		}
	}
	query := `
		SELECT ` + assetSelectColumnsWithJoins() + `
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY a.created_at DESC
		LIMIT 200
	`
	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Asset{}
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func assetSelectColumns(windowDays int) string {
	return `id, code, type, category_id, ''::text AS category_name, name, brand, model, serial_number,
		purchase_date, purchase_price::float8, status, condition, location, assigned_to,
		assigned_user_id, ''::text AS assigned_user_name, notes, warranty_start_date, warranty_expiry_date,
		warranty_notes, archived_at, created_by, updated_by, created_at, updated_at,
		0::int AS document_count,
		CASE
			WHEN warranty_expiry_date IS NULL THEN 'not_set'
			WHEN warranty_expiry_date < current_date THEN 'expired'
			WHEN warranty_expiry_date <= current_date + (` + strconv.Itoa(windowDays) + ` * interval '1 day') THEN 'expiring_soon'
			ELSE 'active'
		END AS warranty_state`
}

func assetSelectColumnsWithJoins() string {
	return `a.id, a.code, a.type, a.category_id, c.name AS category_name, a.name, a.brand, a.model, a.serial_number,
		a.purchase_date, a.purchase_price::float8, a.status, a.condition, a.location, a.assigned_to,
		a.assigned_user_id, COALESCE(u.full_name, '') AS assigned_user_name, a.notes, a.warranty_start_date, a.warranty_expiry_date,
		a.warranty_notes, a.archived_at, a.created_by, a.updated_by, a.created_at, a.updated_at,
		COALESCE(doc_counts.document_count, 0)::int AS document_count,
		CASE
			WHEN a.warranty_expiry_date IS NULL THEN 'not_set'
			WHEN a.warranty_expiry_date < current_date THEN 'expired'
			WHEN a.warranty_expiry_date <= current_date + ($1::int * interval '1 day') THEN 'expiring_soon'
			ELSE 'active'
		END AS warranty_state
		FROM assets a
		JOIN asset_categories c ON c.id = a.category_id
		LEFT JOIN users u ON u.id = a.assigned_user_id
		LEFT JOIN (
			SELECT asset_id, count(*) AS document_count FROM asset_documents GROUP BY asset_id
		) doc_counts ON doc_counts.asset_id = a.id`
}

func scanAsset(row pgx.Row) (domain.Asset, error) {
	var a domain.Asset
	err := row.Scan(
		&a.ID, &a.Code, &a.Type, &a.CategoryID, &a.CategoryName, &a.Name, &a.Brand, &a.Model, &a.SerialNumber,
		&a.PurchaseDate, &a.PurchasePrice, &a.Status, &a.Condition, &a.Location, &a.AssignedTo,
		&a.AssignedUserID, &a.AssignedUserName, &a.Notes, &a.WarrantyStartDate, &a.WarrantyExpiryDate,
		&a.WarrantyNotes, &a.ArchivedAt, &a.CreatedBy, &a.UpdatedBy, &a.CreatedAt, &a.UpdatedAt,
		&a.DocumentCount, &a.WarrantyState,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Asset{}, domain.ErrNotFound
	}
	return a, err
}
