// Package models defines the data structures used throughout the API.
package models

import "time"

// Tournament represents a football competition.
type Tournament struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}
