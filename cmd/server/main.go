// main is the entry-point for the REST API server.
// Configuration is read from environment variables so the binary has no
// hard-coded operational parameters (supports the Layered System principle —
// the same binary can run behind different proxy/load-balancer configurations).
package main

import (
	"log"
	"os"

	"github.com/sc23bd/COMP3011_Coursework1/internal/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := router.New()

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
