DROP INDEX IF EXISTS idx_files_user_folder_complete_created;
DROP INDEX IF EXISTS idx_folders_user_parent_created;

ALTER TABLE files
  DROP COLUMN IF EXISTS folder_id;

DROP TABLE IF EXISTS folders;
