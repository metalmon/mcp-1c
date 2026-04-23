package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeTransport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default empty is stdio", input: "", want: transportStdio},
		{name: "stdio explicit", input: "stdio", want: transportStdio},
		{name: "http explicit", input: "http", want: transportHTTP},
		{name: "trim and case normalize", input: "  HTTP  ", want: transportHTTP},
		{name: "invalid transport", input: "sse", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeTransport(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("transport mismatch: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeListenAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default listen", input: "", want: defaultHTTPListenAddr},
		{name: "explicit listen", input: "127.0.0.1:8765", want: "127.0.0.1:8765"},
		{name: "trim listen", input: "  :9999  ", want: ":9999"},
		{name: "invalid listen", input: "bad addr", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeListenAddr(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("listen addr mismatch: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestWithCORS_Preflight(t *testing.T) {
	t.Parallel()

	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow-origin mismatch: got %q want %q", got, "*")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatalf("allow-methods header is empty")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatalf("allow-headers header is empty")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "" && !containsIgnoreCase(got, "Mcp-Protocol-Version") {
		t.Fatalf("allow-headers missing mcp protocol version: %q", got)
	}
}

func TestWithCORS_GetCompatibility(t *testing.T) {
	t.Parallel()

	nextCalled := false
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusNoContent)
	}
	if nextCalled {
		t.Fatalf("next handler must not be called for GET compatibility path")
	}
}

func containsIgnoreCase(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (stringContainsFold(haystack, needle))
}

func stringContainsFold(s, substr string) bool {
	n := len(substr)
	for i := 0; i+n <= len(s); i++ {
		if equalFoldASCII(s[i:i+n], substr) {
			return true
		}
	}
	return false
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		aa := a[i]
		bb := b[i]
		if 'A' <= aa && aa <= 'Z' {
			aa += 'a' - 'A'
		}
		if 'A' <= bb && bb <= 'Z' {
			bb += 'a' - 'A'
		}
		if aa != bb {
			return false
		}
	}
	return true
}
