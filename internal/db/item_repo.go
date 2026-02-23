package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// ItemRepo is a PostgreSQL-backed implementation of handlers.ItemRepository.
// All queries use parameterized placeholders ($1, $2, …) to prevent SQL
// injection.
type ItemRepo struct {
	db *sql.DB
}

// NewItemRepo constructs an ItemRepo backed by the provided *sql.DB.
func NewItemRepo(db *sql.DB) *ItemRepo {
	return &ItemRepo{db: db}
}

// ListItems returns all items ordered by most-recently-updated descending.
// Rows are iterated with proper resource cleanup via defer rows.Close().
func (r *ItemRepo) ListItems() ([]models.Item, error) {
	const q = `
		SELECT id, name, description, created_at, updated_at
		FROM items
		ORDER BY updated_at DESC`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("itemRepo.ListItems: %w", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var (
			id          int
			name        string
			description string
			createdAt   time.Time
			updatedAt   time.Time
		)
		if err := rows.Scan(&id, &name, &description, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("itemRepo.ListItems scan: %w", err)
		}
		items = append(items, models.Item{
			ID:          strconv.Itoa(id),
			Name:        name,
			Description: description,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("itemRepo.ListItems rows: %w", err)
	}
	return items, nil
}

// GetItem retrieves the item with the given ID.
// Returns ErrNotFound when no matching row exists.
func (r *ItemRepo) GetItem(id string) (models.Item, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return models.Item{}, models.ErrNotFound
	}

	const q = `
		SELECT id, name, description, created_at, updated_at
		FROM items
		WHERE id = $1`

	var (
		name        string
		description string
		createdAt   time.Time
		updatedAt   time.Time
	)
	err = r.db.QueryRow(q, intID).Scan(&intID, &name, &description, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Item{}, models.ErrNotFound
	}
	if err != nil {
		return models.Item{}, fmt.Errorf("itemRepo.GetItem: %w", err)
	}

	return models.Item{
		ID:          strconv.Itoa(intID),
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// CreateItem inserts a new item and returns it with the database-assigned ID
// and timestamps.  The INSERT … RETURNING pattern retrieves the generated
// values in a single round-trip.
func (r *ItemRepo) CreateItem(name, description string) (models.Item, error) {
	const q = `
		INSERT INTO items (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	var (
		id        int
		createdAt time.Time
		updatedAt time.Time
	)
	err := r.db.QueryRow(q, name, description).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return models.Item{}, fmt.Errorf("itemRepo.CreateItem: %w", err)
	}

	return models.Item{
		ID:          strconv.Itoa(id),
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// UpdateItem replaces the name and description of an existing item and
// refreshes its updated_at timestamp.  Returns ErrNotFound when no row with
// the given ID exists.
func (r *ItemRepo) UpdateItem(id, name, description string) (models.Item, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return models.Item{}, models.ErrNotFound
	}

	const q = `
		UPDATE items
		SET name = $2, description = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at`

	var (
		outID     int
		outName   string
		outDesc   string
		createdAt time.Time
		updatedAt time.Time
	)
	err = r.db.QueryRow(q, intID, name, description).
		Scan(&outID, &outName, &outDesc, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Item{}, models.ErrNotFound
	}
	if err != nil {
		return models.Item{}, fmt.Errorf("itemRepo.UpdateItem: %w", err)
	}

	return models.Item{
		ID:          strconv.Itoa(outID),
		Name:        outName,
		Description: outDesc,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// DeleteItem removes the item with the given ID.
// Returns ErrNotFound when no matching row exists.
func (r *ItemRepo) DeleteItem(id string) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return models.ErrNotFound
	}

	const q = `DELETE FROM items WHERE id = $1`

	result, err := r.db.Exec(q, intID)
	if err != nil {
		return fmt.Errorf("itemRepo.DeleteItem: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("itemRepo.DeleteItem rowsAffected: %w", err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}
