package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New creates an MCP server with basic configuration.
func New() *mcp.Server {
	return mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-1c",
			Version: "0.1.0",
		},
		nil,
	)
}
