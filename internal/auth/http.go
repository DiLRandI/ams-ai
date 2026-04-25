package auth

import (
	"context"
	"net/http"

	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type HTTPService interface {
	Login(ctx context.Context, email, password string) (Token, error)
	ListUsers(ctx context.Context, user domain.User) ([]domain.User, error)
	UpdateProfile(ctx context.Context, user domain.User, fullName, password string) (domain.User, error)
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(mux *http.ServeMux, h *Handler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("GET /api/auth/me", requireAuth(h.me))
	mux.HandleFunc("PUT /api/auth/me", requireAuth(h.updateProfile))
	mux.HandleFunc("GET /api/users", requireAuth(h.listUsers))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, token)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, CurrentUser(r))
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context(), CurrentUser(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FullName string `json:"fullName"`
		Password string `json:"password"`
	}
	if !httpx.ReadJSON(w, r, &req) {
		return
	}
	user, err := h.service.UpdateProfile(r.Context(), CurrentUser(r), req.FullName, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, user)
}
