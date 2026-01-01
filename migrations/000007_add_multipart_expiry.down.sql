DROP INDEX IF EXISTS multipart_uploads_expires_at_idx;

ALTER TABLE multipart_uploads
  DROP COLUMN IF EXISTS expires_at;
