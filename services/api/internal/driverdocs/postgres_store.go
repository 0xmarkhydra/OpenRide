package driverdocs

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(doc Document) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO driver_documents (
			id, driver_id, document_type, object_key, filename, content_type,
			size_bytes, review_status, review_note, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, doc.ID, doc.DriverID, doc.DocumentType, doc.ObjectKey, doc.Filename,
		doc.ContentType, doc.SizeBytes, string(doc.ReviewStatus), doc.ReviewNote,
		doc.CreatedAt, doc.UpdatedAt)
	return err
}

func (s *PostgresStore) Get(id string) (Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanDocument(s.pool.QueryRow(ctx, documentSelect+` WHERE id = $1`, id))
}

func (s *PostgresStore) ListForDriver(driverID string) ([]Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, documentSelect+` WHERE driver_id = $1 ORDER BY created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Document, 0)
	for rows.Next() {
		doc, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, doc)
	}
	return result, rows.Err()
}

func (s *PostgresStore) Review(id string, status ReviewStatus, note string) (Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := s.pool.QueryRow(ctx, `
		UPDATE driver_documents
		SET review_status = $2,
		    review_note = $3,
		    reviewed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, driver_id, document_type, object_key, filename, content_type,
		          size_bytes, review_status, review_note, created_at, updated_at, reviewed_at
	`, id, string(status), note)
	return scanDocument(row)
}

const documentSelect = `
	SELECT id, driver_id, document_type, object_key, filename, content_type,
	       size_bytes, review_status, review_note, created_at, updated_at, reviewed_at
	FROM driver_documents`

func scanDocument(row interface{ Scan(dest ...any) error }) (Document, error) {
	var doc Document
	var status string
	if err := row.Scan(
		&doc.ID, &doc.DriverID, &doc.DocumentType, &doc.ObjectKey, &doc.Filename,
		&doc.ContentType, &doc.SizeBytes, &status, &doc.ReviewNote, &doc.CreatedAt,
		&doc.UpdatedAt, &doc.ReviewedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Document{}, ErrNotFound
		}
		return Document{}, err
	}
	doc.ReviewStatus = ReviewStatus(status)
	return doc, nil
}
