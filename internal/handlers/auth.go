// Package handlers implements authentication endpoints for user registration
// and login. Authentication is stateless — all user identity is carried in
// the JWT token returned at login (Stateless principle).
package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/auth"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler holds dependencies for authentication endpoints.
type AuthHandler struct {
	users      db.UserRepository
	jwtService *auth.JWTService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(users db.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		users:      users,
		jwtService: jwtService,
	}
}

// Register handles POST /api/v1/auth/register
// Creates a new user account with hashed password.
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Hash password before calling the repository so the slow bcrypt
	// operation does not block any shared resource (lock, connection, etc.).
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to hash password"})
		return
	}

	user, err := h.users.CreateUser(req.Username, string(hashedPassword))
	if errors.Is(err, models.ErrConflict) {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "username already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "user created successfully",
		"username": user.Username,
		"links": []models.Link{
			{Rel: "login", Href: "/api/v1/auth/login", Method: http.MethodPost},
		},
	})
}

// Login handles POST /api/v1/auth/login
// Validates credentials and returns a JWT token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.users.GetUser(req.Username)
	if errors.Is(err, models.ErrNotFound) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	// Verify password against the stored bcrypt hash.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		Links: []models.Link{
			{Rel: "items", Href: "/api/v1/items", Method: http.MethodGet},
		},
	})
}
