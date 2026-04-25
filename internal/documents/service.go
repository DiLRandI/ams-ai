package documents

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"ams-ai/internal/domain"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type Service struct {
	repo      Repository
	assets    AssetReader
	objects   ObjectStorage
	clock     Clock
	maxUpload int64
}

func NewService(repo Repository, assets AssetReader, objects ObjectStorage, maxUpload int64, clock Clock) *Service {
	if clock == nil {
		clock = realClock{}
	}
	return &Service{repo: repo, assets: assets, objects: objects, maxUpload: maxUpload, clock: clock}
}

func (s *Service) UploadDocument(ctx context.Context, in UploadInput) (domain.AssetDocument, error) {
	if _, err := s.assets.GetAsset(ctx, in.User, in.AssetID); err != nil {
		return domain.AssetDocument{}, err
	}
	if strings.TrimSpace(in.Title) == "" {
		in.Title = in.FileName
	}
	if err := validateUpload(in, s.maxUpload); err != nil {
		return domain.AssetDocument{}, err
	}
	key := fmt.Sprintf("assets/%d/documents/%d/%s", in.AssetID, s.clock.Now().UnixNano(), safeFileName(in.FileName))
	if err := s.objects.Put(ctx, key, in.Reader, in.SizeBytes, in.ContentType); err != nil {
		return domain.AssetDocument{}, err
	}
	doc, err := s.repo.CreateDocument(ctx, domain.AssetDocument{
		AssetID:     in.AssetID,
		Title:       strings.TrimSpace(in.Title),
		Type:        in.Type,
		Notes:       strings.TrimSpace(in.Notes),
		FileName:    safeFileName(in.FileName),
		ContentType: in.ContentType,
		SizeBytes:   in.SizeBytes,
		ObjectKey:   key,
		UploadedBy:  in.User.ID,
	})
	if err != nil {
		_ = s.objects.Delete(ctx, key)
		return domain.AssetDocument{}, err
	}
	return doc, nil
}

func (s *Service) ListDocuments(ctx context.Context, user domain.User, assetID int64) ([]domain.AssetDocument, error) {
	if _, err := s.assets.GetAsset(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.repo.ListDocuments(ctx, assetID)
}

func (s *Service) ReplaceDocument(ctx context.Context, id int64, in UploadInput) (domain.AssetDocument, error) {
	existing, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return domain.AssetDocument{}, err
	}
	if _, err := s.assets.GetAsset(ctx, in.User, existing.AssetID); err != nil {
		return domain.AssetDocument{}, err
	}
	if strings.TrimSpace(in.Title) == "" {
		in.Title = existing.Title
	}
	if strings.TrimSpace(in.Type) == "" {
		in.Type = existing.Type
	}
	if err := validateUpload(in, s.maxUpload); err != nil {
		return domain.AssetDocument{}, err
	}
	key := fmt.Sprintf("assets/%d/documents/%d/%s", existing.AssetID, s.clock.Now().UnixNano(), safeFileName(in.FileName))
	if err := s.objects.Put(ctx, key, in.Reader, in.SizeBytes, in.ContentType); err != nil {
		return domain.AssetDocument{}, err
	}
	replaced, err := s.repo.ReplaceDocument(ctx, id, domain.AssetDocument{
		AssetID:     existing.AssetID,
		Title:       strings.TrimSpace(in.Title),
		Type:        in.Type,
		Notes:       strings.TrimSpace(in.Notes),
		FileName:    safeFileName(in.FileName),
		ContentType: in.ContentType,
		SizeBytes:   in.SizeBytes,
		ObjectKey:   key,
		UploadedBy:  in.User.ID,
	})
	if err != nil {
		_ = s.objects.Delete(ctx, key)
		return domain.AssetDocument{}, err
	}
	_ = s.objects.Delete(ctx, existing.ObjectKey)
	return replaced, nil
}

func (s *Service) DownloadDocument(ctx context.Context, user domain.User, id int64) (domain.AssetDocument, io.ReadCloser, string, int64, error) {
	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return domain.AssetDocument{}, nil, "", 0, err
	}
	if _, err := s.assets.GetAsset(ctx, user, doc.AssetID); err != nil {
		return domain.AssetDocument{}, nil, "", 0, err
	}
	reader, contentType, size, err := s.objects.Get(ctx, doc.ObjectKey)
	if err != nil {
		return domain.AssetDocument{}, nil, "", 0, err
	}
	return doc, reader, contentType, size, nil
}

func (s *Service) DeleteDocument(ctx context.Context, user domain.User, id int64) error {
	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.assets.GetAsset(ctx, user, doc.AssetID); err != nil {
		return err
	}
	if err := s.repo.DeleteDocument(ctx, id); err != nil {
		return err
	}
	return s.objects.Delete(ctx, doc.ObjectKey)
}
