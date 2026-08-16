-- FlashX configurable vehicle-inspection document checklist.
-- Active template can evolve; each inspection job snapshots the template version
-- so later Admin changes never rewrite historical jobs.

CREATE TABLE IF NOT EXISTS inspection_checklist_templates (
    id TEXT PRIMARY KEY,
    version BIGINT NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    created_by TEXT,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS inspection_checklist_one_active_template_idx
    ON inspection_checklist_templates ((TRUE))
    WHERE active = TRUE;

CREATE TABLE IF NOT EXISTS inspection_checklist_template_items (
    template_id TEXT NOT NULL REFERENCES inspection_checklist_templates(id) ON DELETE CASCADE,
    item_key VARCHAR(64) NOT NULL,
    label TEXT NOT NULL,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (template_id, item_key)
);

CREATE TABLE IF NOT EXISTS inspection_checklists (
    trip_id TEXT PRIMARY KEY REFERENCES trips(id) ON DELETE CASCADE,
    template_id TEXT NOT NULL REFERENCES inspection_checklist_templates(id),
    template_version BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inspection_checklist_items (
    id TEXT PRIMARY KEY,
    trip_id TEXT NOT NULL REFERENCES inspection_checklists(trip_id) ON DELETE CASCADE,
    item_key VARCHAR(64) NOT NULL,
    label TEXT NOT NULL,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    customer_status VARCHAR(24) NOT NULL DEFAULT 'pending'
        CHECK (customer_status IN ('pending','present','not_available','not_applicable')),
    customer_note TEXT NOT NULL DEFAULT '',
    customer_updated_at TIMESTAMPTZ,
    driver_status VARCHAR(24) NOT NULL DEFAULT 'pending'
        CHECK (driver_status IN ('pending','received','missing','not_applicable')),
    driver_note TEXT NOT NULL DEFAULT '',
    driver_updated_at TIMESTAMPTZ,
    UNIQUE (trip_id, item_key)
);

CREATE INDEX IF NOT EXISTS inspection_checklist_items_trip_idx
    ON inspection_checklist_items (trip_id, sort_order, item_key);

INSERT INTO inspection_checklist_templates (id, version, active, reason)
VALUES ('inspection_checklist_v1', 1, TRUE, 'FlashX MVP default template')
ON CONFLICT (id) DO NOTHING;

INSERT INTO inspection_checklist_template_items (template_id, item_key, label, required, sort_order)
VALUES
    ('inspection_checklist_v1', 'vehicle_registration', 'Đăng ký xe / giấy tờ xe hợp lệ', TRUE, 10),
    ('inspection_checklist_v1', 'previous_inspection_docs', 'Giấy tờ đăng kiểm trước đây (nếu có)', FALSE, 20),
    ('inspection_checklist_v1', 'authorization_docs', 'Giấy ủy quyền / giấy tờ đại diện (nếu áp dụng)', FALSE, 30),
    ('inspection_checklist_v1', 'other_related_docs', 'Giấy tờ liên quan khác theo trường hợp', FALSE, 40)
ON CONFLICT (template_id, item_key) DO NOTHING;
