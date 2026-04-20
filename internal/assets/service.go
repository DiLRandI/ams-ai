package assets

import (
	"context"

	"ams-ai/internal/domain"
)

type Service struct {
	repo              Repository
	reminderWindowDay int
}

func NewService(repo Repository, reminderWindowDays int) *Service {
	return &Service{repo: repo, reminderWindowDay: reminderWindowDays}
}

func (s *Service) CreateAsset(ctx context.Context, user domain.User, a domain.Asset) (domain.Asset, error) {
	if err := validateAsset(&a); err != nil {
		return domain.Asset{}, err
	}
	a.CreatedBy = user.ID
	a.UpdatedBy = &user.ID
	if user.Role != domain.RoleAdmin {
		a.AssignedUserID = &user.ID
	}
	return s.repo.CreateAsset(ctx, a)
}

func (s *Service) UpdateAsset(ctx context.Context, user domain.User, a domain.Asset) (domain.Asset, error) {
	existing, err := s.repo.GetAsset(ctx, a.ID)
	if err != nil {
		return domain.Asset{}, err
	}
	if !domain.AssetAccessAllowed(user, existing) {
		return domain.Asset{}, domain.ErrForbidden
	}
	if err := validateAsset(&a); err != nil {
		return domain.Asset{}, err
	}
	if user.Role != domain.RoleAdmin {
		a.AssignedUserID = existing.AssignedUserID
		if a.AssignedUserID == nil {
			a.AssignedUserID = &user.ID
		}
	}
	a.UpdatedBy = &user.ID
	return s.repo.UpdateAsset(ctx, a)
}

func (s *Service) ArchiveAsset(ctx context.Context, user domain.User, id int64) error {
	asset, err := s.repo.GetAsset(ctx, id)
	if err != nil {
		return err
	}
	if !domain.AssetAccessAllowed(user, asset) {
		return domain.ErrForbidden
	}
	return s.repo.ArchiveAsset(ctx, id, user.ID)
}

func (s *Service) RestoreAsset(ctx context.Context, user domain.User, id int64) error {
	asset, err := s.repo.GetAsset(ctx, id)
	if err != nil {
		return err
	}
	if !domain.AssetAccessAllowed(user, asset) {
		return domain.ErrForbidden
	}
	return s.repo.RestoreAsset(ctx, id, user.ID)
}

func (s *Service) GetAsset(ctx context.Context, user domain.User, id int64) (domain.Asset, error) {
	asset, err := s.repo.GetAsset(ctx, id)
	if err != nil {
		return domain.Asset{}, err
	}
	if !domain.AssetAccessAllowed(user, asset) {
		return domain.Asset{}, domain.ErrForbidden
	}
	return asset, nil
}

func (s *Service) ListAssets(ctx context.Context, user domain.User, f domain.AssetFilters) ([]domain.Asset, error) {
	f.CurrentUserID = user.ID
	f.CurrentUserRole = user.Role
	f.ReminderWindowDay = s.reminderWindowDay
	return s.repo.ListAssets(ctx, f)
}
