package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

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

// --- CreateMatch -------------------------------------------------------------

func TestCreateMatch_Success(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	tourn := mock.addTournament("FIFA World Cup")

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches", map[string]interface{}{
		"date":         "1966-07-30T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   ger.ID,
		"homeScore":    4,
		"awayScore":    2,
		"tournamentId": tourn.ID,
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

func TestCreateMatch_TournamentNotFound(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")

	w := doRequest(r, http.MethodPost, "/api/v1/football/matches", map[string]interface{}{
		"date":         "1966-07-30T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   ger.ID,
		"homeScore":    0,
		"awayScore":    0,
		"tournamentId": 999,
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
	tourn := mock.addTournament("FIFA World Cup")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID,
		HomeScore: 1, AwayScore: 1, TournamentID: tourn.ID,
	})

	w := doRequest(r, http.MethodPut, "/api/v1/football/matches/"+itoa(m.ID), map[string]interface{}{
		"date":         "1990-07-04T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   ger.ID,
		"homeScore":    1,
		"awayScore":    2,
		"tournamentId": tourn.ID,
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
	tourn := mock.addTournament("FIFA World Cup")

	w := doRequest(r, http.MethodPut, "/api/v1/football/matches/999", map[string]interface{}{
		"date":         "1990-07-04T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   ger.ID,
		"homeScore":    0,
		"awayScore":    0,
		"tournamentId": tourn.ID,
	})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateMatch_HomeTeamNotFound(t *testing.T) {
	r, mock := newFootballRouter()
	ger := mock.addTeam("Germany")
	tourn := mock.addTournament("FIFA World Cup")
	m := mock.addMatch(models.Match{
		HomeTeamID: 1, AwayTeamID: ger.ID, TournamentID: tourn.ID,
	})

	w := doRequest(r, http.MethodPut, "/api/v1/football/matches/"+itoa(m.ID), map[string]interface{}{
		"date":         "1990-07-04T00:00:00Z",
		"homeTeamId":   999,
		"awayTeamId":   ger.ID,
		"homeScore":    0,
		"awayScore":    0,
		"tournamentId": tourn.ID,
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMatch_AwayTeamNotFound(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	tourn := mock.addTournament("FIFA World Cup")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: 1, TournamentID: tourn.ID,
	})

	w := doRequest(r, http.MethodPut, "/api/v1/football/matches/"+itoa(m.ID), map[string]interface{}{
		"date":         "1990-07-04T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   999,
		"homeScore":    0,
		"awayScore":    0,
		"tournamentId": tourn.ID,
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMatch_TournamentNotFound(t *testing.T) {
	r, mock := newFootballRouter()
	eng := mock.addTeam("England")
	ger := mock.addTeam("Germany")
	m := mock.addMatch(models.Match{
		HomeTeamID: eng.ID, AwayTeamID: ger.ID, TournamentID: 1,
	})

	w := doRequest(r, http.MethodPut, "/api/v1/football/matches/"+itoa(m.ID), map[string]interface{}{
		"date":         "1990-07-04T00:00:00Z",
		"homeTeamId":   eng.ID,
		"awayTeamId":   ger.ID,
		"homeScore":    0,
		"awayScore":    0,
		"tournamentId": 999,
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
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
