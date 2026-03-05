package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/feenlace/mcp-1c/internal/config"
	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/feenlace/mcp-1c/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	baseURL := flag.String("base", "", "Base URL of 1C HTTP service")
	user := flag.String("user", "", "1C HTTP service user")
	password := flag.String("password", "", "1C HTTP service password")
	flag.Parse()

	// Load defaults and env var overrides.
	cfg := config.Load()

	// CLI flags take highest priority (override env vars).
	if *baseURL != "" {
		cfg.BaseURL = *baseURL
	}
	if *user != "" {
		cfg.User = *user
	}
	if *password != "" {
		cfg.Password = *password
	}

	client := onec.NewClient(cfg.BaseURL, cfg.User, cfg.Password)
	s := server.New(client)

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "mcp-1c error: %v\n", err)
		os.Exit(1)
	}
}
