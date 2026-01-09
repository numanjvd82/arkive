CREATE TABLE restore_usage (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  size_bytes BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX restore_usage_user_created_idx ON restore_usage (user_id, created_at DESC);

ALTER TABLE restore_usage
  ADD CONSTRAINT restore_usage_size_nonnegative_chk
  CHECK (size_bytes >= 0);
