package server

import (
	"context"
	"testing"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewServer(t *testing.T) {
	client := onec.NewClient("http://localhost:8080/mcp", "", "")
	s := New("test", client, nil)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestParseToolset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		want    Toolset
		wantErr bool
	}{
		{value: "", want: ToolsetAll},
		{value: "all", want: ToolsetAll},
		{value: "developer", want: ToolsetDeveloper},
		{value: "business", want: ToolsetBusiness},
		{value: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			got, err := ParseToolset(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseToolset() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseToolset() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToolsetsAffectRegisteredTools(t *testing.T) {
	t.Parallel()

	client := onec.NewClient("http://localhost:8080/mcp", "", "")

	allTools := listToolNames(t, New("test", client, nil))
	if !allTools["get_metadata_tree"] || !allTools["read_counterparties"] {
		t.Fatalf("expected both developer and business tools, got %v", allTools)
	}

	developerTools := listToolNames(t, New("test", client, nil, Options{
		Toolset: ToolsetDeveloper,
		Profile: "generic",
	}))
	if !developerTools["get_metadata_tree"] {
		t.Fatalf("expected developer tools, got %v", developerTools)
	}
	if developerTools["read_counterparties"] {
		t.Fatalf("did not expect business tools in developer toolset, got %v", developerTools)
	}

	businessTools := listToolNames(t, New("test", client, nil, Options{
		Toolset: ToolsetBusiness,
		Profile: "generic",
	}))
	if businessTools["get_metadata_tree"] {
		t.Fatalf("did not expect developer tools in business toolset, got %v", businessTools)
	}
	if !businessTools["read_counterparties"] || !businessTools["create_counterparty"] {
		t.Fatalf("expected business tools in business toolset, got %v", businessTools)
	}
}

func TestBusinessToolsetRespectsProfileSupport(t *testing.T) {
	t.Parallel()

	client := onec.NewClient("http://localhost:8080/mcp", "", "")

	supported := listToolNames(t, New("test", client, nil, Options{
		Toolset: ToolsetBusiness,
		Profile: "buh_3_0",
	}))
	if !supported["read_counterparties"] || !supported["create_counterparty"] {
		t.Fatalf("expected counterparties tools on supported profile, got %v", supported)
	}

	unsupported := listToolNames(t, New("test", client, nil, Options{
		Toolset: ToolsetBusiness,
		Profile: "unknown",
	}))
	if unsupported["read_counterparties"] || unsupported["create_counterparty"] {
		t.Fatalf("did not expect counterparties tools on unsupported profile, got %v", unsupported)
	}
}

func TestBusinessToolsetDoesNotRegisterDeveloperPrompts(t *testing.T) {
	t.Parallel()

	client := onec.NewClient("http://localhost:8080/mcp", "", "")
	srv := New("test", client, nil, Options{
		Toolset: ToolsetBusiness,
		Profile: "generic",
	})

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	result, err := session.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}
	if len(result.Prompts) != 0 {
		t.Fatalf("expected no prompts in business toolset, got %d", len(result.Prompts))
	}
}

func listToolNames(t *testing.T, srv *mcp.Server) map[string]bool {
	t.Helper()

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	list, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	names := make(map[string]bool, len(list.Tools))
	for _, tool := range list.Tools {
		names[tool.Name] = true
	}
	return names
}
