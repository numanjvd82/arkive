DROP INDEX IF EXISTS files_expires_at_idx;

ALTER TABLE files
  DROP COLUMN IF EXISTS expires_at;
