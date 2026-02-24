package memory

import (
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

func (s *Store) GetUser(username string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[username]
	if !ok {
		return models.User{}, models.ErrNotFound
	}
	return user, nil
}

func (s *Store) CreateUser(username, passwordHash string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[username]; exists {
		return models.User{}, models.ErrConflict
	}
	user := models.User{
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	s.users[username] = user
	return user, nil
}
