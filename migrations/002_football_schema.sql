-- Migration 002: International Football Results schema.
--
-- Apply with:
--   psql "$DATABASE_URL" -f migrations/002_football_schema.sql
--
-- This schema is idempotent; running it multiple times is safe.

-- Teams table: stores each unique national team.
CREATE TABLE IF NOT EXISTS football_teams (
    id         SERIAL       PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Tournaments table: stores each unique competition name.
CREATE TABLE IF NOT EXISTS football_tournaments (
    id         SERIAL       PRIMARY KEY,
    name       VARCHAR(200) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Matches table: stores international match results (from results.csv).
CREATE TABLE IF NOT EXISTS football_matches (
    id            SERIAL      PRIMARY KEY,
    match_date    DATE        NOT NULL,
    home_team_id  INTEGER     NOT NULL REFERENCES football_teams(id),
    away_team_id  INTEGER     NOT NULL REFERENCES football_teams(id),
    home_score    INTEGER     NOT NULL,
    away_score    INTEGER     NOT NULL,
    tournament_id INTEGER     NOT NULL REFERENCES football_tournaments(id),
    city          VARCHAR(100) NOT NULL DEFAULT '',
    country       VARCHAR(100) NOT NULL DEFAULT '',
    neutral       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Composite unique key to prevent duplicate match records.
    UNIQUE (match_date, home_team_id, away_team_id)
);

-- Goalscorers table: stores individual goal events (from goalscorers.csv).
CREATE TABLE IF NOT EXISTS football_goalscorers (
    id         SERIAL      PRIMARY KEY,
    match_id   INTEGER     NOT NULL REFERENCES football_matches(id),
    team_id    INTEGER     NOT NULL REFERENCES football_teams(id),
    scorer     VARCHAR(100) NOT NULL,
    own_goal   BOOLEAN     NOT NULL DEFAULT FALSE,
    penalty    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Shootouts table: stores penalty-shootout results (from shootouts.csv).
CREATE TABLE IF NOT EXISTS football_shootouts (
    id         SERIAL      PRIMARY KEY,
    match_id   INTEGER     NOT NULL REFERENCES football_matches(id) UNIQUE,
    winner_id  INTEGER     NOT NULL REFERENCES football_teams(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Former names table: tracks historical team name changes (from former_names.csv).
CREATE TABLE IF NOT EXISTS football_former_names (
    id           SERIAL      PRIMARY KEY,
    team_id      INTEGER     NOT NULL REFERENCES football_teams(id),
    former_name  VARCHAR(100) NOT NULL,
    start_date   DATE,
    end_date     DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance.
CREATE INDEX IF NOT EXISTS football_matches_date_idx       ON football_matches (match_date DESC);
CREATE INDEX IF NOT EXISTS football_matches_home_team_idx  ON football_matches (home_team_id);
CREATE INDEX IF NOT EXISTS football_matches_away_team_idx  ON football_matches (away_team_id);
CREATE INDEX IF NOT EXISTS football_matches_tournament_idx ON football_matches (tournament_id);
CREATE INDEX IF NOT EXISTS football_goalscorers_match_idx  ON football_goalscorers (match_id);
CREATE INDEX IF NOT EXISTS football_goalscorers_scorer_idx ON football_goalscorers (scorer);
CREATE INDEX IF NOT EXISTS football_former_names_team_idx  ON football_former_names (team_id);
