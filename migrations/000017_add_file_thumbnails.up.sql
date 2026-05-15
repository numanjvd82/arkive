ALTER TABLE files
  ADD COLUMN thumbnail_status TEXT NOT NULL DEFAULT 'none',
  ADD COLUMN thumbnail_size_bytes BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN thumbnail_mime TEXT NOT NULL DEFAULT '',
  ADD COLUMN thumbnail_width INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN thumbnail_height INTEGER NOT NULL DEFAULT 0;

ALTER TABLE files
  ADD CONSTRAINT files_thumbnail_status_chk
  CHECK (thumbnail_status IN ('none', 'complete', 'failed'));

ALTER TABLE files
  ADD CONSTRAINT files_thumbnail_bytes_nonnegative_chk
  CHECK (thumbnail_size_bytes >= 0);

ALTER TABLE files
  ADD CONSTRAINT files_thumbnail_width_nonnegative_chk
  CHECK (thumbnail_width >= 0);

ALTER TABLE files
  ADD CONSTRAINT files_thumbnail_height_nonnegative_chk
  CHECK (thumbnail_height >= 0);
