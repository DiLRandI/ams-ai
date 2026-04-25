package auth

import (
	"context"

	"ams-ai/internal/domain"
)

type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByID(ctx context.Context, id int64) (domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	UpdateUserProfile(ctx context.Context, id int64, fullName string, passwordHash *string) (domain.User, error)
}
