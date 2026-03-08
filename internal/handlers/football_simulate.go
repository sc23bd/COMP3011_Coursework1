package handlers

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/elo"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
	"github.com/sc23bd/COMP3011_Coursework1/internal/simulator"
)
const simulateDateLayout = "2006-01-02"

// concurrencyLimiter tracks concurrent simulation requests for rate limiting.
// A global limit prevents resource exhaustion from parallel heavy workloads.
var concurrencyLimiter struct {
	mu         sync.Mutex
	concurrent int // currently running simulations
}

// maxConcurrentSimulations is the maximum number of simulations allowed to run
// at the same time across all goroutines.
const maxConcurrentSimulations = 5

// SimulateMatch handles POST /api/v1/football/matches/simulate
//
//	@Summary		Simulate a match outcome
//	@Description	Runs a Monte Carlo simulation for a potential match result using historical Elo ratings and scoring patterns.
//	@Tags			matches
//	@Accept			json
//	@Produce		json
//	@Param			body	body		models.SimulateRequest	true	"Simulation parameters"
//	@Success		200		{object}	models.SimulateResponse	"Simulation result"
//	@Failure		400		{object}	models.ErrorResponse	"Invalid input"
//	@Failure		401		{object}	models.ErrorResponse	"Unauthorized"
//	@Failure		429		{object}	models.ErrorResponse	"Too many concurrent simulation requests"
//	@Failure		500		{object}	models.ErrorResponse	"Internal server error"
//	@Security		Bearer
//	@Router			/football/matches/simulate [post]
func (h *FootballHandler) SimulateMatch(c *gin.Context) {
	var req models.SimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	if req.HomeTeamID == req.AwayTeamID {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "home and away teams must be different"})
		return
	}

	// Rate-limit: reject when the concurrency cap is reached.
	concurrencyLimiter.mu.Lock()
	if concurrencyLimiter.concurrent >= maxConcurrentSimulations {
		concurrencyLimiter.mu.Unlock()
		c.Header("Cache-Control", "no-store")
		c.Header("Retry-After", "1")
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Error: "too many concurrent simulation requests; please retry shortly",
		})
		return
	}
	concurrencyLimiter.concurrent++
	concurrencyLimiter.mu.Unlock()

	defer func() {
		concurrencyLimiter.mu.Lock()
		concurrencyLimiter.concurrent--
		concurrencyLimiter.mu.Unlock()
	}()

	// Look up both teams.
	homeTeam, err := h.repo.GetTeamByID(req.HomeTeamID)
	if errors.Is(err, models.ErrNotFound) {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "home team not found"})
		return
	}
	if err != nil {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	awayTeam, err := h.repo.GetTeamByID(req.AwayTeamID)
	if errors.Is(err, models.ErrNotFound) {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "away team not found"})
		return
	}
	if err != nil {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Resolve the point-in-time date for Elo calculations.
	asOf := time.Now().UTC()
	dateStr := req.Date
	if dateStr != "" {
		parsed, parseErr := time.Parse(simulateDateLayout, dateStr)
		if parseErr != nil {
			c.Header("Cache-Control", "no-store")
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid date format; expected YYYY-MM-DD"})
			return
		}
		asOf = parsed
	} else {
		dateStr = asOf.Format(simulateDateLayout)
	}

	cfg := elo.DefaultConfig()

	// Try to use cached Elo ratings for both teams.
	homeElo, _, _, homeErr := h.repo.GetTeamCachedElo(homeTeam.ID, asOf)
	awayElo, _, _, awayErr := h.repo.GetTeamCachedElo(awayTeam.ID, asOf)

	cacheHit := (homeErr == nil && awayErr == nil)
	if cacheHit {
		c.Header("X-Cache-Status", "hit")
	} else {
		// Cache miss: calculate Elos from scratch.
		c.Header("X-Cache-Status", "miss")
		
		// Fetch all matches up to asOf for accurate Elo calculation.
		// Note: We need ALL matches (teamID=0) because each team's Elo depends on
		// the historical Elo of all their opponents.
		allMatches, err := h.repo.GetMatchesChronological(0, asOf)
		if err != nil {
			c.Header("Cache-Control", "no-store")
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
			return
		}

		ratings := elo.CalculateUntil(allMatches, asOf, cfg)
		homeElo = ratings[homeTeam.ID]
		awayElo = ratings[awayTeam.ID]
	}

	if homeElo == 0 {
		homeElo = cfg.DefaultRating
	}
	if awayElo == 0 {
		awayElo = cfg.DefaultRating
	}

	// Fetch individual team matches for goal rate calculation.
	homeMatches, err := h.repo.GetMatchesChronological(homeTeam.ID, asOf)
	if err != nil {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	awayMatches, err := h.repo.GetMatchesChronological(awayTeam.ID, asOf)
	if err != nil {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Compute historical average goals scored per game for each team.
	homeGoalRate := avgGoalsScored(homeTeam.ID, homeMatches)
	awayGoalRate := avgGoalsScored(awayTeam.ID, awayMatches)

	// Resolve venue.
	venueStr := req.Venue
	venue := simulator.VenueNeutral
	switch venueStr {
	case "home":
		venue = simulator.VenueHome
	case "away":
		venue = simulator.VenueAway
	default:
		venueStr = "neutral"
	}

	simInput := simulator.Input{
		HomeElo:       homeElo,
		AwayElo:       awayElo,
		HomeGoalRate:  homeGoalRate,
		AwayGoalRate:  awayGoalRate,
		Venue:         venue,
		Simulations:   req.Simulations,
		HomeAdvantage: cfg.HomeAdvantage,
	}

	result := simulator.Run(simInput, nil)

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, models.SimulateResponse{
		HomeTeam:    homeTeam.Name,
		AwayTeam:    awayTeam.Name,
		Venue:       venueStr,
		AsOf:        dateStr,
		Simulations: result.Simulations,
		HomeElo:     roundElo(homeElo),
		AwayElo:     roundElo(awayElo),
		Outcome: models.SimulationOutcome{
			HomeWinPct: result.HomeWinPct,
			DrawPct:    result.DrawPct,
			AwayWinPct: result.AwayWinPct,
			HomeWinCI:  result.HomeWinCI,
			DrawCI:     result.DrawCI,
			AwayWinCI:  result.AwayWinCI,
		},
		ExpectedScore: models.ExpectedScore{
			HomeGoals: result.ExpectedHomeGoals,
			AwayGoals: result.ExpectedAwayGoals,
		},
		UpsetProbability: result.UpsetProbability,
		Links: []models.Link{
			{Rel: "self", Href: "/api/v1/football/matches/simulate", Method: http.MethodPost},
			{Rel: "home-team", Href: fmt.Sprintf("/api/v1/football/teams/%d", homeTeam.ID), Method: http.MethodGet},
			{Rel: "away-team", Href: fmt.Sprintf("/api/v1/football/teams/%d", awayTeam.ID), Method: http.MethodGet},
			{Rel: "home-elo", Href: fmt.Sprintf("/api/v1/football/teams/%d/elo?date=%s", homeTeam.ID, dateStr), Method: http.MethodGet},
			{Rel: "away-elo", Href: fmt.Sprintf("/api/v1/football/teams/%d/elo?date=%s", awayTeam.ID, dateStr), Method: http.MethodGet},
		},
	})
}

// mergeMatchResults combines two slices of MatchResult, deduplicating by MatchID.
// It uses the slices package (Go 1.21+) to sort and compact the combined slice.
func mergeMatchResults(a, b []elo.MatchResult) []elo.MatchResult {
	merged := append(slices.Clone(a), b...)
	slices.SortFunc(merged, func(x, y elo.MatchResult) int {
		return cmp.Compare(x.MatchID, y.MatchID)
	})
	return slices.CompactFunc(merged, func(x, y elo.MatchResult) bool {
		return x.MatchID == y.MatchID
	})
}

// avgGoalsScored returns the average number of goals scored per game by teamID
// across all provided matches.  Returns 0 if no matches are present (the
// simulator will fall back to BaseGoalRate).
func avgGoalsScored(teamID int, matches []elo.MatchResult) float64 {
	if len(matches) == 0 {
		return 0
	}
	var total float64
	for _, m := range matches {
		switch teamID {
		case m.HomeTeamID:
			total += float64(m.HomeScore)
		case m.AwayTeamID:
			total += float64(m.AwayScore)
		}
	}
	return total / float64(len(matches))
}
