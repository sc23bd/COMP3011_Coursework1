package memory

import (
	"strings"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// FootballStore is an in-memory implementation of db.FootballRepository.
// It is used in tests and local development when no database is available.
type FootballStore struct {
	teams      []models.Team
	matches    []models.Match
	goals      []models.Goal
	shootouts  []models.Shootout
	formerNames []models.FormerName
}

// NewFootballStore returns an empty, initialised FootballStore.
func NewFootballStore() *FootballStore {
	return &FootballStore{}
}

// ListTeams returns all teams ordered by ID.
func (s *FootballStore) ListTeams() ([]models.Team, error) {
	result := make([]models.Team, len(s.teams))
	copy(result, s.teams)
	return result, nil
}

// GetTeamByID returns the team with the given ID.
// Returns ErrNotFound when no matching team exists.
func (s *FootballStore) GetTeamByID(id int) (models.Team, error) {
	for _, t := range s.teams {
		if t.ID == id {
			return t, nil
		}
	}
	return models.Team{}, models.ErrNotFound
}

// GetTeamHistory returns the former names for a team.
func (s *FootballStore) GetTeamHistory(teamID int) ([]models.FormerName, error) {
	var result []models.FormerName
	for _, fn := range s.formerNames {
		if fn.TeamID == teamID {
			result = append(result, fn)
		}
	}
	return result, nil
}

// ListMatches returns a paginated slice of matches.
func (s *FootballStore) ListMatches(limit, offset int) ([]models.Match, error) {
	if offset >= len(s.matches) {
		return []models.Match{}, nil
	}
	end := offset + limit
	if end > len(s.matches) {
		end = len(s.matches)
	}
	result := make([]models.Match, end-offset)
	copy(result, s.matches[offset:end])
	return result, nil
}

// GetMatchByID returns the match with the given ID.
// Returns ErrNotFound when no matching match exists.
func (s *FootballStore) GetMatchByID(id int) (models.Match, error) {
	for _, m := range s.matches {
		if m.ID == id {
			return m, nil
		}
	}
	return models.Match{}, models.ErrNotFound
}

// GetHeadToHead returns all matches between two teams.
func (s *FootballStore) GetHeadToHead(teamA, teamB int) ([]models.Match, error) {
	var result []models.Match
	for _, m := range s.matches {
		if (m.HomeTeamID == teamA && m.AwayTeamID == teamB) ||
			(m.HomeTeamID == teamB && m.AwayTeamID == teamA) {
			result = append(result, m)
		}
	}
	return result, nil
}

// GetMatchGoals returns all goals for a match.
func (s *FootballStore) GetMatchGoals(matchID int) ([]models.Goal, error) {
	var result []models.Goal
	for _, g := range s.goals {
		if g.MatchID == matchID {
			result = append(result, g)
		}
	}
	return result, nil
}

// GetMatchShootout returns the shootout for a match.
// Returns ErrNotFound when no shootout is recorded.
func (s *FootballStore) GetMatchShootout(matchID int) (models.Shootout, error) {
	for _, so := range s.shootouts {
		if so.MatchID == matchID {
			return so, nil
		}
	}
	return models.Shootout{}, models.ErrNotFound
}

// GetPlayerGoals returns all goals scored by the named player (case-insensitive).
func (s *FootballStore) GetPlayerGoals(scorer string) ([]models.Goal, error) {
	lower := strings.ToLower(scorer)
	var result []models.Goal
	for _, g := range s.goals {
		if strings.ToLower(g.Scorer) == lower {
			result = append(result, g)
		}
	}
	return result, nil
}

// --- seeding helpers (used only in tests) ------------------------------------

// AddTeam inserts a team into the store and returns it with an assigned ID.
func (s *FootballStore) AddTeam(name string) models.Team {
	t := models.Team{ID: len(s.teams) + 1, Name: name}
	s.teams = append(s.teams, t)
	return t
}

// AddMatch inserts a match into the store.
func (s *FootballStore) AddMatch(m models.Match) models.Match {
	m.ID = len(s.matches) + 1
	s.matches = append(s.matches, m)
	return m
}

// AddGoal inserts a goal into the store.
func (s *FootballStore) AddGoal(g models.Goal) models.Goal {
	g.ID = len(s.goals) + 1
	s.goals = append(s.goals, g)
	return g
}

// AddShootout inserts a shootout into the store.
func (s *FootballStore) AddShootout(so models.Shootout) models.Shootout {
	so.ID = len(s.shootouts) + 1
	s.shootouts = append(s.shootouts, so)
	return so
}

// AddFormerName inserts a former-name record into the store.
func (s *FootballStore) AddFormerName(fn models.FormerName) models.FormerName {
	fn.ID = len(s.formerNames) + 1
	s.formerNames = append(s.formerNames, fn)
	return fn
}
