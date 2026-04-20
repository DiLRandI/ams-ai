package documents

import (
	"context"

	"ams-ai/internal/domain"
)

type Repository interface {
	CreateDocument(ctx context.Context, document domain.AssetDocument) (domain.AssetDocument, error)
	GetDocument(ctx context.Context, id int64) (domain.AssetDocument, error)
	ListDocuments(ctx context.Context, assetID int64) ([]domain.AssetDocument, error)
	DeleteDocument(ctx context.Context, id int64) error
}

type AssetReader interface {
	GetAsset(ctx context.Context, user domain.User, id int64) (domain.Asset, error)
}
