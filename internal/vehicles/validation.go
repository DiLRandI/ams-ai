package vehicles

import (
	"context"
	"fmt"
	"strings"

	"ams-ai/internal/domain"
)

func (s *Service) ensureVehicleAccess(ctx context.Context, user domain.User, assetID int64) error {
	asset, err := s.assets.GetAsset(ctx, user, assetID)
	if err != nil {
		return err
	}
	if asset.Type != domain.AssetTypeVehicle {
		return fmt.Errorf("%w: operation requires a vehicle asset", domain.ErrInvalid)
	}
	return nil
}

func validateVehicleProfile(asset domain.Asset, p domain.VehicleProfile) error {
	if asset.Type != domain.AssetTypeVehicle {
		return fmt.Errorf("%w: vehicle profile requires a vehicle asset", domain.ErrInvalid)
	}
	if strings.TrimSpace(p.RegistrationNumber) == "" {
		return fmt.Errorf("%w: registration number is required", domain.ErrInvalid)
	}
	return nil
}
