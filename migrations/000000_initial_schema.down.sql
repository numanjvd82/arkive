DROP TABLE IF EXISTS share_snapshot_files;
DROP TABLE IF EXISTS share_items;
DROP TABLE IF EXISTS share_links;
DROP TABLE IF EXISTS restore_usage;
DROP TABLE IF EXISTS upload_usage;
DROP TABLE IF EXISTS upload_parts;
DROP TABLE IF EXISTS upload_sessions;
DROP TABLE IF EXISTS folder_search_tokens;
DROP TABLE IF EXISTS file_search_tokens;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS folders;
DROP TABLE IF EXISTS instance_settings;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS citext;
DROP EXTENSION IF EXISTS pgcrypto;
