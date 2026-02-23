// Package db provides the PostgreSQL data-access layer.
// It contains repository implementations that satisfy the ItemRepository and
// UserRepository interfaces defined in the handlers package, as well as a
// helper for opening a pooled *sql.DB connection.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	// Register the lib/pq PostgreSQL driver as a side-effect import.
	_ "github.com/lib/pq"
)

// Connect opens a *sql.DB connection to the PostgreSQL instance described by
// dsn (a libpq connection string or URL, e.g.
// "postgres://user:pass@localhost:5432/dbname?sslmode=disable").
//
// Connection-pool settings follow common production recommendations:
//   - MaxOpenConns: limits total concurrent connections to the database.
//   - MaxIdleConns: keeps a small pool of ready connections to reduce
//     connection-setup latency.
//   - ConnMaxLifetime: recycles connections periodically so that load-balancer
//     or firewall idle-connection limits are not hit.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return db, nil
}

// ConnectFromEnv is a convenience wrapper that reads the DATABASE_URL
// environment variable and calls Connect.  Returns (nil, nil) when the
// variable is not set so callers can fall back to an in-memory store.
func ConnectFromEnv() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, nil
	}
	return Connect(dsn)
}
