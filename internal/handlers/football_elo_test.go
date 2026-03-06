package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/handlers"
	elomodels "github.com/sc23bd/COMP3011_Coursework1/internal/elo"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// newEloRouter builds a minimal Gin engine for Elo endpoint tests.
func newEloRouter() (*gin.Engine, *footballMock) {
	mock := &footballMock{}
	fh := handlers.NewFootballHandler(mock)

	r := gin.New()
	v1 := r.Group("/api/v1/football")
	{
		v1.GET("/teams/:id/elo", fh.GetTeamElo)
		v1.GET("/teams/:id/elo/timeline", fh.GetTeamEloTimeline)
		v1.GET("/rankings/elo", fh.GetEloRankings)
		v1.POST("/rankings/elo/recalculate", fh.RecalculateEloRankings)
	}
	return r, mock
}

// ---------------------------------------------------------------------------
// GET /teams/:id/elo
// ---------------------------------------------------------------------------

func TestGetTeamElo_NotFound(t *testing.T) {
	r, _ := newEloRouter()
	notFound(t, r, "/api/v1/football/teams/999/elo")
}

func TestGetTeamElo_InvalidID(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/abc/elo", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetTeamElo_InvalidDate(t *testing.T) {
	r, mock := newEloRouter()
	mock.addTeam("Germany")
	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo?date=not-a-date", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetTeamElo_NoMatches(t *testing.T) {
	r, mock := newEloRouter()
	mock.addTeam("Germany")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo", nil)
	assertStatus(t, w, http.StatusOK)

	var resp elomodels.Rating
	decodeJSON(t, w, &resp)

	if resp.TeamID != 1 {
		t.Errorf("expected teamId=1, got %d", resp.TeamID)
	}
	if resp.TeamName != "Germany" {
		t.Errorf("expected teamName=Germany, got %s", resp.TeamName)
	}
	// A team with no matches should start at the default rating (1500).
	if resp.Elo != 1500.0 {
		t.Errorf("expected Elo=1500, got %.2f", resp.Elo)
	}
	if resp.MatchesConsidered != 0 {
		t.Errorf("expected 0 matchesConsidered, got %d", resp.MatchesConsidered)
	}
}

func TestGetTeamElo_HasHATEOASLinks(t *testing.T) {
	r, mock := newEloRouter()
	mock.addTeam("Germany")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo", nil)
	assertStatus(t, w, http.StatusOK)

	var resp elomodels.Rating
	decodeJSON(t, w, &resp)

	if len(resp.Links) == 0 {
		t.Fatal("expected HATEOAS links, got none")
	}
	hasTeamLink := false
	for _, l := range resp.Links {
		if l.Rel == "team" {
			hasTeamLink = true
		}
	}
	if !hasTeamLink {
		t.Error("expected 'team' HATEOAS link")
	}
}

func TestGetTeamElo_WithMatches(t *testing.T) {
	r, mock := newEloRouter()
	germany := mock.addTeam("Germany")
	england := mock.addTeam("England")

	// Germany beat England 2-0.
	mock.addMatch(models.Match{
		ID:         1,
		Date:       time.Date(2010, 7, 4, 0, 0, 0, 0, time.UTC),
		HomeTeamID: germany.ID,
		AwayTeamID: england.ID,
		HomeScore:  4,
		AwayScore:  1,
		Tournament: "FIFA World Cup",
		Neutral:    true,
	})

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo", nil)
	assertStatus(t, w, http.StatusOK)

	var resp elomodels.Rating
	decodeJSON(t, w, &resp)

	// Germany won, so their rating should be above the default.
	if resp.Elo <= 1500 {
		t.Errorf("Germany should have gained Elo after winning; got %.2f", resp.Elo)
	}
	if resp.MatchesConsidered != 1 {
		t.Errorf("expected matchesConsidered=1, got %d", resp.MatchesConsidered)
	}
}

func TestGetTeamElo_HistoricalDate(t *testing.T) {
	r, mock := newEloRouter()
	germany := mock.addTeam("Germany")
	england := mock.addTeam("England")

	// Match before the query date.
	mock.addMatch(models.Match{
		ID:         1,
		Date:       time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		HomeTeamID: germany.ID,
		AwayTeamID: england.ID,
		HomeScore:  2,
		AwayScore:  1,
		Tournament: "Friendly",
		Neutral:    true,
	})
	// Match after the query date (should be ignored).
	mock.addMatch(models.Match{
		ID:         2,
		Date:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		HomeTeamID: england.ID,
		AwayTeamID: germany.ID,
		HomeScore:  3,
		AwayScore:  0,
		Tournament: "Friendly",
		Neutral:    true,
	})

	// Query as of 2005: only the first match should count.
	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo?date=2005-01-01", nil)
	assertStatus(t, w, http.StatusOK)

	var resp elomodels.Rating
	decodeJSON(t, w, &resp)

	if resp.MatchesConsidered != 1 {
		t.Errorf("expected 1 match considered for historical query, got %d", resp.MatchesConsidered)
	}
}

func TestGetTeamElo_XComputedAtHeader(t *testing.T) {
	r, mock := newEloRouter()
	mock.addTeam("Germany")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo", nil)
	assertStatus(t, w, http.StatusOK)

	checkHeader(t, w, "X-Elo-Computed-At")
}

func TestGetTeamElo_MethodologyPresent(t *testing.T) {
	r, mock := newEloRouter()
	mock.addTeam("Germany")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo", nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w, &resp)

	if _, ok := resp["methodology"]; !ok {
		t.Error("expected 'methodology' field in response")
	}
	meth, _ := resp["methodology"].(map[string]interface{})
	if meth["formulaReference"] == "" {
		t.Error("expected non-empty formulaReference in methodology")
	}
}

// ---------------------------------------------------------------------------
// GET /teams/:id/elo/timeline
// ---------------------------------------------------------------------------

func TestGetTeamEloTimeline_NotFound(t *testing.T) {
	r, _ := newEloRouter()
	notFound(t, r, "/api/v1/football/teams/999/elo/timeline")
}

func TestGetTeamEloTimeline_EmptyHistory(t *testing.T) {
	r, mock := newEloRouter()
	mock.addTeam("Germany")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo/timeline", nil)
	assertStatus(t, w, http.StatusOK)

	var resp elomodels.TimelineResponse
	decodeJSON(t, w, &resp)

	if resp.TeamID != 1 {
		t.Errorf("expected teamId=1, got %d", resp.TeamID)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected empty data, got %d entries", len(resp.Data))
	}
}

func TestGetTeamEloTimeline_HasMatchEntries(t *testing.T) {
	r, mock := newEloRouter()
	germany := mock.addTeam("Germany")
	england := mock.addTeam("England")

	mock.addMatch(models.Match{
		ID:         1,
		Date:       time.Date(2010, 7, 4, 0, 0, 0, 0, time.UTC),
		HomeTeamID: germany.ID,
		AwayTeamID: england.ID,
		HomeScore:  4,
		AwayScore:  1,
		Tournament: "FIFA World Cup",
		Neutral:    true,
	})

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo/timeline", nil)
	assertStatus(t, w, http.StatusOK)

	var resp elomodels.TimelineResponse
	decodeJSON(t, w, &resp)

	if len(resp.Data) != 1 {
		t.Errorf("expected 1 timeline entry, got %d", len(resp.Data))
	}
	if resp.Data[0].Result != "W" {
		t.Errorf("expected result=W, got %s", resp.Data[0].Result)
	}
}

func TestGetTeamEloTimeline_InvalidDate(t *testing.T) {
	r, mock := newEloRouter()
	mock.addTeam("Germany")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo/timeline?start_date=bad", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

// ---------------------------------------------------------------------------
// GET /rankings/elo
// ---------------------------------------------------------------------------

func TestGetEloRankings_OK(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/rankings/elo", nil)
	assertStatus(t, w, http.StatusOK)

	var resp elomodels.RankingsResponse
	decodeJSON(t, w, &resp)

	if resp.Limit != 50 {
		t.Errorf("expected default limit=50, got %d", resp.Limit)
	}
}

func TestGetEloRankings_InvalidLimit(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/rankings/elo?limit=abc", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetEloRankings_InvalidOffset(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/rankings/elo?offset=-1", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetEloRankings_InvalidDate(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/rankings/elo?date=notadate", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetEloRankings_HasLinks(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/rankings/elo", nil)
	assertStatus(t, w, http.StatusOK)

	var resp elomodels.RankingsResponse
	decodeJSON(t, w, &resp)

	if len(resp.Links) == 0 {
		t.Error("expected HATEOAS links in rankings response")
	}
}

// ---------------------------------------------------------------------------
// POST /rankings/elo/recalculate
// ---------------------------------------------------------------------------

func TestRecalculateEloRankings_Accepted(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/rankings/elo/recalculate", nil)
	assertStatus(t, w, http.StatusAccepted)

	var resp elomodels.RecalculateResponse
	decodeJSON(t, w, &resp)

	if resp.Message == "" {
		t.Error("expected non-empty message in recalculate response")
	}
}

func TestRecalculateEloRankings_InvalidTeamID(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/rankings/elo/recalculate?team_id=abc", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestRecalculateEloRankings_TeamNotFound(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/rankings/elo/recalculate?team_id=999", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestRecalculateEloRankings_NoCacheHeader(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/football/rankings/elo/recalculate", nil)
	assertStatus(t, w, http.StatusAccepted)

	cc := w.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("expected Cache-Control: no-store, got %q", cc)
	}
}

// ---------------------------------------------------------------------------
// Response shape sanity check
// ---------------------------------------------------------------------------

func TestGetTeamElo_ResponseShape(t *testing.T) {
	r, mock := newEloRouter()
	mock.addTeam("Germany")

	w := doRequest(r, http.MethodGet, "/api/v1/football/teams/1/elo", nil)
	assertStatus(t, w, http.StatusOK)

	var raw map[string]json.RawMessage
	decodeJSON(t, w, &raw)

	requiredFields := []string{"teamId", "teamName", "date", "elo", "matchesConsidered", "methodology", "links"}
	for _, f := range requiredFields {
		if _, ok := raw[f]; !ok {
			t.Errorf("missing required field %q in Elo response", f)
		}
	}
}

// ---------------------------------------------------------------------------
// X-Cache-Status header
// ---------------------------------------------------------------------------

// TestGetEloRankings_CacheStatusMiss verifies that an empty (cache-miss) response
// sets X-Cache-Status: miss so clients know the cache must be pre-warmed.
func TestGetEloRankings_CacheStatusMiss(t *testing.T) {
	r, _ := newEloRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/football/rankings/elo", nil)
	assertStatus(t, w, http.StatusOK)

	status := w.Header().Get("X-Cache-Status")
	if status != "miss" {
		t.Errorf("expected X-Cache-Status: miss for empty rankings, got %q", status)
	}
}

// ---------------------------------------------------------------------------
// Rate limiting on /recalculate
// ---------------------------------------------------------------------------

// TestRecalculateEloRankings_RateLimited verifies that a second recalculation
// request (without ?force=true) within the cooldown window returns 429.
func TestRecalculateEloRankings_RateLimited(t *testing.T) {
	// Use a fresh handler+router so the rate-limit state starts clean.
	mock := &footballMock{}
	fh := handlers.NewFootballHandler(mock)
	r := gin.New()
	r.POST("/api/v1/football/rankings/elo/recalculate", fh.RecalculateEloRankings)

	// First request should be accepted.
	w1 := doRequest(r, http.MethodPost, "/api/v1/football/rankings/elo/recalculate", nil)
	assertStatus(t, w1, http.StatusAccepted)

	// Wait briefly for the background goroutine to complete and mark lastRun.
	time.Sleep(50 * time.Millisecond)

	// Second request without force should be rate-limited (429).
	w2 := doRequest(r, http.MethodPost, "/api/v1/football/rankings/elo/recalculate", nil)
	assertStatus(t, w2, http.StatusTooManyRequests)
}

// TestRecalculateEloRankings_ForceBypassesRateLimit verifies that ?force=true
// skips the rate-limit check and returns 202.
func TestRecalculateEloRankings_ForceBypassesRateLimit(t *testing.T) {
	mock := &footballMock{}
	fh := handlers.NewFootballHandler(mock)
	r := gin.New()
	r.POST("/api/v1/football/rankings/elo/recalculate", fh.RecalculateEloRankings)

	// First request sets lastRun.
	w1 := doRequest(r, http.MethodPost, "/api/v1/football/rankings/elo/recalculate", nil)
	assertStatus(t, w1, http.StatusAccepted)

	// Wait briefly for goroutine to finish.
	time.Sleep(50 * time.Millisecond)

	// Second request with force=true should succeed despite the cooldown.
	w2 := doRequest(r, http.MethodPost, "/api/v1/football/rankings/elo/recalculate?force=true", nil)
	assertStatus(t, w2, http.StatusAccepted)
}
