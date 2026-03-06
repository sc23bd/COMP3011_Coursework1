// Package handlers implements the HTTP handler functions for the Football
// resource.  Handlers are thin adapters between the HTTP layer and the
// repository, keeping business logic separate from transport concerns.
package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// FootballHandler holds the dependencies required by the football HTTP handlers.
type FootballHandler struct {
	repo db.FootballRepository

	// eloRecalc tracks background recalculation state for rate limiting.
	eloRecalc struct {
		mu      sync.Mutex
		lastRun time.Time
		running bool
	}
}

// NewFootballHandler constructs a FootballHandler backed by the provided repository.
func NewFootballHandler(repo db.FootballRepository) *FootballHandler {
	return &FootballHandler{repo: repo}
}

// checkTeamExists looks up a team by ID and writes a 400/500 response if it
// is not found or an error occurs.  Returns true only if the team exists.
func (h *FootballHandler) checkTeamExists(c *gin.Context, id int, label string) bool {
	_, err := h.repo.GetTeamByID(id)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: label + " not found"})
		return false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return false
	}
	return true
}

// checkTournamentExists looks up a tournament by ID and writes a 400/500
// response if it is not found or an error occurs.  Returns true only if it exists.
func (h *FootballHandler) checkTournamentExists(c *gin.Context, id int) bool {
	_, err := h.repo.GetTournamentByID(id)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "tournament not found"})
		return false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return false
	}
	return true
}

func teamLinks(id int) []models.Link {
	base := "/api/v1/football/teams/" + strconv.Itoa(id)
	return []models.Link{
		{Rel: "self", Href: base, Method: http.MethodGet},
		{Rel: "update", Href: base, Method: http.MethodPut},
		{Rel: "delete", Href: base, Method: http.MethodDelete},
		{Rel: "history", Href: base + "/history", Method: http.MethodGet},
	}
}

func matchLinks(id int) []models.Link {
	base := "/api/v1/football/matches/" + strconv.Itoa(id)
	return []models.Link{
		{Rel: "self", Href: base, Method: http.MethodGet},
		{Rel: "update", Href: base, Method: http.MethodPut},
		{Rel: "delete", Href: base, Method: http.MethodDelete},
		{Rel: "goals", Href: base + "/goals", Method: http.MethodGet},
		{Rel: "shootout", Href: base + "/shootout", Method: http.MethodGet},
	}
}
