package dump

import (
	"strings"
	"testing"
)

func TestCachePath_NonEmpty(t *testing.T) {
	p, err := cachePath("/some/dump/dir")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	if p == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestCachePath_Deterministic(t *testing.T) {
	p1, err := cachePath("/some/dump/dir")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	p2, err := cachePath("/some/dump/dir")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	if p1 != p2 {
		t.Errorf("expected deterministic result, got %q and %q", p1, p2)
	}
}

func TestCachePath_DifferentDirs(t *testing.T) {
	p1, err := cachePath("/dir/one")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	p2, err := cachePath("/dir/two")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	if p1 == p2 {
		t.Errorf("expected different paths for different dirs, got same: %q", p1)
	}
}

func TestCachePath_ContainsMcp1c(t *testing.T) {
	p, err := cachePath("/any/dir")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	if !strings.Contains(p, "mcp-1c") {
		t.Errorf("expected path to contain 'mcp-1c', got %q", p)
	}
}
