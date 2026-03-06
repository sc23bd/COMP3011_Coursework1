package postgres

import (
	"fmt"
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/elo"
)

// GetMatchesChronological returns matches involving teamID (or all matches when
// teamID == 0) up to and including endDate, ordered oldest-first.
func (r *FootballRepo) GetMatchesChronological(teamID int, endDate time.Time) ([]elo.MatchResult, error) {
	var (
		rows interface{ Close() error }
		err  error
	)

	if teamID == 0 {
		const q = `
			SELECT m.id, m.match_date,
			       m.home_team_id, m.away_team_id,
			       m.home_score, m.away_score,
			       t.name, m.neutral
			FROM football_matches m
			JOIN football_tournaments t ON t.id = m.tournament_id
			WHERE m.match_date <= $1
			ORDER BY m.match_date ASC, m.id ASC`
		sqlRows, qErr := r.db.Query(q, endDate)
		if qErr != nil {
			return nil, fmt.Errorf("eloRepo.GetMatchesChronological(all): %w", qErr)
		}
		rows = sqlRows
		err = nil
		defer sqlRows.Close()
		return scanMatchResults(sqlRows)
	}

	_ = rows
	_ = err

	const q = `
		SELECT m.id, m.match_date,
		       m.home_team_id, m.away_team_id,
		       m.home_score, m.away_score,
		       t.name, m.neutral
		FROM football_matches m
		JOIN football_tournaments t ON t.id = m.tournament_id
		WHERE (m.home_team_id = $1 OR m.away_team_id = $1)
		  AND m.match_date <= $2
		ORDER BY m.match_date ASC, m.id ASC`

	sqlRows, qErr := r.db.Query(q, teamID, endDate)
	if qErr != nil {
		return nil, fmt.Errorf("eloRepo.GetMatchesChronological(team=%d): %w", teamID, qErr)
	}
	defer sqlRows.Close()
	return scanMatchResults(sqlRows)
}

// SaveEloSnapshot upserts a cached Elo snapshot for one team on one date.
func (r *FootballRepo) SaveEloSnapshot(teamID int, asOf time.Time, rating float64, rank int, matchesPlayed int) error {
	const q = `
		INSERT INTO football_elo_cache (team_id, as_of_date, elo_rating, global_rank, matches_played, computed_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (team_id, as_of_date)
		DO UPDATE SET
		    elo_rating     = EXCLUDED.elo_rating,
		    global_rank    = EXCLUDED.global_rank,
		    matches_played = EXCLUDED.matches_played,
		    computed_at    = NOW()`

	_, err := r.db.Exec(q, teamID, asOf, rating, rank, matchesPlayed)
	if err != nil {
		return fmt.Errorf("eloRepo.SaveEloSnapshot: %w", err)
	}
	return nil
}

// GetEloRankings returns paginated Elo ranking entries computed from cached snapshots.
// When no cached data exists for the requested date the function returns an empty slice.
// region is currently unused (reserved for future filtering by confederation).
func (r *FootballRepo) GetEloRankings(asOf time.Time, _ string, limit, offset int) ([]elo.RankingEntry, error) {
	const q = `
		SELECT c.global_rank, c.team_id, ft.name, c.elo_rating
		FROM football_elo_cache c
		JOIN football_teams ft ON ft.id = c.team_id
		WHERE c.as_of_date = $1
		ORDER BY c.elo_rating DESC
		LIMIT $2 OFFSET $3`

	sqlRows, err := r.db.Query(q, asOf, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("eloRepo.GetEloRankings: %w", err)
	}
	defer sqlRows.Close()

	var entries []elo.RankingEntry
	for sqlRows.Next() {
		var e elo.RankingEntry
		var rank *int
		if err := sqlRows.Scan(&rank, &e.TeamID, &e.TeamName, &e.Elo); err != nil {
			return nil, fmt.Errorf("eloRepo.GetEloRankings scan: %w", err)
		}
		if rank != nil {
			e.Rank = *rank
		}
		entries = append(entries, e)
	}
	if err := sqlRows.Err(); err != nil {
		return nil, fmt.Errorf("eloRepo.GetEloRankings rows: %w", err)
	}
	return entries, nil
}

// --- helpers -----------------------------------------------------------------

// scanMatchResults reads elo.MatchResult rows from an open *sql.Rows cursor.
func scanMatchResults(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}) ([]elo.MatchResult, error) {
	var results []elo.MatchResult
	for rows.Next() {
		var mr elo.MatchResult
		var matchDate time.Time
		if err := rows.Scan(
			&mr.MatchID, &matchDate,
			&mr.HomeTeamID, &mr.AwayTeamID,
			&mr.HomeScore, &mr.AwayScore,
			&mr.Tournament, &mr.Neutral,
		); err != nil {
			return nil, fmt.Errorf("scanMatchResults: %w", err)
		}
		mr.Date = matchDate
		results = append(results, mr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanMatchResults rows: %w", err)
	}
	return results, nil
}
