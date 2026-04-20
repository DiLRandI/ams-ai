package reminders

import (
	"context"

	"ams-ai/internal/domain"
)

type Repository interface {
	Dashboard(ctx context.Context, user domain.User, windowDays int) (domain.Dashboard, error)
	ListReminders(ctx context.Context, user domain.User, limit int) ([]domain.Reminder, error)
	RegenerateReminders(ctx context.Context, windowDays int) error
}
