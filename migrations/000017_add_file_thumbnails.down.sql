ALTER TABLE files
  DROP CONSTRAINT IF EXISTS files_thumbnail_height_nonnegative_chk,
  DROP CONSTRAINT IF EXISTS files_thumbnail_width_nonnegative_chk,
  DROP CONSTRAINT IF EXISTS files_thumbnail_bytes_nonnegative_chk,
  DROP CONSTRAINT IF EXISTS files_thumbnail_status_chk,
  DROP COLUMN IF EXISTS thumbnail_height,
  DROP COLUMN IF EXISTS thumbnail_width,
  DROP COLUMN IF EXISTS thumbnail_mime,
  DROP COLUMN IF EXISTS thumbnail_size_bytes,
  DROP COLUMN IF EXISTS thumbnail_status;
