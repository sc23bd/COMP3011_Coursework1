package handlers_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db/memory"
	"github.com/sc23bd/COMP3011_Coursework1/internal/handlers"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// newFootballRouter builds a minimal Gin engine wired to a fresh football store.
func newFootballRouter() (*gin.Engine, *memory.FootballStore) {
	store := memory.NewFootballStore()
	fh := handlers.NewFootballHandler(store)

	r := gin.New()
	v1 := r.Group("/api/v1/football")
	{
		v1.GET("/teams", fh.ListTeams)
		v1.GET("/teams/:id", fh.GetTeam)
		v1.GET("/teams/:id/history", fh.GetTeamHistory)
		v1.GET("/matches", fh.ListMatches)
		v1.GET("/matches/:id", fh.GetMatch)
		v1.GET("/matches/:id/goals", fh.GetMatchGoals)
		v1.GET("/matches/:id/shootout", fh.GetMatchShootout)
		v1.GET("/head-to-head", fh.GetHeadToHead)
		v1.GET("/players/:name/goals", fh.GetPlayerGoals)
	}
	return r, store
}

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
	r, store := newFootballRouter()
	store.AddTeam("England")
	store.AddTeam("Brazil")

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
	r, store := newFootballRouter()
	team := store.AddTeam("Germany")

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
	r, store := newFootballRouter()
	team := store.AddTeam("France")

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
	r, store := newFootballRouter()
	eng := store.AddTeam("England")
	bra := store.AddTeam("Brazil")
	store.AddMatch(models.Match{
		Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
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
	r, store := newFootballRouter()
	eng := store.AddTeam("England")
	fra := store.AddTeam("France")
	m := store.AddMatch(models.Match{
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
	r, store := newFootballRouter()
	eng := store.AddTeam("England")
	ger := store.AddTeam("Germany")

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
	r, store := newFootballRouter()
	eng := store.AddTeam("England")
	ger := store.AddTeam("Germany")
	m := store.AddMatch(models.Match{
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
	r, store := newFootballRouter()
	eng := store.AddTeam("England")
	ger := store.AddTeam("Germany")
	m := store.AddMatch(models.Match{
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
	r, store := newFootballRouter()
	eng := store.AddTeam("England")
	ger := store.AddTeam("Germany")
	m := store.AddMatch(models.Match{
		Date:       time.Date(1990, 7, 4, 0, 0, 0, 0, time.UTC),
		HomeTeamID: eng.ID, HomeTeam: eng.Name,
		AwayTeamID: ger.ID, AwayTeam: ger.Name,
		HomeScore: 1, AwayScore: 1,
		Tournament: "FIFA World Cup",
	})
	store.AddShootout(models.Shootout{
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
	r, store := newFootballRouter()
	eng := store.AddTeam("England")
	ger := store.AddTeam("Germany")
	m := store.AddMatch(models.Match{
		Date:       time.Date(1966, 7, 30, 0, 0, 0, 0, time.UTC),
		HomeTeamID: eng.ID, HomeTeam: eng.Name,
		AwayTeamID: ger.ID, AwayTeam: ger.Name,
		HomeScore: 4, AwayScore: 2,
		Tournament: "FIFA World Cup",
	})
	store.AddGoal(models.Goal{MatchID: m.ID, TeamID: eng.ID, Team: eng.Name, Scorer: "Hurst"})
	store.AddGoal(models.Goal{MatchID: m.ID, TeamID: eng.ID, Team: eng.Name, Scorer: "Hurst"})
	store.AddGoal(models.Goal{MatchID: m.ID, TeamID: eng.ID, Team: eng.Name, Scorer: "Peters"})

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

// --- helpers -----------------------------------------------------------------

func itoa(n int) string {
	return strconv.Itoa(n)
}
