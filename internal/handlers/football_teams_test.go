package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

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
