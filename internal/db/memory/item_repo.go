package memory

import (
	"sort"
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

func (s *Store) ListItems() ([]models.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Item, 0, len(s.items))
	for _, item := range s.items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func (s *Store) GetItem(id string) (models.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return models.Item{}, models.ErrNotFound
	}
	return item, nil
}

func (s *Store) CreateItem(name, description string) (models.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID()
	now := time.Now()
	item := models.Item{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.items[id] = item
	return item, nil
}

func (s *Store) UpdateItem(id, name, description string) (models.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.items[id]
	if !ok {
		return models.Item{}, models.ErrNotFound
	}
	existing.Name = name
	existing.Description = description
	existing.UpdatedAt = time.Now()
	s.items[id] = existing
	return existing, nil
}

func (s *Store) DeleteItem(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return models.ErrNotFound
	}
	delete(s.items, id)
	return nil
}
