-- Enable UUID generation (Postgres 13+)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Case-insensitive email type
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),

  -- What you collect
  brand_name        TEXT UNIQUE NOT NULL,               -- username / brand name
  email             citext UNIQUE NOT NULL,                    -- case-insensitive email
  password_hash     TEXT NOT NULL,                      -- bcrypt/argon2 hash (never store raw password)

  -- Storage accounting
  quota_bytes       bigint NOT NULL DEFAULT 2147483648, -- 2 GB default (change as you like)
  used_bytes        bigint NOT NULL DEFAULT 0,
  reserved_bytes    bigint NOT NULL DEFAULT 0,          -- for "upload initiated" but not completed

  -- Platform controls
  is_email_verified boolean NOT NULL DEFAULT false,
  is_banned         boolean NOT NULL DEFAULT false,
  ban_reason        text,

  -- Abuse / rate limiting (optional but useful)
  last_login_at     TIMESTAMPTZ,
  last_ip           inet,

  -- Auditing
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Helpful index for admin queries
CREATE INDEX users_created_at_idx ON users (created_at);

-- Safety checks
ALTER TABLE users
  ADD CONSTRAINT users_bytes_nonnegative_chk
  CHECK (quota_bytes >= 0 AND used_bytes >= 0 AND reserved_bytes >= 0);

ALTER TABLE users
  ADD CONSTRAINT users_used_within_quota_chk
  CHECK (used_bytes + reserved_bytes <= quota_bytes);
