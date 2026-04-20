package auth

import (
	"context"
	"net/http"

	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type contextKey string

const userKey contextKey = "user"

type TokenService interface {
	UserFromToken(ctx context.Context, token string) (domain.User, error)
}

func RequireAuth(service TokenService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := BearerToken(r.Header.Get("Authorization"))
			if token == "" {
				httpx.WriteError(w, domain.ErrUnauthorized)
				return
			}
			user, err := service.UserFromToken(r.Context(), token)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			next(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
		}
	}
}

func CurrentUser(r *http.Request) domain.User {
	user, _ := r.Context().Value(userKey).(domain.User)
	return user
}
