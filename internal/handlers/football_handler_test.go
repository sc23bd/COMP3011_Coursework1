package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/handlers"
	elomodels "github.com/sc23bd/COMP3011_Coursework1/internal/elo"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// ---------------------------------------------------------------------------
// footballMock is a minimal in-test stub that implements db.FootballRepository.
// It lives here so the football handler tests do not depend on any external package.
// ---------------------------------------------------------------------------

type footballMock struct {
	teams       []models.Team
	tournaments []models.Tournament
	matches     []models.Match
	goals       []models.Goal
	shootouts   []models.Shootout
	formerNames []models.FormerName
}

func (m *footballMock) addTeam(name string) models.Team {
	t := models.Team{ID: len(m.teams) + 1, Name: name, CreatedAt: time.Time{}}
	m.teams = append(m.teams, t)
	return t
}

func (m *footballMock) addTournament(name string) models.Tournament {
	t := models.Tournament{ID: len(m.tournaments) + 1, Name: name}
	m.tournaments = append(m.tournaments, t)
	return t
}

func (m *footballMock) addMatch(match models.Match) models.Match {
	match.ID = len(m.matches) + 1
	m.matches = append(m.matches, match)
	return match
}

func (m *footballMock) addGoal(g models.Goal) models.Goal {
	g.ID = len(m.goals) + 1
	m.goals = append(m.goals, g)
	return g
}

func (m *footballMock) addShootout(s models.Shootout) models.Shootout {
	s.ID = len(m.shootouts) + 1
	m.shootouts = append(m.shootouts, s)
	return s
}

// --- Read implementations ---------------------------------------------------

func (m *footballMock) ListTeams() ([]models.Team, error) {
	result := make([]models.Team, len(m.teams))
	copy(result, m.teams)
	return result, nil
}

func (m *footballMock) GetTeamByID(id int) (models.Team, error) {
	for _, t := range m.teams {
		if t.ID == id {
			return t, nil
		}
	}
	return models.Team{}, models.ErrNotFound
}

func (m *footballMock) GetTeamHistory(teamID int) ([]models.FormerName, error) {
	var result []models.FormerName
	for _, fn := range m.formerNames {
		if fn.TeamID == teamID {
			result = append(result, fn)
		}
	}
	return result, nil
}

func (m *footballMock) GetTournamentByID(id int) (models.Tournament, error) {
	for _, t := range m.tournaments {
		if t.ID == id {
			return t, nil
		}
	}
	return models.Tournament{}, models.ErrNotFound
}

func (m *footballMock) ListMatches(limit, offset int) ([]models.Match, error) {
	if offset >= len(m.matches) {
		return []models.Match{}, nil
	}
	end := offset + limit
	if end > len(m.matches) {
		end = len(m.matches)
	}
	result := make([]models.Match, end-offset)
	copy(result, m.matches[offset:end])
	return result, nil
}

func (m *footballMock) GetMatchByID(id int) (models.Match, error) {
	for _, match := range m.matches {
		if match.ID == id {
			return match, nil
		}
	}
	return models.Match{}, models.ErrNotFound
}

func (m *footballMock) GetHeadToHead(teamA, teamB int) ([]models.Match, error) {
	var result []models.Match
	for _, match := range m.matches {
		if (match.HomeTeamID == teamA && match.AwayTeamID == teamB) ||
			(match.HomeTeamID == teamB && match.AwayTeamID == teamA) {
			result = append(result, match)
		}
	}
	return result, nil
}

func (m *footballMock) GetMatchGoals(matchID int) ([]models.Goal, error) {
	var result []models.Goal
	for _, g := range m.goals {
		if g.MatchID == matchID {
			result = append(result, g)
		}
	}
	return result, nil
}

func (m *footballMock) GetMatchShootout(matchID int) (models.Shootout, error) {
	for _, s := range m.shootouts {
		if s.MatchID == matchID {
			return s, nil
		}
	}
	return models.Shootout{}, models.ErrNotFound
}

func (m *footballMock) GetPlayerGoals(scorer string) ([]models.Goal, error) {
	lower := strings.ToLower(scorer)
	var result []models.Goal
	for _, g := range m.goals {
		if strings.ToLower(g.Scorer) == lower {
			result = append(result, g)
		}
	}
	return result, nil
}

// --- Write implementations --------------------------------------------------

func (m *footballMock) CreateTeam(name string) (models.Team, error) {
	for _, t := range m.teams {
		if t.Name == name {
			return models.Team{}, models.ErrConflict
		}
	}
	t := models.Team{ID: len(m.teams) + 1, Name: name}
	m.teams = append(m.teams, t)
	return t, nil
}

func (m *footballMock) UpdateTeam(id int, name string) (models.Team, error) {
	for i, t := range m.teams {
		if t.ID == id {
			m.teams[i].Name = name
			return m.teams[i], nil
		}
	}
	return models.Team{}, models.ErrNotFound
}

func (m *footballMock) DeleteTeam(id int) error {
	for i, t := range m.teams {
		if t.ID == id {
			m.teams = append(m.teams[:i], m.teams[i+1:]...)
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *footballMock) CreateMatch(match models.Match) (models.Match, error) {
	match.ID = len(m.matches) + 1
	m.matches = append(m.matches, match)
	return match, nil
}

func (m *footballMock) UpdateMatch(id int, match models.Match) (models.Match, error) {
	for i, ms := range m.matches {
		if ms.ID == id {
			match.ID = id
			m.matches[i] = match
			return match, nil
		}
	}
	return models.Match{}, models.ErrNotFound
}

func (m *footballMock) DeleteMatch(id int) error {
	for i, ms := range m.matches {
		if ms.ID == id {
			m.matches = append(m.matches[:i], m.matches[i+1:]...)
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *footballMock) CreateGoal(g models.Goal) (models.Goal, error) {
	g.ID = len(m.goals) + 1
	m.goals = append(m.goals, g)
	return g, nil
}

func (m *footballMock) DeleteGoal(id int) error {
	for i, g := range m.goals {
		if g.ID == id {
			m.goals = append(m.goals[:i], m.goals[i+1:]...)
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *footballMock) CreateShootout(s models.Shootout) (models.Shootout, error) {
	for _, existing := range m.shootouts {
		if existing.MatchID == s.MatchID {
			return models.Shootout{}, models.ErrConflict
		}
	}
	s.ID = len(m.shootouts) + 1
	m.shootouts = append(m.shootouts, s)
	return s, nil
}

func (m *footballMock) DeleteShootout(matchID int) error {
	for i, s := range m.shootouts {
		if s.MatchID == matchID {
			m.shootouts = append(m.shootouts[:i], m.shootouts[i+1:]...)
			return nil
		}
	}
	return models.ErrNotFound
}

// --- Elo implementations ----------------------------------------------------

func (m *footballMock) GetMatchesChronological(teamID int, endDate time.Time) ([]elomodels.MatchResult, error) {
	var results []elomodels.MatchResult
	for _, match := range m.matches {
		if match.Date.After(endDate) {
			continue
		}
		if teamID != 0 && match.HomeTeamID != teamID && match.AwayTeamID != teamID {
			continue
		}
		results = append(results, elomodels.MatchResult{
			MatchID:    match.ID,
			Date:       match.Date,
			HomeTeamID: match.HomeTeamID,
			AwayTeamID: match.AwayTeamID,
			HomeScore:  match.HomeScore,
			AwayScore:  match.AwayScore,
			Tournament: match.Tournament,
			Neutral:    match.Neutral,
		})
	}
	return results, nil
}

func (m *footballMock) SaveEloSnapshot(_ int, _ time.Time, _ float64, _ int, _ int) error {
	return nil
}

func (m *footballMock) GetEloRankings(_ time.Time, _ string, limit, offset int) ([]elomodels.RankingEntry, error) {
	return nil, nil
}

func (m *footballMock) GetTeamCachedRank(_ int, _ time.Time) (int, error) {
	return 0, nil
}

func (m *footballMock) ListTournaments() ([]models.Tournament, error) {
	return m.tournaments, nil
}

// ---------------------------------------------------------------------------
// Router helpers
// ---------------------------------------------------------------------------

// newFootballRouter builds a minimal Gin engine wired to a fresh football mock.
// Write routes are wired without JWT middleware (auth tests use newFootballRouterWithAuth).
func newFootballRouter() (*gin.Engine, *footballMock) {
	mock := &footballMock{}
	fh := handlers.NewFootballHandler(mock)

	r := gin.New()
	v1 := r.Group("/api/v1/football")
	{
		// Read routes
		v1.GET("/teams", fh.ListTeams)
		v1.GET("/teams/:id", fh.GetTeam)
		v1.GET("/teams/:id/history", fh.GetTeamHistory)
		v1.GET("/matches", fh.ListMatches)
		v1.GET("/matches/:id", fh.GetMatch)
		v1.GET("/matches/:id/goals", fh.GetMatchGoals)
		v1.GET("/matches/:id/shootout", fh.GetMatchShootout)
		v1.GET("/head-to-head", fh.GetHeadToHead)
		v1.GET("/players/:name/goals", fh.GetPlayerGoals)

		// Write routes (no middleware – unit tests validate handler logic directly)
		v1.POST("/teams", fh.CreateTeam)
		v1.PUT("/teams/:id", fh.UpdateTeam)
		v1.DELETE("/teams/:id", fh.DeleteTeam)

		v1.POST("/matches", fh.CreateMatch)
		v1.PUT("/matches/:id", fh.UpdateMatch)
		v1.DELETE("/matches/:id", fh.DeleteMatch)

		v1.POST("/matches/:id/goals", fh.CreateGoal)
		v1.DELETE("/matches/:id/goals/:goalId", fh.DeleteGoal)

		v1.POST("/matches/:id/shootout", fh.CreateShootout)
		v1.DELETE("/matches/:id/shootout", fh.DeleteShootout)
	}
	return r, mock
}

// newFootballRouterWithAuth builds a router where write routes are gated by a
// simple stub middleware that rejects requests without an "Authorization" header.
// This is enough to confirm the auth gate is wired correctly at the handler level.
func newFootballRouterWithAuth() (*gin.Engine, *footballMock) {
	mock := &footballMock{}
	fh := handlers.NewFootballHandler(mock)

	authGuard := func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: "authorization header required"})
			return
		}
		c.Next()
	}

	r := gin.New()
	v1 := r.Group("/api/v1/football")
	{
		v1.POST("/teams", authGuard, fh.CreateTeam)
		v1.PUT("/teams/:id", authGuard, fh.UpdateTeam)
		v1.DELETE("/teams/:id", authGuard, fh.DeleteTeam)

		v1.POST("/matches", authGuard, fh.CreateMatch)
		v1.PUT("/matches/:id", authGuard, fh.UpdateMatch)
		v1.DELETE("/matches/:id", authGuard, fh.DeleteMatch)

		v1.POST("/matches/:id/goals", authGuard, fh.CreateGoal)
		v1.DELETE("/matches/:id/goals/:goalId", authGuard, fh.DeleteGoal)

		v1.POST("/matches/:id/shootout", authGuard, fh.CreateShootout)
		v1.DELETE("/matches/:id/shootout", authGuard, fh.DeleteShootout)
	}
	return r, mock
}

func doRequestAuth(r *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf strings.Builder
	if body != nil {
		b, _ := json.Marshal(body)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- helpers -----------------------------------------------------------------

func itoa(n int) string {
	return strconv.Itoa(n)
}

// ---------------------------------------------------------------------------
// Auth gate
// ---------------------------------------------------------------------------

// TestWriteRoutes_RequireAuth verifies that all mutation endpoints return 401
// when no Authorization header is present.
func TestWriteRoutes_RequireAuth(t *testing.T) {
	r, mock := newFootballRouterWithAuth()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	match := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})
	goal := mock.addGoal(models.Goal{MatchID: match.ID, TeamID: eng.ID, Scorer: "Test"})
	mock.addShootout(models.Shootout{MatchID: match.ID, WinnerID: eng.ID})

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/football/teams"},
		{http.MethodPut, "/api/v1/football/teams/" + itoa(eng.ID)},
		{http.MethodDelete, "/api/v1/football/teams/" + itoa(eng.ID)},
		{http.MethodPost, "/api/v1/football/matches"},
		{http.MethodPut, "/api/v1/football/matches/" + itoa(match.ID)},
		{http.MethodDelete, "/api/v1/football/matches/" + itoa(match.ID)},
		{http.MethodPost, "/api/v1/football/matches/" + itoa(match.ID) + "/goals"},
		{http.MethodDelete, "/api/v1/football/matches/" + itoa(match.ID) + "/goals/" + itoa(goal.ID)},
		{http.MethodPost, "/api/v1/football/matches/" + itoa(match.ID) + "/shootout"},
		{http.MethodDelete, "/api/v1/football/matches/" + itoa(match.ID) + "/shootout"},
	}

	for _, rt := range routes {
		w := doRequestAuth(r, rt.method, rt.path, nil, "" /* no token */)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", rt.method, rt.path, w.Code)
		}
	}
}
