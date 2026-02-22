package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/auth"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

// JWTAuth validates JWT tokens from the Authorization header.
// This middleware enforces the Stateless principle — all authentication state
// is contained in the self-describing JWT token, not in server-side sessions.
func JWTAuth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "authorization header required",
			})
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "authorization header format must be 'Bearer {token}'",
			})
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid or expired token",
			})
			return
		}

		// Attach username to context for handlers to use
		c.Set("username", claims.Username)
		c.Next()
	}
}
