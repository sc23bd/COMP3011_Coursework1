// Package db provides repository interfaces for data access.
// Implementations are provided by the postgres subpackage.
package db

import "github.com/sc23bd/COMP3011_Coursework1/internal/models"

// FootballRepository abstracts the data-access layer for the football feature.
// It is currently implemented by the PostgreSQL repository.
type FootballRepository interface {
	// Teams - read
	ListTeams() ([]models.Team, error)
	GetTeamByID(id int) (models.Team, error)
	GetTeamHistory(teamID int) ([]models.FormerName, error)

	// Tournaments - read
	GetTournamentByID(id int) (models.Tournament, error)

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
}

// UserRepository abstracts the data-access layer for users.
// The PostgreSQL UserRepo satisfies this interface.
type UserRepository interface {
	GetUser(username string) (models.User, error)
	CreateUser(username, passwordHash string) (models.User, error)
}
