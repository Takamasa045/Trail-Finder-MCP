package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestRunRejectsUnknownFlag(t *testing.T) {
	err := run([]string{"-nope"})
	if err == nil {
		t.Fatal("expected flag error")
	}
	if !strings.Contains(err.Error(), "flag provided but not defined") && !strings.Contains(err.Error(), "nope") {
		t.Fatalf("err=%v", err)
	}
}

// Use the test binary as a subprocess so the actual stdio entry point is exercised.
func TestStdioHelper(t *testing.T) {
	if os.Getenv("TRAILFINDER_TEST_STDIO") != "1" {
		return
	}
	if err := run(nil); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestStdioMCPStartup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStdioHelper$")
	cmd.Env = append(os.Environ(), "TRAILFINDER_TEST_STDIO=1")
	client := mcp.NewClient(&mcp.Implementation{Name: "stdio-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	result, err := session.ListTools(ctx, nil)
	if err != nil || len(result.Tools) != 6 {
		t.Fatalf("tools=%+v err=%v", result, err)
	}
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "geocode", Arguments: map[string]any{"query": " "}})
	if err != nil || !res.IsError {
		t.Fatalf("validation=%+v err=%v", res, err)
	}
}
