// Package router wires together all middleware and route handlers to produce
// a ready-to-serve *gin.Engine.
//
// REST principles addressed here:
//
//   - Uniform Interface: all resources are identified by versioned URI paths
//     (/api/v1/…) and accessed through standard HTTP methods.
//   - Layered System: middleware (RequestID, Logger, CacheControl,
//     NoSessionState, JWTAuth) runs transparently between the client and handler,
//     just as a proxy or gateway would.
//   - Stateless: JWT authentication carries all user identity in the token itself,
//     eliminating server-side session state.
package router

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/auth"
	dbpkg "github.com/sc23bd/COMP3011_Coursework1/internal/db"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db/memory"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db/postgres"
	"github.com/sc23bd/COMP3011_Coursework1/internal/handlers"
	"github.com/sc23bd/COMP3011_Coursework1/internal/middleware"
)

// New returns a configured *gin.Engine.
//
// When db is non-nil the router uses the PostgreSQL-backed repositories;
// otherwise it falls back to the in-memory Store.  Pass a nil *sql.DB for
// local development without a running database (e.g. in tests).
//
// jwtSecret is used to sign and verify JWT tokens.
func New(jwtSecret string, db *sql.DB) *gin.Engine {
	var items dbpkg.ItemRepository
	var users dbpkg.UserRepository
	var football dbpkg.FootballRepository

	if db != nil {
		items = postgres.NewItemRepo(db)
		users = postgres.NewUserRepo(db)
		football = postgres.NewFootballRepo(db)
	} else {
		store := memory.NewStore()
		items = store
		users = store
		football = memory.NewFootballStore()
	}

	h := handlers.NewHandler(items)
	fh := handlers.NewFootballHandler(football)

	// Initialize JWT service
	jwtService := auth.NewJWTService(jwtSecret, "COMP3011_API")
	authHandler := handlers.NewAuthHandler(users, jwtService)

	r := gin.New()

	// Global middleware — applied to every route (Layered System principle).
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CacheControl())
	r.Use(gin.Recovery())

	// API v1 route group — versioned URI prefix (Uniform Interface principle).
	v1 := r.Group("/api/v1")
	{
		// Public authentication routes (no JWT required)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
		}

		// Items routes - read operations are public, mutations require JWT
		items := v1.Group("/items")
		{
			// Public read endpoints (no auth required)
			items.GET("", h.ListItems)
			items.HEAD("", h.ListItems)
			items.GET("/:id", h.GetItem)
			items.HEAD("/:id", h.GetItem)

			// Protected mutation endpoints (JWT required)
			items.POST("", middleware.JWTAuth(jwtService), h.CreateItem)
			items.PUT("/:id", middleware.JWTAuth(jwtService), h.UpdateItem)
			items.DELETE("/:id", middleware.JWTAuth(jwtService), h.DeleteItem)
		}

		// Football routes - all read-only and public
		football := v1.Group("/football")
		{
			football.GET("/teams", fh.ListTeams)
			football.GET("/teams/:id", fh.GetTeam)
			football.GET("/teams/:id/history", fh.GetTeamHistory)

			football.GET("/matches", fh.ListMatches)
			football.GET("/matches/:id", fh.GetMatch)
			football.GET("/matches/:id/goals", fh.GetMatchGoals)
			football.GET("/matches/:id/shootout", fh.GetMatchShootout)

			football.GET("/head-to-head", fh.GetHeadToHead)

			football.GET("/players/:name/goals", fh.GetPlayerGoals)
		}
	}

	return r
}
