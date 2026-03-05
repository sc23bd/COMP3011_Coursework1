// Package handlers implements the HTTP handler functions for the Football
// resource.  Handlers are thin adapters between the HTTP layer and the
// repository, keeping business logic separate from transport concerns.
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// FootballHandler holds the dependencies required by the football HTTP handlers.
type FootballHandler struct {
	repo db.FootballRepository
}

// NewFootballHandler constructs a FootballHandler backed by the provided repository.
func NewFootballHandler(repo db.FootballRepository) *FootballHandler {
	return &FootballHandler{repo: repo}
}

// --- Teams -------------------------------------------------------------------

// ListTeams handles GET /api/v1/football/teams
// Returns all national teams with HATEOAS links.
func (h *FootballHandler) ListTeams(c *gin.Context) {
	teams, err := h.repo.ListTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	responses := make([]models.TeamResponse, 0, len(teams))
	for _, t := range teams {
		responses = append(responses, models.TeamResponse{
			Team:  t,
			Links: teamLinks(t.ID),
		})
	}

	c.JSON(http.StatusOK, models.TeamsResponse{
		Data: responses,
		Links: []models.Link{
			{Rel: "self", Href: "/api/v1/football/teams", Method: http.MethodGet},
		},
	})
}

// GetTeam handles GET /api/v1/football/teams/:id
// Returns the requested team or 404 if it does not exist.
func (h *FootballHandler) GetTeam(c *gin.Context) {
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

	c.JSON(http.StatusOK, models.TeamResponse{
		Team:  team,
		Links: teamLinks(team.ID),
	})
}

// GetTeamHistory handles GET /api/v1/football/teams/:id/history
// Returns the historical names for the given team.
func (h *FootballHandler) GetTeamHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid team id"})
		return
	}

	// Verify the team exists first.
	if _, err := h.repo.GetTeamByID(id); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "team not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	history, err := h.repo.GetTeamHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}
	if history == nil {
		history = []models.FormerName{}
	}

	c.JSON(http.StatusOK, models.FormerNamesResponse{
		Data: history,
		Links: []models.Link{
			{Rel: "team", Href: "/api/v1/football/teams/" + c.Param("id"), Method: http.MethodGet},
		},
	})
}

// --- Matches -----------------------------------------------------------------

// defaultLimit is the default number of matches returned per page.
const defaultLimit = 50

// ListMatches handles GET /api/v1/football/matches
// Accepts optional ?limit= and ?offset= query parameters for pagination.
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

// GetMatchGoals handles GET /api/v1/football/matches/:id/goals
// Returns all goals for the specified match.
func (h *FootballHandler) GetMatchGoals(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid match id"})
		return
	}

	// Verify the match exists first.
	if _, err := h.repo.GetMatchByID(id); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "match not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	goals, err := h.repo.GetMatchGoals(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}
	if goals == nil {
		goals = []models.Goal{}
	}

	c.JSON(http.StatusOK, models.GoalsResponse{
		Data: goals,
		Links: []models.Link{
			{Rel: "match", Href: "/api/v1/football/matches/" + c.Param("id"), Method: http.MethodGet},
		},
	})
}

// GetMatchShootout handles GET /api/v1/football/matches/:id/shootout
// Returns the shootout result for the match, or 404 if there was none.
func (h *FootballHandler) GetMatchShootout(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid match id"})
		return
	}

	// Verify the match exists first.
	if _, err := h.repo.GetMatchByID(id); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "match not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	shootout, err := h.repo.GetMatchShootout(id)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "no shootout recorded for this match"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, models.ShootoutResponse{
		Shootout: shootout,
		Links: []models.Link{
			{Rel: "match", Href: "/api/v1/football/matches/" + c.Param("id"), Method: http.MethodGet},
		},
	})
}

// --- Players -----------------------------------------------------------------

// GetPlayerGoals handles GET /api/v1/football/players/:name/goals
// Returns all goals scored by the named player across all matches.
func (h *FootballHandler) GetPlayerGoals(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "player name is required"})
		return
	}

	goals, err := h.repo.GetPlayerGoals(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}
	if goals == nil {
		goals = []models.Goal{}
	}

	c.JSON(http.StatusOK, models.GoalsResponse{
		Data: goals,
		Links: []models.Link{
			{Rel: "self", Href: "/api/v1/football/players/" + name + "/goals", Method: http.MethodGet},
		},
	})
}

// --- link builders -----------------------------------------------------------

func teamLinks(id int) []models.Link {
	base := "/api/v1/football/teams/" + strconv.Itoa(id)
	return []models.Link{
		{Rel: "self", Href: base, Method: http.MethodGet},
		{Rel: "history", Href: base + "/history", Method: http.MethodGet},
	}
}

func matchLinks(id int) []models.Link {
	base := "/api/v1/football/matches/" + strconv.Itoa(id)
	return []models.Link{
		{Rel: "self", Href: base, Method: http.MethodGet},
		{Rel: "goals", Href: base + "/goals", Method: http.MethodGet},
		{Rel: "shootout", Href: base + "/shootout", Method: http.MethodGet},
	}
}
