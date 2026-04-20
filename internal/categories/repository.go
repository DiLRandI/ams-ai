package categories

import (
	"context"

	"ams-ai/internal/domain"
)

type Repository interface {
	ListCategories(ctx context.Context) ([]domain.Category, error)
	CreateCategory(ctx context.Context, name, description string) (domain.Category, error)
	UpdateCategory(ctx context.Context, id int64, name, description string) (domain.Category, error)
}
