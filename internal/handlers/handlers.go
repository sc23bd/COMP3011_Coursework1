// Package handlers implements the HTTP handler functions for the Items
// resource.  Each handler is a thin adapter between the HTTP layer and the
// repository, deliberately keeping business logic separate from transport
// concerns (Client–Server principle).
package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// Handler holds the dependencies required by the HTTP handlers.
type Handler struct {
	items db.ItemRepository
}

// NewHandler constructs a Handler backed by the provided ItemRepository.
func NewHandler(items db.ItemRepository) *Handler {
	return &Handler{items: items}
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
	items, err := h.items.ListItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	var mostRecent time.Time
	responses := make([]models.ItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toResponse(item))
		if item.UpdatedAt.After(mostRecent) {
			mostRecent = item.UpdatedAt
		}
	}

	if !mostRecent.IsZero() {
		c.Header("Last-Modified", mostRecent.UTC().Format(http.TimeFormat))
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

	item, err := h.items.GetItem(id)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Header("Last-Modified", item.UpdatedAt.UTC().Format(http.TimeFormat))
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

	item, err := h.items.CreateItem(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Header("Location", "/api/v1/items/"+item.ID)
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

	item, err := h.items.UpdateItem(id, req.Name, req.Description)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toResponse(item))
}

// DeleteItem handles DELETE /api/v1/items/:id
// Removes the item and returns 204 No Content.
func (h *Handler) DeleteItem(c *gin.Context) {
	id := c.Param("id")

	err := h.items.DeleteItem(id)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
