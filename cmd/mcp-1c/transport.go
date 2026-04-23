package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	transportStdio        = "stdio"
	transportHTTP         = "http"
	defaultHTTPListenAddr = "127.0.0.1:8765"
)

func normalizeTransport(value string) (string, error) {
	transport := strings.ToLower(strings.TrimSpace(value))
	if transport == "" {
		return transportStdio, nil
	}
	switch transport {
	case transportStdio, transportHTTP:
		return transport, nil
	default:
		return "", fmt.Errorf("unsupported transport %q (allowed: stdio|http)", value)
	}
}

func normalizeListenAddr(value string) (string, error) {
	addr := strings.TrimSpace(value)
	if addr == "" {
		addr = defaultHTTPListenAddr
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return "", fmt.Errorf("invalid listen address %q: %w", value, err)
	}
	return addr, nil
}

func runServer(ctx context.Context, s *mcp.Server, transport, listenAddr string) error {
	switch transport {
	case transportStdio:
		return s.Run(ctx, &mcp.StdioTransport{})
	case transportHTTP:
		return runHTTPServer(ctx, s, listenAddr)
	default:
		return fmt.Errorf("unsupported transport %q", transport)
	}
}

func runHTTPServer(ctx context.Context, s *mcp.Server, listenAddr string) error {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s
	}, &mcp.StreamableHTTPOptions{
		Stateless: true,
	})
	server := &http.Server{
		Addr:    listenAddr,
		Handler: withCORS(handler),
	}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Last-Event-ID, Mcp-Session-Id, Mcp-Protocol-Version")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodGet {
			// Compatibility path for clients that probe MCP endpoint with GET/SSE first.
			// Streamable HTTP flow works via POST, so return 204 instead of noisy 405.
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
