package postgres

import (
	"context"
	"fmt"

	"ams-ai/internal/domain"
)

type ReminderRepository struct {
	db *DB
}

func NewReminderRepository(db *DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

func (r *ReminderRepository) Dashboard(ctx context.Context, user domain.User, windowDays int) (domain.Dashboard, error) {
	total, err := r.countVisibleAssets(ctx, user)
	if err != nil {
		return domain.Dashboard{}, err
	}
	byCategory, err := r.assetsByCategory(ctx, user)
	if err != nil {
		return domain.Dashboard{}, err
	}
	assetRepo := NewAssetRepository(r.db, windowDays)
	expiring, err := assetRepo.ListAssets(ctx, domain.AssetFilters{
		WarrantyState:     domain.WarrantyExpiringSoon,
		CurrentUserID:     user.ID,
		CurrentUserRole:   user.Role,
		ReminderWindowDay: windowDays,
	})
	if err != nil {
		return domain.Dashboard{}, err
	}
	recent, err := r.recentAssets(ctx, user, windowDays)
	if err != nil {
		return domain.Dashboard{}, err
	}
	reminders, err := r.ListReminders(ctx, user, 20)
	if err != nil {
		return domain.Dashboard{}, err
	}
	var insurance, licenses, service []domain.Reminder
	for _, reminder := range reminders {
		switch reminder.SourceType {
		case "insurance":
			insurance = append(insurance, reminder)
		case "license":
			licenses = append(licenses, reminder)
		case "service":
			service = append(service, reminder)
		}
	}
	return domain.Dashboard{
		TotalAssets:              total,
		AssetsByCategory:         byCategory,
		ExpiringWarranties:       takeAssets(expiring, 8),
		ExpiringVehicleInsurance: takeReminders(insurance, 8),
		ExpiringVehicleLicenses:  takeReminders(licenses, 8),
		RecentlyAddedAssets:      recent,
		ServiceDueSoon:           takeReminders(service, 8),
		UpcomingReminders:        reminders,
	}, nil
}

func (r *ReminderRepository) countVisibleAssets(ctx context.Context, user domain.User) (int, error) {
	args := []any{}
	where := "archived_at IS NULL"
	if user.Role != domain.RoleAdmin {
		args = append(args, user.ID)
		where += " AND (created_by = $1 OR assigned_user_id = $1)"
	}
	var count int
	err := r.db.pool.QueryRow(ctx, "SELECT count(*) FROM assets WHERE "+where, args...).Scan(&count)
	return count, err
}

func (r *ReminderRepository) assetsByCategory(ctx context.Context, user domain.User) ([]domain.CategoryCount, error) {
	args := []any{}
	where := "a.archived_at IS NULL"
	if user.Role != domain.RoleAdmin {
		args = append(args, user.ID)
		where += " AND (a.created_by = $1 OR a.assigned_user_id = $1)"
	}
	rows, err := r.db.pool.Query(ctx, `
		SELECT c.id, c.name, count(a.id)::int
		FROM asset_categories c
		LEFT JOIN assets a ON a.category_id = c.id AND `+where+`
		GROUP BY c.id, c.name
		ORDER BY c.name
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CategoryCount{}
	for rows.Next() {
		var c domain.CategoryCount
		if err := rows.Scan(&c.CategoryID, &c.CategoryName, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ReminderRepository) recentAssets(ctx context.Context, user domain.User, windowDays int) ([]domain.Asset, error) {
	return NewAssetRepository(r.db, windowDays).ListAssets(ctx, domain.AssetFilters{
		CurrentUserID:     user.ID,
		CurrentUserRole:   user.Role,
		ReminderWindowDay: windowDays,
	})
}

func (r *ReminderRepository) RegenerateReminders(ctx context.Context, windowDays int) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM reminders`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		SELECT id, 'warranty', id, name || ' warranty expires', warranty_expiry_date,
			CASE WHEN warranty_expiry_date < current_date THEN 'overdue'
				 WHEN warranty_expiry_date = current_date THEN 'due'
				 ELSE 'upcoming' END
		FROM assets
		WHERE archived_at IS NULL
		  AND warranty_expiry_date IS NOT NULL
		  AND warranty_expiry_date <= current_date + ($1::int * interval '1 day')
	`, windowDays); err != nil {
		return err
	}
	for _, q := range []string{
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'insurance', r.id, a.name || ' insurance expires', r.expiry_date,
		 CASE WHEN r.expiry_date < current_date THEN 'overdue' WHEN r.expiry_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM vehicle_insurance_records r JOIN assets a ON a.id = r.asset_id
		 WHERE a.archived_at IS NULL AND r.expiry_date <= current_date + ($1::int * interval '1 day')`,
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'license', r.id, a.name || ' license expires', r.expiry_date,
		 CASE WHEN r.expiry_date < current_date THEN 'overdue' WHEN r.expiry_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM vehicle_license_records r JOIN assets a ON a.id = r.asset_id
		 WHERE a.archived_at IS NULL AND r.expiry_date <= current_date + ($1::int * interval '1 day')`,
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'emission', r.id, a.name || ' emission/inspection expires', r.expiry_date,
		 CASE WHEN r.expiry_date < current_date THEN 'overdue' WHEN r.expiry_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM vehicle_emission_records r JOIN assets a ON a.id = r.asset_id
		 WHERE a.archived_at IS NULL AND r.expiry_date <= current_date + ($1::int * interval '1 day')`,
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'service', s.id, a.name || ' service due', s.next_service_date,
		 CASE WHEN s.next_service_date < current_date THEN 'overdue' WHEN s.next_service_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM service_records s JOIN assets a ON a.id = s.asset_id
		 WHERE a.archived_at IS NULL AND s.next_service_date IS NOT NULL AND s.next_service_date <= current_date + ($1::int * interval '1 day')`,
		`INSERT INTO reminders (asset_id, source_type, source_id, title, due_date, state)
		 SELECT a.id, 'service', -vp.asset_id, a.name || ' service due', vp.next_service_date,
		 CASE WHEN vp.next_service_date < current_date THEN 'overdue' WHEN vp.next_service_date = current_date THEN 'due' ELSE 'upcoming' END
		 FROM vehicle_profiles vp JOIN assets a ON a.id = vp.asset_id
		 WHERE a.archived_at IS NULL
		   AND vp.next_service_date IS NOT NULL
		   AND vp.next_service_date <= current_date + ($1::int * interval '1 day')
		   AND NOT EXISTS (
		   	SELECT 1 FROM service_records sr
		   	WHERE sr.asset_id = vp.asset_id AND sr.next_service_date IS NOT NULL
		   )`,
	} {
		if _, err := tx.Exec(ctx, q, windowDays); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *ReminderRepository) ListReminders(ctx context.Context, user domain.User, limit int) ([]domain.Reminder, error) {
	if limit <= 0 {
		limit = 100
	}
	args := []any{limit}
	where := "a.archived_at IS NULL"
	if user.Role != domain.RoleAdmin {
		args = append(args, user.ID)
		where += fmt.Sprintf(" AND (a.created_by = $%d OR a.assigned_user_id = $%d)", len(args), len(args))
	}
	rows, err := r.db.pool.Query(ctx, `
		SELECT r.id, r.asset_id, a.code, a.name, r.source_type, r.source_id, r.title, r.due_date, r.state, r.created_at
		FROM reminders r
		JOIN assets a ON a.id = r.asset_id
		WHERE `+where+`
		ORDER BY r.due_date ASC, r.id ASC
		LIMIT $1
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Reminder{}
	for rows.Next() {
		var reminder domain.Reminder
		if err := rows.Scan(&reminder.ID, &reminder.AssetID, &reminder.AssetCode, &reminder.AssetName, &reminder.SourceType, &reminder.SourceID, &reminder.Title, &reminder.DueDate, &reminder.State, &reminder.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, reminder)
	}
	return out, rows.Err()
}
