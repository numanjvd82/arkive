ALTER TABLE files
  ADD COLUMN expires_at TIMESTAMPTZ;

CREATE INDEX files_expires_at_idx
  ON files (expires_at)
  WHERE expires_at IS NOT NULL;
