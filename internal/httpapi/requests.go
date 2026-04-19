package httpapi

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"ams-ai/internal/domain"
)

const dateLayout = "2006-01-02"

type categoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type assetRequest struct {
	Type               string   `json:"type"`
	CategoryID         int64    `json:"categoryId"`
	Name               string   `json:"name"`
	Brand              string   `json:"brand"`
	Model              string   `json:"model"`
	SerialNumber       string   `json:"serialNumber"`
	PurchaseDate       string   `json:"purchaseDate"`
	PurchasePrice      *float64 `json:"purchasePrice"`
	Status             string   `json:"status"`
	Condition          string   `json:"condition"`
	Location           string   `json:"location"`
	AssignedTo         string   `json:"assignedTo"`
	AssignedUserID     *int64   `json:"assignedUserId"`
	Notes              string   `json:"notes"`
	WarrantyStartDate  string   `json:"warrantyStartDate"`
	WarrantyExpiryDate string   `json:"warrantyExpiryDate"`
	WarrantyNotes      string   `json:"warrantyNotes"`
}

func (r assetRequest) toAsset() (domain.Asset, error) {
	purchaseDate, err := parseOptionalDate(r.PurchaseDate)
	if err != nil {
		return domain.Asset{}, err
	}
	warrantyStart, err := parseOptionalDate(r.WarrantyStartDate)
	if err != nil {
		return domain.Asset{}, err
	}
	warrantyExpiry, err := parseOptionalDate(r.WarrantyExpiryDate)
	if err != nil {
		return domain.Asset{}, err
	}
	return domain.Asset{
		Type:               r.Type,
		CategoryID:         r.CategoryID,
		Name:               r.Name,
		Brand:              r.Brand,
		Model:              r.Model,
		SerialNumber:       r.SerialNumber,
		PurchaseDate:       purchaseDate,
		PurchasePrice:      r.PurchasePrice,
		Status:             r.Status,
		Condition:          r.Condition,
		Location:           r.Location,
		AssignedTo:         r.AssignedTo,
		AssignedUserID:     r.AssignedUserID,
		Notes:              r.Notes,
		WarrantyStartDate:  warrantyStart,
		WarrantyExpiryDate: warrantyExpiry,
		WarrantyNotes:      r.WarrantyNotes,
	}, nil
}

type vehicleProfileRequest struct {
	RegistrationNumber string `json:"registrationNumber"`
	VehicleType        string `json:"vehicleType"`
	ChassisNumber      string `json:"chassisNumber"`
	EngineNumber       string `json:"engineNumber"`
	Odometer           *int   `json:"odometer"`
	AssignedDriver     string `json:"assignedDriver"`
	NextServiceDate    string `json:"nextServiceDate"`
	NextServiceMileage *int   `json:"nextServiceMileage"`
	Notes              string `json:"notes"`
}

func (r vehicleProfileRequest) toDomain(assetID int64) (domain.VehicleProfile, error) {
	nextServiceDate, err := parseOptionalDate(r.NextServiceDate)
	if err != nil {
		return domain.VehicleProfile{}, err
	}
	return domain.VehicleProfile{
		AssetID:            assetID,
		RegistrationNumber: r.RegistrationNumber,
		VehicleType:        r.VehicleType,
		ChassisNumber:      r.ChassisNumber,
		EngineNumber:       r.EngineNumber,
		Odometer:           r.Odometer,
		AssignedDriver:     r.AssignedDriver,
		NextServiceDate:    nextServiceDate,
		NextServiceMileage: r.NextServiceMileage,
		Notes:              r.Notes,
	}, nil
}

type insuranceRequest struct {
	Provider     string   `json:"provider"`
	PolicyNumber string   `json:"policyNumber"`
	Cost         *float64 `json:"cost"`
	StartDate    string   `json:"startDate"`
	ExpiryDate   string   `json:"expiryDate"`
	DocumentID   *int64   `json:"documentId"`
	Notes        string   `json:"notes"`
}

func (r insuranceRequest) toDomain(assetID int64) (domain.VehicleInsuranceRecord, error) {
	start, err := parseOptionalDate(r.StartDate)
	if err != nil {
		return domain.VehicleInsuranceRecord{}, err
	}
	expiry, err := parseRequiredDate(r.ExpiryDate, "expiryDate")
	if err != nil {
		return domain.VehicleInsuranceRecord{}, err
	}
	return domain.VehicleInsuranceRecord{AssetID: assetID, Provider: r.Provider, PolicyNumber: r.PolicyNumber, Cost: r.Cost, StartDate: start, ExpiryDate: expiry, DocumentID: r.DocumentID, Notes: r.Notes}, nil
}

type licenseRequest struct {
	RenewalType     string   `json:"renewalType"`
	ReferenceNumber string   `json:"referenceNumber"`
	Cost            *float64 `json:"cost"`
	IssueDate       string   `json:"issueDate"`
	ExpiryDate      string   `json:"expiryDate"`
	DocumentID      *int64   `json:"documentId"`
	Notes           string   `json:"notes"`
}

func (r licenseRequest) toDomain(assetID int64) (domain.VehicleLicenseRecord, error) {
	issue, err := parseOptionalDate(r.IssueDate)
	if err != nil {
		return domain.VehicleLicenseRecord{}, err
	}
	expiry, err := parseRequiredDate(r.ExpiryDate, "expiryDate")
	if err != nil {
		return domain.VehicleLicenseRecord{}, err
	}
	return domain.VehicleLicenseRecord{AssetID: assetID, RenewalType: r.RenewalType, ReferenceNumber: r.ReferenceNumber, Cost: r.Cost, IssueDate: issue, ExpiryDate: expiry, DocumentID: r.DocumentID, Notes: r.Notes}, nil
}

type emissionRequest struct {
	InspectionType  string   `json:"inspectionType"`
	ReferenceNumber string   `json:"referenceNumber"`
	Cost            *float64 `json:"cost"`
	IssueDate       string   `json:"issueDate"`
	ExpiryDate      string   `json:"expiryDate"`
	DocumentID      *int64   `json:"documentId"`
	Notes           string   `json:"notes"`
}

func (r emissionRequest) toDomain(assetID int64) (domain.VehicleEmissionRecord, error) {
	issue, err := parseOptionalDate(r.IssueDate)
	if err != nil {
		return domain.VehicleEmissionRecord{}, err
	}
	expiry, err := parseRequiredDate(r.ExpiryDate, "expiryDate")
	if err != nil {
		return domain.VehicleEmissionRecord{}, err
	}
	return domain.VehicleEmissionRecord{AssetID: assetID, InspectionType: r.InspectionType, ReferenceNumber: r.ReferenceNumber, Cost: r.Cost, IssueDate: issue, ExpiryDate: expiry, DocumentID: r.DocumentID, Notes: r.Notes}, nil
}

type serviceRequest struct {
	ServiceType        string   `json:"serviceType"`
	ServiceDate        string   `json:"serviceDate"`
	Cost               *float64 `json:"cost"`
	Vendor             string   `json:"vendor"`
	Description        string   `json:"description"`
	Notes              string   `json:"notes"`
	Mileage            *int     `json:"mileage"`
	NextServiceDate    string   `json:"nextServiceDate"`
	NextServiceMileage *int     `json:"nextServiceMileage"`
}

func (r serviceRequest) toDomain(assetID int64) (domain.ServiceRecord, error) {
	serviceDate, err := parseRequiredDate(r.ServiceDate, "serviceDate")
	if err != nil {
		return domain.ServiceRecord{}, err
	}
	nextDate, err := parseOptionalDate(r.NextServiceDate)
	if err != nil {
		return domain.ServiceRecord{}, err
	}
	return domain.ServiceRecord{AssetID: assetID, ServiceType: r.ServiceType, ServiceDate: serviceDate, Cost: r.Cost, Vendor: r.Vendor, Description: r.Description, Notes: r.Notes, Mileage: r.Mileage, NextServiceDate: nextDate, NextServiceMileage: r.NextServiceMileage}, nil
}

type fuelRequest struct {
	FuelDate string  `json:"fuelDate"`
	FuelType string  `json:"fuelType"`
	Quantity float64 `json:"quantity"`
	Cost     float64 `json:"cost"`
	Odometer *int    `json:"odometer"`
	Notes    string  `json:"notes"`
}

func (r fuelRequest) toDomain(assetID int64) (domain.FuelLog, error) {
	fuelDate, err := parseRequiredDate(r.FuelDate, "fuelDate")
	if err != nil {
		return domain.FuelLog{}, err
	}
	return domain.FuelLog{AssetID: assetID, FuelDate: fuelDate, FuelType: r.FuelType, Quantity: r.Quantity, Cost: r.Cost, Odometer: r.Odometer, Notes: r.Notes}, nil
}

func parseOptionalDate(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, raw)
	if err != nil {
		return nil, fmt.Errorf("%w: dates must use YYYY-MM-DD", domain.ErrInvalid)
	}
	return &t, nil
}

func parseRequiredDate(raw, field string) (time.Time, error) {
	t, err := parseOptionalDate(raw)
	if err != nil {
		return time.Time{}, err
	}
	if t == nil {
		return time.Time{}, fmt.Errorf("%w: %s is required", domain.ErrInvalid, field)
	}
	return *t, nil
}

func normalizeOptionalStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return ""
	}
	return domain.NormalizeStatus(status)
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateLayout)
}

func formatFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

func formatInt(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}
