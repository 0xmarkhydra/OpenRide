-- Direct-to-object-storage KYC uploads. The API stores metadata only; file bytes never pass through FlashX backend.
ALTER TABLE driver_documents
    ADD COLUMN IF NOT EXISTS filename VARCHAR(180) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS content_type VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS size_bytes BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE UNIQUE INDEX IF NOT EXISTS driver_documents_object_key_unique_idx
    ON driver_documents (object_key);

CREATE INDEX IF NOT EXISTS driver_documents_driver_type_created_idx
    ON driver_documents (driver_id, document_type, created_at DESC);

CREATE INDEX IF NOT EXISTS driver_documents_review_idx
    ON driver_documents (review_status, created_at DESC);
