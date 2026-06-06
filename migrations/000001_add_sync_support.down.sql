DROP INDEX IF EXISTS idx_folders_sync_listing;
DROP INDEX IF EXISTS idx_files_sync_listing;
DROP INDEX IF EXISTS idx_files_tombstone_cleanup;
DROP INDEX IF EXISTS idx_files_deleted_cleanup;

ALTER TABLE files
DROP COLUMN IF EXISTS tombstone_purged_at,
DROP COLUMN IF EXISTS purged_at,
DROP COLUMN IF EXISTS deleted_at;
