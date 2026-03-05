// Package models defines the data structures used throughout the API.
package models

import "time"

// Team represents a national football team.
type Team struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// TeamResponse wraps a Team with hypermedia links (HATEOAS).
type TeamResponse struct {
	Team
	Links []Link `json:"links"`
}

// TeamsResponse wraps a list of teams with a collection-level link.
type TeamsResponse struct {
	Data  []TeamResponse `json:"data"`
	Links []Link         `json:"links"`
}

// FormerName represents a historical name used by a team.
type FormerName struct {
	ID         int        `json:"id"`
	TeamID     int        `json:"teamId"`
	FormerName string     `json:"formerName"`
	StartDate  *time.Time `json:"startDate,omitempty"`
	EndDate    *time.Time `json:"endDate,omitempty"`
}

// FormerNamesResponse wraps a list of former names with collection-level links.
type FormerNamesResponse struct {
	Data  []FormerName `json:"data"`
	Links []Link       `json:"links"`
}
