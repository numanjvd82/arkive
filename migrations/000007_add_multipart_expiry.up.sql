ALTER TABLE multipart_uploads
  ADD COLUMN expires_at TIMESTAMPTZ;

CREATE INDEX multipart_uploads_expires_at_idx
  ON multipart_uploads (expires_at)
  WHERE expires_at IS NOT NULL;
