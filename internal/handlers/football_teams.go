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
//
// @Summary      List all teams
// @Description  Get all national teams with HATEOAS links
// @Tags         teams
// @Produce      json
// @Success      200  {object}  models.TeamsResponse    "List of teams"
// @Failure      500  {object}  models.ErrorResponse    "Internal server error"
// @Router       /football/teams [get]
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
//
// @Summary      Get a team by ID
// @Description  Get detailed information about a specific team
// @Tags         teams
// @Produce      json
// @Param        id   path      int  true  "Team ID"
// @Success      200  {object}  models.TeamResponse     "Team details"
// @Failure      400  {object}  models.ErrorResponse    "Invalid team ID"
// @Failure      404  {object}  models.ErrorResponse    "Team not found"
// @Failure      500  {object}  models.ErrorResponse    "Internal server error"
// @Router       /football/teams/{id} [get]
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
//
// @Summary      Get team history
// @Description  Get historical names for a specific team
// @Tags         teams
// @Produce      json
// @Param        id   path      int  true  "Team ID"
// @Success      200  {object}  models.FormerNamesResponse  "Team history"
// @Failure      400  {object}  models.ErrorResponse        "Invalid team ID"
// @Failure      404  {object}  models.ErrorResponse        "Team not found"
// @Failure      500  {object}  models.ErrorResponse        "Internal server error"
// @Router       /football/teams/{id}/history [get]
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
//
// @Summary      Create a new team
// @Description  Create a new national team (requires authentication)
// @Tags         teams
// @Accept       json
// @Produce      json
// @Param        request body models.CreateTeamRequest true "Team details"
// @Success      201  {object}  models.TeamResponse     "Team created"
// @Failure      400  {object}  models.ErrorResponse    "Invalid request"
// @Failure      401  {object}  models.ErrorResponse    "Unauthorized"
// @Failure      409  {object}  models.ErrorResponse    "Team already exists"
// @Failure      500  {object}  models.ErrorResponse    "Internal server error"
// @Security     Bearer
// @Router       /football/teams [post]
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
//
// @Summary      Update a team
// @Description  Update team name (requires authentication)
// @Tags         teams
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Team ID"
// @Param        request body models.UpdateTeamRequest true "Updated team details"
// @Success      200  {object}  models.TeamResponse     "Team updated"
// @Failure      400  {object}  models.ErrorResponse    "Invalid request"
// @Failure      401  {object}  models.ErrorResponse    "Unauthorized"
// @Failure      404  {object}  models.ErrorResponse    "Team not found"
// @Failure      409  {object}  models.ErrorResponse    "Team name already in use"
// @Failure      500  {object}  models.ErrorResponse    "Internal server error"
// @Security     Bearer
// @Router       /football/teams/{id} [put]
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
//
// @Summary      Delete a team
// @Description  Delete a team by ID (requires authentication)
// @Tags         teams
// @Produce      json
// @Param        id   path      int  true  "Team ID"
// @Success      204  "Team deleted successfully"
// @Failure      400  {object}  models.ErrorResponse    "Invalid team ID"
// @Failure      401  {object}  models.ErrorResponse    "Unauthorized"
// @Failure      404  {object}  models.ErrorResponse    "Team not found"
// @Failure      500  {object}  models.ErrorResponse    "Internal server error"
// @Security     Bearer
// @Router       /football/teams/{id} [delete]
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
