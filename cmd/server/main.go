// main is the entry-point for the REST API server.
// Configuration is read from environment variables so the binary has no
// hard-coded operational parameters (supports the Layered System principle —
// the same binary can run behind different proxy/load-balancer configurations).
package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"

	"github.com/sc23bd/COMP3011_Coursework1/internal/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// In development, allow falling back to a random secret when DEV_MODE is explicitly enabled.
		if os.Getenv("DEV_MODE") == "true" {
			randomBytes := make([]byte, 32)
			if _, err := rand.Read(randomBytes); err != nil {
				log.Fatalf("failed to generate random JWT secret: %v", err)
			}
			jwtSecret = base64.StdEncoding.EncodeToString(randomBytes)
			log.Println("WARNING: Using randomly generated JWT_SECRET because DEV_MODE=true. Do not use this configuration in production; set the JWT_SECRET environment variable instead.")
		} else {
			log.Fatal("JWT_SECRET environment variable is required but not set. Refusing to start without a stable JWT secret.")
		}
	}

	r := router.New(jwtSecret)

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

