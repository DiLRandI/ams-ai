package categories

import (
	"context"
	"net/http"

	"ams-ai/internal/auth"
	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type HTTPService interface {
	ListCategories(ctx context.Context) ([]domain.Category, error)
	CreateCategory(ctx context.Context, user domain.User, name, description string) (domain.Category, error)
	UpdateCategory(ctx context.Context, user domain.User, id int64, name, description string) (domain.Category, error)
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(mux *http.ServeMux, h *Handler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/categories", requireAuth(h.listCategories))
	mux.HandleFunc("POST /api/categories", requireAuth(h.createCategory))
	mux.HandleFunc("PUT /api/categories/{id}", requireAuth(h.updateCategory))
}

type categoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, categories)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	c, err := h.service.CreateCategory(r.Context(), auth.CurrentUser(r), req.Name, req.Description)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, c)
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	var req categoryRequest
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	c, err := h.service.UpdateCategory(r.Context(), auth.CurrentUser(r), id, req.Name, req.Description)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}
