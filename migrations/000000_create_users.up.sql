-- Enable UUID generation (Postgres 13+)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Case-insensitive email type
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),

  -- What you collect
  brand_name        TEXT UNIQUE NOT NULL,               -- username / brand name
  email             citext UNIQUE NOT NULL,                    -- case-insensitive email
  password_hash          TEXT,                               -- bcrypt/argon2 hash (never store raw password)
  vault_salt             BYTEA,
  encrypted_master_key   BYTEA,
  google_sub             TEXT UNIQUE,                        -- google user id
  google_given_name      TEXT,
  google_family_name     TEXT,
  google_picture_url     TEXT,

  -- Storage accounting
  quota_bytes       bigint NOT NULL DEFAULT 5368709120, -- 5 GB default (change as you like)
  used_bytes        bigint NOT NULL DEFAULT 0,
  reserved_bytes    bigint NOT NULL DEFAULT 0,          -- for "upload initiated" but not completed

  last_login_at     TIMESTAMPTZ,

  -- Setup recovery flow
  recovery_setup_token            TEXT UNIQUE,
  recovery_setup_token_expires_at TIMESTAMPTZ,

  -- Auditing
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Helpful index for admin queries
CREATE INDEX users_created_at_idx ON users (created_at);
CREATE INDEX users_vault_ready_idx
  ON users (id)
  WHERE vault_salt IS NOT NULL AND encrypted_master_key IS NOT NULL;

-- Safety checks
ALTER TABLE users
  ADD CONSTRAINT users_bytes_nonnegative_chk
  CHECK (quota_bytes >= 0 AND used_bytes >= 0 AND reserved_bytes >= 0);

ALTER TABLE users
  ADD CONSTRAINT users_used_within_quota_chk
  CHECK (used_bytes + reserved_bytes <= quota_bytes);
