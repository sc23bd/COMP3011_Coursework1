// Package db provides repository interfaces and implementations for data access.
// It supports both in-memory (memory package) and PostgreSQL (postgres package)
// backends through a common interface.
package db

import "github.com/sc23bd/COMP3011_Coursework1/internal/models"

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
