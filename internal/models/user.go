package models

import "time"

// User represents a user account in the system.
type User struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time `json:"createdAt"`
}

// RegisterRequest is the payload for creating a new user account.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// LoginRequest is the payload for authenticating a user.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse contains the JWT token returned after successful authentication.
type LoginResponse struct {
	Token string `json:"token"`
	Links []Link `json:"links"`
}
