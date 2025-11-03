package main

import (
	"context"
	"log"

	"trail-finder-mcp/internal/mcpserver"
)

func main() {
	if err := mcpserver.Run(context.Background()); err != nil {
		log.Fatalf("mcp server error: %v", err)
	}
}
