package documents

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"ams-ai/internal/domain"
)

type fakeDocumentRepo struct {
	createErr error
	doc       domain.AssetDocument
}

func (r *fakeDocumentRepo) CreateDocument(ctx context.Context, document domain.AssetDocument) (domain.AssetDocument, error) {
	r.doc = document
	if r.createErr != nil {
		return domain.AssetDocument{}, r.createErr
	}
	document.ID = 1
	return document, nil
}

func (r *fakeDocumentRepo) GetDocument(ctx context.Context, id int64) (domain.AssetDocument, error) {
	return r.doc, nil
}

func (r *fakeDocumentRepo) ListDocuments(ctx context.Context, assetID int64) ([]domain.AssetDocument, error) {
	return []domain.AssetDocument{r.doc}, nil
}

func (r *fakeDocumentRepo) DeleteDocument(ctx context.Context, id int64) error {
	return nil
}

type fakeAssetReader struct{}

func (fakeAssetReader) GetAsset(ctx context.Context, user domain.User, id int64) (domain.Asset, error) {
	return domain.Asset{ID: id, CreatedBy: user.ID}, nil
}

type fakeObjects struct {
	putKey    string
	deleted   string
	content   string
	sizeBytes int64
}

func (o *fakeObjects) Put(ctx context.Context, key string, content io.Reader, size int64, contentType string) error {
	o.putKey = key
	o.sizeBytes = size
	return nil
}

func (o *fakeObjects) Get(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
	return io.NopCloser(strings.NewReader(o.content)), "application/pdf", int64(len(o.content)), nil
}

func (o *fakeObjects) Delete(ctx context.Context, key string) error {
	o.deleted = key
	return nil
}

type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time {
	return c.t
}

func TestUploadDocumentDeletesObjectWhenRepositoryFails(t *testing.T) {
	repoErr := errors.New("insert failed")
	repo := &fakeDocumentRepo{createErr: repoErr}
	objects := &fakeObjects{}
	service := NewService(repo, fakeAssetReader{}, objects, 1024, fixedClock{t: time.Unix(100, 0)})

	_, err := service.UploadDocument(context.Background(), UploadInput{
		AssetID:     9,
		Title:       "",
		Type:        "warranty",
		FileName:    "warranty final.pdf",
		ContentType: "application/pdf",
		SizeBytes:   10,
		Reader:      strings.NewReader("0123456789"),
		User:        domain.User{ID: 3, Role: domain.RoleUser},
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("UploadDocument() error = %v, want %v", err, repoErr)
	}
	if objects.putKey == "" {
		t.Fatal("object was not written before repository insert")
	}
	if objects.deleted != objects.putKey {
		t.Fatalf("deleted key = %q, want put key %q", objects.deleted, objects.putKey)
	}
	if repo.doc.Title != "warranty final.pdf" {
		t.Fatalf("document title = %q, want filename fallback", repo.doc.Title)
	}
	if repo.doc.FileName != "warranty_final.pdf" {
		t.Fatalf("document filename = %q, want sanitized filename", repo.doc.FileName)
	}
}
