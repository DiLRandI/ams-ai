package postgres

import (
	"context"
	"errors"
	"strings"

	"ams-ai/internal/domain"

	"github.com/jackc/pgx/v5"
)

type CategoryRepository struct {
	db *DB
}

func NewCategoryRepository(db *DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, name, description, is_system, created_at, updated_at
		FROM asset_categories
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Category{}
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepository) CreateCategory(ctx context.Context, name, description string) (domain.Category, error) {
	row := r.db.pool.QueryRow(ctx, `
		INSERT INTO asset_categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, is_system, created_at, updated_at
	`, strings.TrimSpace(name), strings.TrimSpace(description))
	var c domain.Category
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt)
	return c, mapPgErr(err)
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, id int64, name, description string) (domain.Category, error) {
	row := r.db.pool.QueryRow(ctx, `
		UPDATE asset_categories
		SET name = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, name, description, is_system, created_at, updated_at
	`, id, strings.TrimSpace(name), strings.TrimSpace(description))
	var c domain.Category
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Category{}, domain.ErrNotFound
	}
	return c, mapPgErr(err)
}
