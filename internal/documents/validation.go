package documents

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"unicode"

	"ams-ai/internal/domain"
)

func validDocumentType(t string) bool {
	switch t {
	case "bill_invoice", "warranty", "insurance", "license_registration", "service_receipt", "manual", "other":
		return true
	default:
		return false
	}
}

func validContentType(t string) bool {
	if mediaType, _, err := mime.ParseMediaType(t); err == nil {
		t = mediaType
	}
	switch t {
	case "image/jpeg", "image/png", "application/pdf":
		return true
	default:
		return false
	}
}

func safeFileName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == "/" || name == "" {
		return "upload"
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "._-")
	if out == "" {
		return "upload"
	}
	return out
}

func formatByteLimit(limit int64) string {
	if limit%(1024*1024) == 0 {
		return fmt.Sprintf("%d MB", limit/(1024*1024))
	}
	return fmt.Sprintf("%d bytes", limit)
}

func validateUpload(in UploadInput, maxUpload int64) error {
	if strings.TrimSpace(in.Title) == "" && strings.TrimSpace(in.FileName) == "" {
		return fmt.Errorf("%w: file is required", domain.ErrInvalid)
	}
	if !validDocumentType(in.Type) {
		return fmt.Errorf("%w: unsupported document type", domain.ErrInvalid)
	}
	if !validContentType(in.ContentType) {
		return fmt.Errorf("%w: unsupported file type; only JPG, PNG, and PDF files are supported", domain.ErrInvalid)
	}
	if in.SizeBytes <= 0 || in.SizeBytes > maxUpload {
		return fmt.Errorf("%w: file size must be greater than 0 bytes and no larger than %s", domain.ErrInvalid, formatByteLimit(maxUpload))
	}
	return nil
}
