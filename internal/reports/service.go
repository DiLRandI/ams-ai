package reports

import (
	"context"
	"strconv"

	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type Service struct {
	assets   AssetLister
	vehicles VehicleRecords
}

func NewService(assets AssetLister, vehicles VehicleRecords) *Service {
	return &Service{assets: assets, vehicles: vehicles}
}

func (s *Service) AssetRows(ctx context.Context, user domain.User) ([][]string, error) {
	assets, err := s.assets.ListAssets(ctx, user, domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	rows := [][]string{{"Code", "Name", "Type", "Category", "Status", "Location", "Assigned To", "Warranty Expiry", "Documents"}}
	for _, a := range assets {
		rows = append(rows, []string{a.Code, a.Name, a.Type, a.CategoryName, a.Status, a.Location, a.AssignedTo, httpx.FormatDate(a.WarrantyExpiryDate), strconv.Itoa(a.DocumentCount)})
	}
	return rows, nil
}

func (s *Service) WarrantyRows(ctx context.Context, user domain.User) ([][]string, error) {
	assets, err := s.assets.ListAssets(ctx, user, domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	rows := [][]string{{"Code", "Name", "Warranty Expiry", "Warranty State", "Notes"}}
	for _, a := range assets {
		if a.WarrantyExpiryDate == nil {
			continue
		}
		rows = append(rows, []string{a.Code, a.Name, httpx.FormatDate(a.WarrantyExpiryDate), a.WarrantyState, a.WarrantyNotes})
	}
	return rows, nil
}

func (s *Service) VehicleRenewalRows(ctx context.Context, user domain.User) ([][]string, error) {
	assets, err := s.assets.ListAssets(ctx, user, domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	rows := [][]string{{"Code", "Name", "Renewal Type", "Reference", "Cost", "Expiry Date", "Notes"}}
	for _, a := range assets {
		if a.Type != domain.AssetTypeVehicle {
			continue
		}
		insurance, _ := s.vehicles.ListInsuranceRecords(ctx, user, a.ID)
		for _, rec := range insurance {
			rows = append(rows, []string{a.Code, a.Name, "insurance", rec.PolicyNumber, httpx.FormatFloat(rec.Cost), rec.ExpiryDate.Format(httpx.DateLayout), rec.Notes})
		}
		licenses, _ := s.vehicles.ListLicenseRecords(ctx, user, a.ID)
		for _, rec := range licenses {
			rows = append(rows, []string{a.Code, a.Name, "license", rec.ReferenceNumber, httpx.FormatFloat(rec.Cost), rec.ExpiryDate.Format(httpx.DateLayout), rec.Notes})
		}
		emissions, _ := s.vehicles.ListEmissionRecords(ctx, user, a.ID)
		for _, rec := range emissions {
			rows = append(rows, []string{a.Code, a.Name, "emission", rec.ReferenceNumber, httpx.FormatFloat(rec.Cost), rec.ExpiryDate.Format(httpx.DateLayout), rec.Notes})
		}
	}
	return rows, nil
}

func (s *Service) ServiceRows(ctx context.Context, user domain.User) ([][]string, error) {
	assets, err := s.assets.ListAssets(ctx, user, domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	rows := [][]string{{"Asset ID", "Date", "Type", "Vendor", "Cost", "Description"}}
	for _, a := range assets {
		records, err := s.vehicles.ListServiceRecords(ctx, user, a.ID)
		if err != nil {
			continue
		}
		for _, rec := range records {
			rows = append(rows, []string{a.Code, rec.ServiceDate.Format(httpx.DateLayout), rec.ServiceType, rec.Vendor, httpx.FormatFloat(rec.Cost), rec.Description})
		}
	}
	return rows, nil
}

func (s *Service) FuelRows(ctx context.Context, user domain.User) ([][]string, error) {
	assets, err := s.assets.ListAssets(ctx, user, domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	rows := [][]string{{"Asset ID", "Date", "Fuel Type", "Quantity", "Cost", "Odometer"}}
	for _, a := range assets {
		if a.Type != domain.AssetTypeVehicle {
			continue
		}
		records, err := s.vehicles.ListFuelLogs(ctx, user, a.ID)
		if err != nil {
			continue
		}
		for _, rec := range records {
			rows = append(rows, []string{a.Code, rec.FuelDate.Format(httpx.DateLayout), rec.FuelType, strconv.FormatFloat(rec.Quantity, 'f', 3, 64), strconv.FormatFloat(rec.Cost, 'f', 2, 64), httpx.FormatInt(rec.Odometer)})
		}
	}
	return rows, nil
}
