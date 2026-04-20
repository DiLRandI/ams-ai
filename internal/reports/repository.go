package reports

import (
	"context"

	"ams-ai/internal/domain"
)

type AssetLister interface {
	ListAssets(ctx context.Context, user domain.User, filters domain.AssetFilters) ([]domain.Asset, error)
}

type VehicleRecords interface {
	ListInsuranceRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleInsuranceRecord, error)
	ListLicenseRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleLicenseRecord, error)
	ListEmissionRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleEmissionRecord, error)
	ListServiceRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.ServiceRecord, error)
	ListFuelLogs(ctx context.Context, user domain.User, assetID int64) ([]domain.FuelLog, error)
}
