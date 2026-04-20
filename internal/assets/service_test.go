package assets

import (
	"context"
	"testing"

	"ams-ai/internal/domain"
)

type fakeAssetRepo struct {
	existing domain.Asset
	created  domain.Asset
	updated  domain.Asset
	filters  domain.AssetFilters
}

func (r *fakeAssetRepo) CreateAsset(ctx context.Context, asset domain.Asset) (domain.Asset, error) {
	r.created = asset
	asset.ID = 1
	return asset, nil
}

func (r *fakeAssetRepo) UpdateAsset(ctx context.Context, asset domain.Asset) (domain.Asset, error) {
	r.updated = asset
	return asset, nil
}

func (r *fakeAssetRepo) ArchiveAsset(ctx context.Context, id, userID int64) error {
	return nil
}

func (r *fakeAssetRepo) RestoreAsset(ctx context.Context, id, userID int64) error {
	return nil
}

func (r *fakeAssetRepo) GetAsset(ctx context.Context, id int64) (domain.Asset, error) {
	return r.existing, nil
}

func (r *fakeAssetRepo) ListAssets(ctx context.Context, filters domain.AssetFilters) ([]domain.Asset, error) {
	r.filters = filters
	return []domain.Asset{{ID: 1}}, nil
}

func TestCreateAssetAssignsNonAdminToSelf(t *testing.T) {
	repo := &fakeAssetRepo{}
	service := NewService(repo, 45)
	user := domain.User{ID: 7, Role: domain.RoleUser}

	created, err := service.CreateAsset(context.Background(), user, domain.Asset{CategoryID: 2, Name: " Laptop "})
	if err != nil {
		t.Fatalf("CreateAsset() error = %v", err)
	}
	if created.Type != domain.AssetTypeGeneral {
		t.Fatalf("type = %q, want %q", created.Type, domain.AssetTypeGeneral)
	}
	if created.AssignedUserID == nil || *created.AssignedUserID != user.ID {
		t.Fatalf("assigned user = %v, want %d", created.AssignedUserID, user.ID)
	}
	if repo.created.CreatedBy != user.ID {
		t.Fatalf("created by = %d, want %d", repo.created.CreatedBy, user.ID)
	}
}

func TestUpdateAssetNonAdminCannotReassignAsset(t *testing.T) {
	assignedID := int64(7)
	requestedID := int64(99)
	repo := &fakeAssetRepo{existing: domain.Asset{ID: 3, CreatedBy: 7, AssignedUserID: &assignedID}}
	service := NewService(repo, 45)

	updated, err := service.UpdateAsset(context.Background(), domain.User{ID: 7, Role: domain.RoleUser}, domain.Asset{
		ID:             3,
		CategoryID:     2,
		Name:           "Laptop",
		AssignedUserID: &requestedID,
	})
	if err != nil {
		t.Fatalf("UpdateAsset() error = %v", err)
	}
	if updated.AssignedUserID == nil || *updated.AssignedUserID != assignedID {
		t.Fatalf("assigned user = %v, want %d", updated.AssignedUserID, assignedID)
	}
}

func TestListAssetsInjectsCallerAndReminderWindow(t *testing.T) {
	repo := &fakeAssetRepo{}
	service := NewService(repo, 60)
	user := domain.User{ID: 8, Role: domain.RoleUser}

	if _, err := service.ListAssets(context.Background(), user, domain.AssetFilters{IncludeArchived: true}); err != nil {
		t.Fatalf("ListAssets() error = %v", err)
	}
	if repo.filters.CurrentUserID != user.ID {
		t.Fatalf("CurrentUserID = %d, want %d", repo.filters.CurrentUserID, user.ID)
	}
	if repo.filters.CurrentUserRole != user.Role {
		t.Fatalf("CurrentUserRole = %q, want %q", repo.filters.CurrentUserRole, user.Role)
	}
	if repo.filters.ReminderWindowDay != 60 {
		t.Fatalf("ReminderWindowDay = %d, want 60", repo.filters.ReminderWindowDay)
	}
	if !repo.filters.IncludeArchived {
		t.Fatal("IncludeArchived should pass through")
	}
}
