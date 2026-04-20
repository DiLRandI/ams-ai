package assets

import (
	"fmt"
	"strings"

	"ams-ai/internal/domain"
)

func validateAsset(a *domain.Asset) error {
	a.Type = strings.TrimSpace(strings.ToLower(a.Type))
	if a.Type == "" {
		a.Type = domain.AssetTypeGeneral
	}
	if !domain.IsValidAssetType(a.Type) {
		return fmt.Errorf("%w: asset type must be general or vehicle", domain.ErrInvalid)
	}
	a.Status = domain.NormalizeStatus(a.Status)
	if !domain.IsValidStatus(a.Status) {
		return fmt.Errorf("%w: invalid asset status", domain.ErrInvalid)
	}
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return fmt.Errorf("%w: asset name is required", domain.ErrInvalid)
	}
	if a.CategoryID <= 0 {
		return fmt.Errorf("%w: category is required", domain.ErrInvalid)
	}
	a.Brand = strings.TrimSpace(a.Brand)
	a.Model = strings.TrimSpace(a.Model)
	a.SerialNumber = strings.TrimSpace(a.SerialNumber)
	a.Condition = strings.TrimSpace(a.Condition)
	a.Location = strings.TrimSpace(a.Location)
	a.AssignedTo = strings.TrimSpace(a.AssignedTo)
	a.Notes = strings.TrimSpace(a.Notes)
	a.WarrantyNotes = strings.TrimSpace(a.WarrantyNotes)
	if a.PurchasePrice != nil && *a.PurchasePrice < 0 {
		return fmt.Errorf("%w: purchase price cannot be negative", domain.ErrInvalid)
	}
	if a.WarrantyStartDate != nil && a.WarrantyExpiryDate != nil && a.WarrantyExpiryDate.Before(*a.WarrantyStartDate) {
		return fmt.Errorf("%w: warranty expiry cannot be before warranty start", domain.ErrInvalid)
	}
	return nil
}
