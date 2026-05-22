ALTER TABLE share_links
  DROP COLUMN IF EXISTS consumed_at,
  DROP COLUMN IF EXISTS max_access_count,
  DROP COLUMN IF EXISTS access_count;
