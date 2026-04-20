package profile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/feenlace/mcp-1c/onec"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default to auto", input: "", want: Auto},
		{name: "generic", input: "generic", want: Generic},
		{name: "buh", input: "buh_3_0", want: Buh30},
		{name: "unknown", input: "unknown", want: Unknown},
		{name: "invalid", input: "erp", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Normalize(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected validation error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveAutoDetectsBuh30(t *testing.T) {
	t.Parallel()

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/configuration" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":    "БухгалтерияПредприятия",
			"version": "3.0.150.1",
		})
	}))
	defer mock.Close()

	client := onec.NewClient(mock.URL, "", "")
	got, err := Resolve(context.Background(), client, Auto)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got != Buh30 {
		t.Fatalf("Resolve() = %q, want %q", got, Buh30)
	}
}

func TestResolveAutoFallsBackToGeneric(t *testing.T) {
	t.Parallel()

	client := onec.NewClient("http://127.0.0.1:1", "", "")
	got, err := Resolve(context.Background(), client, Auto)
	if got != Generic {
		t.Fatalf("Resolve() = %q, want %q", got, Generic)
	}
	if err == nil {
		t.Fatal("expected auto-resolve error for unavailable service")
	}
}

func TestResolveExplicitProfileSkipsHttpCall(t *testing.T) {
	t.Parallel()

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("HTTP call is not expected for explicit profile")
	}))
	defer mock.Close()

	client := onec.NewClient(mock.URL, "", "")
	got, err := Resolve(context.Background(), client, Generic)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got != Generic {
		t.Fatalf("Resolve() = %q, want %q", got, Generic)
	}
}

func TestDetect(t *testing.T) {
	t.Parallel()

	if got := Detect("БухгалтерияПредприятия", "3.0.1"); got != Buh30 {
		t.Fatalf("Detect() = %q, want %q", got, Buh30)
	}
	if got := Detect("УправлениеТорговлей", "11.5"); got != Generic {
		t.Fatalf("Detect() = %q, want %q", got, Generic)
	}
}

func TestResolveReturnsValidationError(t *testing.T) {
	t.Parallel()

	client := onec.NewClient("http://example.invalid", "", "")
	_, err := Resolve(context.Background(), client, "invalid_profile")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
