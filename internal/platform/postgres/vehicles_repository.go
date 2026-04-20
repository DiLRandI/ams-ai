package postgres

import (
	"context"
	"errors"

	"ams-ai/internal/domain"

	"github.com/jackc/pgx/v5"
)

type VehicleRepository struct {
	db *DB
}

func NewVehicleRepository(db *DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) UpsertVehicleProfile(ctx context.Context, p domain.VehicleProfile) (domain.VehicleProfile, error) {
	row := r.db.pool.QueryRow(ctx, `
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

func (r *VehicleRepository) GetVehicleProfile(ctx context.Context, assetID int64) (domain.VehicleProfile, error) {
	row := r.db.pool.QueryRow(ctx, `
		SELECT asset_id, registration_number, vehicle_type, chassis_number, engine_number,
			odometer, assigned_driver, next_service_date, next_service_mileage, notes, created_at, updated_at
		FROM vehicle_profiles
		WHERE asset_id = $1
	`, assetID)
	return scanVehicleProfile(row)
}

func (r *VehicleRepository) CreateInsuranceRecord(ctx context.Context, rec domain.VehicleInsuranceRecord) (domain.VehicleInsuranceRecord, error) {
	row := r.db.pool.QueryRow(ctx, `
		INSERT INTO vehicle_insurance_records (asset_id, provider, policy_number, cost, start_date, expiry_date, document_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, asset_id, provider, policy_number, cost::float8, start_date, expiry_date, document_id, notes, created_at
	`, rec.AssetID, rec.Provider, rec.PolicyNumber, rec.Cost, rec.StartDate, rec.ExpiryDate, rec.DocumentID, rec.Notes)
	return scanInsurance(row)
}

func (r *VehicleRepository) ListInsuranceRecords(ctx context.Context, assetID int64) ([]domain.VehicleInsuranceRecord, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, asset_id, provider, policy_number, cost::float8, start_date, expiry_date, document_id, notes, created_at
		FROM vehicle_insurance_records
		WHERE asset_id = $1
		ORDER BY expiry_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.VehicleInsuranceRecord{}
	for rows.Next() {
		rec, err := scanInsurance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *VehicleRepository) CreateLicenseRecord(ctx context.Context, rec domain.VehicleLicenseRecord) (domain.VehicleLicenseRecord, error) {
	row := r.db.pool.QueryRow(ctx, `
		INSERT INTO vehicle_license_records (asset_id, renewal_type, reference_number, cost, issue_date, expiry_date, document_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, asset_id, renewal_type, reference_number, cost::float8, issue_date, expiry_date, document_id, notes, created_at
	`, rec.AssetID, rec.RenewalType, rec.ReferenceNumber, rec.Cost, rec.IssueDate, rec.ExpiryDate, rec.DocumentID, rec.Notes)
	return scanLicense(row)
}

func (r *VehicleRepository) ListLicenseRecords(ctx context.Context, assetID int64) ([]domain.VehicleLicenseRecord, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, asset_id, renewal_type, reference_number, cost::float8, issue_date, expiry_date, document_id, notes, created_at
		FROM vehicle_license_records
		WHERE asset_id = $1
		ORDER BY expiry_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.VehicleLicenseRecord{}
	for rows.Next() {
		rec, err := scanLicense(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *VehicleRepository) CreateEmissionRecord(ctx context.Context, rec domain.VehicleEmissionRecord) (domain.VehicleEmissionRecord, error) {
	row := r.db.pool.QueryRow(ctx, `
		INSERT INTO vehicle_emission_records (asset_id, inspection_type, reference_number, cost, issue_date, expiry_date, document_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, asset_id, inspection_type, reference_number, cost::float8, issue_date, expiry_date, document_id, notes, created_at
	`, rec.AssetID, rec.InspectionType, rec.ReferenceNumber, rec.Cost, rec.IssueDate, rec.ExpiryDate, rec.DocumentID, rec.Notes)
	return scanEmission(row)
}

func (r *VehicleRepository) ListEmissionRecords(ctx context.Context, assetID int64) ([]domain.VehicleEmissionRecord, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, asset_id, inspection_type, reference_number, cost::float8, issue_date, expiry_date, document_id, notes, created_at
		FROM vehicle_emission_records
		WHERE asset_id = $1
		ORDER BY expiry_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.VehicleEmissionRecord{}
	for rows.Next() {
		rec, err := scanEmission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *VehicleRepository) CreateServiceRecord(ctx context.Context, rec domain.ServiceRecord) (domain.ServiceRecord, error) {
	row := r.db.pool.QueryRow(ctx, `
		INSERT INTO service_records (
			asset_id, service_type, service_date, cost, vendor, description, notes,
			mileage, next_service_date, next_service_mileage, created_by
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, asset_id, service_type, service_date, cost::float8, vendor, description, notes,
			mileage, next_service_date, next_service_mileage, created_by, created_at
	`, rec.AssetID, rec.ServiceType, rec.ServiceDate, rec.Cost, rec.Vendor, rec.Description, rec.Notes,
		rec.Mileage, rec.NextServiceDate, rec.NextServiceMileage, rec.CreatedBy)
	return scanService(row)
}

func (r *VehicleRepository) ListServiceRecords(ctx context.Context, assetID int64) ([]domain.ServiceRecord, error) {
	rows, err := r.db.pool.Query(ctx, `
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
	out := []domain.ServiceRecord{}
	for rows.Next() {
		rec, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *VehicleRepository) CreateFuelLog(ctx context.Context, rec domain.FuelLog) (domain.FuelLog, error) {
	row := r.db.pool.QueryRow(ctx, `
		INSERT INTO fuel_logs (asset_id, fuel_date, fuel_type, quantity, cost, odometer, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, asset_id, fuel_date, fuel_type, quantity::float8, cost::float8, odometer, notes, created_by, created_at
	`, rec.AssetID, rec.FuelDate, rec.FuelType, rec.Quantity, rec.Cost, rec.Odometer, rec.Notes, rec.CreatedBy)
	return scanFuel(row)
}

func (r *VehicleRepository) ListFuelLogs(ctx context.Context, assetID int64) ([]domain.FuelLog, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, asset_id, fuel_date, fuel_type, quantity::float8, cost::float8, odometer, notes, created_by, created_at
		FROM fuel_logs
		WHERE asset_id = $1
		ORDER BY fuel_date DESC, created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.FuelLog{}
	for rows.Next() {
		rec, err := scanFuel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
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

func scanInsurance(row pgx.Row) (domain.VehicleInsuranceRecord, error) {
	var rec domain.VehicleInsuranceRecord
	err := row.Scan(&rec.ID, &rec.AssetID, &rec.Provider, &rec.PolicyNumber, &rec.Cost, &rec.StartDate, &rec.ExpiryDate, &rec.DocumentID, &rec.Notes, &rec.CreatedAt)
	return rec, mapPgErr(err)
}

func scanLicense(row pgx.Row) (domain.VehicleLicenseRecord, error) {
	var rec domain.VehicleLicenseRecord
	err := row.Scan(&rec.ID, &rec.AssetID, &rec.RenewalType, &rec.ReferenceNumber, &rec.Cost, &rec.IssueDate, &rec.ExpiryDate, &rec.DocumentID, &rec.Notes, &rec.CreatedAt)
	return rec, mapPgErr(err)
}

func scanEmission(row pgx.Row) (domain.VehicleEmissionRecord, error) {
	var rec domain.VehicleEmissionRecord
	err := row.Scan(&rec.ID, &rec.AssetID, &rec.InspectionType, &rec.ReferenceNumber, &rec.Cost, &rec.IssueDate, &rec.ExpiryDate, &rec.DocumentID, &rec.Notes, &rec.CreatedAt)
	return rec, mapPgErr(err)
}

func scanService(row pgx.Row) (domain.ServiceRecord, error) {
	var rec domain.ServiceRecord
	err := row.Scan(&rec.ID, &rec.AssetID, &rec.ServiceType, &rec.ServiceDate, &rec.Cost, &rec.Vendor, &rec.Description, &rec.Notes,
		&rec.Mileage, &rec.NextServiceDate, &rec.NextServiceMileage, &rec.CreatedBy, &rec.CreatedAt)
	return rec, mapPgErr(err)
}

func scanFuel(row pgx.Row) (domain.FuelLog, error) {
	var rec domain.FuelLog
	err := row.Scan(&rec.ID, &rec.AssetID, &rec.FuelDate, &rec.FuelType, &rec.Quantity, &rec.Cost, &rec.Odometer, &rec.Notes, &rec.CreatedBy, &rec.CreatedAt)
	return rec, mapPgErr(err)
}
