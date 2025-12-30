CREATE TABLE shares (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token TEXT NOT NULL UNIQUE,
  password_hash TEXT,
  expires_at TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'active',
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX shares_file_id_idx ON shares (file_id);
CREATE INDEX shares_owner_user_id_idx ON shares (owner_user_id);

ALTER TABLE shares
  ADD CONSTRAINT shares_status_chk
  CHECK (status IN ('active', 'revoked'));
