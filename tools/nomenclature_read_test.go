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

func TestReadNomenclatureTool(t *testing.T) {
	tool := ReadNomenclatureTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "read_nomenclature" {
		t.Fatalf("unexpected tool name: %s", tool.Name)
	}
}

func TestReadNomenclatureHandler(t *testing.T) {
	const resp = `{
		"nomenclature":[
			{
				"ref":"22222222-2222-2222-2222-222222222222",
				"code":"00-000001",
				"name":"Тестовый товар",
				"article":"ART-001",
				"nomenclature_type":"Товары",
				"unit":"шт",
				"is_service":false
			}
		],
		"total":1,
		"truncated":false
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nomenclatures" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewReadNomenclatureHandler(onec.NewClient(srv.URL, "", ""))
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "read_nomenclature",
			Arguments: []byte(`{"search":"Тест","limit":10}`),
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Тестовый товар") {
		t.Fatalf("expected row in result, got:\n%s", text)
	}
}

func TestReadNomenclatureHandler_InvalidRefError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nomenclatures" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"ref must be valid UUID"}`))
	}))
	defer srv.Close()

	handler := NewReadNomenclatureHandler(onec.NewClient(srv.URL, "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "read_nomenclature",
			Arguments: []byte(`{"ref":"not-a-uuid"}`),
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}
