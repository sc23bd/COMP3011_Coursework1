package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/elo"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

const eloDateLayout = "2006-01-02"

// eloLinks returns the standard HATEOAS links for a team's Elo resource.
func eloLinks(teamID int, dateStr string) []models.Link {
	base := fmt.Sprintf("/api/v1/football/teams/%d/elo", teamID)
	selfHref := base
	if dateStr != "" {
		selfHref = base + "?date=" + dateStr
	}
	return []models.Link{
		{Rel: "self", Href: selfHref, Method: http.MethodGet},
		{Rel: "timeline", Href: fmt.Sprintf("/api/v1/football/teams/%d/elo/timeline", teamID), Method: http.MethodGet},
		{Rel: "team", Href: fmt.Sprintf("/api/v1/football/teams/%d", teamID), Method: http.MethodGet},
	}
}

// GetTeamElo handles GET /api/v1/football/teams/:id/elo
// Returns the current or historical Elo rating for a team.
//
//	@Summary		Get team Elo rating
//	@Description	Returns the World Football Elo rating for a team, optionally at a historical date
//	@Tags			elo
//	@Produce		json
//	@Param			id				path		int						true	"Team ID"
//	@Param			date			query		string					false	"Point-in-time date (YYYY-MM-DD); defaults to today"
//	@Param			include_history	query		bool					false	"Include full rating history"
//	@Success		200				{object}	elo.Rating				"Team Elo rating"
//	@Failure		400				{object}	models.ErrorResponse	"Invalid team ID or date"
//	@Failure		404				{object}	models.ErrorResponse	"Team not found"
//	@Failure		500				{object}	models.ErrorResponse	"Internal server error"
//	@Router			/football/teams/{id}/elo [get]
func (h *FootballHandler) GetTeamElo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid team id"})
		return
	}

	team, err := h.repo.GetTeamByID(id)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "team not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	asOf := time.Now().UTC()
	dateStr := c.Query("date")
	if dateStr != "" {
		parsed, parseErr := time.Parse(eloDateLayout, dateStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid date format; expected YYYY-MM-DD"})
			return
		}
		asOf = parsed
	} else {
		dateStr = asOf.Format(eloDateLayout)
	}

	cfg := elo.DefaultConfig()

	// Fetch all matches for this team up to the requested date.
	matches, err := h.repo.GetMatchesChronological(id, asOf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	ratings := elo.CalculateUntil(matches, asOf, cfg)
	currentElo := ratings[id]
	if currentElo == 0 {
		currentElo = cfg.DefaultRating
	}

	// Compute previous rating (all matches before last match date) for delta.
	var prevElo float64
	if len(matches) > 0 {
		prevMatches := matches[:len(matches)-1]
		prevRatings := elo.Calculate(prevMatches, cfg)
		prevElo = prevRatings[id]
		if prevElo == 0 {
			prevElo = cfg.DefaultRating
		}
	} else {
		prevElo = cfg.DefaultRating
	}

	c.Header("X-Elo-Computed-At", time.Now().UTC().Format(time.RFC3339))
	c.JSON(http.StatusOK, elo.Rating{
		TeamID:            team.ID,
		TeamName:          team.Name,
		Date:              asOf,
		Elo:               roundElo(currentElo),
		ChangeFromPrev:    roundElo(currentElo - prevElo),
		MatchesConsidered: len(matches),
		Methodology: elo.Methodology{
			KFactor:          cfg.DefaultKFactor,
			HomeAdvantage:    cfg.HomeAdvantage,
			WeightMultiplier: 1.0,
			FormulaReference: cfg.FormulaRef(),
		},
		Links: eloLinks(id, dateStr),
	})
}

// GetTeamEloTimeline handles GET /api/v1/football/teams/:id/elo/timeline
// Returns the time-series of Elo changes for a team.
//
//	@Summary		Get team Elo timeline
//	@Description	Returns the full time-series of Elo rating changes for a team
//	@Tags			elo
//	@Produce		json
//	@Param			id			path		int						true	"Team ID"
//	@Param			start_date	query		string					false	"Start date (YYYY-MM-DD)"
//	@Param			end_date	query		string					false	"End date (YYYY-MM-DD)"
//	@Param			resolution	query		string					false	"Aggregation: match|month|year (default: match)"
//	@Success		200			{object}	elo.TimelineResponse	"Team Elo timeline"
//	@Failure		400			{object}	models.ErrorResponse	"Invalid team ID or date"
//	@Failure		404			{object}	models.ErrorResponse	"Team not found"
//	@Failure		500			{object}	models.ErrorResponse	"Internal server error"
//	@Router			/football/teams/{id}/elo/timeline [get]
func (h *FootballHandler) GetTeamEloTimeline(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid team id"})
		return
	}

	team, err := h.repo.GetTeamByID(id)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "team not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	endDate := time.Now().UTC()
	if s := c.Query("end_date"); s != "" {
		parsed, parseErr := time.Parse(eloDateLayout, s)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid end_date format; expected YYYY-MM-DD"})
			return
		}
		endDate = parsed
	}

	var startDate *time.Time
	if s := c.Query("start_date"); s != "" {
		parsed, parseErr := time.Parse(eloDateLayout, s)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid start_date format; expected YYYY-MM-DD"})
			return
		}
		startDate = &parsed
	}

	cfg := elo.DefaultConfig()

	// Fetch all matches for this team up to endDate.
	matches, err := h.repo.GetMatchesChronological(id, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	timeline := elo.CalculateTimeline(id, matches, cfg)

	// Filter by start_date if provided.
	if startDate != nil {
		var filtered []elo.TimelineEntry
		for _, entry := range timeline {
			if !entry.Date.Before(*startDate) {
				filtered = append(filtered, entry)
			}
		}
		timeline = filtered
	}

	if timeline == nil {
		timeline = []elo.TimelineEntry{}
	}

	c.Header("X-Elo-Computed-At", time.Now().UTC().Format(time.RFC3339))
	c.JSON(http.StatusOK, elo.TimelineResponse{
		TeamID:   team.ID,
		TeamName: team.Name,
		Data:     timeline,
		Links: []models.Link{
			{Rel: "self", Href: fmt.Sprintf("/api/v1/football/teams/%d/elo/timeline", id), Method: http.MethodGet},
			{Rel: "elo", Href: fmt.Sprintf("/api/v1/football/teams/%d/elo", id), Method: http.MethodGet},
			{Rel: "team", Href: fmt.Sprintf("/api/v1/football/teams/%d", id), Method: http.MethodGet},
		},
	})
}

// GetEloRankings handles GET /api/v1/football/rankings/elo
// Returns a global Elo rankings snapshot with pagination.
//
//	@Summary		Get global Elo rankings
//	@Description	Returns a paginated snapshot of global Elo rankings, optionally filtered by region
//	@Tags			elo
//	@Produce		json
//	@Param			date	query		string					false	"Point-in-time date (YYYY-MM-DD)"
//	@Param			region	query		string					false	"Filter by region (e.g. europe, asia)"
//	@Param			limit	query		int						false	"Page size (default 50)"
//	@Param			offset	query		int						false	"Page offset (default 0)"
//	@Success		200		{object}	elo.RankingsResponse	"Global Elo rankings"
//	@Failure		400		{object}	models.ErrorResponse	"Invalid query parameters"
//	@Failure		500		{object}	models.ErrorResponse	"Internal server error"
//	@Router			/football/rankings/elo [get]
func (h *FootballHandler) GetEloRankings(c *gin.Context) {
	asOf := time.Now().UTC()
	dateStr := c.Query("date")
	if dateStr != "" {
		parsed, parseErr := time.Parse(eloDateLayout, dateStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid date format; expected YYYY-MM-DD"})
			return
		}
		asOf = parsed
	} else {
		dateStr = asOf.Format(eloDateLayout)
	}

	limit := 50
	if s := c.Query("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "limit must be a positive integer"})
			return
		}
		limit = v
	}

	offset := 0
	if s := c.Query("offset"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "offset must be a non-negative integer"})
			return
		}
		offset = v
	}

	region := c.Query("region")

	rankings, err := h.repo.GetEloRankings(asOf, region, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Determine cache status immediately and set the header before any further
	// slice manipulation, so the intent is clear.
	cacheStatus := "hit"
	if len(rankings) == 0 {
		cacheStatus = "miss"
	}
	c.Header("X-Cache-Status", cacheStatus)

	if rankings == nil {
		rankings = []elo.RankingEntry{}
	}

	// Attach HATEOAS links to each entry.
	for i := range rankings {
		rankings[i].Links = []models.Link{
			{Rel: "elo", Href: fmt.Sprintf("/api/v1/football/teams/%d/elo?date=%s", rankings[i].TeamID, dateStr), Method: http.MethodGet},
			{Rel: "team", Href: fmt.Sprintf("/api/v1/football/teams/%d", rankings[i].TeamID), Method: http.MethodGet},
		}
	}

	selfHref := fmt.Sprintf("/api/v1/football/rankings/elo?date=%s&limit=%d&offset=%d", dateStr, limit, offset)
	c.JSON(http.StatusOK, elo.RankingsResponse{
		Date:   dateStr,
		Data:   rankings,
		Total:  len(rankings),
		Limit:  limit,
		Offset: offset,
		Links: []models.Link{
			{Rel: "self", Href: selfHref, Method: http.MethodGet},
		},
	})
}

// RecalculateEloRankings handles POST /api/v1/football/rankings/elo/recalculate
// Triggers a background recalculation of Elo ratings for all (or one) team.
// Requests are rate-limited to one run per 5 minutes; concurrent runs return 429.
//
//	@Summary		Recalculate Elo rankings
//	@Description	Triggers background Elo recalculation (admin). Optionally scoped to one team.
//	@Tags			elo
//	@Produce		json
//	@Param			team_id	query		int						false	"Limit recalculation to this team"
//	@Param			force	query		bool					false	"Force recalculation even if cache is current"
//	@Success		202		{object}	elo.RecalculateResponse	"Recalculation started"
//	@Failure		400		{object}	models.ErrorResponse	"Invalid parameters"
//	@Failure		401		{object}	models.ErrorResponse	"Unauthorized"
//	@Failure		429		{object}	models.ErrorResponse	"Recalculation already running or rate limit exceeded"
//	@Failure		500		{object}	models.ErrorResponse	"Internal server error"
//	@Security		Bearer
//	@Router			/football/rankings/elo/recalculate [post]
func (h *FootballHandler) RecalculateEloRankings(c *gin.Context) {
	var teamID int
	if s := c.Query("team_id"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "team_id must be a positive integer"})
			return
		}
		// Verify team exists.
		if _, err := h.repo.GetTeamByID(v); errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "team not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
			return
		}
		teamID = v
	}

	// Rate limiting: reject if a recalculation is already running or if the last
	// one completed less than 5 minutes ago (unless ?force=true is set).
	force := c.Query("force") == "true"
	h.eloRecalc.mu.Lock()
	if h.eloRecalc.running {
		h.eloRecalc.mu.Unlock()
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{Error: "recalculation already in progress"})
		return
	}
	if !force && !h.eloRecalc.lastRun.IsZero() && time.Since(h.eloRecalc.lastRun) < 5*time.Minute {
		h.eloRecalc.mu.Unlock()
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{Error: "recalculation rate limit: wait 5 minutes between runs or use ?force=true"})
		return
	}
	h.eloRecalc.running = true
	h.eloRecalc.mu.Unlock()

	// Launch background goroutine; respond immediately with 202.
	go h.runEloRecalculation(teamID)

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusAccepted, elo.RecalculateResponse{
		Message: "Elo recalculation started in the background",
		Links: []models.Link{
			{Rel: "rankings", Href: "/api/v1/football/rankings/elo", Method: http.MethodGet},
		},
	})
}

// runEloRecalculation performs the full Elo recalculation for all teams (teamID=0)
// or a single team and persists snapshots into the cache table.
// It always marks the recalculation as complete (running=false) when done.
func (h *FootballHandler) runEloRecalculation(teamID int) {
	start := time.Now()
	log.Printf("Elo recalculation started (teamID=%d)", teamID)

	defer func() {
		h.eloRecalc.mu.Lock()
		h.eloRecalc.running = false
		h.eloRecalc.lastRun = time.Now()
		h.eloRecalc.mu.Unlock()
		log.Printf("Elo recalculation finished (teamID=%d, duration=%s)", teamID, time.Since(start))
	}()

	cfg := elo.DefaultConfig()
	endDate := time.Now().UTC()

	matches, err := h.repo.GetMatchesChronological(teamID, endDate)
	if err != nil {
		log.Printf("Elo recalculation error fetching matches (teamID=%d): %v", teamID, err)
		return
	}

	ratings := elo.Calculate(matches, cfg)

	// Count matches per team.
	matchCount := make(map[int]int)
	for _, m := range matches {
		matchCount[m.HomeTeamID]++
		matchCount[m.AwayTeamID]++
	}

	// Build sorted rank list.
	type ranked struct {
		id  int
		elo float64
	}
	sortable := make([]ranked, 0, len(ratings))
	for id, r := range ratings {
		sortable = append(sortable, ranked{id, r})
	}
	// Sort descending by Elo using an efficient O(n log n) algorithm.
	sort.Slice(sortable, func(i, j int) bool {
		return sortable[i].elo > sortable[j].elo
	})

	var saveErrors int
	for rank, entry := range sortable {
		if saveErr := h.repo.SaveEloSnapshot(entry.id, endDate, entry.elo, rank+1, matchCount[entry.id]); saveErr != nil {
			log.Printf("Elo recalculation error saving snapshot (teamID=%d, rank=%d): %v", entry.id, rank+1, saveErr)
			saveErrors++
		}
	}
	if saveErrors > 0 {
		// Log as a warning: the calculation succeeded but some snapshots were not
		// persisted.  Rankings for the affected teams will be stale until the
		// next successful recalculation.
		log.Printf("Elo recalculation PARTIAL FAILURE: %d/%d snapshot(s) not saved (teamID=%d); rankings may be incomplete",
			saveErrors, len(sortable), teamID)
	}
}

// roundElo rounds a float to 2 decimal places.
func roundElo(v float64) float64 {
	const factor = 100
	if v < 0 {
		return float64(int(v*factor-0.5)) / factor
	}
	return float64(int(v*factor+0.5)) / factor
}
