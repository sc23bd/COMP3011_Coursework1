package models

// SimulateRequest is the JSON payload accepted by the match-outcome simulator.
type SimulateRequest struct {
	// HomeTeamID is the ID of the team designated as "home" for the simulation.
	HomeTeamID int `json:"homeTeamId" binding:"required,min=1"`
	// AwayTeamID is the ID of the team designated as "away" for the simulation.
	AwayTeamID int `json:"awayTeamId" binding:"required,min=1"`
	// Date is an optional point-in-time date (YYYY-MM-DD) used to derive Elo
	// ratings from historical data.  Defaults to today when omitted.
	Date string `json:"date"`
	// Venue describes where the match is to be played: "home" (home team's
	// ground), "away" (away team's ground / home team away), or "neutral".
	// Defaults to "neutral" when omitted or invalid.
	Venue string `json:"venue"`
	// Simulations is the number of Monte Carlo iterations to run.  Values ≤ 0
	// default to 1 000; values > 10 000 are capped at 10 000.
	Simulations int `json:"simulations"`
}

// SimulationOutcome groups the win/draw/loss probability estimates and their
// 95 % Wilson-score confidence intervals.
type SimulationOutcome struct {
	HomeWinPct float64    `json:"homeWinPct"`
	DrawPct    float64    `json:"drawPct"`
	AwayWinPct float64    `json:"awayWinPct"`
	HomeWinCI  [2]float64 `json:"homeWinCI"`
	DrawCI     [2]float64 `json:"drawCI"`
	AwayWinCI  [2]float64 `json:"awayWinCI"`
}

// ExpectedScore holds the mean goals per iteration for each side.
type ExpectedScore struct {
	HomeGoals float64 `json:"homeGoals"`
	AwayGoals float64 `json:"awayGoals"`
}

// SimulateResponse is the response envelope returned by the simulate endpoint.
type SimulateResponse struct {
	HomeTeam         string            `json:"homeTeam"`
	AwayTeam         string            `json:"awayTeam"`
	Venue            string            `json:"venue"`
	AsOf             string            `json:"asOf"`
	Simulations      int               `json:"simulations"`
	HomeElo          float64           `json:"homeElo"`
	AwayElo          float64           `json:"awayElo"`
	Outcome          SimulationOutcome `json:"outcome"`
	ExpectedScore    ExpectedScore     `json:"expectedScore"`
	UpsetProbability float64           `json:"upsetProbability"`
	Links            []Link            `json:"links"`
}
