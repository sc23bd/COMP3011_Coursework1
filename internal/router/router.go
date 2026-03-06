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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/sc23bd/COMP3011_Coursework1/docs"
	"github.com/sc23bd/COMP3011_Coursework1/internal/auth"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db/postgres"
	"github.com/sc23bd/COMP3011_Coursework1/internal/handlers"
	"github.com/sc23bd/COMP3011_Coursework1/internal/middleware"
	docs "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// New returns a configured *gin.Engine.
//
// When db is non-nil the router registers authentication and football routes
// backed by PostgreSQL.  Pass a nil *sql.DB only when running without a
// database (no routes requiring persistence will be registered).
//
// jwtSecret is used to sign and verify JWT tokens.
func New(jwtSecret string, db *sql.DB) *gin.Engine {
	// Initialize JWT service
	jwtService := auth.NewJWTService(jwtSecret, "COMP3011_API")

	r := gin.New()

	// Global middleware — applied to every route (Layered System principle).
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CacheControl())
	r.Use(gin.Recovery())

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(docs.Handler))

	// API v1 route group — versioned URI prefix (Uniform Interface principle).
	v1 := r.Group("/api/v1")

	// All routes require a database connection.
	if db != nil {
		users := postgres.NewUserRepo(db)
		authHandler := handlers.NewAuthHandler(users, jwtService)

		// Public authentication routes (no JWT required)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
		}

		// Football routes - read operations are public, mutations require JWT.
		fh := handlers.NewFootballHandler(postgres.NewFootballRepo(db))
		football := v1.Group("/football")
		{
			// Public read endpoints
			football.GET("/teams", fh.ListTeams)
			football.GET("/teams/:id", fh.GetTeam)
			football.GET("/teams/:id/history", fh.GetTeamHistory)
			football.GET("/teams/:id/elo", fh.GetTeamElo)
			football.GET("/teams/:id/elo/timeline", fh.GetTeamEloTimeline)

			football.GET("/tournaments", fh.ListTournaments)

			football.GET("/matches", fh.ListMatches)
			football.GET("/matches/:id", fh.GetMatch)
			football.GET("/matches/:id/goals", fh.GetMatchGoals)
			football.GET("/matches/:id/shootout", fh.GetMatchShootout)

			football.GET("/head-to-head", fh.GetHeadToHead)

			football.GET("/players/:name/goals", fh.GetPlayerGoals)

			football.GET("/rankings/elo", fh.GetEloRankings)

			// Protected mutation endpoints (JWT required)
			football.POST("/teams", middleware.JWTAuth(jwtService), fh.CreateTeam)
			football.PUT("/teams/:id", middleware.JWTAuth(jwtService), fh.UpdateTeam)
			football.DELETE("/teams/:id", middleware.JWTAuth(jwtService), fh.DeleteTeam)

			football.POST("/matches", middleware.JWTAuth(jwtService), fh.CreateMatch)
			football.PUT("/matches/:id", middleware.JWTAuth(jwtService), fh.UpdateMatch)
			football.DELETE("/matches/:id", middleware.JWTAuth(jwtService), fh.DeleteMatch)

			football.POST("/matches/:id/goals", middleware.JWTAuth(jwtService), fh.CreateGoal)
			football.DELETE("/matches/:id/goals/:goalId", middleware.JWTAuth(jwtService), fh.DeleteGoal)

			football.POST("/matches/:id/shootout", middleware.JWTAuth(jwtService), fh.CreateShootout)
			football.DELETE("/matches/:id/shootout", middleware.JWTAuth(jwtService), fh.DeleteShootout)

			football.POST("/rankings/elo/recalculate", middleware.JWTAuth(jwtService), fh.RecalculateEloRankings)
		}
	}

	// Serve the built frontend static files if the dist directory exists.
	// In production (Docker), the frontend is built via the node:alpine stage
	// and copied to ./frontend/dist alongside the server binary.
	const frontendDist = "./frontend/dist"
	if _, err := os.Stat(frontendDist); err == nil {
		r.Static("/assets", filepath.Join(frontendDist, "assets"))
		r.StaticFile("/vite.svg", filepath.Join(frontendDist, "vite.svg"))
		// Catch-all: serve index.html for any non-API path to support
		// client-side (React Router) navigation.
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/swagger/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File(filepath.Join(frontendDist, "index.html"))
		})
	}

	return r
}
