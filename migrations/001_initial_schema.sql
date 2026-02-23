-- Migration 001: Initial schema for items and users.
--
-- Apply with:
--   psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
--
-- This schema is idempotent; running it multiple times is safe.

-- Users table: stores registered accounts.
-- Passwords are NEVER stored in plain text — only bcrypt hashes.
CREATE TABLE IF NOT EXISTS users (
    username      VARCHAR(50)  PRIMARY KEY,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Items table: stores the application resources managed by the API.
-- Serial primary key provides an auto-incrementing integer ID.
CREATE TABLE IF NOT EXISTS items (
    id          SERIAL       PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index for fast lookups ordered by most-recently-updated (used by ListItems
-- to set the Last-Modified response header).
CREATE INDEX IF NOT EXISTS items_updated_at_idx ON items (updated_at DESC);

-- Index for name-based searching (future pagination / filtering support).
CREATE INDEX IF NOT EXISTS items_name_idx ON items (name);
