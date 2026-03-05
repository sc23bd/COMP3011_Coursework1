//go:build ignore
// +build ignore

// import_football_data.go is a standalone script that downloads and imports the
// "International football results from 1872 to 2025" dataset from Kaggle into
// the PostgreSQL database.
//
// Prerequisites:
//   - DATABASE_URL environment variable must point to a running PostgreSQL instance
//     with the football schema already applied (migrations/002_football_schema.sql).
//   - KAGGLE_USERNAME and KAGGLE_KEY environment variables must be set with valid
//     Kaggle API credentials, OR the ZIP file can be placed at ./football_data.zip
//     to skip the download step.
//
// Usage:
//
//	go run scripts/import_football_data.go
//
// The script is idempotent: running it multiple times will not create duplicates
// because INSERT … ON CONFLICT DO NOTHING is used throughout.
package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	// kaggleDownloadURL is the Kaggle API endpoint for the dataset.
	// Note: the slug contains "1872-to-2017" because that was the original
	// dataset name; the actual data has been updated and covers up to 2025.
	kaggleDownloadURL = "https://www.kaggle.com/api/v1/datasets/download/martj42/international-football-results-from-1872-to-2017"
	localZipPath      = "./football_data.zip"
	downloadTimeout   = 5 * time.Minute
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("import failed: %v", err)
	}
}

func run() error {
	// --- Connect to the database -------------------------------------------
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Connected to database")

	// --- Obtain the ZIP file -----------------------------------------------
	zipData, err := getZipData()
	if err != nil {
		return fmt.Errorf("failed to get ZIP data: %w", err)
	}
	log.Printf("ZIP data ready (%d bytes)", len(zipData))

	// --- Parse CSV files ---------------------------------------------------
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to open ZIP: %w", err)
	}

	csvFiles := make(map[string][]byte)
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		// We only need the four known CSV files.
		switch name {
		case "results.csv", "goalscorers.csv", "shootouts.csv", "former_names.csv":
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open %s in ZIP: %w", f.Name, err)
			}
			data, readErr := io.ReadAll(rc)
			rc.Close()
			if readErr != nil {
				return fmt.Errorf("failed to read %s: %w", f.Name, readErr)
			}
			csvFiles[name] = data
			log.Printf("  Read %s (%d bytes)", f.Name, len(data))
		}
	}

	// --- Import data -------------------------------------------------------
	// All imports run inside a single database transaction so that a partial
	// failure leaves the database unchanged.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			log.Printf("transaction rolled back due to error: %v", err)
		}
	}()

	// Step 1: Insert unique teams and tournaments from results.csv.
	// We build in-memory maps (name → id) to avoid repeated round-trips for
	// each match row during the match-insertion step.
	teamIDs, tournamentIDs, err := insertTeamsAndTournaments(tx, csvFiles["results.csv"])
	if err != nil {
		return fmt.Errorf("failed to insert teams/tournaments: %w", err)
	}
	log.Printf("Teams: %d unique | Tournaments: %d unique", len(teamIDs), len(tournamentIDs))

	// Step 2: Insert matches and build a match-key → id map for child tables.
	matchIDs, err := insertMatches(tx, csvFiles["results.csv"], teamIDs, tournamentIDs)
	if err != nil {
		return fmt.Errorf("failed to insert matches: %w", err)
	}
	log.Printf("Matches imported: %d rows", len(matchIDs))

	// Step 3: Import goalscorers.csv (optional — file may be absent).
	if data, ok := csvFiles["goalscorers.csv"]; ok {
		count, err := insertGoalscorers(tx, data, matchIDs, teamIDs)
		if err != nil {
			return fmt.Errorf("failed to insert goalscorers: %w", err)
		}
		log.Printf("Goals imported: %d rows", count)
	}

	// Step 4: Import shootouts.csv (optional — file may be absent).
	if data, ok := csvFiles["shootouts.csv"]; ok {
		count, err := insertShootouts(tx, data, matchIDs, teamIDs)
		if err != nil {
			return fmt.Errorf("failed to insert shootouts: %w", err)
		}
		log.Printf("Shootouts imported: %d rows", count)
	}

	// Step 5: Import former_names.csv (optional — file may be absent).
	if data, ok := csvFiles["former_names.csv"]; ok {
		// Ensure any team in former_names.csv that is not already in teamIDs
		// gets inserted first.
		if err = ensureFormerNameTeams(tx, data, teamIDs); err != nil {
			return fmt.Errorf("failed to ensure former-name teams: %w", err)
		}
		count, err := insertFormerNames(tx, data, teamIDs)
		if err != nil {
			return fmt.Errorf("failed to insert former names: %w", err)
		}
		log.Printf("Former names imported: %d rows", count)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Println("Import completed successfully")
	return nil
}

// --- download / file helpers -------------------------------------------------

// getZipData returns the raw bytes of the dataset ZIP.
func getZipData() ([]byte, error) {
	if _, err := os.Stat(localZipPath); err != nil {
		log.Printf("Local File not found")
		return nil, err
	}
	log.Printf("Using local ZIP file: %s", localZipPath)
	return os.ReadFile(localZipPath)
}

// --- CSV helpers -------------------------------------------------------------

// parseCSV reads all records from raw CSV bytes, stripping the header row.
func parseCSV(data []byte) (header []string, records [][]string, err error) {
	r := csv.NewReader(bytes.NewReader(data))
	header, err = r.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("reading CSV header: %w", err)
	}
	records, err = r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("reading CSV records: %w", err)
	}
	return header, records, nil
}

// colIndex returns the column index for a given header name.
func colIndex(header []string, name string) int {
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), name) {
			return i
		}
	}
	return -1
}

// parseBool converts "TRUE"/"FALSE" CSV strings to a Go bool.
func parseBool(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "true")
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(s))
}

// matchKey produces a unique string key for a match based on date + teams.
type matchKey struct {
	date     string
	homeTeam string
	awayTeam string
}

// --- insertion functions -----------------------------------------------------

// insertTeamsAndTournaments inserts unique team and tournament names from
// results.csv and returns maps of name → database ID.
func insertTeamsAndTournaments(tx *sql.Tx, data []byte) (map[string]int, map[string]int, error) {
	header, records, err := parseCSV(data)
	if err != nil {
		return nil, nil, err
	}

	homeCol := colIndex(header, "home_team")
	awayCol := colIndex(header, "away_team")
	tournCol := colIndex(header, "tournament")

	// Collect unique names.
	teamSet := make(map[string]struct{})
	tournSet := make(map[string]struct{})
	for _, row := range records {
		if homeCol >= 0 && homeCol < len(row) {
			teamSet[strings.TrimSpace(row[homeCol])] = struct{}{}
		}
		if awayCol >= 0 && awayCol < len(row) {
			teamSet[strings.TrimSpace(row[awayCol])] = struct{}{}
		}
		if tournCol >= 0 && tournCol < len(row) {
			tournSet[strings.TrimSpace(row[tournCol])] = struct{}{}
		}
	}

	// Insert teams.
	teamIDs := make(map[string]int, len(teamSet))
	for name := range teamSet {
		if name == "" {
			continue
		}
		var id int
		err := tx.QueryRow(
			`INSERT INTO football_teams (name) VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`, name).Scan(&id)
		if err != nil {
			return nil, nil, fmt.Errorf("inserting team %q: %w", name, err)
		}
		teamIDs[name] = id
	}

	// Insert tournaments.
	tournamentIDs := make(map[string]int, len(tournSet))
	for name := range tournSet {
		if name == "" {
			continue
		}
		var id int
		err := tx.QueryRow(
			`INSERT INTO football_tournaments (name) VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`, name).Scan(&id)
		if err != nil {
			return nil, nil, fmt.Errorf("inserting tournament %q: %w", name, err)
		}
		tournamentIDs[name] = id
	}

	return teamIDs, tournamentIDs, nil
}

// insertMatches inserts match records from results.csv.
// Returns a map of matchKey → database match ID.
func insertMatches(tx *sql.Tx, data []byte, teamIDs, tournamentIDs map[string]int) (map[matchKey]int, error) {
	header, records, err := parseCSV(data)
	if err != nil {
		return nil, err
	}

	dateCol := colIndex(header, "date")
	homeCol := colIndex(header, "home_team")
	awayCol := colIndex(header, "away_team")
	homeScoreCol := colIndex(header, "home_score")
	awayScoreCol := colIndex(header, "away_score")
	tournCol := colIndex(header, "tournament")
	cityCol := colIndex(header, "city")
	countryCol := colIndex(header, "country")
	neutralCol := colIndex(header, "neutral")

	matchIDs := make(map[matchKey]int, len(records))

	const q = `
		INSERT INTO football_matches
			(match_date, home_team_id, away_team_id, home_score, away_score,
			 tournament_id, city, country, neutral)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (match_date, home_team_id, away_team_id) DO NOTHING
		RETURNING id`

	const selectID = `
		SELECT id FROM football_matches
		WHERE match_date = $1 AND home_team_id = $2 AND away_team_id = $3`

	for _, row := range records {
		get := func(col int) string {
			if col < 0 || col >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[col])
		}

		dateStr := get(dateCol)
		date, err := parseDate(dateStr)
		if err != nil {
			log.Printf("WARN: skipping row with unparseable date %q: %v", dateStr, err)
			continue
		}

		homeTeam := get(homeCol)
		awayTeam := get(awayCol)
		tournament := get(tournCol)

		homeTeamID, ok := teamIDs[homeTeam]
		if !ok {
			log.Printf("WARN: unknown home team %q, skipping", homeTeam)
			continue
		}
		awayTeamID, ok := teamIDs[awayTeam]
		if !ok {
			log.Printf("WARN: unknown away team %q, skipping", awayTeam)
			continue
		}
		tournamentID, ok := tournamentIDs[tournament]
		if !ok {
			log.Printf("WARN: unknown tournament %q, skipping", tournament)
			continue
		}

		homeScore, _ := strconv.Atoi(get(homeScoreCol))
		awayScore, _ := strconv.Atoi(get(awayScoreCol))
		neutral := parseBool(get(neutralCol))

		var id int
		err = tx.QueryRow(q,
			date.Format("2006-01-02"),
			homeTeamID, awayTeamID,
			homeScore, awayScore,
			tournamentID,
			get(cityCol), get(countryCol),
			neutral,
		).Scan(&id)
		if err == sql.ErrNoRows {
			// Row already existed; fetch its ID so downstream inserts can reference it.
			err = tx.QueryRow(selectID,
				date.Format("2006-01-02"), homeTeamID, awayTeamID,
			).Scan(&id)
		}
		if err != nil {
			return nil, fmt.Errorf("inserting match %s %s vs %s: %w", dateStr, homeTeam, awayTeam, err)
		}

		matchIDs[matchKey{dateStr, homeTeam, awayTeam}] = id
	}

	return matchIDs, nil
}

// insertGoalscorers inserts goal records from goalscorers.csv.
func insertGoalscorers(tx *sql.Tx, data []byte, matchIDs map[matchKey]int, teamIDs map[string]int) (int, error) {
	header, records, err := parseCSV(data)
	if err != nil {
		return 0, err
	}

	dateCol := colIndex(header, "date")
	homeCol := colIndex(header, "home_team")
	awayCol := colIndex(header, "away_team")
	teamCol := colIndex(header, "team")
	scorerCol := colIndex(header, "scorer")
	ownGoalCol := colIndex(header, "own_goal")
	penaltyCol := colIndex(header, "penalty")

	const q = `
		INSERT INTO football_goalscorers (match_id, team_id, scorer, own_goal, penalty)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING`

	count := 0
	for _, row := range records {
		get := func(col int) string {
			if col < 0 || col >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[col])
		}

		key := matchKey{get(dateCol), get(homeCol), get(awayCol)}
		matchID, ok := matchIDs[key]
		if !ok {
			log.Printf("WARN: goalscorer row references unknown match %v, skipping", key)
			continue
		}

		teamName := get(teamCol)
		teamID, ok := teamIDs[teamName]
		if !ok {
			log.Printf("WARN: goalscorer row references unknown team %q, skipping", teamName)
			continue
		}

		scorer := get(scorerCol)
		ownGoal := parseBool(get(ownGoalCol))
		penalty := parseBool(get(penaltyCol))

		_, err := tx.Exec(q, matchID, teamID, scorer, ownGoal, penalty)
		if err != nil {
			return count, fmt.Errorf("inserting goal (scorer=%q): %w", scorer, err)
		}
		count++
	}
	return count, nil
}

// insertShootouts inserts penalty-shootout records from shootouts.csv.
func insertShootouts(tx *sql.Tx, data []byte, matchIDs map[matchKey]int, teamIDs map[string]int) (int, error) {
	header, records, err := parseCSV(data)
	if err != nil {
		return 0, err
	}

	dateCol := colIndex(header, "date")
	homeCol := colIndex(header, "home_team")
	awayCol := colIndex(header, "away_team")
	winnerCol := colIndex(header, "winner")

	const q = `
		INSERT INTO football_shootouts (match_id, winner_id)
		VALUES ($1, $2)
		ON CONFLICT (match_id) DO NOTHING`

	count := 0
	for _, row := range records {
		get := func(col int) string {
			if col < 0 || col >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[col])
		}

		key := matchKey{get(dateCol), get(homeCol), get(awayCol)}
		matchID, ok := matchIDs[key]
		if !ok {
			log.Printf("WARN: shootout row references unknown match %v, skipping", key)
			continue
		}

		winnerName := get(winnerCol)
		winnerID, ok := teamIDs[winnerName]
		if !ok {
			log.Printf("WARN: shootout row references unknown winner team %q, skipping", winnerName)
			continue
		}

		_, err := tx.Exec(q, matchID, winnerID)
		if err != nil {
			return count, fmt.Errorf("inserting shootout (matchID=%d): %w", matchID, err)
		}
		count++
	}
	return count, nil
}

// ensureFormerNameTeams inserts any teams from former_names.csv that are not
// already present in teamIDs (the "current" name column references new teams
// that may not appear in results.csv).
func ensureFormerNameTeams(tx *sql.Tx, data []byte, teamIDs map[string]int) error {
	header, records, err := parseCSV(data)
	if err != nil {
		return err
	}

	currentCol := colIndex(header, "current")

	for _, row := range records {
		if currentCol < 0 || currentCol >= len(row) {
			continue
		}
		name := strings.TrimSpace(row[currentCol])
		if name == "" {
			continue
		}
		if _, exists := teamIDs[name]; exists {
			continue
		}
		var id int
		err := tx.QueryRow(
			`INSERT INTO football_teams (name) VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`, name).Scan(&id)
		if err != nil {
			return fmt.Errorf("inserting team %q: %w", name, err)
		}
		teamIDs[name] = id
	}
	return nil
}

// insertFormerNames inserts historical team-name records from former_names.csv.
func insertFormerNames(tx *sql.Tx, data []byte, teamIDs map[string]int) (int, error) {
	header, records, err := parseCSV(data)
	if err != nil {
		return 0, err
	}

	currentCol := colIndex(header, "current")
	formerCol := colIndex(header, "former")
	startCol := colIndex(header, "start_date")
	endCol := colIndex(header, "end_date")

	const q = `
		INSERT INTO football_former_names (team_id, former_name, start_date, end_date)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`

	count := 0
	for _, row := range records {
		get := func(col int) string {
			if col < 0 || col >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[col])
		}

		currentName := get(currentCol)
		teamID, ok := teamIDs[currentName]
		if !ok {
			log.Printf("WARN: former_names row references unknown team %q, skipping", currentName)
			continue
		}

		formerName := get(formerCol)
		if formerName == "" {
			continue
		}

		// Dates are optional — store as NULL when empty.
		var startDate, endDate interface{}
		if s := get(startCol); s != "" {
			if t, err := parseDate(s); err == nil {
				startDate = t.Format("2006-01-02")
			}
		}
		if s := get(endCol); s != "" {
			if t, err := parseDate(s); err == nil {
				endDate = t.Format("2006-01-02")
			}
		}

		_, err := tx.Exec(q, teamID, formerName, startDate, endDate)
		if err != nil {
			return count, fmt.Errorf("inserting former name %q for team %q: %w", formerName, currentName, err)
		}
		count++
	}
	return count, nil
}
