// Package elo provides the data types used by the dynamic Elo rating system.
package elo

import (
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// Rating holds the Elo rating and supporting metadata for one team at a point in time.
type Rating struct {
	TeamID           int       `json:"teamId"`
	TeamName         string    `json:"teamName"`
	Date             time.Time `json:"date"`
	Elo              float64   `json:"elo"`
	Rank             int       `json:"rank,omitempty"`
	ChangeFromPrev   float64   `json:"changeFromPrevious"`
	MatchesConsidered int      `json:"matchesConsidered"`
	Methodology      Methodology `json:"methodology"`
	Links            []models.Link `json:"links"`
}

// Methodology describes the parameters used to produce a rating.
type Methodology struct {
	KFactor          float64 `json:"kFactor"`
	HomeAdvantage    float64 `json:"homeAdvantage"`
	WeightMultiplier float64 `json:"weightMultiplier"`
	FormulaReference string  `json:"formulaReference"`
}

// TimelineEntry represents the Elo value at a single point in the team's history.
type TimelineEntry struct {
	Date      time.Time `json:"date"`
	Elo       float64   `json:"elo"`
	Change    float64   `json:"change"`
	MatchID   int       `json:"matchId,omitempty"`
	Opponent  string    `json:"opponent,omitempty"`
	Result    string    `json:"result,omitempty"`   // "W", "D", or "L"
	HomeAway  string    `json:"homeAway,omitempty"` // "H", "A", or "N"
}

// TimelineResponse wraps a slice of TimelineEntry values with HATEOAS links.
type TimelineResponse struct {
	TeamID   int             `json:"teamId"`
	TeamName string          `json:"teamName"`
	Data     []TimelineEntry `json:"data"`
	Links    []models.Link   `json:"links"`
}

// RankingEntry is one row in the global Elo rankings snapshot.
type RankingEntry struct {
	Rank     int           `json:"rank"`
	TeamID   int           `json:"teamId"`
	TeamName string        `json:"teamName"`
	Elo      float64       `json:"elo"`
	Links    []models.Link `json:"links"`
}

// RankingsResponse is the paginated response for the global rankings endpoint.
type RankingsResponse struct {
	Date   string         `json:"date"`
	Data   []RankingEntry `json:"data"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Links  []models.Link  `json:"links"`
}

// RecalculateResponse is returned by the background-recalculation trigger endpoint.
type RecalculateResponse struct {
	Message string        `json:"message"`
	Links   []models.Link `json:"links"`
}

// MatchResult carries the information needed to update ratings after one match.
type MatchResult struct {
	MatchID    int
	Date       time.Time
	HomeTeamID int
	AwayTeamID int
	HomeScore  int
	AwayScore  int
	Tournament string
	Neutral    bool
}
