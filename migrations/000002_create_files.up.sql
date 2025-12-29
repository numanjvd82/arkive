CREATE TABLE files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  bucket TEXT NOT NULL,
  object_key TEXT NOT NULL UNIQUE,
  filename TEXT NOT NULL,
  content_type TEXT,
  size_bytes BIGINT NOT NULL,
  video_width INT NOT NULL DEFAULT 0,
  video_height INT NOT NULL DEFAULT 0,
  video_duration_seconds BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX files_user_id_idx ON files (user_id);

ALTER TABLE files
  ADD CONSTRAINT files_status_chk
  CHECK (status IN ('pending', 'uploading', 'complete', 'failed', 'aborted'));

CREATE TABLE multipart_uploads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  upload_id TEXT NOT NULL,
  bucket TEXT NOT NULL,
  object_key TEXT NOT NULL,
  chunk_size INT NOT NULL,
  total_parts INT NOT NULL,
  uploaded_parts JSONB NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'initiated',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX multipart_uploads_file_id_idx ON multipart_uploads (file_id);
CREATE INDEX multipart_uploads_status_idx ON multipart_uploads (status);

ALTER TABLE multipart_uploads
  ADD CONSTRAINT multipart_uploads_status_chk
  CHECK (status IN ('initiated', 'uploading', 'completed', 'aborted', 'failed'));
