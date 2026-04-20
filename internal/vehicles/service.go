package vehicles

import (
	"context"
	"fmt"

	"ams-ai/internal/domain"
)

type Service struct {
	repo   Repository
	assets AssetReader
}

func NewService(repo Repository, assets AssetReader) *Service {
	return &Service{repo: repo, assets: assets}
}

func (s *Service) UpsertVehicleProfile(ctx context.Context, user domain.User, p domain.VehicleProfile) (domain.VehicleProfile, error) {
	asset, err := s.assets.GetAsset(ctx, user, p.AssetID)
	if err != nil {
		return domain.VehicleProfile{}, err
	}
	if err := validateVehicleProfile(asset, p); err != nil {
		return domain.VehicleProfile{}, err
	}
	return s.repo.UpsertVehicleProfile(ctx, p)
}

func (s *Service) GetVehicleProfile(ctx context.Context, user domain.User, assetID int64) (domain.VehicleProfile, error) {
	asset, err := s.assets.GetAsset(ctx, user, assetID)
	if err != nil {
		return domain.VehicleProfile{}, err
	}
	if asset.Type != domain.AssetTypeVehicle {
		return domain.VehicleProfile{}, fmt.Errorf("%w: asset is not a vehicle", domain.ErrInvalid)
	}
	return s.repo.GetVehicleProfile(ctx, assetID)
}

func (s *Service) CreateInsuranceRecord(ctx context.Context, user domain.User, r domain.VehicleInsuranceRecord) (domain.VehicleInsuranceRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, r.AssetID); err != nil {
		return domain.VehicleInsuranceRecord{}, err
	}
	if r.ExpiryDate.IsZero() {
		return domain.VehicleInsuranceRecord{}, fmt.Errorf("%w: insurance expiry date is required", domain.ErrInvalid)
	}
	return s.repo.CreateInsuranceRecord(ctx, r)
}

func (s *Service) ListInsuranceRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleInsuranceRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.repo.ListInsuranceRecords(ctx, assetID)
}

func (s *Service) CreateLicenseRecord(ctx context.Context, user domain.User, r domain.VehicleLicenseRecord) (domain.VehicleLicenseRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, r.AssetID); err != nil {
		return domain.VehicleLicenseRecord{}, err
	}
	if r.ExpiryDate.IsZero() {
		return domain.VehicleLicenseRecord{}, fmt.Errorf("%w: license expiry date is required", domain.ErrInvalid)
	}
	return s.repo.CreateLicenseRecord(ctx, r)
}

func (s *Service) ListLicenseRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleLicenseRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.repo.ListLicenseRecords(ctx, assetID)
}

func (s *Service) CreateEmissionRecord(ctx context.Context, user domain.User, r domain.VehicleEmissionRecord) (domain.VehicleEmissionRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, r.AssetID); err != nil {
		return domain.VehicleEmissionRecord{}, err
	}
	if r.ExpiryDate.IsZero() {
		return domain.VehicleEmissionRecord{}, fmt.Errorf("%w: emission/inspection expiry date is required", domain.ErrInvalid)
	}
	return s.repo.CreateEmissionRecord(ctx, r)
}

func (s *Service) ListEmissionRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleEmissionRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.repo.ListEmissionRecords(ctx, assetID)
}

func (s *Service) CreateServiceRecord(ctx context.Context, user domain.User, r domain.ServiceRecord) (domain.ServiceRecord, error) {
	if _, err := s.assets.GetAsset(ctx, user, r.AssetID); err != nil {
		return domain.ServiceRecord{}, err
	}
	if r.ServiceDate.IsZero() {
		return domain.ServiceRecord{}, fmt.Errorf("%w: service date is required", domain.ErrInvalid)
	}
	if r.ServiceType == "" {
		r.ServiceType = "service"
	}
	if r.ServiceType != "service" && r.ServiceType != "repair" {
		return domain.ServiceRecord{}, fmt.Errorf("%w: service type must be service or repair", domain.ErrInvalid)
	}
	r.CreatedBy = user.ID
	return s.repo.CreateServiceRecord(ctx, r)
}

func (s *Service) ListServiceRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.ServiceRecord, error) {
	if _, err := s.assets.GetAsset(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.repo.ListServiceRecords(ctx, assetID)
}

func (s *Service) CreateFuelLog(ctx context.Context, user domain.User, r domain.FuelLog) (domain.FuelLog, error) {
	if err := s.ensureVehicleAccess(ctx, user, r.AssetID); err != nil {
		return domain.FuelLog{}, err
	}
	if r.FuelDate.IsZero() {
		return domain.FuelLog{}, fmt.Errorf("%w: fuel date is required", domain.ErrInvalid)
	}
	if r.Quantity <= 0 || r.Cost < 0 {
		return domain.FuelLog{}, fmt.Errorf("%w: fuel quantity and cost must be valid", domain.ErrInvalid)
	}
	r.CreatedBy = user.ID
	return s.repo.CreateFuelLog(ctx, r)
}

func (s *Service) ListFuelLogs(ctx context.Context, user domain.User, assetID int64) ([]domain.FuelLog, error) {
	if err := s.ensureVehicleAccess(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.repo.ListFuelLogs(ctx, assetID)
}
