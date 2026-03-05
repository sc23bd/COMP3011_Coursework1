package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// FootballRepo is a PostgreSQL-backed implementation of db.FootballRepository.
// All queries use parameterised placeholders ($1, $2, …) to prevent SQL injection.
type FootballRepo struct {
	db *sql.DB
}

// NewFootballRepo constructs a FootballRepo backed by the provided *sql.DB.
func NewFootballRepo(db *sql.DB) *FootballRepo {
	return &FootballRepo{db: db}
}

// ListTeams returns all teams ordered alphabetically.
func (r *FootballRepo) ListTeams() ([]models.Team, error) {
	const q = `
		SELECT id, name, created_at
		FROM football_teams
		ORDER BY name ASC`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("footballRepo.ListTeams: %w", err)
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("footballRepo.ListTeams scan: %w", err)
		}
		teams = append(teams, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("footballRepo.ListTeams rows: %w", err)
	}
	return teams, nil
}

// GetTeamByID returns the team with the given ID.
// Returns ErrNotFound when no matching row exists.
func (r *FootballRepo) GetTeamByID(id int) (models.Team, error) {
	const q = `SELECT id, name, created_at FROM football_teams WHERE id = $1`

	var t models.Team
	err := r.db.QueryRow(q, id).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Team{}, models.ErrNotFound
	}
	if err != nil {
		return models.Team{}, fmt.Errorf("footballRepo.GetTeamByID: %w", err)
	}
	return t, nil
}

// GetTeamHistory returns the former names recorded for a team.
func (r *FootballRepo) GetTeamHistory(teamID int) ([]models.FormerName, error) {
	const q = `
		SELECT id, team_id, former_name, start_date, end_date
		FROM football_former_names
		WHERE team_id = $1
		ORDER BY start_date ASC NULLS LAST`

	rows, err := r.db.Query(q, teamID)
	if err != nil {
		return nil, fmt.Errorf("footballRepo.GetTeamHistory: %w", err)
	}
	defer rows.Close()

	var history []models.FormerName
	for rows.Next() {
		var fn models.FormerName
		var start, end sql.NullTime
		if err := rows.Scan(&fn.ID, &fn.TeamID, &fn.FormerName, &start, &end); err != nil {
			return nil, fmt.Errorf("footballRepo.GetTeamHistory scan: %w", err)
		}
		if start.Valid {
			t := start.Time
			fn.StartDate = &t
		}
		if end.Valid {
			t := end.Time
			fn.EndDate = &t
		}
		history = append(history, fn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("footballRepo.GetTeamHistory rows: %w", err)
	}
	return history, nil
}

// ListMatches returns a paginated list of matches ordered by date descending.
func (r *FootballRepo) ListMatches(limit, offset int) ([]models.Match, error) {
	const q = `
		SELECT
			m.id, m.match_date,
			ht.id, ht.name,
			at.id, at.name,
			m.home_score, m.away_score,
			t.id, t.name,
			m.city, m.country, m.neutral
		FROM football_matches m
		JOIN football_teams ht      ON ht.id = m.home_team_id
		JOIN football_teams at      ON at.id = m.away_team_id
		JOIN football_tournaments t ON t.id  = m.tournament_id
		ORDER BY m.match_date DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("footballRepo.ListMatches: %w", err)
	}
	defer rows.Close()

	return scanMatchRows(rows)
}

// GetMatchByID returns the match with the given ID.
// Returns ErrNotFound when no matching row exists.
func (r *FootballRepo) GetMatchByID(id int) (models.Match, error) {
	const q = `
		SELECT
			m.id, m.match_date,
			ht.id, ht.name,
			at.id, at.name,
			m.home_score, m.away_score,
			t.id, t.name,
			m.city, m.country, m.neutral
		FROM football_matches m
		JOIN football_teams ht      ON ht.id = m.home_team_id
		JOIN football_teams at      ON at.id = m.away_team_id
		JOIN football_tournaments t ON t.id  = m.tournament_id
		WHERE m.id = $1`

	var m models.Match
	var matchDate time.Time
	err := r.db.QueryRow(q, id).Scan(
		&m.ID, &matchDate,
		&m.HomeTeamID, &m.HomeTeam,
		&m.AwayTeamID, &m.AwayTeam,
		&m.HomeScore, &m.AwayScore,
		&m.TournamentID, &m.Tournament,
		&m.City, &m.Country, &m.Neutral,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Match{}, models.ErrNotFound
	}
	if err != nil {
		return models.Match{}, fmt.Errorf("footballRepo.GetMatchByID: %w", err)
	}
	m.Date = matchDate
	return m, nil
}

// GetHeadToHead returns all matches between two teams ordered by date descending.
func (r *FootballRepo) GetHeadToHead(teamA, teamB int) ([]models.Match, error) {
	const q = `
		SELECT
			m.id, m.match_date,
			ht.id, ht.name,
			at.id, at.name,
			m.home_score, m.away_score,
			t.id, t.name,
			m.city, m.country, m.neutral
		FROM football_matches m
		JOIN football_teams ht      ON ht.id = m.home_team_id
		JOIN football_teams at      ON at.id = m.away_team_id
		JOIN football_tournaments t ON t.id  = m.tournament_id
		WHERE (m.home_team_id = $1 AND m.away_team_id = $2)
		   OR (m.home_team_id = $2 AND m.away_team_id = $1)
		ORDER BY m.match_date DESC`

	rows, err := r.db.Query(q, teamA, teamB)
	if err != nil {
		return nil, fmt.Errorf("footballRepo.GetHeadToHead: %w", err)
	}
	defer rows.Close()

	return scanMatchRows(rows)
}

// GetMatchGoals returns all goals recorded for a match.
func (r *FootballRepo) GetMatchGoals(matchID int) ([]models.Goal, error) {
	const q = `
		SELECT g.id, g.match_id, t.id, t.name, g.scorer, g.own_goal, g.penalty
		FROM football_goalscorers g
		JOIN football_teams t ON t.id = g.team_id
		WHERE g.match_id = $1
		ORDER BY g.id ASC`

	rows, err := r.db.Query(q, matchID)
	if err != nil {
		return nil, fmt.Errorf("footballRepo.GetMatchGoals: %w", err)
	}
	defer rows.Close()

	return scanGoalRows(rows)
}

// GetMatchShootout returns the shootout result for a match.
// Returns ErrNotFound when the match had no shootout.
func (r *FootballRepo) GetMatchShootout(matchID int) (models.Shootout, error) {
	const q = `
		SELECT s.id, s.match_id, t.id, t.name
		FROM football_shootouts s
		JOIN football_teams t ON t.id = s.winner_id
		WHERE s.match_id = $1`

	var s models.Shootout
	err := r.db.QueryRow(q, matchID).Scan(&s.ID, &s.MatchID, &s.WinnerID, &s.Winner)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Shootout{}, models.ErrNotFound
	}
	if err != nil {
		return models.Shootout{}, fmt.Errorf("footballRepo.GetMatchShootout: %w", err)
	}
	return s, nil
}

// GetPlayerGoals returns all goals scored by the named player.
func (r *FootballRepo) GetPlayerGoals(scorer string) ([]models.Goal, error) {
	const q = `
		SELECT g.id, g.match_id, t.id, t.name, g.scorer, g.own_goal, g.penalty
		FROM football_goalscorers g
		JOIN football_teams t ON t.id = g.team_id
		WHERE g.scorer = $1
		ORDER BY g.match_id ASC`

	rows, err := r.db.Query(q, scorer)
	if err != nil {
		return nil, fmt.Errorf("footballRepo.GetPlayerGoals: %w", err)
	}
	defer rows.Close()

	return scanGoalRows(rows)
}

// --- scan helpers ------------------------------------------------------------

// scanMatchRows reads Match rows from a *sql.Rows cursor.
func scanMatchRows(rows *sql.Rows) ([]models.Match, error) {
	var matches []models.Match
	for rows.Next() {
		var m models.Match
		var matchDate time.Time
		if err := rows.Scan(
			&m.ID, &matchDate,
			&m.HomeTeamID, &m.HomeTeam,
			&m.AwayTeamID, &m.AwayTeam,
			&m.HomeScore, &m.AwayScore,
			&m.TournamentID, &m.Tournament,
			&m.City, &m.Country, &m.Neutral,
		); err != nil {
			return nil, fmt.Errorf("scanMatchRows: %w", err)
		}
		m.Date = matchDate
		matches = append(matches, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanMatchRows rows: %w", err)
	}
	return matches, nil
}

// scanGoalRows reads Goal rows from a *sql.Rows cursor.
func scanGoalRows(rows *sql.Rows) ([]models.Goal, error) {
	var goals []models.Goal
	for rows.Next() {
		var g models.Goal
		if err := rows.Scan(&g.ID, &g.MatchID, &g.TeamID, &g.Team, &g.Scorer, &g.OwnGoal, &g.Penalty); err != nil {
			return nil, fmt.Errorf("scanGoalRows: %w", err)
		}
		goals = append(goals, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanGoalRows rows: %w", err)
	}
	return goals, nil
}
