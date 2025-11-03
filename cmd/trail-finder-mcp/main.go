package main

import (
	"log"
	"net/http"
	"os"

	"trail-finder-mcp/internal/httpserver"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := httpserver.New()

	log.Printf("[trail-finder-mcp] listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
