package handlers_test

import (
"encoding/json"
"net/http"
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
// It lives here so the football handler tests do not depend on any in-memory
// package.
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

// newFootballRouter builds a minimal Gin engine wired to a fresh football mock.
func newFootballRouter() (*gin.Engine, *footballMock) {
mock := &footballMock{}
fh := handlers.NewFootballHandler(mock)

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
return r, mock
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

// --- helpers -----------------------------------------------------------------

func itoa(n int) string {
return strconv.Itoa(n)
}
