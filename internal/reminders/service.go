package reminders

import (
	"context"

	"ams-ai/internal/domain"
)

type Service struct {
	repo       Repository
	windowDays int
}

func NewService(repo Repository, windowDays int) *Service {
	return &Service{repo: repo, windowDays: windowDays}
}

func (s *Service) Dashboard(ctx context.Context, user domain.User) (domain.Dashboard, error) {
	if err := s.repo.RegenerateReminders(ctx, s.windowDays); err != nil {
		return domain.Dashboard{}, err
	}
	return s.repo.Dashboard(ctx, user, s.windowDays)
}

func (s *Service) ListReminders(ctx context.Context, user domain.User) ([]domain.Reminder, error) {
	if err := s.repo.RegenerateReminders(ctx, s.windowDays); err != nil {
		return nil, err
	}
	return s.repo.ListReminders(ctx, user, 100)
}

func (s *Service) RegenerateReminders(ctx context.Context, user domain.User) error {
	if user.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	return s.repo.RegenerateReminders(ctx, s.windowDays)
}
