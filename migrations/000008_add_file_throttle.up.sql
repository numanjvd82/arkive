ALTER TABLE files
  ADD COLUMN throttle_ms INT NOT NULL DEFAULT 0;

ALTER TABLE files
  ADD CONSTRAINT files_throttle_nonnegative_chk
  CHECK (throttle_ms >= 0);
