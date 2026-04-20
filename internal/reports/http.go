package reports

import (
	"context"
	"encoding/csv"
	"net/http"

	"ams-ai/internal/auth"
	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type HTTPService interface {
	AssetRows(ctx context.Context, user domain.User) ([][]string, error)
	WarrantyRows(ctx context.Context, user domain.User) ([][]string, error)
	VehicleRenewalRows(ctx context.Context, user domain.User) ([][]string, error)
	ServiceRows(ctx context.Context, user domain.User) ([][]string, error)
	FuelRows(ctx context.Context, user domain.User) ([][]string, error)
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(mux *http.ServeMux, h *Handler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/reports/assets.csv", requireAuth(h.assetReport))
	mux.HandleFunc("GET /api/reports/warranties.csv", requireAuth(h.warrantyReport))
	mux.HandleFunc("GET /api/reports/vehicle-renewals.csv", requireAuth(h.vehicleRenewalReport))
	mux.HandleFunc("GET /api/reports/service-history.csv", requireAuth(h.serviceReport))
	mux.HandleFunc("GET /api/reports/fuel-logs.csv", requireAuth(h.fuelReport))
}

func (h *Handler) assetReport(w http.ResponseWriter, r *http.Request) {
	h.writeCSV(w, r, "assets.csv", h.service.AssetRows)
}

func (h *Handler) warrantyReport(w http.ResponseWriter, r *http.Request) {
	h.writeCSV(w, r, "warranties.csv", h.service.WarrantyRows)
}

func (h *Handler) vehicleRenewalReport(w http.ResponseWriter, r *http.Request) {
	h.writeCSV(w, r, "vehicle-renewals.csv", h.service.VehicleRenewalRows)
}

func (h *Handler) serviceReport(w http.ResponseWriter, r *http.Request) {
	h.writeCSV(w, r, "service-history.csv", h.service.ServiceRows)
}

func (h *Handler) fuelReport(w http.ResponseWriter, r *http.Request) {
	h.writeCSV(w, r, "fuel-logs.csv", h.service.FuelRows)
}

func (h *Handler) writeCSV(w http.ResponseWriter, r *http.Request, filename string, rowsFunc func(context.Context, domain.User) ([][]string, error)) {
	rows, err := rowsFunc(r.Context(), auth.CurrentUser(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.StartCSV(w, filename)
	cw := csv.NewWriter(w)
	for _, row := range rows {
		_ = cw.Write(row)
	}
	cw.Flush()
}
