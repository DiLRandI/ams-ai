package postgres

import (
	"errors"
	"fmt"

	"ams-ai/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapPgErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: duplicate value", domain.ErrConflict)
		case "23503":
			return fmt.Errorf("%w: referenced record not found", domain.ErrInvalid)
		case "23514":
			return fmt.Errorf("%w: invalid value", domain.ErrInvalid)
		}
	}
	return err
}

func takeAssets(in []domain.Asset, n int) []domain.Asset {
	if in == nil {
		return []domain.Asset{}
	}
	if len(in) <= n {
		return in
	}
	return in[:n]
}

func takeReminders(in []domain.Reminder, n int) []domain.Reminder {
	if in == nil {
		return []domain.Reminder{}
	}
	if len(in) <= n {
		return in
	}
	return in[:n]
}
