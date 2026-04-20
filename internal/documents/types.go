package documents

import (
	"io"

	"ams-ai/internal/domain"
)

type UploadInput struct {
	AssetID     int64
	Title       string
	Type        string
	Notes       string
	FileName    string
	ContentType string
	SizeBytes   int64
	Reader      io.Reader
	User        domain.User
}
