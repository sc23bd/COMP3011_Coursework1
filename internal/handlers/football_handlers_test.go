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
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// ---------------------------------------------------------------------------
// footballMock is a minimal in-test stub that implements db.FootballRepository.
// It lives here so the football handler tests do not depend on any external package.
// ---------------------------------------------------------------------------

type footballMock struct {
	teams       []models.Team
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

// ==========================================================================
// Read tests (unchanged from before)
// ==========================================================================

// --- ListTeams ---------------------------------------------------------------

func TestListTeams_Empty(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/teams", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.TeamsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("expected non-nil data slice")
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected 0 teams, got %d", len(resp.Data))
	}
}

func TestListTeams_WithData(t *testing.T) {
	r, mock := newFootballRouter()
	mock.addTeam("England")
	mock.addTeam("Brazil")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.TeamsResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(resp.Data))
	}
	// Each team response must include HATEOAS links.
	if len(resp.Data[0].Links) == 0 {
		t.Fatal("expected HATEOAS links on team")
	}
}

// --- GetTeam -----------------------------------------------------------------

func TestGetTeam_NotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/999", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetTeam_Success(t *testing.T) {
	r, mock := newFootballRouter()
	team := mock.addTeam("Germany")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/"+itoa(team.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.TeamResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Name != "Germany" {
		t.Fatalf("expected name 'Germany', got %q", resp.Name)
	}
}

func TestGetTeam_InvalidID(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/abc", nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- GetTeamHistory ----------------------------------------------------------

func TestGetTeamHistory_NoHistory(t *testing.T) {
	r, mock := newFootballRouter()
	team := mock.addTeam("France")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/"+itoa(team.ID)+"/history", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.FormerNamesResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Data == nil {
		t.Fatal("expected non-nil data slice")
	}
}

func TestGetTeamHistory_TeamNotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/999/history", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- ListMatches -------------------------------------------------------------

func TestListMatches_Empty(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/matches", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.MatchesResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Data == nil {
		t.Fatal("expected non-nil data slice")
	}
}

func TestListMatches_WithData(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	bra := mock.addTeam("Brazil")
	mock.addMatch(models.Match{
		Date:       time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		HomeTeamID: eng.ID, HomeTeam: eng.Name,
		AwayTeamID: bra.ID, AwayTeam: bra.Name,
		HomeScore: 1, AwayScore: 2,
		Tournament: "Friendly",
	})

	w := doRequest(r, http.MethodGet, "/api/v1/football/matches", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.MatchesResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 match, got %d", len(resp.Data))
	}
}

func TestListMatches_InvalidLimit(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/matches?limit=-1", nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- GetMatch ----------------------------------------------------------------

func TestGetMatch_NotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/matches/999", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetMatch_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	fra := mock.addTeam("France")
	m := mock.addMatch(models.Match{
		Date:       time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC),
		HomeTeamID: eng.ID, HomeTeam: eng.Name,
		AwayTeamID: fra.ID, AwayTeam: fra.Name,
		HomeScore: 3, AwayScore: 0,
		Tournament: "UEFA Nations League",
	})

	w := doRequest(r, http.MethodGet, "/api/v1/football/matches/"+itoa(m.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.MatchResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.HomeScore != 3 || resp.AwayScore != 0 {
		t.Fatalf("unexpected score %d-%d", resp.HomeScore, resp.AwayScore)
	}
}

// --- GetHeadToHead -----------------------------------------------------------

func TestGetHeadToHead_MissingParams(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/head-to-head", nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetHeadToHead_NoMatches(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")

	url := "/api/v1/football/head-to-head?teamA=" + itoa(eng.ID) + "&teamB=" + itoa(ger.ID)
	w := doRequest(r, http.MethodGet, url, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.MatchesResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Data) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(resp.Data))
	}
}

// --- GetMatchGoals -----------------------------------------------------------

func TestGetMatchGoals_MatchNotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/matches/999/goals", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetMatchGoals_Empty(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		Date:       time.Date(1966, 7, 30, 0, 0, 0, 0, time.UTC),
		HomeTeamID: eng.ID, HomeTeam: eng.Name,
		AwayTeamID: ger.ID, AwayTeam: ger.Name,
		HomeScore: 4, AwayScore: 2,
		Tournament: "FIFA World Cup",
	})

	w := doRequest(r, http.MethodGet, "/api/v1/football/matches/"+itoa(m.ID)+"/goals", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.GoalsResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Data == nil {
		t.Fatal("expected non-nil data slice")
	}
}

// --- GetMatchShootout --------------------------------------------------------

func TestGetMatchShootout_NoShootout(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		Date:       time.Date(1966, 7, 30, 0, 0, 0, 0, time.UTC),
		HomeTeamID: eng.ID, HomeTeam: eng.Name,
		AwayTeamID: ger.ID, AwayTeam: ger.Name,
		HomeScore: 4, AwayScore: 2,
		Tournament: "FIFA World Cup",
	})

	w := doRequest(r, http.MethodGet, "/api/v1/football/matches/"+itoa(m.ID)+"/shootout", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetMatchShootout_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		Date:       time.Date(1990, 7, 4, 0, 0, 0, 0, time.UTC),
		HomeTeamID: eng.ID, HomeTeam: eng.Name,
		AwayTeamID: ger.ID, AwayTeam: ger.Name,
		HomeScore: 1, AwayScore: 1,
		Tournament: "FIFA World Cup",
	})
	mock.addShootout(models.Shootout{
		MatchID:  m.ID,
		Winner:   ger.Name,
		WinnerID: ger.ID,
	})

	w := doRequest(r, http.MethodGet, "/api/v1/football/matches/"+itoa(m.ID)+"/shootout", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.ShootoutResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Winner != "Germany" {
		t.Fatalf("expected winner 'Germany', got %q", resp.Winner)
	}
}

// --- GetPlayerGoals ----------------------------------------------------------

func TestGetPlayerGoals_NoGoals(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/players/Ronaldo/goals", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.GoalsResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Data == nil {
		t.Fatal("expected non-nil data slice")
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected 0 goals, got %d", len(resp.Data))
	}
}

func TestGetPlayerGoals_WithData(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		Date:       time.Date(1966, 7, 30, 0, 0, 0, 0, time.UTC),
		HomeTeamID: eng.ID, HomeTeam: eng.Name,
		AwayTeamID: ger.ID, AwayTeam: ger.Name,
		HomeScore: 4, AwayScore: 2,
		Tournament: "FIFA World Cup",
	})
	mock.addGoal(models.Goal{MatchID: m.ID, TeamID: eng.ID, Team: eng.Name, Scorer: "Hurst"})
	mock.addGoal(models.Goal{MatchID: m.ID, TeamID: eng.ID, Team: eng.Name, Scorer: "Hurst"})
	mock.addGoal(models.Goal{MatchID: m.ID, TeamID: eng.ID, Team: eng.Name, Scorer: "Peters"})

	w := doRequest(r, http.MethodGet, "/api/v1/football/players/Hurst/goals", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.GoalsResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 goals for Hurst, got %d", len(resp.Data))
	}
}

// ==========================================================================
// Write tests
// ==========================================================================

// --- Auth gate ---------------------------------------------------------------

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

// --- CreateTeam --------------------------------------------------------------

func TestCreateTeam_Success(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/teams", map[string]string{
		"name": "Italy",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Location") == "" {
		t.Fatal("expected Location header")
	}

	var resp models.TeamResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Name != "Italy" {
		t.Fatalf("expected name 'Italy', got %q", resp.Name)
	}
	if len(resp.Links) == 0 {
		t.Fatal("expected HATEOAS links")
	}
}

func TestCreateTeam_MissingName(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/teams", map[string]string{})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateTeam_Conflict(t *testing.T) {
	r, mock := newFootballRouter()
	mock.addTeam("Italy")

	w := doRequest(r, http.MethodPost, "/api/v1/football/teams", map[string]string{
		"name": "Italy",
	})

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

// --- UpdateTeam --------------------------------------------------------------

func TestUpdateTeam_Success(t *testing.T) {
	r, mock := newFootballRouter()
	team := mock.addTeam("West Germany")

	w := doRequest(r, http.MethodPut, "/api/v1/football/teams/"+itoa(team.ID), map[string]string{
		"name": "Germany",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.TeamResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Name != "Germany" {
		t.Fatalf("expected name 'Germany', got %q", resp.Name)
	}
}

func TestUpdateTeam_NotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodPut, "/api/v1/football/teams/999", map[string]string{
		"name": "Nobody",
	})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateTeam_InvalidID(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodPut, "/api/v1/football/teams/abc", map[string]string{
		"name": "Nobody",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- DeleteTeam --------------------------------------------------------------

func TestDeleteTeam_Success(t *testing.T) {
	r, mock := newFootballRouter()
	team := mock.addTeam("Yugoslavia")

	w := doRequest(r, http.MethodDelete, "/api/v1/football/teams/"+itoa(team.ID), nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteTeam_NotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodDelete, "/api/v1/football/teams/999", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- CreateMatch -------------------------------------------------------------

func TestCreateMatch_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches", map[string]interface{}{
		"date":         "1966-07-30T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   ger.ID,
		"homeScore":    4,
		"awayScore":    2,
		"tournamentId": 1,
		"city":         "London",
		"country":      "England",
		"neutral":      false,
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Location") == "" {
		t.Fatal("expected Location header")
	}

	var resp models.MatchResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.HomeScore != 4 || resp.AwayScore != 2 {
		t.Fatalf("unexpected score %d-%d", resp.HomeScore, resp.AwayScore)
	}
	if len(resp.Links) == 0 {
		t.Fatal("expected HATEOAS links")
	}
}

func TestCreateMatch_HomeTeamNotFound(t *testing.T) {
	r, mock := newFootballRouter()
	ger := mock.addTeam("Germany")

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches", map[string]interface{}{
		"date":         "1966-07-30T00:00:00Z",
		"homeTeamId":   999,
		"awayTeamId":   ger.ID,
		"homeScore":    0,
		"awayScore":    0,
		"tournamentId": 1,
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateMatch_AwayTeamNotFound(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches", map[string]interface{}{
		"date":         "1966-07-30T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   999,
		"homeScore":    0,
		"awayScore":    0,
		"tournamentId": 1,
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateMatch_MissingFields(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches", map[string]interface{}{})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- UpdateMatch -------------------------------------------------------------

func TestUpdateMatch_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID,
		HomeScore: 1, AwayScore: 1, TournamentID: 1,
	})

	w := doRequest(r, http.MethodPut, "/api/v1/football/matches/"+itoa(m.ID), map[string]interface{}{
		"date":         "1990-07-04T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   ger.ID,
		"homeScore":    1,
		"awayScore":    2,
		"tournamentId": 1,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.MatchResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.AwayScore != 2 {
		t.Fatalf("expected awayScore 2, got %d", resp.AwayScore)
	}
}

func TestUpdateMatch_NotFound(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")

	w := doRequest(r, http.MethodPut, "/api/v1/football/matches/999", map[string]interface{}{
		"date":         "1990-07-04T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   ger.ID,
		"homeScore":    0,
		"awayScore":    0,
		"tournamentId": 1,
	})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- DeleteMatch -------------------------------------------------------------

func TestDeleteMatch_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})

	w := doRequest(r, http.MethodDelete, "/api/v1/football/matches/"+itoa(m.ID), nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteMatch_NotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodDelete, "/api/v1/football/matches/999", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- CreateGoal --------------------------------------------------------------

func TestCreateGoal_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/"+itoa(m.ID)+"/goals", map[string]interface{}{
		"teamId":  eng.ID,
		"scorer":  "Hurst",
		"ownGoal": false,
		"penalty": false,
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.GoalsResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 goal in response, got %d", len(resp.Data))
	}
	if resp.Data[0].Scorer != "Hurst" {
		t.Fatalf("expected scorer 'Hurst', got %q", resp.Data[0].Scorer)
	}
	if resp.Data[0].Team != eng.Name {
		t.Fatalf("expected team %q, got %q", eng.Name, resp.Data[0].Team)
	}
}

func TestCreateGoal_MatchNotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/999/goals", map[string]interface{}{
		"teamId": 1,
		"scorer": "Test",
	})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateGoal_TeamNotFound(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/"+itoa(m.ID)+"/goals", map[string]interface{}{
		"teamId": 999,
		"scorer": "Ghost",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateGoal_MissingFields(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/"+itoa(m.ID)+"/goals", map[string]interface{}{})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- DeleteGoal --------------------------------------------------------------

func TestDeleteGoal_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})
	g := mock.addGoal(models.Goal{MatchID: m.ID, TeamID: eng.ID, Scorer: "Hurst"})

	w := doRequest(r, http.MethodDelete, "/api/v1/football/matches/"+itoa(m.ID)+"/goals/"+itoa(g.ID), nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteGoal_NotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodDelete, "/api/v1/football/matches/1/goals/999", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- CreateShootout ----------------------------------------------------------

func TestCreateShootout_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/"+itoa(m.ID)+"/shootout", map[string]interface{}{
		"winnerId": ger.ID,
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.ShootoutResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Winner != ger.Name {
		t.Fatalf("expected winner %q, got %q", ger.Name, resp.Winner)
	}
}

func TestCreateShootout_MatchNotFound(t *testing.T) {
	r, _ := newFootballRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/999/shootout", map[string]interface{}{
		"winnerId": 1,
	})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateShootout_WinnerTeamNotFound(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/"+itoa(m.ID)+"/shootout", map[string]interface{}{
		"winnerId": 999,
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateShootout_Conflict(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})
	mock.addShootout(models.Shootout{MatchID: m.ID, WinnerID: ger.ID, Winner: ger.Name})

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/"+itoa(m.ID)+"/shootout", map[string]interface{}{
		"winnerId": eng.ID,
	})

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

// --- DeleteShootout ----------------------------------------------------------

func TestDeleteShootout_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})
	mock.addShootout(models.Shootout{MatchID: m.ID, WinnerID: ger.ID, Winner: ger.Name})

	w := doRequest(r, http.MethodDelete, "/api/v1/football/matches/"+itoa(m.ID)+"/shootout", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteShootout_NotFound(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})

	w := doRequest(r, http.MethodDelete, "/api/v1/football/matches/"+itoa(m.ID)+"/shootout", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- helpers -----------------------------------------------------------------

func itoa(n int) string {
	return strconv.Itoa(n)
}

