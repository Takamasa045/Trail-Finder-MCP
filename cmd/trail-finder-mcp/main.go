package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"trail-finder-mcp/internal/httpserver"
	"trail-finder-mcp/internal/mcpserver"
)

func main() {
	defaultMode := os.Getenv("TRAILFINDER_MODE")
	if defaultMode == "" {
		defaultMode = "http"
	}

	mode := flag.String("mode", defaultMode, "server mode: http or mcp")
	flag.Parse()

	switch *mode {
	case "http":
		runHTTP()
	case "mcp":
		if err := mcpserver.Run(context.Background()); err != nil {
			log.Fatalf("mcp server error: %v", err)
		}
	default:
		log.Fatalf("unknown mode %q, use http or mcp", *mode)
	}
}

func runHTTP() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := httpserver.New()

	log.Printf("[trail-finder-mcp] listening on :%s (http)", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
