package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"ams-ai/internal/config"
	"ams-ai/internal/domain"
	"ams-ai/internal/platform/postgres"

	"github.com/jackc/pgx/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		exit(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		exit(err)
	}
	defer db.Close()

	authRepo := postgres.NewAuthRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	assetRepo := postgres.NewAssetRepository(db, cfg.ReminderWindowDays)
	vehicleRepo := postgres.NewVehicleRepository(db)
	reminderRepo := postgres.NewReminderRepository(db)

	if err := authRepo.UpsertSeedUser(ctx, "admin@example.com", "admin123", "Demo Admin", domain.RoleAdmin); err != nil {
		exit(err)
	}
	if err := authRepo.UpsertSeedUser(ctx, "user@example.com", "user123", "Demo User", domain.RoleUser); err != nil {
		exit(err)
	}

	admin, err := authRepo.GetUserByEmail(ctx, "admin@example.com")
	if err != nil {
		exit(err)
	}
	user, err := authRepo.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		exit(err)
	}
	categories, err := categoryRepo.ListCategories(ctx)
	if err != nil {
		exit(err)
	}
	categoryID := func(name string) int64 {
		for _, category := range categories {
			if category.Name == name {
				return category.ID
			}
		}
		return categories[0].ID
	}

	laptopID, err := ensureAsset(ctx, db, assetRepo, domain.Asset{
		Type:               domain.AssetTypeGeneral,
		CategoryID:         categoryID("IT devices"),
		Name:               "Demo Laptop",
		Brand:              "Lenovo",
		Model:              "ThinkPad T14",
		SerialNumber:       "SN-DEMO-001",
		PurchaseDate:       datePtr("2025-01-10"),
		PurchasePrice:      floatPtr(1250),
		Status:             domain.StatusActive,
		Condition:          "Good",
		Location:           "Home office",
		AssignedTo:         "Demo User",
		AssignedUserID:     &user.ID,
		Notes:              "Seeded laptop asset with warranty and service dates.",
		WarrantyStartDate:  datePtr("2025-01-10"),
		WarrantyExpiryDate: datePtr(time.Now().AddDate(0, 0, 20).Format("2006-01-02")),
		WarrantyNotes:      "Manufacturer warranty.",
		CreatedBy:          admin.ID,
		UpdatedBy:          &admin.ID,
	})
	if err != nil {
		exit(err)
	}
	if err := ensureLaptopService(ctx, db, vehicleRepo, laptopID, admin.ID); err != nil {
		exit(err)
	}

	vehicleID, err := ensureAsset(ctx, db, assetRepo, domain.Asset{
		Type:               domain.AssetTypeVehicle,
		CategoryID:         categoryID("Vehicles"),
		Name:               "Demo Company Car",
		Brand:              "Toyota",
		Model:              "Corolla",
		SerialNumber:       "VIN-DEMO-001",
		PurchaseDate:       datePtr("2024-05-01"),
		PurchasePrice:      floatPtr(22000),
		Status:             domain.StatusActive,
		Condition:          "Good",
		Location:           "Office parking",
		AssignedTo:         "Demo Driver",
		AssignedUserID:     &admin.ID,
		Notes:              "Seeded vehicle asset with renewal and fuel records.",
		WarrantyStartDate:  datePtr("2024-05-01"),
		WarrantyExpiryDate: datePtr("2027-05-01"),
		CreatedBy:          admin.ID,
		UpdatedBy:          &admin.ID,
	})
	if err != nil {
		exit(err)
	}
	if _, err := vehicleRepo.UpsertVehicleProfile(ctx, domain.VehicleProfile{
		AssetID:            vehicleID,
		RegistrationNumber: "WP-CAB-1234",
		VehicleType:        "Car",
		ChassisNumber:      "CH-DEMO-001",
		EngineNumber:       "EN-DEMO-001",
		Odometer:           new(18500),
		AssignedDriver:     "Demo Driver",
		NextServiceDate:    datePtr(time.Now().AddDate(0, 0, 14).Format("2006-01-02")),
		NextServiceMileage: new(20000),
		Notes:              "Vehicle profile seed data.",
	}); err != nil {
		exit(err)
	}
	if err := ensureVehicleRecords(ctx, db, vehicleRepo, vehicleID, admin.ID); err != nil {
		exit(err)
	}

	if err := reminderRepo.RegenerateReminders(ctx, cfg.ReminderWindowDays); err != nil {
		exit(err)
	}
	fmt.Println("seed complete")
}

func ensureAsset(ctx context.Context, db *postgres.DB, assets *postgres.AssetRepository, a domain.Asset) (int64, error) {
	var id int64
	err := db.Pool().QueryRow(ctx, `SELECT id FROM assets WHERE name = $1 LIMIT 1`, a.Name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != pgx.ErrNoRows {
		return 0, err
	}
	created, err := assets.CreateAsset(ctx, a)
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

func ensureLaptopService(ctx context.Context, db *postgres.DB, vehicles *postgres.VehicleRepository, assetID, userID int64) error {
	var count int
	if err := db.Pool().QueryRow(ctx, `SELECT count(*) FROM service_records WHERE asset_id = $1`, assetID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := vehicles.CreateServiceRecord(ctx, domain.ServiceRecord{
		AssetID:         assetID,
		ServiceType:     "service",
		ServiceDate:     date("2025-11-15"),
		Cost:            floatPtr(75),
		Vendor:          "Demo Repairs",
		Description:     "Fan cleaning and diagnostics",
		Notes:           "Seeded general asset service record.",
		CreatedBy:       userID,
		NextServiceDate: datePtr(time.Now().AddDate(0, 0, 28).Format("2006-01-02")),
	})
	return err
}

func ensureVehicleRecords(ctx context.Context, db *postgres.DB, vehicles *postgres.VehicleRepository, assetID, userID int64) error {
	var count int
	if err := db.Pool().QueryRow(ctx, `SELECT count(*) FROM vehicle_insurance_records WHERE asset_id = $1`, assetID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := vehicles.CreateInsuranceRecord(ctx, domain.VehicleInsuranceRecord{
			AssetID:      assetID,
			Provider:     "Demo Insurance",
			PolicyNumber: "POL-123",
			Cost:         floatPtr(450),
			StartDate:    datePtr("2026-01-01"),
			ExpiryDate:   date(time.Now().AddDate(0, 0, 25).Format("2006-01-02")),
			Notes:        "Seeded insurance record.",
		}); err != nil {
			return err
		}
	}

	if err := db.Pool().QueryRow(ctx, `SELECT count(*) FROM vehicle_license_records WHERE asset_id = $1`, assetID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := vehicles.CreateLicenseRecord(ctx, domain.VehicleLicenseRecord{
			AssetID:         assetID,
			RenewalType:     "Annual license",
			ReferenceNumber: "LIC-123",
			Cost:            floatPtr(120),
			IssueDate:       datePtr("2026-01-01"),
			ExpiryDate:      date(time.Now().AddDate(0, 0, 35).Format("2006-01-02")),
			Notes:           "Seeded license record.",
		}); err != nil {
			return err
		}
	}

	if err := db.Pool().QueryRow(ctx, `SELECT count(*) FROM vehicle_emission_records WHERE asset_id = $1`, assetID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := vehicles.CreateEmissionRecord(ctx, domain.VehicleEmissionRecord{
			AssetID:         assetID,
			InspectionType:  "Emission test",
			ReferenceNumber: "EM-123",
			Cost:            floatPtr(30),
			IssueDate:       datePtr("2026-01-01"),
			ExpiryDate:      date(time.Now().AddDate(0, 0, 18).Format("2006-01-02")),
			Notes:           "Seeded emission record.",
		}); err != nil {
			return err
		}
	}

	if err := db.Pool().QueryRow(ctx, `SELECT count(*) FROM service_records WHERE asset_id = $1`, assetID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := vehicles.CreateServiceRecord(ctx, domain.ServiceRecord{
			AssetID:            assetID,
			ServiceType:        "service",
			ServiceDate:        date("2026-01-10"),
			Cost:               floatPtr(180),
			Vendor:             "Demo Auto Service",
			Description:        "Oil and filter change",
			Mileage:            new(18000),
			NextServiceDate:    datePtr(time.Now().AddDate(0, 0, 14).Format("2006-01-02")),
			NextServiceMileage: new(20000),
			CreatedBy:          userID,
		}); err != nil {
			return err
		}
	}

	if err := db.Pool().QueryRow(ctx, `SELECT count(*) FROM fuel_logs WHERE asset_id = $1`, assetID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := vehicles.CreateFuelLog(ctx, domain.FuelLog{
			AssetID:   assetID,
			FuelDate:  date("2026-02-01"),
			FuelType:  "Petrol",
			Quantity:  35.5,
			Cost:      62.40,
			Odometer:  new(18600),
			Notes:     "Seeded fuel log.",
			CreatedBy: userID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func date(raw string) time.Time {
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		panic(err)
	}
	return t
}

func datePtr(raw string) *time.Time {
	t := date(raw)
	return &t
}

//go:fix inline
func floatPtr(v float64) *float64 {
	return new(v)
}

//go:fix inline
func intPtr(v int) *int {
	return new(v)
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
