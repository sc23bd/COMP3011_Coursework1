// Package db provides repository interfaces for data access.
// Implementations are provided by the memory and postgres subpackages.
package db

import "github.com/sc23bd/COMP3011_Coursework1/internal/models"

// FootballRepository abstracts the data-access layer for the football feature.
// Both the in-memory and PostgreSQL implementations satisfy this interface.
type FootballRepository interface {
	// Teams
	ListTeams() ([]models.Team, error)
	GetTeamByID(id int) (models.Team, error)
	GetTeamHistory(teamID int) ([]models.FormerName, error)

	// Matches
	ListMatches(limit, offset int) ([]models.Match, error)
	GetMatchByID(id int) (models.Match, error)
	GetHeadToHead(teamA, teamB int) ([]models.Match, error)

	// Goals & Shootouts
	GetMatchGoals(matchID int) ([]models.Goal, error)
	GetMatchShootout(matchID int) (models.Shootout, error)

	// Players
	GetPlayerGoals(scorer string) ([]models.Goal, error)
}

// ItemRepository abstracts the data-access layer for items.
// Both the in-memory Store and the PostgreSQL ItemRepo satisfy this interface.
type ItemRepository interface {
	ListItems() ([]models.Item, error)
	GetItem(id string) (models.Item, error)
	CreateItem(name, description string) (models.Item, error)
	UpdateItem(id, name, description string) (models.Item, error)
	DeleteItem(id string) error
}

// UserRepository abstracts the data-access layer for users.
// Both the in-memory Store and the PostgreSQL UserRepo satisfy this interface.
type UserRepository interface {
	GetUser(username string) (models.User, error)
	CreateUser(username, passwordHash string) (models.User, error)
}
