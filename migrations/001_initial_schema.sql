-- Migration 001: Initial schema — users table.
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
