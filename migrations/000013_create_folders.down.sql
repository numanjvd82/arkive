DROP TABLE IF EXISTS folders;

ALTER TABLE files
  DROP COLUMN IF EXISTS folder_path;

ALTER TABLE files
  ADD COLUMN throttle_ms INT NOT NULL DEFAULT 0;

ALTER TABLE files
  ADD CONSTRAINT files_throttle_nonnegative_chk
  CHECK (throttle_ms >= 0);
