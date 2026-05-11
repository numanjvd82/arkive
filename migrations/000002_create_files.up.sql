CREATE TABLE files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  encrypted_metadata BYTEA NOT NULL,
  encrypted_file_key BYTEA NOT NULL,
  encrypted_manifest BYTEA NOT NULL,
  encryption_version SMALLINT NOT NULL DEFAULT 1,
  chunk_size BIGINT NOT NULL,
  chunk_count INTEGER NOT NULL,
  plaintext_size BIGINT NOT NULL,
  encrypted_size BIGINT,
  encrypted_hash BYTEA,
  upload_status TEXT NOT NULL DEFAULT 'pending',
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX files_user_id_idx ON files (user_id);

ALTER TABLE files
  ADD CONSTRAINT files_upload_status_chk
  CHECK (upload_status IN ('pending', 'uploading', 'complete', 'failed', 'aborted'));

CREATE TABLE file_chunks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  chunk_index INTEGER NOT NULL,
  storage_key TEXT NOT NULL,
  plaintext_size BIGINT NOT NULL,
  encrypted_size BIGINT NOT NULL,
  encrypted_hash BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX file_chunks_file_id_idx ON file_chunks (file_id);

ALTER TABLE file_chunks
  ADD CONSTRAINT file_chunks_file_id_chunk_index_key
  UNIQUE (file_id, chunk_index);

CREATE TABLE upload_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  storage_key TEXT NOT NULL,
  provider_upload_id TEXT,
  status TEXT NOT NULL DEFAULT 'active',
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX upload_sessions_file_id_idx ON upload_sessions (file_id);
CREATE INDEX upload_sessions_status_idx ON upload_sessions (status);

ALTER TABLE upload_sessions
  ADD CONSTRAINT upload_sessions_status_chk
  CHECK (status IN ('active', 'completed', 'aborted', 'expired', 'failed'));

CREATE TABLE upload_parts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  upload_session_id UUID NOT NULL REFERENCES upload_sessions(id) ON DELETE CASCADE,
  part_number INTEGER NOT NULL,
  etag TEXT,
  encrypted_hash BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX upload_parts_upload_session_id_idx ON upload_parts (upload_session_id);

ALTER TABLE upload_parts
  ADD CONSTRAINT upload_parts_upload_session_id_part_number_key
  UNIQUE (upload_session_id, part_number);
