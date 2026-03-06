-- Migration 005: Elo Rating System
-- Adds cache and configuration tables for the dynamic Elo rating system.
-- This migration is idempotent: all statements use IF NOT EXISTS / ON CONFLICT.

-- Cache table for pre-computed Elo snapshots.
-- Populated on-demand and via the POST /rankings/elo/recalculate endpoint.
CREATE TABLE IF NOT EXISTS football_elo_cache (
    id              SERIAL PRIMARY KEY,
    team_id         INTEGER NOT NULL REFERENCES football_teams(id) ON DELETE CASCADE,
    as_of_date      DATE NOT NULL,
    elo_rating      NUMERIC(8,2) NOT NULL,
    global_rank     INTEGER,
    matches_played  INTEGER NOT NULL DEFAULT 0,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (team_id, as_of_date)
);

CREATE INDEX IF NOT EXISTS idx_elo_cache_date      ON football_elo_cache(as_of_date);
CREATE INDEX IF NOT EXISTS idx_elo_cache_team_date ON football_elo_cache(team_id, as_of_date);

-- Configuration table for Elo parameters (allows runtime tuning without redeployment).
CREATE TABLE IF NOT EXISTS football_elo_config (
    key         VARCHAR(50) PRIMARY KEY,
    value       JSONB NOT NULL,
    description TEXT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO football_elo_config (key, value, description) VALUES
  ('home_advantage',    '100',      'Points added to home team expected result calculation'),
  ('default_rating',    '1500',     'Starting Elo rating for teams with no prior matches'),
  ('k_factors',         '{"friendly":5,"qualifier":5,"nations_cup":5,"quarter_final":10,"semi_final":15,"final_non_wc":25,"world_cup":30,"world_cup_final":30}',
                                    'K multipliers by tournament type/stage'),
  ('goal_margin_factor','0.1',      'Multiplier applied to ln(|goal_diff|+1) for goal-margin adjustment')
ON CONFLICT (key) DO NOTHING;
