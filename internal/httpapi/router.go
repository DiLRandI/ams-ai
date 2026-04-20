package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"ams-ai/internal/assets"
	"ams-ai/internal/auth"
	"ams-ai/internal/categories"
	"ams-ai/internal/config"
	"ams-ai/internal/documents"
	"ams-ai/internal/platform/httpx"
	"ams-ai/internal/reminders"
	"ams-ai/internal/reports"
	"ams-ai/internal/vehicles"
)

type Dependencies struct {
	Auth       *auth.Service
	Categories *categories.Service
	Assets     *assets.Service
	Vehicles   *vehicles.Service
	Documents  *documents.Service
	Reminders  *reminders.Service
	Reports    *reports.Service
}

func New(cfg config.Config, deps Dependencies, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)

	requireAuth := auth.RequireAuth(deps.Auth)
	auth.RegisterRoutes(mux, auth.NewHandler(deps.Auth), requireAuth)
	categories.RegisterRoutes(mux, categories.NewHandler(deps.Categories), requireAuth)
	assets.RegisterRoutes(mux, assets.NewHandler(deps.Assets), requireAuth)
	documents.RegisterRoutes(mux, documents.NewHandler(deps.Documents, cfg.Storage.MaxUploadBytes), requireAuth)
	vehicles.RegisterRoutes(mux, vehicles.NewHandler(deps.Vehicles), requireAuth)
	reminders.RegisterRoutes(mux, reminders.NewHandler(deps.Reminders), requireAuth)
	reports.RegisterRoutes(mux, reports.NewHandler(deps.Reports), requireAuth)

	return recoverPanic(log)(cors(cfg, mux))
}

func health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func cors(cfg config.Config, next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, origin := range cfg.CORSAllowedOrigins {
		allowed[origin] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func recoverPanic(log *slog.Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", "panic", rec)
					httpx.WriteError(w, errors.New("internal server error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
