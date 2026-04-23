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

func TestParseAccessMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		want    AccessMode
		wantErr bool
	}{
		{value: "", want: AccessReadWrite},
		{value: "read_write", want: AccessReadWrite},
		{value: "read_only", want: AccessReadOnly},
		{value: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			got, err := ParseAccessMode(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseAccessMode() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseAccessMode() = %q, want %q", got, tt.want)
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
	if !businessTools["read_nomenclature"] || !businessTools["create_nomenclature"] {
		t.Fatalf("expected nomenclature tools in business toolset, got %v", businessTools)
	}
	for _, tool := range []string{
		"read_organizations", "read_contracts", "read_sales_invoices",
		"create_sales_invoice", "update_sales_invoice",
		"read_sales_documents", "create_sales_document", "update_sales_document",
	} {
		if !businessTools[tool] {
			t.Fatalf("expected %s in business toolset, got %v", tool, businessTools)
		}
	}
}

func TestBusinessToolsetRespectsProfileSupport(t *testing.T) {
	t.Parallel()

	client := onec.NewClient("http://localhost:8080/mcp", "", "")

	supported := listToolNames(t, New("test", client, nil, Options{
		Toolset: ToolsetBusiness,
		Profile: "buh_3_0",
	}))
	if !supported["read_counterparties"] || !supported["create_counterparty"] ||
		!supported["read_nomenclature"] || !supported["create_nomenclature"] ||
		!supported["read_organizations"] || !supported["read_contracts"] ||
		!supported["read_sales_invoices"] || !supported["create_sales_invoice"] || !supported["update_sales_invoice"] ||
		!supported["read_sales_documents"] || !supported["create_sales_document"] || !supported["update_sales_document"] {
		t.Fatalf("expected business tools on supported profile, got %v", supported)
	}

	unsupported := listToolNames(t, New("test", client, nil, Options{
		Toolset: ToolsetBusiness,
		Profile: "unknown",
	}))
	if unsupported["read_counterparties"] || unsupported["create_counterparty"] ||
		unsupported["read_nomenclature"] || unsupported["create_nomenclature"] ||
		unsupported["read_organizations"] || unsupported["read_contracts"] ||
		unsupported["read_sales_invoices"] || unsupported["create_sales_invoice"] || unsupported["update_sales_invoice"] ||
		unsupported["read_sales_documents"] || unsupported["create_sales_document"] || unsupported["update_sales_document"] {
		t.Fatalf("did not expect business tools on unsupported profile, got %v", unsupported)
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

func TestBusinessReadOnlyModeDoesNotRegisterWriteTools(t *testing.T) {
	t.Parallel()

	client := onec.NewClient("http://localhost:8080/mcp", "", "")
	toolsList := listToolNames(t, New("test", client, nil, Options{
		Toolset: ToolsetBusiness,
		Profile: "generic",
		Access:  AccessReadOnly,
	}))

	for _, writeTool := range []string{
		"create_counterparty",
		"create_nomenclature",
		"create_sales_invoice",
		"update_sales_invoice",
		"create_sales_document",
		"update_sales_document",
	} {
		if toolsList[writeTool] {
			t.Fatalf("did not expect write tool %s in read_only mode, got %v", writeTool, toolsList)
		}
	}
	for _, readTool := range []string{
		"read_counterparties",
		"read_nomenclature",
		"read_organizations",
		"read_contracts",
		"read_sales_invoices",
		"read_sales_documents",
	} {
		if !toolsList[readTool] {
			t.Fatalf("expected read tool %s in read_only mode, got %v", readTool, toolsList)
		}
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
