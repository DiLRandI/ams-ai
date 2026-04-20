package categories

import (
	"context"
	"fmt"
	"strings"

	"ams-ai/internal/domain"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, user domain.User, name, description string) (domain.Category, error) {
	if user.Role != domain.RoleAdmin {
		return domain.Category{}, domain.ErrForbidden
	}
	if strings.TrimSpace(name) == "" {
		return domain.Category{}, fmt.Errorf("%w: category name is required", domain.ErrInvalid)
	}
	return s.repo.CreateCategory(ctx, name, description)
}

func (s *Service) UpdateCategory(ctx context.Context, user domain.User, id int64, name, description string) (domain.Category, error) {
	if user.Role != domain.RoleAdmin {
		return domain.Category{}, domain.ErrForbidden
	}
	if strings.TrimSpace(name) == "" {
		return domain.Category{}, fmt.Errorf("%w: category name is required", domain.ErrInvalid)
	}
	return s.repo.UpdateCategory(ctx, id, name, description)
}
