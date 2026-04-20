package httpapi

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ams-ai/internal/config"
)

func TestHealth(t *testing.T) {
	handler := New(config.Config{}, Dependencies{}, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestProtectedRouteRequiresBearerToken(t *testing.T) {
	handler := New(config.Config{}, Dependencies{}, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
