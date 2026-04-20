package documents

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Put(ctx context.Context, key string, content io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, string, int64, error)
	Delete(ctx context.Context, key string) error
}
