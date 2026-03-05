package main

import (
	"context"
	"fmt"
	"os"

	"github.com/feenlace/mcp-1c/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	s := server.New()

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "mcp-1c error: %v\n", err)
		os.Exit(1)
	}
}
