ALTER TABLE files
ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL,
ADD COLUMN IF NOT EXISTS purged_at timestamptz NULL,
ADD COLUMN IF NOT EXISTS tombstone_purged_at timestamptz NULL;

CREATE INDEX IF NOT EXISTS idx_files_deleted_cleanup
ON files(deleted_at)
WHERE deleted_at IS NOT NULL
  AND purged_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_files_tombstone_cleanup
ON files(deleted_at)
WHERE deleted_at IS NOT NULL
  AND purged_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_files_sync_listing
ON files(user_id, folder_id, updated_at DESC, id)
WHERE upload_status = 'complete';

CREATE INDEX IF NOT EXISTS idx_folders_sync_listing
ON folders(user_id, parent_folder_id, updated_at DESC, id);
