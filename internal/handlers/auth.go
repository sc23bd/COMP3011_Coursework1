// Package handlers implements authentication endpoints for user registration
// and login. Authentication is stateless — all user identity is carried in
// the JWT token returned at login (Stateless principle).
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/auth"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler holds dependencies for authentication endpoints.
type AuthHandler struct {
	store      *Store
	jwtService *auth.JWTService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(store *Store, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		store:      store,
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

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	// Check if username already exists
	if _, exists := h.store.users[req.Username]; exists {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}
	h.store.users[req.Username] = user

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
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

	h.store.mu.RLock()
	user, exists := h.store.users[req.Username]
	h.store.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	// Verify password
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
