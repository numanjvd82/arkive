ALTER TABLE files
  ADD COLUMN folder_path TEXT NOT NULL DEFAULT '';

ALTER TABLE files
  DROP COLUMN IF EXISTS throttle_ms;

CREATE TABLE folders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  path TEXT NOT NULL,
  name TEXT NOT NULL,
  parent_path TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX folders_user_path_uidx ON folders (user_id, path);
CREATE INDEX folders_user_parent_idx ON folders (user_id, parent_path);
