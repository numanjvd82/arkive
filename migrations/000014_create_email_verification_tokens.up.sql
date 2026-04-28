CREATE TABLE email_verification_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash bytea NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX email_verification_tokens_user_id_idx ON email_verification_tokens (user_id);
CREATE INDEX email_verification_tokens_expires_at_idx ON email_verification_tokens (expires_at);
CREATE INDEX email_verification_tokens_used_at_idx ON email_verification_tokens (used_at);
