package auth

import (
	"time"

	"ams-ai/internal/domain"
)

type Token struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expiresAt"`
	User      domain.User `json:"user"`
}
