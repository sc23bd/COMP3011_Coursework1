// Package handlers implements the HTTP handler functions for the Items
// resource.  Each handler is a thin adapter between the HTTP layer and the
// in-memory store, deliberately keeping business logic separate from
// transport concerns (Client–Server principle).
package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// Store is the in-memory data store shared by all handlers.
// Replace with a real database adapter in production.
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

// nextID generates a simple sequential string ID.
func (s *Store) nextID() string {
	s.counter++
	return fmt.Sprintf("%d", s.counter)
}

// Handler holds the dependencies required by the HTTP handlers.
type Handler struct {
	store *Store
}

// NewHandler constructs a Handler backed by the provided store.
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// itemLinks builds the HATEOAS link set for a single item (Uniform Interface
// principle — self-descriptive messages and hypermedia as the engine of
// application state).
func itemLinks(id string) []models.Link {
	base := "/api/v1/items/" + id
	return []models.Link{
		{Rel: "self", Href: base, Method: http.MethodGet},
		{Rel: "update", Href: base, Method: http.MethodPut},
		{Rel: "delete", Href: base, Method: http.MethodDelete},
	}
}

// toResponse wraps an Item in an ItemResponse with HATEOAS links.
func toResponse(item models.Item) models.ItemResponse {
	return models.ItemResponse{Item: item, Links: itemLinks(item.ID)}
}

// ListItems handles GET /api/v1/items
// Returns all items together with a collection-level hypermedia link.
func (h *Handler) ListItems(c *gin.Context) {
	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	responses := make([]models.ItemResponse, 0, len(h.store.items))
	for _, item := range h.store.items {
		responses = append(responses, toResponse(item))
	}

	c.JSON(http.StatusOK, models.ItemsResponse{
		Data: responses,
		Links: []models.Link{
			{Rel: "self", Href: "/api/v1/items", Method: http.MethodGet},
			{Rel: "create", Href: "/api/v1/items", Method: http.MethodPost},
		},
	})
}

// GetItem handles GET /api/v1/items/:id
// Returns the requested item or 404 if it does not exist.
func (h *Handler) GetItem(c *gin.Context) {
	id := c.Param("id")

	h.store.mu.RLock()
	item, ok := h.store.items[id]
	h.store.mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "item not found"})
		return
	}

	c.JSON(http.StatusOK, toResponse(item))
}

// CreateItem handles POST /api/v1/items
// Validates the request body, creates a new item, and returns 201 Created.
// The handler is stateless — all information required to fulfil the request
// is present in the request itself.
func (h *Handler) CreateItem(c *gin.Context) {
	var req models.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.store.mu.Lock()
	id := h.store.nextID()
	now := time.Now()
	item := models.Item{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	h.store.items[id] = item
	h.store.mu.Unlock()

	c.Header("Location", "/api/v1/items/"+id)
	c.JSON(http.StatusCreated, toResponse(item))
}

// UpdateItem handles PUT /api/v1/items/:id
// Replaces an existing item and returns the updated representation.
func (h *Handler) UpdateItem(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.store.mu.Lock()
	existing, ok := h.store.items[id]
	if !ok {
		h.store.mu.Unlock()
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "item not found"})
		return
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.UpdatedAt = time.Now()
	h.store.items[id] = existing
	h.store.mu.Unlock()

	c.JSON(http.StatusOK, toResponse(existing))
}

// DeleteItem handles DELETE /api/v1/items/:id
// Removes the item and returns 204 No Content.
func (h *Handler) DeleteItem(c *gin.Context) {
	id := c.Param("id")

	h.store.mu.Lock()
	_, ok := h.store.items[id]
	if !ok {
		h.store.mu.Unlock()
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "item not found"})
		return
	}
	delete(h.store.items, id)
	h.store.mu.Unlock()

	c.Status(http.StatusNoContent)
}
