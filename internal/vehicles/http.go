package vehicles

import (
	"context"
	"net/http"

	"ams-ai/internal/auth"
	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type HTTPService interface {
	GetVehicleProfile(ctx context.Context, user domain.User, assetID int64) (domain.VehicleProfile, error)
	UpsertVehicleProfile(ctx context.Context, user domain.User, profile domain.VehicleProfile) (domain.VehicleProfile, error)
	ListInsuranceRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleInsuranceRecord, error)
	CreateInsuranceRecord(ctx context.Context, user domain.User, record domain.VehicleInsuranceRecord) (domain.VehicleInsuranceRecord, error)
	ListLicenseRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleLicenseRecord, error)
	CreateLicenseRecord(ctx context.Context, user domain.User, record domain.VehicleLicenseRecord) (domain.VehicleLicenseRecord, error)
	ListEmissionRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleEmissionRecord, error)
	CreateEmissionRecord(ctx context.Context, user domain.User, record domain.VehicleEmissionRecord) (domain.VehicleEmissionRecord, error)
	ListServiceRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.ServiceRecord, error)
	CreateServiceRecord(ctx context.Context, user domain.User, record domain.ServiceRecord) (domain.ServiceRecord, error)
	ListFuelLogs(ctx context.Context, user domain.User, assetID int64) ([]domain.FuelLog, error)
	CreateFuelLog(ctx context.Context, user domain.User, record domain.FuelLog) (domain.FuelLog, error)
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(mux *http.ServeMux, h *Handler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/assets/{id}/vehicle", requireAuth(h.getVehicleProfile))
	mux.HandleFunc("PUT /api/assets/{id}/vehicle", requireAuth(h.upsertVehicleProfile))
	mux.HandleFunc("GET /api/assets/{id}/vehicle/insurance", requireAuth(h.listInsurance))
	mux.HandleFunc("POST /api/assets/{id}/vehicle/insurance", requireAuth(h.createInsurance))
	mux.HandleFunc("GET /api/assets/{id}/vehicle/licenses", requireAuth(h.listLicenses))
	mux.HandleFunc("POST /api/assets/{id}/vehicle/licenses", requireAuth(h.createLicense))
	mux.HandleFunc("GET /api/assets/{id}/vehicle/emissions", requireAuth(h.listEmissions))
	mux.HandleFunc("POST /api/assets/{id}/vehicle/emissions", requireAuth(h.createEmission))
	mux.HandleFunc("GET /api/assets/{id}/services", requireAuth(h.listServices))
	mux.HandleFunc("POST /api/assets/{id}/services", requireAuth(h.createService))
	mux.HandleFunc("GET /api/assets/{id}/fuel-logs", requireAuth(h.listFuelLogs))
	mux.HandleFunc("POST /api/assets/{id}/fuel-logs", requireAuth(h.createFuelLog))
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
	nextServiceDate, err := httpx.ParseOptionalDate(r.NextServiceDate)
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
	start, err := httpx.ParseOptionalDate(r.StartDate)
	if err != nil {
		return domain.VehicleInsuranceRecord{}, err
	}
	expiry, err := httpx.ParseRequiredDate(r.ExpiryDate, "expiryDate")
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
	issue, err := httpx.ParseOptionalDate(r.IssueDate)
	if err != nil {
		return domain.VehicleLicenseRecord{}, err
	}
	expiry, err := httpx.ParseRequiredDate(r.ExpiryDate, "expiryDate")
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
	issue, err := httpx.ParseOptionalDate(r.IssueDate)
	if err != nil {
		return domain.VehicleEmissionRecord{}, err
	}
	expiry, err := httpx.ParseRequiredDate(r.ExpiryDate, "expiryDate")
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
	serviceDate, err := httpx.ParseRequiredDate(r.ServiceDate, "serviceDate")
	if err != nil {
		return domain.ServiceRecord{}, err
	}
	nextDate, err := httpx.ParseOptionalDate(r.NextServiceDate)
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
	fuelDate, err := httpx.ParseRequiredDate(r.FuelDate, "fuelDate")
	if err != nil {
		return domain.FuelLog{}, err
	}
	return domain.FuelLog{AssetID: assetID, FuelDate: fuelDate, FuelType: r.FuelType, Quantity: r.Quantity, Cost: r.Cost, Odometer: r.Odometer, Notes: r.Notes}, nil
}

func (h *Handler) getVehicleProfile(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	profile, err := h.service.GetVehicleProfile(r.Context(), auth.CurrentUser(r), assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) upsertVehicleProfile(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	var req vehicleProfileRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	profile, err := req.toDomain(assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.service.UpsertVehicleProfile(r.Context(), auth.CurrentUser(r), profile)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) listInsurance(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	out, err := h.service.ListInsuranceRecords(r.Context(), auth.CurrentUser(r), assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) createInsurance(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	var req insuranceRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.service.CreateInsuranceRecord(r.Context(), auth.CurrentUser(r), rec)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) listLicenses(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	out, err := h.service.ListLicenseRecords(r.Context(), auth.CurrentUser(r), assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) createLicense(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	var req licenseRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.service.CreateLicenseRecord(r.Context(), auth.CurrentUser(r), rec)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) listEmissions(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	out, err := h.service.ListEmissionRecords(r.Context(), auth.CurrentUser(r), assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) createEmission(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	var req emissionRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.service.CreateEmissionRecord(r.Context(), auth.CurrentUser(r), rec)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) listServices(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	out, err := h.service.ListServiceRecords(r.Context(), auth.CurrentUser(r), assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) createService(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	var req serviceRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.service.CreateServiceRecord(r.Context(), auth.CurrentUser(r), rec)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) listFuelLogs(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	out, err := h.service.ListFuelLogs(r.Context(), auth.CurrentUser(r), assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) createFuelLog(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	var req fuelRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.service.CreateFuelLog(r.Context(), auth.CurrentUser(r), rec)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}
