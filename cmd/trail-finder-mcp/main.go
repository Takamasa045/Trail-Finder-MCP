package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/httpserver"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/mcpserver"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("trail-finder-mcp: %v", err)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("trail-finder-mcp", flag.ContinueOnError)
	httpAddr := fs.String("http", "", "optional JSON HTTP listen address (e.g. :8080). Default is MCP stdio.")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *httpAddr != "" {
		addr := *httpAddr
		if strings.HasPrefix(addr, ":") {
			addr = "127.0.0.1" + addr
		}
		log.Printf("http tools listening on %s", addr)
		srv := &http.Server{
			Addr:              addr,
			Handler:           httpserver.New(),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
		}
		return srv.ListenAndServe()
	}
	if err := mcpserver.Run(context.Background()); err != nil {
		return fmt.Errorf("mcp server: %w", err)
	}
	return nil
}
