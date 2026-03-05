package server

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}
