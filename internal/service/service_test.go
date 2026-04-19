package service

import "testing"

func TestSafeFileName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "base name only", in: "../../invoice final.pdf", want: "invoice_final.pdf"},
		{name: "keeps safe ascii", in: "warranty-2026_04.png", want: "warranty-2026_04.png"},
		{name: "fallback for punctuation only", in: "...", want: "upload"},
		{name: "fallback for empty", in: "", want: "upload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeFileName(tt.in); got != tt.want {
				t.Fatalf("safeFileName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidContentType(t *testing.T) {
	allowed := []string{
		"image/jpeg",
		"image/png",
		"application/pdf",
		"application/pdf; charset=binary",
	}
	for _, contentType := range allowed {
		if !validContentType(contentType) {
			t.Fatalf("validContentType(%q) = false, want true", contentType)
		}
	}

	rejected := []string{
		"text/plain",
		"application/octet-stream",
		"image/gif",
	}
	for _, contentType := range rejected {
		if validContentType(contentType) {
			t.Fatalf("validContentType(%q) = true, want false", contentType)
		}
	}
}
