CREATE TABLE upload_usage (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  size_bytes BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX upload_usage_user_created_idx ON upload_usage (user_id, created_at DESC);

ALTER TABLE upload_usage
  ADD CONSTRAINT upload_usage_size_nonnegative_chk
  CHECK (size_bytes >= 0);
