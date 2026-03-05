// Package models defines the data structures used throughout the API.
package models

// Link represents a hypermedia link used to satisfy HATEOAS (Uniform Interface).
type Link struct {
	Rel    string `json:"rel"`
	Href   string `json:"href"`
	Method string `json:"method"`
}

// ErrorResponse is the standard error envelope returned by all handlers.
type ErrorResponse struct {
	Error string `json:"error"`
}
