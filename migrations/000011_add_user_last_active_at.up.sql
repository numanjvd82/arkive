ALTER TABLE users
  ADD COLUMN last_active_at TIMESTAMPTZ;

UPDATE users
SET last_active_at = COALESCE(last_login_at, created_at);

ALTER TABLE users
  ALTER COLUMN last_active_at SET NOT NULL,
  ALTER COLUMN last_active_at SET DEFAULT now();

CREATE INDEX users_last_active_at_idx ON users (last_active_at);
