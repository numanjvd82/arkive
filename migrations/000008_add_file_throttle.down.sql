ALTER TABLE files
  DROP CONSTRAINT IF EXISTS files_throttle_nonnegative_chk;

ALTER TABLE files
  DROP COLUMN IF EXISTS throttle_ms;
