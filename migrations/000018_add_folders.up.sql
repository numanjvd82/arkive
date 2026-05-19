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

ALTER TABLE files
  ADD COLUMN folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;

CREATE INDEX idx_folders_user_parent_created
  ON folders (user_id, parent_folder_id, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_files_user_folder_complete_created
  ON files (user_id, folder_id, created_at DESC)
  WHERE upload_status = 'complete'
    AND expires_at IS NULL;
