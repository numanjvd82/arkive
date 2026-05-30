CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  brand_name TEXT UNIQUE NOT NULL,
  email citext UNIQUE NOT NULL,
  password_hash TEXT,
  vault_salt BYTEA,
  encrypted_master_key BYTEA,
  used_bytes BIGINT NOT NULL DEFAULT 0,
  reserved_bytes BIGINT NOT NULL DEFAULT 0,
  last_login_at TIMESTAMPTZ,
  last_active_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  recovery_setup_token TEXT UNIQUE,
  recovery_setup_token_expires_at TIMESTAMPTZ,
  encrypted_master_key_recovery BYTEA,
  password_reset_token_hash TEXT UNIQUE,
  password_reset_token_expires_at TIMESTAMPTZ,
  password_reset_consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT users_bytes_nonnegative_chk
    CHECK (used_bytes >= 0 AND reserved_bytes >= 0)
);



CREATE INDEX users_created_at_idx ON users (created_at);
CREATE INDEX users_last_active_at_idx ON users (last_active_at);
CREATE INDEX users_vault_ready_idx
  ON users (id)
  WHERE vault_salt IS NOT NULL AND encrypted_master_key IS NOT NULL;
CREATE INDEX users_password_reset_token_hash_idx
  ON users (password_reset_token_hash)
  WHERE password_reset_token_hash IS NOT NULL;

CREATE TABLE sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

CREATE TABLE instance_settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE folders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  vault_id UUID NOT NULL,
  parent_folder_id UUID REFERENCES folders(id) ON DELETE SET NULL,
  encrypted_name BYTEA NOT NULL,
  encrypted_metadata BYTEA,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_folders_user_parent_created
  ON folders (user_id, parent_folder_id, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE TABLE files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  folder_id UUID REFERENCES folders(id) ON DELETE SET NULL,
  encrypted_metadata BYTEA NOT NULL,
  encrypted_file_key BYTEA NOT NULL,
  encrypted_manifest BYTEA NOT NULL,
  encryption_version SMALLINT NOT NULL DEFAULT 1,
  chunk_size BIGINT NOT NULL,
  chunk_count INTEGER NOT NULL,
  plaintext_size BIGINT NOT NULL,
  actual_encrypted_size BIGINT NOT NULL DEFAULT 0,
  encrypted_hash BYTEA,
  upload_status TEXT NOT NULL DEFAULT 'pending',
  completed_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ,
  throttle_ms INT NOT NULL DEFAULT 0,
  thumbnail_status TEXT NOT NULL DEFAULT 'none',
  thumbnail_size_bytes BIGINT NOT NULL DEFAULT 0,
  thumbnail_mime TEXT NOT NULL DEFAULT '',
  thumbnail_width INTEGER NOT NULL DEFAULT 0,
  thumbnail_height INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT files_upload_status_chk
    CHECK (upload_status IN ('pending', 'uploading', 'complete', 'failed', 'aborted')),
  CONSTRAINT files_throttle_nonnegative_chk
    CHECK (throttle_ms >= 0),
  CONSTRAINT files_thumbnail_status_chk
    CHECK (thumbnail_status IN ('none', 'complete', 'failed')),
  CONSTRAINT files_thumbnail_bytes_nonnegative_chk
    CHECK (thumbnail_size_bytes >= 0),
  CONSTRAINT files_thumbnail_width_nonnegative_chk
    CHECK (thumbnail_width >= 0),
  CONSTRAINT files_thumbnail_height_nonnegative_chk
    CHECK (thumbnail_height >= 0)
);

CREATE INDEX files_user_id_idx ON files (user_id);
CREATE INDEX files_expires_at_idx
  ON files (expires_at)
  WHERE expires_at IS NOT NULL;
CREATE INDEX idx_files_user_folder_complete_created
  ON files (user_id, folder_id, created_at DESC)
  WHERE upload_status = 'complete'
    AND expires_at IS NULL;

CREATE TABLE file_search_tokens (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  vault_id UUID NOT NULL,
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  token_hash BYTEA NOT NULL,
  field TEXT NOT NULL,
  weight INT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, vault_id, token_hash, file_id, field)
);

CREATE INDEX idx_file_search_tokens_lookup
  ON file_search_tokens (user_id, vault_id, token_hash, weight DESC);

CREATE TABLE folder_search_tokens (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  vault_id UUID NOT NULL,
  folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
  token_hash BYTEA NOT NULL,
  field TEXT NOT NULL,
  weight INT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, vault_id, token_hash, folder_id, field)
);

CREATE INDEX idx_folder_search_tokens_lookup
  ON folder_search_tokens (user_id, vault_id, token_hash, weight DESC);

CREATE TABLE upload_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  provider_upload_id TEXT,
  status TEXT NOT NULL DEFAULT 'active',
  expires_at TIMESTAMPTZ NOT NULL,
  upload_part_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT upload_sessions_status_chk
    CHECK (status IN ('active', 'completed', 'aborted', 'expired', 'failed'))
);

CREATE INDEX upload_sessions_file_id_idx ON upload_sessions (file_id);
CREATE INDEX upload_sessions_status_idx ON upload_sessions (status);
CREATE INDEX upload_sessions_expires_at_idx ON upload_sessions (expires_at);

CREATE TABLE upload_parts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  upload_session_id UUID NOT NULL REFERENCES upload_sessions(id) ON DELETE CASCADE,
  part_number INTEGER NOT NULL,
  etag TEXT,
  encrypted_hash BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT upload_parts_upload_session_id_part_number_key
    UNIQUE (upload_session_id, part_number)
);

CREATE INDEX upload_parts_upload_session_id_idx ON upload_parts (upload_session_id);

CREATE TABLE upload_usage (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  size_bytes BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT upload_usage_size_nonnegative_chk
    CHECK (size_bytes >= 0)
);

CREATE INDEX upload_usage_user_created_idx ON upload_usage (user_id, created_at DESC);

CREATE TABLE restore_usage (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  size_bytes BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT restore_usage_size_nonnegative_chk
    CHECK (size_bytes >= 0)
);

CREATE INDEX restore_usage_user_created_idx ON restore_usage (user_id, created_at DESC);

CREATE TABLE share_links (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token TEXT NOT NULL UNIQUE,
  slug TEXT UNIQUE,
  status TEXT NOT NULL DEFAULT 'active',
  title_encrypted BYTEA,
  description_encrypted BYTEA,
  encrypted_share_key BYTEA NOT NULL,
  crypto_version SMALLINT NOT NULL DEFAULT 1,
  password_hash TEXT,
  password_mode TEXT NOT NULL DEFAULT 'access_gate',
  access_count INTEGER NOT NULL DEFAULT 0,
  max_access_count INTEGER,
  consumed_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  allow_preview BOOLEAN NOT NULL DEFAULT true,
  allow_download BOOLEAN NOT NULL DEFAULT true,
  comments_enabled BOOLEAN NOT NULL DEFAULT false,
  reactions_enabled BOOLEAN NOT NULL DEFAULT false,
  burn_after_read BOOLEAN NOT NULL DEFAULT false,
  show_exif BOOLEAN NOT NULL DEFAULT false,
  show_location BOOLEAN NOT NULL DEFAULT false,
  strip_exif_download BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT share_links_status_chk
    CHECK (status IN ('active', 'expired', 'revoked', 'burned', 'disabled')),
  CONSTRAINT share_links_password_mode_chk
    CHECK (password_mode IN ('access_gate', 'decrypt_gate'))
);

CREATE INDEX share_links_owner_user_id_idx ON share_links (owner_user_id);

CREATE TABLE share_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  share_link_id UUID NOT NULL REFERENCES share_links(id) ON DELETE CASCADE,
  item_type TEXT NOT NULL,
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  display_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT share_items_item_type_chk
    CHECK (item_type = 'file')
);

CREATE INDEX share_items_share_link_id_idx ON share_items (share_link_id);
CREATE INDEX share_items_file_id_idx ON share_items (file_id);

CREATE TABLE share_snapshot_files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  share_item_id UUID NOT NULL REFERENCES share_items(id) ON DELETE CASCADE,
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  encrypted_relative_path BYTEA,
  encrypted_file_key_for_share BYTEA NOT NULL,
  display_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX share_snapshot_files_share_item_id_idx ON share_snapshot_files (share_item_id);
CREATE INDEX share_snapshot_files_file_id_idx ON share_snapshot_files (file_id);
