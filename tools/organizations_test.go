package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestReadOrganizationsTool(t *testing.T) {
	tool := ReadOrganizationsTool()
	if tool == nil || tool.Name != "read_organizations" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestReadOrganizationsHandler(t *testing.T) {
	const resp = `{
		"organizations":[{"ref":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","code":"0001","name":"ООО Тест","inn":"7701000001","kpp":"770101001"}],
		"total":1,
		"truncated":false
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/organizations" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewReadOrganizationsHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Name: "read_organizations", Arguments: []byte(`{"limit":10}`)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "ООО Тест") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

