ALTER TABLE share_links
  ADD COLUMN access_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN max_access_count INTEGER,
  ADD COLUMN consumed_at TIMESTAMPTZ;
