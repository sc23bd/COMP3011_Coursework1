package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// defaultLimit is the default number of matches returned per page.
const defaultLimit = 50

// --- Tournaments (read) -------------------------------------------------------

// ListTournaments handles GET /api/v1/football/tournaments
// Returns all tournaments ordered alphabetically.
//
//	@Summary		List all tournaments
//	@Description	Get all football tournaments ordered by name
//	@Tags			tournaments
//	@Produce		json
//	@Success		200	{object}	models.TournamentsResponse	"List of tournaments"
//	@Failure		500	{object}	models.ErrorResponse		"Internal server error"
//	@Router			/football/tournaments [get]
func (h *FootballHandler) ListTournaments(c *gin.Context) {
	tournaments, err := h.repo.ListTournaments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}
	if tournaments == nil {
		tournaments = []models.Tournament{}
	}
	c.JSON(http.StatusOK, models.TournamentsResponse{Data: tournaments})
}

// --- Matches (read) ----------------------------------------------------------

// ListMatches handles GET /api/v1/football/matches
// Accepts optional ?limit= and ?offset= query parameters for pagination.
//
//	@Summary		List all matches
//	@Description	Get all matches with pagination support
//	@Tags			matches
//	@Produce		json
//	@Param			limit	query		int						false	"Number of results per page"	default(50)
//	@Param			offset	query		int						false	"Offset for pagination"			default(0)
//	@Success		200		{object}	models.MatchesResponse	"List of matches"
//	@Failure		400		{object}	models.ErrorResponse	"Invalid query parameters"
//	@Failure		500		{object}	models.ErrorResponse	"Internal server error"
//	@Router			/football/matches [get]
func (h *FootballHandler) ListMatches(c *gin.Context) {
	limit := defaultLimit
	offset := 0

	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "limit must be a positive integer"})
			return
		}
		limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "offset must be a non-negative integer"})
			return
		}
		offset = n
	}

	matches, err := h.repo.ListMatches(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	responses := make([]models.MatchResponse, 0, len(matches))
	for _, m := range matches {
		responses = append(responses, models.MatchResponse{
			Match: m,
			Links: matchLinks(m.ID),
		})
	}

	c.JSON(http.StatusOK, models.MatchesResponse{
		Data: responses,
		Links: []models.Link{
			{Rel: "self", Href: "/api/v1/football/matches", Method: http.MethodGet},
		},
	})
}

// GetMatch handles GET /api/v1/football/matches/:id
// Returns the requested match or 404 if it does not exist.
//
//	@Summary		Get a match by ID
//	@Description	Get detailed information about a specific match
//	@Tags			matches
//	@Produce		json
//	@Param			id	path		int						true	"Match ID"
//	@Success		200	{object}	models.MatchResponse	"Match details"
//	@Failure		400	{object}	models.ErrorResponse	"Invalid match ID"
//	@Failure		404	{object}	models.ErrorResponse	"Match not found"
//	@Failure		500	{object}	models.ErrorResponse	"Internal server error"
//	@Router			/football/matches/{id} [get]
func (h *FootballHandler) GetMatch(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid match id"})
		return
	}

	match, err := h.repo.GetMatchByID(id)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "match not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, models.MatchResponse{
		Match: match,
		Links: matchLinks(match.ID),
	})
}

// GetHeadToHead handles GET /api/v1/football/head-to-head?teamA=:id&teamB=:id
// Returns all matches between the two teams.
//
//	@Summary		Get head-to-head matches
//	@Description	Get all matches between two specific teams
//	@Tags			matches
//	@Produce		json
//	@Param			teamA	query		int						true	"First team ID"
//	@Param			teamB	query		int						true	"Second team ID"
//	@Success		200		{object}	models.MatchesResponse	"Head-to-head matches"
//	@Failure		400		{object}	models.ErrorResponse	"Invalid query parameters"
//	@Failure		500		{object}	models.ErrorResponse	"Internal server error"
//	@Router			/football/head-to-head [get]
func (h *FootballHandler) GetHeadToHead(c *gin.Context) {
	aStr := c.Query("teamA")
	bStr := c.Query("teamB")
	if aStr == "" || bStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "teamA and teamB query parameters are required"})
		return
	}

	teamA, err := strconv.Atoi(aStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "teamA must be an integer"})
		return
	}
	teamB, err := strconv.Atoi(bStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "teamB must be an integer"})
		return
	}

	matches, err := h.repo.GetHeadToHead(teamA, teamB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	responses := make([]models.MatchResponse, 0, len(matches))
	for _, m := range matches {
		responses = append(responses, models.MatchResponse{
			Match: m,
			Links: matchLinks(m.ID),
		})
	}

	c.JSON(http.StatusOK, models.MatchesResponse{
		Data: responses,
		Links: []models.Link{
			{Rel: "self", Href: "/api/v1/football/head-to-head", Method: http.MethodGet},
		},
	})
}

// --- Matches (write) ---------------------------------------------------------

// CreateMatch handles POST /api/v1/football/matches
// Creates a new match result. Requires JWT authorisation.
//
//	@Summary		Create a new match
//	@Description	Create a new match result (requires authentication)
//	@Tags			matches
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateMatchRequest	true	"Match details"
//	@Success		201		{object}	models.MatchResponse		"Match created"
//	@Failure		400		{object}	models.ErrorResponse		"Invalid request"
//	@Failure		401		{object}	models.ErrorResponse		"Unauthorized"
//	@Failure		409		{object}	models.ErrorResponse		"Match already exists"
//	@Failure		500		{object}	models.ErrorResponse		"Internal server error"
//	@Security		Bearer
//	@Router			/football/matches [post]
func (h *FootballHandler) CreateMatch(c *gin.Context) {
	var req models.CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Verify the home and away teams exist before inserting.
	if !h.checkTeamExists(c, req.HomeTeamID, "home team") {
		return
	}
	if !h.checkTeamExists(c, req.AwayTeamID, "away team") {
		return
	}
	// Verify the tournament exists before inserting.
	if !h.checkTournamentExists(c, req.TournamentID) {
		return
	}

	m := models.Match{
		Date:         req.Date,
		HomeTeamID:   req.HomeTeamID,
		AwayTeamID:   req.AwayTeamID,
		HomeScore:    req.HomeScore,
		AwayScore:    req.AwayScore,
		TournamentID: req.TournamentID,
		City:         req.City,
		Country:      req.Country,
		Neutral:      req.Neutral,
	}

	created, err := h.repo.CreateMatch(m)
	if errors.Is(err, models.ErrConflict) {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "match already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Header("Location", "/api/v1/football/matches/"+strconv.Itoa(created.ID))
	c.JSON(http.StatusCreated, models.MatchResponse{
		Match: created,
		Links: matchLinks(created.ID),
	})
}

// UpdateMatch handles PUT /api/v1/football/matches/:id
// Replaces an existing match record. Requires JWT authorisation.
//
//	@Summary		Update a match
//	@Description	Update an existing match record (requires authentication)
//	@Tags			matches
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Match ID"
//	@Param			request	body		models.UpdateMatchRequest	true	"Updated match details"
//	@Success		200		{object}	models.MatchResponse		"Match updated"
//	@Failure		400		{object}	models.ErrorResponse		"Invalid request"
//	@Failure		401		{object}	models.ErrorResponse		"Unauthorized"
//	@Failure		404		{object}	models.ErrorResponse		"Match not found"
//	@Failure		409		{object}	models.ErrorResponse		"Match already exists"
//	@Failure		500		{object}	models.ErrorResponse		"Internal server error"
//	@Security		Bearer
//	@Router			/football/matches/{id} [put]
func (h *FootballHandler) UpdateMatch(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid match id"})
		return
	}

	var req models.UpdateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Verify teams and tournament exist before updating.
	if !h.checkTeamExists(c, req.HomeTeamID, "home team") {
		return
	}
	if !h.checkTeamExists(c, req.AwayTeamID, "away team") {
		return
	}
	if !h.checkTournamentExists(c, req.TournamentID) {
		return
	}

	m := models.Match{
		Date:         req.Date,
		HomeTeamID:   req.HomeTeamID,
		AwayTeamID:   req.AwayTeamID,
		HomeScore:    req.HomeScore,
		AwayScore:    req.AwayScore,
		TournamentID: req.TournamentID,
		City:         req.City,
		Country:      req.Country,
		Neutral:      req.Neutral,
	}

	updated, err := h.repo.UpdateMatch(id, m)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "match not found"})
		return
	}
	if errors.Is(err, models.ErrConflict) {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "match already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, models.MatchResponse{
		Match: updated,
		Links: matchLinks(updated.ID),
	})
}

// DeleteMatch handles DELETE /api/v1/football/matches/:id
// Removes a match record. Requires JWT authorisation.
//
//	@Summary		Delete a match
//	@Description	Delete a match by ID (requires authentication)
//	@Tags			matches
//	@Produce		json
//	@Param			id	path	int	true	"Match ID"
//	@Success		204	"Match deleted successfully"
//	@Failure		400	{object}	models.ErrorResponse	"Invalid match ID"
//	@Failure		401	{object}	models.ErrorResponse	"Unauthorized"
//	@Failure		404	{object}	models.ErrorResponse	"Match not found"
//	@Failure		500	{object}	models.ErrorResponse	"Internal server error"
//	@Security		Bearer
//	@Router			/football/matches/{id} [delete]
func (h *FootballHandler) DeleteMatch(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid match id"})
		return
	}

	if err := h.repo.DeleteMatch(id); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "match not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
