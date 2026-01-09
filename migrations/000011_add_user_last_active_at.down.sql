DROP INDEX IF EXISTS users_last_active_at_idx;

ALTER TABLE users
  DROP COLUMN IF EXISTS last_active_at;
