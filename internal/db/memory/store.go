package memory

import (
	"fmt"
	"sync"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// Store is the in-memory data store that implements both ItemRepository and
// UserRepository.  It is used when no DATABASE_URL is configured (e.g. tests,
// local development without PostgreSQL).
type Store struct {
	mu      sync.RWMutex
	items   map[string]models.Item
	users   map[string]models.User
	counter int
}

// NewStore returns an initialised, empty store.
func NewStore() *Store {
	return &Store{
		items: make(map[string]models.Item),
		users: make(map[string]models.User),
	}
}

// nextID generates a simple sequential string ID (must be called under lock).
func (s *Store) nextID() string {
	s.counter++
	return fmt.Sprintf("%d", s.counter)
}
