package postgres

import (
	"context"
	"errors"

	"ams-ai/internal/domain"

	"github.com/jackc/pgx/v5"
)

type DocumentRepository struct {
	db *DB
}

func NewDocumentRepository(db *DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) CreateDocument(ctx context.Context, d domain.AssetDocument) (domain.AssetDocument, error) {
	row := r.db.pool.QueryRow(ctx, `
		INSERT INTO asset_documents (asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by, created_at
	`, d.AssetID, d.Title, d.Type, d.Notes, d.FileName, d.ContentType, d.SizeBytes, d.ObjectKey, d.UploadedBy)
	return scanDocument(row)
}

func (r *DocumentRepository) GetDocument(ctx context.Context, id int64) (domain.AssetDocument, error) {
	row := r.db.pool.QueryRow(ctx, `
		SELECT id, asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by, created_at
		FROM asset_documents
		WHERE id = $1
	`, id)
	return scanDocument(row)
}

func (r *DocumentRepository) ListDocuments(ctx context.Context, assetID int64) ([]domain.AssetDocument, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by, created_at
		FROM asset_documents
		WHERE asset_id = $1
		ORDER BY created_at DESC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AssetDocument{}
	for rows.Next() {
		d, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DocumentRepository) ReplaceDocument(ctx context.Context, id int64, d domain.AssetDocument) (domain.AssetDocument, error) {
	row := r.db.pool.QueryRow(ctx, `
		UPDATE asset_documents
		SET title = $2,
			type = $3,
			notes = $4,
			file_name = $5,
			content_type = $6,
			size_bytes = $7,
			object_key = $8,
			uploaded_by = $9
		WHERE id = $1
		RETURNING id, asset_id, title, type, notes, file_name, content_type, size_bytes, object_key, uploaded_by, created_at
	`, id, d.Title, d.Type, d.Notes, d.FileName, d.ContentType, d.SizeBytes, d.ObjectKey, d.UploadedBy)
	return scanDocument(row)
}

func (r *DocumentRepository) DeleteDocument(ctx context.Context, id int64) error {
	cmd, err := r.db.pool.Exec(ctx, `DELETE FROM asset_documents WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanDocument(row pgx.Row) (domain.AssetDocument, error) {
	var d domain.AssetDocument
	err := row.Scan(&d.ID, &d.AssetID, &d.Title, &d.Type, &d.Notes, &d.FileName, &d.ContentType, &d.SizeBytes, &d.ObjectKey, &d.UploadedBy, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AssetDocument{}, domain.ErrNotFound
	}
	return d, mapPgErr(err)
}
