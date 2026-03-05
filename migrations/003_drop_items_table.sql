-- Migration 003: Drop the items table (and its indexes) introduced in 001.
--
-- The Item resource has been removed from the API.  This migration is safe
-- to apply on databases that were provisioned before this change.
--
-- Apply with:
--   psql "$DATABASE_URL" -f migrations/003_drop_items_table.sql
--
-- This migration is idempotent; running it multiple times is safe.

DROP INDEX IF EXISTS items_name_idx;
DROP INDEX IF EXISTS items_updated_at_idx;
DROP TABLE IF EXISTS items;
