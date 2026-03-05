-- query/football.sql
-- SQL reference queries for the International Football Results feature.
-- These queries are hand-written and directly embedded in the repository
-- implementations in internal/db/postgres/football_repo.go.
-- They are collected here for documentation and review.

-- name: ListTeams :many
-- Returns all teams ordered alphabetically.
SELECT id, name, created_at
FROM football_teams
ORDER BY name ASC;

-- name: GetTeamByID :one
-- Returns a single team by primary key.
SELECT id, name, created_at
FROM football_teams
WHERE id = $1;

-- name: GetTeamHistory :many
-- Returns all former names for a team ordered by start date.
SELECT id, team_id, former_name, start_date, end_date
FROM football_former_names
WHERE team_id = $1
ORDER BY start_date ASC NULLS LAST;

-- name: ListMatches :many
-- Returns a paginated list of matches with denormalised team and tournament names.
SELECT
    m.id, m.match_date,
    ht.id AS home_team_id, ht.name AS home_team,
    at.id AS away_team_id, at.name AS away_team,
    m.home_score, m.away_score,
    t.id  AS tournament_id, t.name AS tournament,
    m.city, m.country, m.neutral
FROM football_matches m
JOIN football_teams      ht ON ht.id = m.home_team_id
JOIN football_teams      at ON at.id = m.away_team_id
JOIN football_tournaments t ON t.id  = m.tournament_id
ORDER BY m.match_date DESC
LIMIT $1 OFFSET $2;

-- name: GetMatchByID :one
-- Returns a single match with denormalised team and tournament names.
SELECT
    m.id, m.match_date,
    ht.id AS home_team_id, ht.name AS home_team,
    at.id AS away_team_id, at.name AS away_team,
    m.home_score, m.away_score,
    t.id  AS tournament_id, t.name AS tournament,
    m.city, m.country, m.neutral
FROM football_matches m
JOIN football_teams      ht ON ht.id = m.home_team_id
JOIN football_teams      at ON at.id = m.away_team_id
JOIN football_tournaments t ON t.id  = m.tournament_id
WHERE m.id = $1;

-- name: GetHeadToHead :many
-- Returns all matches between two teams (in either direction).
SELECT
    m.id, m.match_date,
    ht.id AS home_team_id, ht.name AS home_team,
    at.id AS away_team_id, at.name AS away_team,
    m.home_score, m.away_score,
    t.id  AS tournament_id, t.name AS tournament,
    m.city, m.country, m.neutral
FROM football_matches m
JOIN football_teams      ht ON ht.id = m.home_team_id
JOIN football_teams      at ON at.id = m.away_team_id
JOIN football_tournaments t ON t.id  = m.tournament_id
WHERE (m.home_team_id = $1 AND m.away_team_id = $2)
   OR (m.home_team_id = $2 AND m.away_team_id = $1)
ORDER BY m.match_date DESC;

-- name: GetMatchGoals :many
-- Returns all goals for a match.
SELECT g.id, g.match_id, t.id AS team_id, t.name AS team,
       g.scorer, g.own_goal, g.penalty
FROM football_goalscorers g
JOIN football_teams t ON t.id = g.team_id
WHERE g.match_id = $1
ORDER BY g.id ASC;

-- name: GetMatchShootout :one
-- Returns the shootout winner for a match.
SELECT s.id, s.match_id, t.id AS winner_id, t.name AS winner
FROM football_shootouts s
JOIN football_teams t ON t.id = s.winner_id
WHERE s.match_id = $1;

-- name: GetPlayerGoals :many
-- Returns all goals scored by a specific player.
SELECT g.id, g.match_id, t.id AS team_id, t.name AS team,
       g.scorer, g.own_goal, g.penalty
FROM football_goalscorers g
JOIN football_teams t ON t.id = g.team_id
WHERE g.scorer = $1
ORDER BY g.match_id ASC;
