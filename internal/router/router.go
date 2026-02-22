// Package router wires together all middleware and route handlers to produce
// a ready-to-serve *gin.Engine.
//
// REST principles addressed here:
//
//   - Uniform Interface: all resources are identified by versioned URI paths
//     (/api/v1/…) and accessed through standard HTTP methods.
//   - Layered System: middleware (RequestID, Logger, CacheControl,
//     NoSessionState) runs transparently between the client and handler,
//     just as a proxy or gateway would.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/handlers"
	"github.com/sc23bd/COMP3011_Coursework1/internal/middleware"
)

// New returns a configured *gin.Engine.
// A fresh in-memory store is created; swap it for a real DB adapter as needed.
func New() *gin.Engine {
	store := handlers.NewStore()
	h := handlers.NewHandler(store)

	r := gin.New()

	// Global middleware — applied to every route (Layered System principle).
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.NoSessionState())
	r.Use(middleware.CacheControl())
	r.Use(gin.Recovery())

	// API v1 route group — versioned URI prefix (Uniform Interface principle).
	v1 := r.Group("/api/v1")
	{
		items := v1.Group("/items")
		{
			items.GET("", h.ListItems)
			items.POST("", h.CreateItem)
			items.GET("/:id", h.GetItem)
			items.PUT("/:id", h.UpdateItem)
			items.DELETE("/:id", h.DeleteItem)
		}
	}

	return r
}
