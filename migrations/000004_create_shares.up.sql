CREATE TABLE shares (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token TEXT NOT NULL UNIQUE,
  password_hash TEXT,
  expires_at TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'active',
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX shares_file_id_idx ON shares (file_id);
CREATE INDEX shares_owner_user_id_idx ON shares (owner_user_id);

ALTER TABLE shares
  ADD CONSTRAINT shares_status_chk
  CHECK (status IN ('active', 'revoked'));

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
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX share_links_owner_user_id_idx ON share_links (owner_user_id);

ALTER TABLE share_links
  ADD CONSTRAINT share_links_status_chk
  CHECK (status IN ('active', 'expired', 'revoked', 'burned', 'disabled'));

ALTER TABLE share_links
  ADD CONSTRAINT share_links_password_mode_chk
  CHECK (password_mode IN ('access_gate', 'decrypt_gate'));

CREATE TABLE share_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  share_link_id UUID NOT NULL REFERENCES share_links(id) ON DELETE CASCADE,
  item_type TEXT NOT NULL,
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  display_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX share_items_share_link_id_idx ON share_items (share_link_id);
CREATE INDEX share_items_file_id_idx ON share_items (file_id);

ALTER TABLE share_items
  ADD CONSTRAINT share_items_item_type_chk
  CHECK (item_type = 'file');

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
