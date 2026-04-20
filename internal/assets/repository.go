package assets

import (
	"context"

	"ams-ai/internal/domain"
)

type Repository interface {
	CreateAsset(ctx context.Context, asset domain.Asset) (domain.Asset, error)
	UpdateAsset(ctx context.Context, asset domain.Asset) (domain.Asset, error)
	ArchiveAsset(ctx context.Context, id, userID int64) error
	RestoreAsset(ctx context.Context, id, userID int64) error
	GetAsset(ctx context.Context, id int64) (domain.Asset, error)
	ListAssets(ctx context.Context, filters domain.AssetFilters) ([]domain.Asset, error)
}
