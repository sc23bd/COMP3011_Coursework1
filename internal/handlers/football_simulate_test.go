package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/handlers"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// newSimulateRouter builds a minimal Gin engine for the simulate endpoint tests.
func newSimulateRouter() (*gin.Engine, *footballMock) {
	mock := &footballMock{}
	fh := handlers.NewFootballHandler(mock)

	r := gin.New()
	r.POST("/api/v1/football/matches/simulate", fh.SimulateMatch)
	return r, mock
}

// ---------------------------------------------------------------------------
// Basic request/response
// ---------------------------------------------------------------------------

func TestSimulateMatch_OK(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Germany")
	away := mock.addTeam("France")

	// Give both teams some match history.
	mock.addMatch(models.Match{
		ID:         1,
		Date:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		HomeTeamID: home.ID,
		AwayTeamID: away.ID,
		HomeScore:  2,
		AwayScore:  1,
		Tournament: "Friendly",
		Neutral:    true,
	})

	body := models.SimulateRequest{
		HomeTeamID:  home.ID,
		AwayTeamID:  away.ID,
		Venue:       "neutral",
		Simulations: 500,
	}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var resp models.SimulateResponse
	decodeJSON(t, w, &resp)

	if resp.HomeTeam != "Germany" {
		t.Errorf("expected HomeTeam=Germany, got %s", resp.HomeTeam)
	}
	if resp.AwayTeam != "France" {
		t.Errorf("expected AwayTeam=France, got %s", resp.AwayTeam)
	}
	if resp.Simulations != 500 {
		t.Errorf("expected Simulations=500, got %d", resp.Simulations)
	}
}

// TestSimulateMatch_ProbabilitiesSumToOne verifies the probability invariant.
func TestSimulateMatch_ProbabilitiesSumToOne(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Spain")
	away := mock.addTeam("England")

	body := models.SimulateRequest{
		HomeTeamID:  home.ID,
		AwayTeamID:  away.ID,
		Simulations: 2000,
	}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var resp models.SimulateResponse
	decodeJSON(t, w, &resp)

	sum := resp.Outcome.HomeWinPct + resp.Outcome.DrawPct + resp.Outcome.AwayWinPct
	if sum < 0.9998 || sum > 1.0002 {
		t.Errorf("outcome probabilities should sum to 1.0, got %.4f", sum)
	}
}

// TestSimulateMatch_HasHATEOASLinks verifies that HATEOAS links are included.
func TestSimulateMatch_HasHATEOASLinks(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Brazil")
	away := mock.addTeam("Argentina")

	body := models.SimulateRequest{
		HomeTeamID:  home.ID,
		AwayTeamID:  away.ID,
		Simulations: 200,
	}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var resp models.SimulateResponse
	decodeJSON(t, w, &resp)

	if len(resp.Links) == 0 {
		t.Fatal("expected HATEOAS links, got none")
	}
	hasSelf := false
	for _, l := range resp.Links {
		if l.Rel == "self" {
			hasSelf = true
		}
	}
	if !hasSelf {
		t.Error("expected 'self' HATEOAS link")
	}
}

// TestSimulateMatch_NoCacheHeader verifies Cache-Control: no-store is set.
func TestSimulateMatch_NoCacheHeader(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Italy")
	away := mock.addTeam("Belgium")

	body := models.SimulateRequest{HomeTeamID: home.ID, AwayTeamID: away.ID, Simulations: 100}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	cc := w.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("expected Cache-Control: no-store, got %q", cc)
	}
}

// ---------------------------------------------------------------------------
// Input validation
// ---------------------------------------------------------------------------

func TestSimulateMatch_MissingBody(t *testing.T) {
	r, _ := newSimulateRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSimulateMatch_MissingHomeTeam(t *testing.T) {
	r, mock := newSimulateRouter()
	away := mock.addTeam("Portugal")
	body := models.SimulateRequest{AwayTeamID: away.ID, Simulations: 100}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSimulateMatch_MissingAwayTeam(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Portugal")
	body := models.SimulateRequest{HomeTeamID: home.ID, Simulations: 100}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSimulateMatch_SameTeam(t *testing.T) {
	r, mock := newSimulateRouter()
	team := mock.addTeam("Netherlands")
	body := models.SimulateRequest{HomeTeamID: team.ID, AwayTeamID: team.ID, Simulations: 100}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSimulateMatch_HomeTeamNotFound(t *testing.T) {
	r, mock := newSimulateRouter()
	away := mock.addTeam("Sweden")
	body := models.SimulateRequest{HomeTeamID: 9999, AwayTeamID: away.ID, Simulations: 100}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSimulateMatch_AwayTeamNotFound(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Sweden")
	body := models.SimulateRequest{HomeTeamID: home.ID, AwayTeamID: 9999, Simulations: 100}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSimulateMatch_InvalidDate(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Chile")
	away := mock.addTeam("Colombia")
	body := models.SimulateRequest{
		HomeTeamID:  home.ID,
		AwayTeamID:  away.ID,
		Date:        "not-a-date",
		Simulations: 100,
	}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusBadRequest)
}

// ---------------------------------------------------------------------------
// Response shape
// ---------------------------------------------------------------------------

func TestSimulateMatch_ResponseShape(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Japan")
	away := mock.addTeam("South Korea")

	body := models.SimulateRequest{HomeTeamID: home.ID, AwayTeamID: away.ID, Simulations: 200}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var raw map[string]json.RawMessage
	decodeJSON(t, w, &raw)

	required := []string{
		"homeTeam", "awayTeam", "venue", "asOf", "simulations",
		"homeElo", "awayElo", "outcome", "expectedScore", "upsetProbability", "links",
	}
	for _, f := range required {
		if _, ok := raw[f]; !ok {
			t.Errorf("missing required field %q in simulate response", f)
		}
	}
}

func TestSimulateMatch_OutcomeShape(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Mexico")
	away := mock.addTeam("USA")

	body := models.SimulateRequest{HomeTeamID: home.ID, AwayTeamID: away.ID, Simulations: 200}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var resp struct {
		Outcome map[string]json.RawMessage `json:"outcome"`
	}
	decodeJSON(t, w, &resp)

	requiredOutcomeFields := []string{
		"homeWinPct", "drawPct", "awayWinPct",
		"homeWinCI", "drawCI", "awayWinCI",
	}
	for _, f := range requiredOutcomeFields {
		if _, ok := resp.Outcome[f]; !ok {
			t.Errorf("missing required outcome field %q", f)
		}
	}
}

// ---------------------------------------------------------------------------
// Venue variants
// ---------------------------------------------------------------------------

func TestSimulateMatch_VenueHome(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Australia")
	away := mock.addTeam("New Zealand")
	body := models.SimulateRequest{HomeTeamID: home.ID, AwayTeamID: away.ID, Venue: "home", Simulations: 500}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var resp models.SimulateResponse
	decodeJSON(t, w, &resp)
	if resp.Venue != "home" {
		t.Errorf("expected venue=home, got %s", resp.Venue)
	}
}

func TestSimulateMatch_VenueAway(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Canada")
	away := mock.addTeam("Costa Rica")
	body := models.SimulateRequest{HomeTeamID: home.ID, AwayTeamID: away.ID, Venue: "away", Simulations: 500}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var resp models.SimulateResponse
	decodeJSON(t, w, &resp)
	if resp.Venue != "away" {
		t.Errorf("expected venue=away, got %s", resp.Venue)
	}
}

func TestSimulateMatch_DefaultsToNeutralVenue(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Morocco")
	away := mock.addTeam("Senegal")
	body := models.SimulateRequest{HomeTeamID: home.ID, AwayTeamID: away.ID, Simulations: 500}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var resp models.SimulateResponse
	decodeJSON(t, w, &resp)
	if resp.Venue != "neutral" {
		t.Errorf("expected venue=neutral (default), got %s", resp.Venue)
	}
}

// ---------------------------------------------------------------------------
// EloRating fields present
// ---------------------------------------------------------------------------

func TestSimulateMatch_EloFieldsPresent(t *testing.T) {
	r, mock := newSimulateRouter()
	home := mock.addTeam("Iran")
	away := mock.addTeam("Saudi Arabia")
	body := models.SimulateRequest{HomeTeamID: home.ID, AwayTeamID: away.ID, Simulations: 100}
	w := doRequest(r, http.MethodPost, "/api/v1/football/matches/simulate", body)
	assertStatus(t, w, http.StatusOK)

	var resp models.SimulateResponse
	decodeJSON(t, w, &resp)
	if resp.HomeElo <= 0 {
		t.Errorf("expected positive homeElo, got %.2f", resp.HomeElo)
	}
	if resp.AwayElo <= 0 {
		t.Errorf("expected positive awayElo, got %.2f", resp.AwayElo)
	}
}
