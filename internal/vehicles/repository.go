package vehicles

import (
	"context"

	"ams-ai/internal/domain"
)

type Repository interface {
	UpsertVehicleProfile(ctx context.Context, profile domain.VehicleProfile) (domain.VehicleProfile, error)
	GetVehicleProfile(ctx context.Context, assetID int64) (domain.VehicleProfile, error)
	CreateInsuranceRecord(ctx context.Context, record domain.VehicleInsuranceRecord) (domain.VehicleInsuranceRecord, error)
	ListInsuranceRecords(ctx context.Context, assetID int64) ([]domain.VehicleInsuranceRecord, error)
	CreateLicenseRecord(ctx context.Context, record domain.VehicleLicenseRecord) (domain.VehicleLicenseRecord, error)
	ListLicenseRecords(ctx context.Context, assetID int64) ([]domain.VehicleLicenseRecord, error)
	CreateEmissionRecord(ctx context.Context, record domain.VehicleEmissionRecord) (domain.VehicleEmissionRecord, error)
	ListEmissionRecords(ctx context.Context, assetID int64) ([]domain.VehicleEmissionRecord, error)
	CreateServiceRecord(ctx context.Context, record domain.ServiceRecord) (domain.ServiceRecord, error)
	ListServiceRecords(ctx context.Context, assetID int64) ([]domain.ServiceRecord, error)
	CreateFuelLog(ctx context.Context, log domain.FuelLog) (domain.FuelLog, error)
	ListFuelLogs(ctx context.Context, assetID int64) ([]domain.FuelLog, error)
}

type AssetReader interface {
	GetAsset(ctx context.Context, user domain.User, id int64) (domain.Asset, error)
}
