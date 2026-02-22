// Package middleware provides reusable Gin middleware that enforces REST
// architectural constraints across all routes.
package middleware

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestID attaches a unique identifier to every incoming request and echoes
// it in the response via the X-Request-ID header.  This supports the
// Layered System and Uniform Interface principles by making requests
// traceable through any intermediary proxy or load-balancer.
func RequestID() gin.HandlerFunc {
	var counter int64
	return func(c *gin.Context) {
		n := atomic.AddInt64(&counter, 1)
		id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), n)
		c.Set("requestID", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// CacheControl sets appropriate Cache-Control headers so that clients and
// intermediate caches know whether a response may be stored (Cacheable
// principle).
//
//   - Safe, idempotent GET/HEAD responses are marked as cacheable for 60 s.
//   - All other methods are marked no-store to prevent stale mutations.
func CacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			c.Header("Cache-Control", "public, max-age=60")
		} else {
			c.Header("Cache-Control", "no-store")
		}
	}
}

// NoSessionState validates that no session cookie is present in the request.
// REST requires each request to be self-contained (Stateless principle) —
// server-side session state is therefore not allowed.
func NoSessionState() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Cookie") != "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "session cookies are not supported; the API is stateless",
			})
			return
		}
		c.Next()
	}
}

// Logger prints a structured log line for every request, including the
// request-ID injected by RequestID().  Logging middleware is a classic
// example of the Layered System principle — the handler never knows whether
// an additional layer is observing its traffic.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		id, _ := c.Get("requestID")
		fmt.Printf("[GIN] %s | %3d | %12v | %-7s %s | req-id=%v\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			c.Writer.Status(),
			time.Since(start),
			c.Request.Method,
			c.Request.URL.Path,
			id,
		)
	}
}
