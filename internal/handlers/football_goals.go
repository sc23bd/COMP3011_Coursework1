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
//
// @Summary      Get goals for a match
// @Description  Get all goals scored in the specified match
// @Tags         goals
// @Produce      json
// @Param        id   path      int  true  "Match ID"
// @Success      200  {object}  models.GoalsResponse     "List of goals"
// @Failure      400  {object}  models.ErrorResponse     "Invalid match ID"
// @Failure      404  {object}  models.ErrorResponse     "Match not found"
// @Failure      500  {object}  models.ErrorResponse     "Internal server error"
// @Router       /football/matches/{id}/goals [get]
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
//
// @Summary      Get shootout result
// @Description  Get the penalty shootout result for a match
// @Tags         shootouts
// @Produce      json
// @Param        id   path      int  true  "Match ID"
// @Success      200  {object}  models.ShootoutResponse  "Shootout result"
// @Failure      400  {object}  models.ErrorResponse     "Invalid match ID"
// @Failure      404  {object}  models.ErrorResponse     "Match or shootout not found"
// @Failure      500  {object}  models.ErrorResponse     "Internal server error"
// @Router       /football/matches/{id}/shootout [get]
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
//
// @Summary      Get goals by player
// @Description  Get all goals scored by a specific player
// @Tags         players
// @Produce      json
// @Param        name  path      string  true  "Player name"
// @Success      200  {object}  models.GoalsResponse     "List of goals"
// @Failure      400  {object}  models.ErrorResponse     "Invalid player name"
// @Failure      500  {object}  models.ErrorResponse     "Internal server error"
// @Router       /football/players/{name}/goals [get]
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
//
// @Summary      Create a goal
// @Description  Record a new goal for a match
// @Tags         goals
// @Accept       json
// @Produce      json
// @Param        id    path      int                        true  "Match ID"
// @Param        goal  body      models.CreateGoalRequest   true  "Goal details"
// @Success      201  {object}  models.GoalsResponse     "Goal created"
// @Failure      400  {object}  models.ErrorResponse     "Invalid input"
// @Failure      404  {object}  models.ErrorResponse     "Match or team not found"
// @Failure      500  {object}  models.ErrorResponse     "Internal server error"
// @Security     BearerAuth
// @Router       /football/matches/{id}/goals [post]
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
//
// @Summary      Delete a goal
// @Description  Remove a goal record from a match
// @Tags         goals
// @Param        id      path  int  true  "Match ID"
// @Param        goalId  path  int  true  "Goal ID"
// @Success      204  "Goal deleted"
// @Failure      400  {object}  models.ErrorResponse     "Invalid ID"
// @Failure      404  {object}  models.ErrorResponse     "Goal not found"
// @Failure      500  {object}  models.ErrorResponse     "Internal server error"
// @Security     BearerAuth
// @Router       /football/matches/{id}/goals/{goalId} [delete]
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
//
// @Summary      Create a shootout
// @Description  Record a penalty shootout result for a match
// @Tags         shootouts
// @Accept       json
// @Produce      json
// @Param        id        path      int                             true  "Match ID"
// @Param        shootout  body      models.CreateShootoutRequest    true  "Shootout details"
// @Success      201  {object}  models.ShootoutResponse  "Shootout created"
// @Failure      400  {object}  models.ErrorResponse     "Invalid input"
// @Failure      404  {object}  models.ErrorResponse     "Match or team not found"
// @Failure      409  {object}  models.ErrorResponse     "Shootout already exists"
// @Failure      500  {object}  models.ErrorResponse     "Internal server error"
// @Security     BearerAuth
// @Router       /football/matches/{id}/shootout [post]
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
//
// @Summary      Delete a shootout
// @Description  Remove a penalty shootout record from a match
// @Tags         shootouts
// @Param        id   path  int  true  "Match ID"
// @Success      204  "Shootout deleted"
// @Failure      400  {object}  models.ErrorResponse     "Invalid match ID"
// @Failure      404  {object}  models.ErrorResponse     "Shootout not found"
// @Failure      500  {object}  models.ErrorResponse     "Internal server error"
// @Security     BearerAuth
// @Router       /football/matches/{id}/shootout [delete]
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
