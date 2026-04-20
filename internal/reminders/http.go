package reminders

import (
	"context"
	"net/http"

	"ams-ai/internal/auth"
	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type HTTPService interface {
	Dashboard(ctx context.Context, user domain.User) (domain.Dashboard, error)
	ListReminders(ctx context.Context, user domain.User) ([]domain.Reminder, error)
	RegenerateReminders(ctx context.Context, user domain.User) error
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(mux *http.ServeMux, h *Handler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/dashboard", requireAuth(h.dashboard))
	mux.HandleFunc("GET /api/reminders", requireAuth(h.listReminders))
	mux.HandleFunc("POST /api/reminders/regenerate", requireAuth(h.regenerateReminders))
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	out, err := h.service.Dashboard(r.Context(), auth.CurrentUser(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) listReminders(w http.ResponseWriter, r *http.Request) {
	out, err := h.service.ListReminders(r.Context(), auth.CurrentUser(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) regenerateReminders(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RegenerateReminders(r.Context(), auth.CurrentUser(r)); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
