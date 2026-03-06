// Package db provides repository interfaces for data access.
// Implementations are provided by the postgres subpackage.
package db

import (
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/elo"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// FootballRepository abstracts the data-access layer for the football feature.
// It is currently implemented by the PostgreSQL repository.
type FootballRepository interface {
	// Teams - read
	ListTeams() ([]models.Team, error)
	GetTeamByID(id int) (models.Team, error)
	GetTeamHistory(teamID int) ([]models.FormerName, error)

	// Tournaments - read
	GetTournamentByID(id int) (models.Tournament, error)
	ListTournaments() ([]models.Tournament, error)

	// Teams - write
	CreateTeam(name string) (models.Team, error)
	UpdateTeam(id int, name string) (models.Team, error)
	DeleteTeam(id int) error

	// Matches - read
	ListMatches(limit, offset int) ([]models.Match, error)
	GetMatchByID(id int) (models.Match, error)
	GetHeadToHead(teamA, teamB int) ([]models.Match, error)

	// Matches - write
	CreateMatch(m models.Match) (models.Match, error)
	UpdateMatch(id int, m models.Match) (models.Match, error)
	DeleteMatch(id int) error

	// Goals & Shootouts - read
	GetMatchGoals(matchID int) ([]models.Goal, error)
	GetMatchShootout(matchID int) (models.Shootout, error)

	// Goals - write
	CreateGoal(g models.Goal) (models.Goal, error)
	DeleteGoal(id int) error

	// Shootouts - write
	CreateShootout(s models.Shootout) (models.Shootout, error)
	DeleteShootout(matchID int) error

	// Players
	GetPlayerGoals(scorer string) ([]models.Goal, error)

	// Elo – read
	// GetMatchesChronological returns all matches involving teamID up to and
	// including endDate, ordered oldest-first.  Pass teamID = 0 to fetch all matches.
	GetMatchesChronological(teamID int, endDate time.Time) ([]elo.MatchResult, error)
	// GetEloRankings returns a paginated global Elo ranking snapshot.
	// region is an optional filter (empty = all regions); limit/offset control pagination.
	GetEloRankings(asOf time.Time, region string, limit, offset int) ([]elo.RankingEntry, error)
	// GetTeamCachedRank returns the most-recently cached global rank for a team
	// on or before asOf. Returns 0 if no cached rank exists.
	GetTeamCachedRank(teamID int, asOf time.Time) (int, error)

	// Elo – write
	// SaveEloSnapshot upserts a cached Elo rating for one team on one date.
	SaveEloSnapshot(teamID int, asOf time.Time, rating float64, rank int, matchesPlayed int) error
}

// UserRepository abstracts the data-access layer for users.
// The PostgreSQL UserRepo satisfies this interface.
type UserRepository interface {
	GetUser(username string) (models.User, error)
	CreateUser(username, passwordHash string) (models.User, error)
}
