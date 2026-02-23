package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// UserRepo is a PostgreSQL-backed implementation of handlers.UserRepository.
// Passwords are stored exclusively as bcrypt hashes — plain-text passwords
// never touch the database layer.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo constructs a UserRepo backed by the provided *sql.DB.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// GetUser retrieves the user record for the given username.
// Returns models.ErrNotFound when the username does not exist.
func (r *UserRepo) GetUser(username string) (models.User, error) {
	const q = `
		SELECT username, password_hash, created_at
		FROM users
		WHERE username = $1`

	var (
		uname        string
		passwordHash string
		createdAt    time.Time
	)
	err := r.db.QueryRow(q, username).Scan(&uname, &passwordHash, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, models.ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("userRepo.GetUser: %w", err)
	}

	return models.User{
		Username:     uname,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
	}, nil
}

// CreateUser inserts a new user with the given bcrypt-hashed password.
// Returns models.ErrConflict when the username is already taken (PostgreSQL
// unique_violation error code 23505).
func (r *UserRepo) CreateUser(username, passwordHash string) (models.User, error) {
	const q = `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING created_at`

	var createdAt time.Time
	err := r.db.QueryRow(q, username, passwordHash).Scan(&createdAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return models.User{}, models.ErrConflict
		}
		return models.User{}, fmt.Errorf("userRepo.CreateUser: %w", err)
	}

	return models.User{
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
	}, nil
}
