package inspectionchecklist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) EnsureDefaultTemplate(template Template) error {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM inspection_checklist_templates WHERE active=TRUE)`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return tx.Commit(ctx)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO inspection_checklist_templates (id,version,active,created_by,reason,created_at) VALUES ($1,$2,TRUE,$3,$4,$5) ON CONFLICT (id) DO UPDATE SET active=TRUE`, template.ID, template.Version, template.CreatedBy, template.Reason, template.CreatedAt); err != nil {
		return err
	}
	for _, item := range template.Items {
		if _, err := tx.Exec(ctx, `INSERT INTO inspection_checklist_template_items (template_id,item_key,label,required,sort_order) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (template_id,item_key) DO NOTHING`, template.ID, item.Key, item.Label, item.Required, item.SortOrder); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) GetActiveTemplate() (Template, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var template Template
	if err := s.pool.QueryRow(ctx, `SELECT id,version,active,COALESCE(created_by,''),reason,created_at FROM inspection_checklist_templates WHERE active=TRUE LIMIT 1`).Scan(&template.ID, &template.Version, &template.Active, &template.CreatedBy, &template.Reason, &template.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Template{}, ErrNotFound
		}
		return Template{}, err
	}
	items, err := s.templateItems(ctx, template.ID)
	if err != nil {
		return Template{}, err
	}
	template.Items = items
	return template, nil
}

func (s *PostgresStore) CreateTemplate(items []TemplateItem, actorID, reason string, now time.Time) (Template, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Template{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('flashx_inspection_checklist_template'))`); err != nil {
		return Template{}, err
	}
	var version int64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(version),0)+1 FROM inspection_checklist_templates`).Scan(&version); err != nil {
		return Template{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE inspection_checklist_templates SET active=FALSE WHERE active=TRUE`); err != nil {
		return Template{}, err
	}
	template := Template{ID: templateID(version), Version: version, Active: true, Items: cloneTemplateItems(items), CreatedBy: actorID, Reason: reason, CreatedAt: now}
	if _, err := tx.Exec(ctx, `INSERT INTO inspection_checklist_templates (id,version,active,created_by,reason,created_at) VALUES ($1,$2,TRUE,$3,$4,$5)`, template.ID, template.Version, template.CreatedBy, template.Reason, template.CreatedAt); err != nil {
		return Template{}, err
	}
	for _, item := range items {
		if _, err := tx.Exec(ctx, `INSERT INTO inspection_checklist_template_items (template_id,item_key,label,required,sort_order) VALUES ($1,$2,$3,$4,$5)`, template.ID, item.Key, item.Label, item.Required, item.SortOrder); err != nil {
			return Template{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Template{}, err
	}
	return template, nil
}

func (s *PostgresStore) GetChecklist(tripID string) (Checklist, []Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var checklist Checklist
	if err := s.pool.QueryRow(ctx, `SELECT trip_id,template_id,template_version,created_at,updated_at FROM inspection_checklists WHERE trip_id=$1`, tripID).Scan(&checklist.TripID, &checklist.TemplateID, &checklist.TemplateVersion, &checklist.CreatedAt, &checklist.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Checklist{}, nil, ErrNotFound
		}
		return Checklist{}, nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT id,trip_id,item_key,label,required,sort_order,customer_status,customer_note,customer_updated_at,driver_status,driver_note,driver_updated_at FROM inspection_checklist_items WHERE trip_id=$1 ORDER BY sort_order,item_key`, tripID)
	if err != nil {
		return Checklist{}, nil, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.TripID, &item.Key, &item.Label, &item.Required, &item.SortOrder, &item.CustomerStatus, &item.CustomerNote, &item.CustomerUpdatedAt, &item.DriverStatus, &item.DriverNote, &item.DriverUpdatedAt); err != nil {
			return Checklist{}, nil, err
		}
		items = append(items, item)
	}
	return checklist, items, rows.Err()
}

func (s *PostgresStore) CreateChecklist(checklist Checklist, items []Item) error {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `INSERT INTO inspection_checklists (trip_id,template_id,template_version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (trip_id) DO NOTHING`, checklist.TripID, checklist.TemplateID, checklist.TemplateVersion, checklist.CreatedAt, checklist.UpdatedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}
	for _, item := range items {
		if _, err := tx.Exec(ctx, `INSERT INTO inspection_checklist_items (id,trip_id,item_key,label,required,sort_order,customer_status,customer_note,driver_status,driver_note) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, item.ID, item.TripID, item.Key, item.Label, item.Required, item.SortOrder, item.CustomerStatus, item.CustomerNote, item.DriverStatus, item.DriverNote); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) SaveItem(item Item) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `UPDATE inspection_checklist_items SET customer_status=$3,customer_note=$4,customer_updated_at=$5,driver_status=$6,driver_note=$7,driver_updated_at=$8 WHERE trip_id=$1 AND item_key=$2`, item.TripID, item.Key, item.CustomerStatus, item.CustomerNote, item.CustomerUpdatedAt, item.DriverStatus, item.DriverNote, item.DriverUpdatedAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	_, err = s.pool.Exec(ctx, `UPDATE inspection_checklists SET updated_at=NOW() WHERE trip_id=$1`, item.TripID)
	return err
}

func (s *PostgresStore) templateItems(ctx context.Context, templateID string) ([]TemplateItem, error) {
	rows, err := s.pool.Query(ctx, `SELECT item_key,label,required,sort_order FROM inspection_checklist_template_items WHERE template_id=$1 ORDER BY sort_order,item_key`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TemplateItem, 0)
	for rows.Next() {
		var item TemplateItem
		if err := rows.Scan(&item.Key, &item.Label, &item.Required, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: template has no items", ErrNotFound)
	}
	return items, rows.Err()
}
