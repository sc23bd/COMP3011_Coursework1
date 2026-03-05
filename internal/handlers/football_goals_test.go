package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

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
