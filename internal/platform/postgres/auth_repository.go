package postgres

import (
	"context"
	"errors"
	"strings"

	"ams-ai/internal/domain"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository struct {
	db *DB
}

func NewAuthRepository(db *DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.db.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)
	`, strings.TrimSpace(email))
	return scanUser(row)
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id int64) (domain.User, error) {
	row := r.db.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *AuthRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		ORDER BY full_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []domain.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *AuthRepository) UpsertSeedUser(ctx context.Context, email, password, fullName, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = r.db.pool.Exec(ctx, `
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = EXCLUDED.role,
			updated_at = now()
	`, email, string(hash), fullName, role)
	return err
}

func scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	return u, err
}
