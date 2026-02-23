// Package models defines the data structures used throughout the API.
package models

import "time"

// Item represents a resource in the system.
// Each field is annotated for JSON serialisation and input validation.
type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"        binding:"required,min=1,max=100"`
	Description string    `json:"description" binding:"max=500"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateItemRequest is the payload accepted when creating a new Item.
type CreateItemRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateItemRequest is the payload accepted when replacing an existing Item.
type UpdateItemRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// Link represents a hypermedia link used to satisfy HATEOAS (Uniform Interface).
type Link struct {
	Rel    string `json:"rel"`
	Href   string `json:"href"`
	Method string `json:"method"`
}

// ItemResponse wraps an Item with hypermedia links (HATEOAS).
type ItemResponse struct {
	Item
	Links []Link `json:"links"`
}

// ItemsResponse wraps a list of items with a collection-level hypermedia link.
type ItemsResponse struct {
	Data  []ItemResponse `json:"data"`
	Links []Link         `json:"links"`
}

// ErrorResponse is the standard error envelope returned by all handlers.
type ErrorResponse struct {
	Error string `json:"error"`
}
