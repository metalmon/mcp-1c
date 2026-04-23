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

func TestReadContractsTool(t *testing.T) {
	tool := ReadContractsTool()
	if tool == nil || tool.Name != "read_contracts" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestReadContractsHandler(t *testing.T) {
	const resp = `{
		"contracts":[{"ref":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb","code":"0001","name":"Основной","number":"Д-1","counterparty":"ООО Тест","organization":"ООО Наша","currency":"RUB"}],
		"total":1,
		"truncated":false
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/contracts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewReadContractsHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Name: "read_contracts", Arguments: []byte(`{"limit":10}`)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Д-1") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

