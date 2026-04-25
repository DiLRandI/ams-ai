package documents

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"ams-ai/internal/auth"
	"ams-ai/internal/domain"
	"ams-ai/internal/platform/httpx"
)

type HTTPService interface {
	ListDocuments(ctx context.Context, user domain.User, assetID int64) ([]domain.AssetDocument, error)
	UploadDocument(ctx context.Context, input UploadInput) (domain.AssetDocument, error)
	ReplaceDocument(ctx context.Context, id int64, input UploadInput) (domain.AssetDocument, error)
	DownloadDocument(ctx context.Context, user domain.User, id int64) (domain.AssetDocument, io.ReadCloser, string, int64, error)
	DeleteDocument(ctx context.Context, user domain.User, id int64) error
}

type Handler struct {
	service   HTTPService
	maxUpload int64
}

func NewHandler(service HTTPService, maxUpload int64) *Handler {
	return &Handler{service: service, maxUpload: maxUpload}
}

func RegisterRoutes(mux *http.ServeMux, h *Handler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/assets/{id}/documents", requireAuth(h.listDocuments))
	mux.HandleFunc("POST /api/assets/{id}/documents", requireAuth(h.uploadDocument))
	mux.HandleFunc("GET /api/documents/{id}/download", requireAuth(h.downloadDocument))
	mux.HandleFunc("PUT /api/documents/{id}", requireAuth(h.replaceDocument))
	mux.HandleFunc("DELETE /api/documents/{id}", requireAuth(h.deleteDocument))
}

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	docs, err := h.service.ListDocuments(r.Context(), auth.CurrentUser(r), assetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, docs)
}

func (h *Handler) uploadDocument(w http.ResponseWriter, r *http.Request) {
	assetID, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUpload+1024*1024)
	if err := r.ParseMultipartForm(h.maxUpload); err != nil {
		httpx.WriteError(w, fmt.Errorf("%w: invalid multipart upload", domain.ErrInvalid))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, fmt.Errorf("%w: file is required", domain.ErrInvalid))
		return
	}
	defer file.Close()
	contentType, reader, err := detectUploadContentType(file, header.Filename)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	doc, err := h.service.UploadDocument(r.Context(), UploadInput{
		AssetID:     assetID,
		Title:       r.FormValue("title"),
		Type:        r.FormValue("type"),
		Notes:       r.FormValue("notes"),
		FileName:    header.Filename,
		ContentType: contentType,
		SizeBytes:   header.Size,
		Reader:      reader,
		User:        auth.CurrentUser(r),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, doc)
}

func (h *Handler) replaceDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUpload+1024*1024)
	if err := r.ParseMultipartForm(h.maxUpload); err != nil {
		httpx.WriteError(w, fmt.Errorf("%w: invalid multipart upload", domain.ErrInvalid))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, fmt.Errorf("%w: file is required", domain.ErrInvalid))
		return
	}
	defer file.Close()
	contentType, reader, err := detectUploadContentType(file, header.Filename)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	doc, err := h.service.ReplaceDocument(r.Context(), id, UploadInput{
		Title:       r.FormValue("title"),
		Type:        r.FormValue("type"),
		Notes:       r.FormValue("notes"),
		FileName:    header.Filename,
		ContentType: contentType,
		SizeBytes:   header.Size,
		Reader:      reader,
		User:        auth.CurrentUser(r),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, doc)
}

func (h *Handler) downloadDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	doc, reader, contentType, size, err := h.service.DownloadDocument(r.Context(), auth.CurrentUser(r), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	defer reader.Close()
	w.Header().Set("Content-Type", firstNonEmpty(contentType, doc.ContentType))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(doc.FileName, `"`, "")))
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	_, _ = io.Copy(w, reader)
}

func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.PathID(w, r)
	if !ok {
		return
	}
	if err := h.service.DeleteDocument(r.Context(), auth.CurrentUser(r), id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "application/octet-stream"
}

func detectUploadContentType(file io.Reader, filename string) (string, io.Reader, error) {
	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", nil, fmt.Errorf("%w: could not read uploaded file", domain.ErrInvalid)
	}
	sniff = sniff[:n]
	contentType := http.DetectContentType(sniff)
	if contentType == "application/octet-stream" {
		contentType = mime.TypeByExtension(strings.ToLower(filepathExt(filename)))
	}
	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
		contentType = mediaType
	}
	return contentType, io.MultiReader(bytes.NewReader(sniff), file), nil
}

func filepathExt(name string) string {
	idx := strings.LastIndex(name, ".")
	if idx < 0 {
		return ""
	}
	return name[idx:]
}
