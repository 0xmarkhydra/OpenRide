ALTER TABLE admin_users
    ADD COLUMN IF NOT EXISTS phone VARCHAR(32);

CREATE UNIQUE INDEX IF NOT EXISTS admin_users_phone_unique_idx
    ON admin_users (phone)
    WHERE phone IS NOT NULL;
