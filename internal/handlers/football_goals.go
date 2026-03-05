package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// --- Goals & shootouts (read) ------------------------------------------------

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

// --- Players (read) ----------------------------------------------------------

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

// --- Goals (write) -----------------------------------------------------------

// CreateGoal handles POST /api/v1/football/matches/:id/goals
// Records a new goal for the specified match. Requires JWT authorisation.
func (h *FootballHandler) CreateGoal(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid match id"})
		return
	}

	var req models.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Verify the match exists.
	if _, err := h.repo.GetMatchByID(matchID); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "match not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Look up the team to populate the team name on the goal.
	team, err := h.repo.GetTeamByID(req.TeamID)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "team not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	goal, err := h.repo.CreateGoal(models.Goal{
		MatchID: matchID,
		TeamID:  req.TeamID,
		Team:    team.Name,
		Scorer:  req.Scorer,
		OwnGoal: req.OwnGoal,
		Penalty: req.Penalty,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, models.GoalsResponse{
		Data: []models.Goal{goal},
		Links: []models.Link{
			{Rel: "match", Href: "/api/v1/football/matches/" + c.Param("id"), Method: http.MethodGet},
		},
	})
}

// DeleteGoal handles DELETE /api/v1/football/matches/:id/goals/:goalId
// Removes a goal record. Requires JWT authorisation.
func (h *FootballHandler) DeleteGoal(c *gin.Context) {
	goalID, err := strconv.Atoi(c.Param("goalId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid goal id"})
		return
	}

	if err := h.repo.DeleteGoal(goalID); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "goal not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Shootouts (write) -------------------------------------------------------

// CreateShootout handles POST /api/v1/football/matches/:id/shootout
// Records the penalty-shootout result for a match. Requires JWT authorisation.
func (h *FootballHandler) CreateShootout(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid match id"})
		return
	}

	var req models.CreateShootoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Verify the match exists.
	if _, err := h.repo.GetMatchByID(matchID); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "match not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Look up the winning team to populate the winner name.
	winner, err := h.repo.GetTeamByID(req.WinnerID)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "winner team not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	shootout, err := h.repo.CreateShootout(models.Shootout{
		MatchID:  matchID,
		WinnerID: req.WinnerID,
		Winner:   winner.Name,
	})
	if errors.Is(err, models.ErrConflict) {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "shootout already recorded for this match"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, models.ShootoutResponse{
		Shootout: shootout,
		Links: []models.Link{
			{Rel: "match", Href: "/api/v1/football/matches/" + c.Param("id"), Method: http.MethodGet},
		},
	})
}

// DeleteShootout handles DELETE /api/v1/football/matches/:id/shootout
// Removes the shootout record for a match. Requires JWT authorisation.
func (h *FootballHandler) DeleteShootout(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid match id"})
		return
	}

	if err := h.repo.DeleteShootout(matchID); errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "no shootout found for this match"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
