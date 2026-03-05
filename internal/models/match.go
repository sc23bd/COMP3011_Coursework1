// Package models defines the data structures used throughout the API.
package models

import "time"

// Match represents an international football match result.
type Match struct {
	ID           int       `json:"id"`
	Date         time.Time `json:"date"`
	HomeTeam     string    `json:"homeTeam"`
	AwayTeam     string    `json:"awayTeam"`
	HomeTeamID   int       `json:"homeTeamId"`
	AwayTeamID   int       `json:"awayTeamId"`
	HomeScore    int       `json:"homeScore"`
	AwayScore    int       `json:"awayScore"`
	Tournament   string    `json:"tournament"`
	TournamentID int       `json:"tournamentId"`
	City         string    `json:"city"`
	Country      string    `json:"country"`
	Neutral      bool      `json:"neutral"`
}

// MatchResponse wraps a Match with hypermedia links (HATEOAS).
type MatchResponse struct {
	Match
	Links []Link `json:"links"`
}

// MatchesResponse wraps a list of matches with a collection-level link.
type MatchesResponse struct {
	Data  []MatchResponse `json:"data"`
	Links []Link          `json:"links"`
}

// Goal represents a single goal event in a match.
type Goal struct {
	ID      int    `json:"id"`
	MatchID int    `json:"matchId"`
	Team    string `json:"team"`
	TeamID  int    `json:"teamId"`
	Scorer  string `json:"scorer"`
	OwnGoal bool   `json:"ownGoal"`
	Penalty bool   `json:"penalty"`
}

// GoalsResponse wraps a list of goals with collection-level links.
type GoalsResponse struct {
	Data  []Goal `json:"data"`
	Links []Link `json:"links"`
}

// Shootout represents the penalty-shootout result for a match.
type Shootout struct {
	ID       int    `json:"id"`
	MatchID  int    `json:"matchId"`
	Winner   string `json:"winner"`
	WinnerID int    `json:"winnerId"`
}

// ShootoutResponse wraps a Shootout with hypermedia links (HATEOAS).
type ShootoutResponse struct {
	Shootout
	Links []Link `json:"links"`
}
