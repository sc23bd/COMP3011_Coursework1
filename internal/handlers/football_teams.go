package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// --- Teams (read) ------------------------------------------------------------

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

// --- Teams (write) -----------------------------------------------------------

// CreateTeam handles POST /api/v1/football/teams
// Creates a new national team. Requires JWT authorisation.
func (h *FootballHandler) CreateTeam(c *gin.Context) {
	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	team, err := h.repo.CreateTeam(req.Name)
	if errors.Is(err, models.ErrConflict) {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "team already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Header("Location", "/api/v1/football/teams/"+strconv.Itoa(team.ID))
	c.JSON(http.StatusCreated, models.TeamResponse{
		Team:  team,
		Links: teamLinks(team.ID),
	})
}

// UpdateTeam handles PUT /api/v1/football/teams/:id
// Replaces the name of an existing team. Requires JWT authorisation.
func (h *FootballHandler) UpdateTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid team id"})
		return
	}

	var req models.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	team, err := h.repo.UpdateTeam(id, req.Name)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "team not found"})
		return
	}
	if errors.Is(err, models.ErrConflict) {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "team name already in use"})
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

// DeleteTeam handles DELETE /api/v1/football/teams/:id
// Removes a team. Requires JWT authorisation.
func (h *FootballHandler) DeleteTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid team id"})
		return
	}

	if err := h.repo.DeleteTeam(id); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "team not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
