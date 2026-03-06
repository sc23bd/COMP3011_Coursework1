-- Migration 004: Add unique constraint to football_goalscorers to prevent duplicates.
--
-- Apply with:
--   psql "$DATABASE_URL" -f migrations/004_add_goalscorers_unique_constraint.sql
--
-- This prevents the same goal from being inserted multiple times when the
-- import script is run repeatedly.

-- First, remove any existing duplicates (keeping the first occurrence of each).
DELETE FROM football_goalscorers
WHERE id NOT IN (
    SELECT MIN(id)
    FROM football_goalscorers
    GROUP BY match_id, team_id, scorer, own_goal, penalty
);

-- Add the unique constraint.
ALTER TABLE football_goalscorers
ADD CONSTRAINT football_goalscorers_unique_goal 
UNIQUE (match_id, team_id, scorer, own_goal, penalty);
