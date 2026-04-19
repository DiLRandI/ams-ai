package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ams-ai/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)
	`, strings.TrimSpace(email))
	return scanUser(row)
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (domain.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)
	return scanUser(row)
}

func (s *Store) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		ORDER BY full_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Store) UpsertSeedUser(ctx context.Context, email, password, fullName, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = EXCLUDED.role,
			updated_at = now()
	`, email, string(hash), fullName, role)
	return err
}

func scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	return u, err
}

func (s *Store) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, description, is_system, created_at, updated_at
		FROM asset_categories
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) CreateCategory(ctx context.Context, name, description string) (domain.Category, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO asset_categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, is_system, created_at, updated_at
	`, strings.TrimSpace(name), strings.TrimSpace(description))
	var c domain.Category
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt)
	return c, mapPgErr(err)
}

func (s *Store) UpdateCategory(ctx context.Context, id int64, name, description string) (domain.Category, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE asset_categories
		SET name = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, name, description, is_system, created_at, updated_at
	`, id, strings.TrimSpace(name), strings.TrimSpace(description))
	var c domain.Category
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Category{}, domain.ErrNotFound
	}
	return c, mapPgErr(err)
}

func (s *Store) CreateAsset(ctx context.Context, a domain.Asset) (domain.Asset, error) {
	tx, err := s.pool.Begin(ctx)
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
		RETURNING `+assetSelectColumns()+`
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

func (s *Store) UpdateAsset(ctx context.Context, a domain.Asset) (domain.Asset, error) {
	row := s.pool.QueryRow(ctx, `
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
		RETURNING `+assetSelectColumns()+`
	`, a.ID, a.Type, a.CategoryID, a.Name, a.Brand, a.Model, a.SerialNumber,
		a.PurchaseDate, a.PurchasePrice, a.Status, a.Condition, a.Location, a.AssignedTo,
		a.AssignedUserID, a.Notes, a.WarrantyStartDate, a.WarrantyExpiryDate, a.WarrantyNotes, a.UpdatedBy)
	updated, err := scanAsset(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Asset{}, domain.ErrNotFound
	}
	return updated, mapPgErr(err)
}

func (s *Store) ArchiveAsset(ctx context.Context, id, userID int64) error {
	cmd, err := s.pool.Exec(ctx, `
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

func (s *Store) RestoreAsset(ctx context.Context, id, userID int64) error {
	cmd, err := s.pool.Exec(ctx, `
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

func (s *Store) GetAsset(ctx context.Context, id int64) (domain.Asset, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+assetSelectColumnsWithJoins()+`
		WHERE a.id = $2
	`, 30, id)
	a, err := scanAsset(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Asset{}, domain.ErrNotFound
	}
	return a, err
}

func (s *Store) ListAssets(ctx context.Context, f domain.AssetFilters) ([]domain.Asset, error) {
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
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Asset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func assetSelectColumns() string {
	return `id, code, type, category_id, ''::text AS category_name, name, brand, model, serial_number,
		purchase_date, purchase_price::float8, status, condition, location, assigned_to,
		assigned_user_id, ''::text AS assigned_user_name, notes, warranty_start_date, warranty_expiry_date,
		warranty_notes, archived_at, created_by, updated_by, created_at, updated_at,
		0::int AS document_count,
		CASE
			WHEN warranty_expiry_date IS NULL THEN 'not_set'
			WHEN warranty_expiry_date < current_date THEN 'expired'
			WHEN warranty_expiry_date <= current_date + interval '30 day' THEN 'expiring_soon'
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

func (s *Store) UpsertVehicleProfile(ctx context.Context, p domain.VehicleProfile) (domain.VehicleProfile, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_profiles (
			asset_id, registration_number, vehicle_type, chassis_number, engine_number,
			odometer, assigned_driver, next_service_date, next_service_mileage, notes
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (asset_id) DO UPDATE SET
			registration_number = EXCLUDED.registration_number,
			vehicle_type = EXCLUDED.vehicle_type,
			chassis_number = EXCLUDED.chassis_number,
			engine_number = EXCLUDED.engine_number,
			odometer = EXCLUDED.odometer,
			assigned_driver = EXCLUDED.assigned_driver,
			next_service_date = EXCLUDED.next_service_date,
			next_service_mileage = EXCLUDED.next_service_mileage,
			notes = EXCLUDED.notes,
			updated_at = now()
		RETURNING asset_id, registration_number, vehicle_type, chassis_number, engine_number,
			odometer, assigned_driver, next_service_date, next_service_mileage, notes, created_at, updated_at
	`, p.AssetID, p.RegistrationNumber, p.VehicleType, p.ChassisNumber, p.EngineNumber, p.Odometer,
		p.AssignedDriver, p.NextServiceDate, p.NextServiceMileage, p.Notes)
	return scanVehicleProfile(row)
}

func (s *Store) GetVehicleProfile(ctx context.Context, assetID int64) (domain.VehicleProfile, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT asset_id, registration_number, vehicle_type, chassis_number, engine_number,
			odometer, assigned_driver, next_service_date, next_service_mileage, notes, created_at, updated_at
		FROM vehicle_profiles
		WHERE asset_id = $1
	`, assetID)
	return scanVehicleProfile(row)
}

func scanVehicleProfile(row pgx.Row) (domain.VehicleProfile, error) {
	var p domain.VehicleProfile
	err := row.Scan(&p.AssetID, &p.RegistrationNumber, &p.VehicleType, &p.ChassisNumber, &p.EngineNumber,
		&p.Odometer, &p.AssignedDriver, &p.NextServiceDate, &p.NextServiceMileage, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VehicleProfile{}, domain.ErrNotFound
	}
	return p, mapPgErr(err)
}

func (s *Store) CreateInsuranceRecord(ctx context.Context, r domain.VehicleInsuranceRecord) (domain.VehicleInsuranceRecord, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_insurance_records (asset_id, provider, policy_number, cost, start_date, expiry_date, document_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, asset_id, provider, policy_number, cost::float8, start_date, expiry_date, document_id, notes, created_at
	`, r.AssetID, r.Provider, r.PolicyNumber, r.Cost, r.StartDate, r.ExpiryDate, r.DocumentID, r.Notes)
	return scanInsurance(row)
}

func (s *Store) ListInsuranceRecords(ctx context.Context, assetID int64) ([]domain.VehicleInsuranceRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, asset_id, provider, policy_number, cost::float8, start_date, expiry_date, document_id, notes, created_at
		FROM vehicle_insurance_records
		WHERE asset_id = $1
		ORDER BY expiry_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.VehicleInsuranceRecord
	for rows.Next() {
		r, err := scanInsurance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanInsurance(row pgx.Row) (domain.VehicleInsuranceRecord, error) {
	var r domain.VehicleInsuranceRecord
	err := row.Scan(&r.ID, &r.AssetID, &r.Provider, &r.PolicyNumber, &r.Cost, &r.StartDate, &r.ExpiryDate, &r.DocumentID, &r.Notes, &r.CreatedAt)
	return r, mapPgErr(err)
}

func (s *Store) CreateLicenseRecord(ctx context.Context, r domain.VehicleLicenseRecord) (domain.VehicleLicenseRecord, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_license_records (asset_id, renewal_type, reference_number, cost, issue_date, expiry_date, document_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, asset_id, renewal_type, reference_number, cost::float8, issue_date, expiry_date, document_id, notes, created_at
	`, r.AssetID, r.RenewalType, r.ReferenceNumber, r.Cost, r.IssueDate, r.ExpiryDate, r.DocumentID, r.Notes)
	return scanLicense(row)
}

func (s *Store) ListLicenseRecords(ctx context.Context, assetID int64) ([]domain.VehicleLicenseRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, asset_id, renewal_type, reference_number, cost::float8, issue_date, expiry_date, document_id, notes, created_at
		FROM vehicle_license_records
		WHERE asset_id = $1
		ORDER BY expiry_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.VehicleLicenseRecord
	for rows.Next() {
		r, err := scanLicense(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanLicense(row pgx.Row) (domain.VehicleLicenseRecord, error) {
	var r domain.VehicleLicenseRecord
	err := row.Scan(&r.ID, &r.AssetID, &r.RenewalType, &r.ReferenceNumber, &r.Cost, &r.IssueDate, &r.ExpiryDate, &r.DocumentID, &r.Notes, &r.CreatedAt)
	return r, mapPgErr(err)
}

func (s *Store) CreateEmissionRecord(ctx context.Context, r domain.VehicleEmissionRecord) (domain.VehicleEmissionRecord, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_emission_records (asset_id, inspection_type, reference_number, cost, issue_date, expiry_date, document_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, asset_id, inspection_type, reference_number, cost::float8, issue_date, expiry_date, document_id, notes, created_at
	`, r.AssetID, r.InspectionType, r.ReferenceNumber, r.Cost, r.IssueDate, r.ExpiryDate, r.DocumentID, r.Notes)
	return scanEmission(row)
}

func (s *Store) ListEmissionRecords(ctx context.Context, assetID int64) ([]domain.VehicleEmissionRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, asset_id, inspection_type, reference_number, cost::float8, issue_date, expiry_date, document_id, notes, created_at
		FROM vehicle_emission_records
		WHERE asset_id = $1
		ORDER BY expiry_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.VehicleEmissionRecord
	for rows.Next() {
		r, err := scanEmission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanEmission(row pgx.Row) (domain.VehicleEmissionRecord, error) {
	var r domain.VehicleEmissionRecord
	err := row.Scan(&r.ID, &r.AssetID, &r.InspectionType, &r.ReferenceNumber, &r.Cost, &r.IssueDate, &r.ExpiryDate, &r.DocumentID, &r.Notes, &r.CreatedAt)
	return r, mapPgErr(err)
}

func (s *Store) CreateServiceRecord(ctx context.Context, r domain.ServiceRecord) (domain.ServiceRecord, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO service_records (
			asset_id, service_type, service_date, cost, vendor, description, notes,
			mileage, next_service_date, next_service_mileage, created_by
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, asset_id, service_type, service_date, cost::float8, vendor, description, notes,
			mileage, next_service_date, next_service_mileage, created_by, created_at
	`, r.AssetID, r.ServiceType, r.ServiceDate, r.Cost, r.Vendor, r.Description, r.Notes,
		r.Mileage, r.NextServiceDate, r.NextServiceMileage, r.CreatedBy)
	return scanService(row)
}

func (s *Store) ListServiceRecords(ctx context.Context, assetID int64) ([]domain.ServiceRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, asset_id, service_type, service_date, cost::float8, vendor, description, notes,
			mileage, next_service_date, next_service_mileage, created_by, created_at
		FROM service_records
		WHERE asset_id = $1
		ORDER BY service_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ServiceRecord
	for rows.Next() {
		r, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanService(row pgx.Row) (domain.ServiceRecord, error) {
	var r domain.ServiceRecord
	err := row.Scan(&r.ID, &r.AssetID, &r.ServiceType, &r.ServiceDate, &r.Cost, &r.Vendor, &r.Description, &r.Notes,
		&r.Mileage, &r.NextServiceDate, &r.NextServiceMileage, &r.CreatedBy, &r.CreatedAt)
	return r, mapPgErr(err)
}

func (s *Store) CreateFuelLog(ctx context.Context, r domain.FuelLog) (domain.FuelLog, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO fuel_logs (asset_id, fuel_date, fuel_type, quantity, cost, odometer, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, asset_id, fuel_date, fuel_type, quantity::float8, cost::float8, odometer, notes, created_by, created_at
	`, r.AssetID, r.FuelDate, r.FuelType, r.Quantity, r.Cost, r.Odometer, r.Notes, r.CreatedBy)
	return scanFuel(row)
}

func (s *Store) ListFuelLogs(ctx context.Context, assetID int64) ([]domain.FuelLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, asset_id, fuel_date, fuel_type, quantity::float8, cost::float8, odometer, notes, created_by, created_at
		FROM fuel_logs
		WHERE asset_id = $1
		ORDER BY fuel_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.FuelLog
	for rows.Next() {
		r, err := scanFuel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanFuel(row pgx.Row) (domain.FuelLog, error) {
	var r domain.FuelLog
	err := row.Scan(&r.ID, &r.AssetID, &r.FuelDate, &r.FuelType, &r.Quantity, &r.Cost, &r.Odometer, &r.Notes, &r.CreatedBy, &r.CreatedAt)
	return r, mapPgErr(err)
}

func (s *Store) CreateDocument(ctx context.Context, d domain.AssetDocument) (domain.AssetDocument, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO asset_documents (asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by, created_at
	`, d.AssetID, d.Title, d.Type, d.Notes, d.FileName, d.ContentType, d.SizeBytes, d.ObjectKey, d.UploadedBy)
	return scanDocument(row)
}

func (s *Store) GetDocument(ctx context.Context, id int64) (domain.AssetDocument, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by, created_at
		FROM asset_documents
		WHERE id = $1
	`, id)
	return scanDocument(row)
}

func (s *Store) ListDocuments(ctx context.Context, assetID int64) ([]domain.AssetDocument, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by, created_at
		FROM asset_documents
		WHERE asset_id = $1
		ORDER BY created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AssetDocument
	for rows.Next() {
		d, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) DeleteDocument(ctx context.Context, id int64) error {
	cmd, err := s.pool.Exec(ctx, `DELETE FROM asset_documents WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanDocument(row pgx.Row) (domain.AssetDocument, error) {
	var d domain.AssetDocument
	err := row.Scan(&d.ID, &d.AssetID, &d.Title, &d.Type, &d.Notes, &d.FileName, &d.ContentType, &d.SizeBytes, &d.ObjectKey, &d.UploadedBy, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AssetDocument{}, domain.ErrNotFound
	}
	return d, mapPgErr(err)
}

func (s *Store) Dashboard(ctx context.Context, user domain.User, windowDays int) (domain.Dashboard, error) {
	total, err := s.countVisibleAssets(ctx, user)
	if err != nil {
		return domain.Dashboard{}, err
	}
	byCategory, err := s.assetsByCategory(ctx, user)
	if err != nil {
		return domain.Dashboard{}, err
	}
	expiring, err := s.ListAssets(ctx, domain.AssetFilters{
		WarrantyState:     domain.WarrantyExpiringSoon,
		CurrentUserID:     user.ID,
		CurrentUserRole:   user.Role,
		ReminderWindowDay: windowDays,
	})
	if err != nil {
		return domain.Dashboard{}, err
	}
	recent, err := s.recentAssets(ctx, user, windowDays)
	if err != nil {
		return domain.Dashboard{}, err
	}
	reminders, err := s.ListReminders(ctx, user, 20)
	if err != nil {
		return domain.Dashboard{}, err
	}
	var insurance, licenses, service []domain.Reminder
	for _, r := range reminders {
		switch r.SourceType {
		case "insurance":
			insurance = append(insurance, r)
		case "license":
			licenses = append(licenses, r)
		case "service":
			service = append(service, r)
		}
	}
	return domain.Dashboard{
		TotalAssets:              total,
		AssetsByCategory:         byCategory,
		ExpiringWarranties:       takeAssets(expiring, 8),
		ExpiringVehicleInsurance: takeReminders(insurance, 8),
		ExpiringVehicleLicenses:  takeReminders(licenses, 8),
		RecentlyAddedAssets:      recent,
		ServiceDueSoon:           takeReminders(service, 8),
		UpcomingReminders:        reminders,
	}, nil
}

func (s *Store) countVisibleAssets(ctx context.Context, user domain.User) (int, error) {
	args := []any{}
	where := "archived_at IS NULL"
	if user.Role != domain.RoleAdmin {
		args = append(args, user.ID)
		where += " AND (created_by = $1 OR assigned_user_id = $1)"
	}
	var count int
	err := s.pool.QueryRow(ctx, "SELECT count(*) FROM assets WHERE "+where, args...).Scan(&count)
	return count, err
}

func (s *Store) assetsByCategory(ctx context.Context, user domain.User) ([]domain.CategoryCount, error) {
	args := []any{}
	where := "a.archived_at IS NULL"
	if user.Role != domain.RoleAdmin {
		args = append(args, user.ID)
		where += " AND (a.created_by = $1 OR a.assigned_user_id = $1)"
	}
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.name, count(a.id)::int
		FROM asset_categories c
		LEFT JOIN assets a ON a.category_id = c.id AND `+where+`
		GROUP BY c.id, c.name
		ORDER BY c.name
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CategoryCount
	for rows.Next() {
		var c domain.CategoryCount
		if err := rows.Scan(&c.CategoryID, &c.CategoryName, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) recentAssets(ctx context.Context, user domain.User, windowDays int) ([]domain.Asset, error) {
	return s.ListAssets(ctx, domain.AssetFilters{
		CurrentUserID:     user.ID,
		CurrentUserRole:   user.Role,
		ReminderWindowDay: windowDays,
	})
}

func (s *Store) RegenerateReminders(ctx context.Context, windowDays int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM reminders`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		SELECT id, 'warranty', id, name || ' warranty expires', warranty_expiry_date,
			CASE WHEN warranty_expiry_date < current_date THEN 'overdue'
				 WHEN warranty_expiry_date = current_date THEN 'due'
				 ELSE 'upcoming' END
		FROM assets
		WHERE archived_at IS NULL
		  AND warranty_expiry_date IS NOT NULL
		  AND warranty_expiry_date <= current_date + ($1::int * interval '1 day')
	`, windowDays); err != nil {
		return err
	}
	for _, q := range []string{
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'insurance', r.id, a.name || ' insurance expires', r.expiry_date,
		 CASE WHEN r.expiry_date < current_date THEN 'overdue' WHEN r.expiry_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM vehicle_insurance_records r JOIN assets a ON a.id = r.asset_id
		 WHERE a.archived_at IS NULL AND r.expiry_date <= current_date + ($1::int * interval '1 day')`,
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'license', r.id, a.name || ' license expires', r.expiry_date,
		 CASE WHEN r.expiry_date < current_date THEN 'overdue' WHEN r.expiry_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM vehicle_license_records r JOIN assets a ON a.id = r.asset_id
		 WHERE a.archived_at IS NULL AND r.expiry_date <= current_date + ($1::int * interval '1 day')`,
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'emission', r.id, a.name || ' emission/inspection expires', r.expiry_date,
		 CASE WHEN r.expiry_date < current_date THEN 'overdue' WHEN r.expiry_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM vehicle_emission_records r JOIN assets a ON a.id = r.asset_id
		 WHERE a.archived_at IS NULL AND r.expiry_date <= current_date + ($1::int * interval '1 day')`,
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'service', s.id, a.name || ' service due', s.next_service_date,
		 CASE WHEN s.next_service_date < current_date THEN 'overdue' WHEN s.next_service_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM service_records s JOIN assets a ON a.id = s.asset_id
		 WHERE a.archived_at IS NULL AND s.next_service_date IS NOT NULL AND s.next_service_date <= current_date + ($1::int * interval '1 day')`,
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'service', -vp.asset_id, a.name || ' service due', vp.next_service_date,
		 CASE WHEN vp.next_service_date < current_date THEN 'overdue' WHEN vp.next_service_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM vehicle_profiles vp JOIN assets a ON a.id = vp.asset_id
		 WHERE a.archived_at IS NULL
		   AND vp.next_service_date IS NOT NULL
		   AND vp.next_service_date <= current_date + ($1::int * interval '1 day')
		   AND NOT EXISTS (
		   	SELECT 1 FROM service_records sr
		   	WHERE sr.asset_id = vp.asset_id AND sr.next_service_date IS NOT NULL
		   )`,
	} {
		if _, err := tx.Exec(ctx, q, windowDays); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) ListReminders(ctx context.Context, user domain.User, limit int) ([]domain.Reminder, error) {
	if limit <= 0 {
		limit = 100
	}
	args := []any{limit}
	where := "a.archived_at IS NULL"
	if user.Role != domain.RoleAdmin {
		args = append(args, user.ID)
		where += fmt.Sprintf(" AND (a.created_by = $%d OR a.assigned_user_id = $%d)", len(args), len(args))
	}
	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.asset_id, a.code, a.name, r.source_type, r.source_id, r.title, r.due_date, r.state, r.created_at
		FROM reminders r
		JOIN assets a ON a.id = r.asset_id
		WHERE `+where+`
		ORDER BY r.due_date ASC, r.id ASC
		LIMIT $1
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Reminder
	for rows.Next() {
		var r domain.Reminder
		if err := rows.Scan(&r.ID, &r.AssetID, &r.AssetCode, &r.AssetName, &r.SourceType, &r.SourceID, &r.Title, &r.DueDate, &r.State, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func takeAssets(in []domain.Asset, n int) []domain.Asset {
	if in == nil {
		return []domain.Asset{}
	}
	if len(in) <= n {
		return in
	}
	return in[:n]
}

func takeReminders(in []domain.Reminder, n int) []domain.Reminder {
	if in == nil {
		return []domain.Reminder{}
	}
	if len(in) <= n {
		return in
	}
	return in[:n]
}

func mapPgErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: duplicate value", domain.ErrConflict)
		case "23503":
			return fmt.Errorf("%w: referenced record not found", domain.ErrInvalid)
		case "23514":
			return fmt.Errorf("%w: invalid value", domain.ErrInvalid)
		}
	}
	return err
}

func (s *Store) ExportRows(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return s.pool.Query(ctx, query, args...)
}

func (s *Store) Now(ctx context.Context) (time.Time, error) {
	var now time.Time
	err := s.pool.QueryRow(ctx, `SELECT now()`).Scan(&now)
	return now, err
}
