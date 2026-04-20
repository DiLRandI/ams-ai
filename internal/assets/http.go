package assets

import (
	"context"
	"net/http"
	"strconv"

	"ams-ai/internal/auth"
	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type HTTPService interface {
	ListAssets(ctx context.Context, user domain.User, filters domain.AssetFilters) ([]domain.Asset, error)
	CreateAsset(ctx context.Context, user domain.User, asset domain.Asset) (domain.Asset, error)
	GetAsset(ctx context.Context, user domain.User, id int64) (domain.Asset, error)
	UpdateAsset(ctx context.Context, user domain.User, asset domain.Asset) (domain.Asset, error)
	ArchiveAsset(ctx context.Context, user domain.User, id int64) error
	RestoreAsset(ctx context.Context, user domain.User, id int64) error
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(mux *http.ServeMux, h *Handler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/assets", requireAuth(h.listAssets))
	mux.HandleFunc("POST /api/assets", requireAuth(h.createAsset))
	mux.HandleFunc("GET /api/assets/{id}", requireAuth(h.getAsset))
	mux.HandleFunc("PUT /api/assets/{id}", requireAuth(h.updateAsset))
	mux.HandleFunc("POST /api/assets/{id}/archive", requireAuth(h.archiveAsset))
	mux.HandleFunc("POST /api/assets/{id}/restore", requireAuth(h.restoreAsset))
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
	purchaseDate, err := httpx.ParseOptionalDate(r.PurchaseDate)
	if err != nil {
		return domain.Asset{}, err
	}
	warrantyStart, err := httpx.ParseOptionalDate(r.WarrantyStartDate)
	if err != nil {
		return domain.Asset{}, err
	}
	warrantyExpiry, err := httpx.ParseOptionalDate(r.WarrantyExpiryDate)
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

func (h *Handler) listAssets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var hasDocs *bool
	if raw := q.Get("hasDocuments"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err == nil {
			hasDocs = &v
		}
	}
	filters := domain.AssetFilters{
		Query:           q.Get("q"),
		Status:          httpx.NormalizeOptionalStatus(q.Get("status")),
		Location:        q.Get("location"),
		WarrantyState:   q.Get("warrantyState"),
		HasDocuments:    hasDocs,
		IncludeArchived: q.Get("includeArchived") == "true",
	}
	filters.CategoryID, _ = strconv.ParseInt(q.Get("categoryId"), 10, 64)
	filters.AssignedUserID, _ = strconv.ParseInt(q.Get("assignedUserId"), 10, 64)
	assets, err := h.service.ListAssets(r.Context(), auth.CurrentUser(r), filters)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, assets)
}

func (h *Handler) createAsset(w http.ResponseWriter, r *http.Request) {
	var req assetRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	asset, err := req.toAsset()
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	created, err := h.service.CreateAsset(r.Context(), auth.CurrentUser(r), asset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) getAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	asset, err := h.service.GetAsset(r.Context(), auth.CurrentUser(r), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, asset)
}

func (h *Handler) updateAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	var req assetRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	asset, err := req.toAsset()
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	asset.ID = id
	updated, err := h.service.UpdateAsset(r.Context(), auth.CurrentUser(r), asset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, updated)
}

func (h *Handler) archiveAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	if err := h.service.ArchiveAsset(r.Context(), auth.CurrentUser(r), id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) restoreAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	if err := h.service.RestoreAsset(r.Context(), auth.CurrentUser(r), id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
