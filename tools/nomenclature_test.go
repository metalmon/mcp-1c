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

func TestCreateNomenclatureTool(t *testing.T) {
	tool := CreateNomenclatureTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "create_nomenclature" {
		t.Fatalf("unexpected tool name: %s", tool.Name)
	}
}

func TestCreateNomenclatureHandler(t *testing.T) {
	const resp = `{
		"success": true,
		"nomenclature": {
			"ref":"22222222-2222-2222-2222-222222222222",
			"code":"00-000001",
			"name":"Тестовый товар",
			"full_name":"Тестовый товар полный",
			"article":"ART-001",
			"nomenclature_type":"Товары",
			"unit":"шт",
			"is_service": false
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nomenclature" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewCreateNomenclatureHandler(onec.NewClient(srv.URL, "", ""))
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name: "create_nomenclature",
			Arguments: []byte(`{
				"name":"Тестовый товар",
				"full_name":"Тестовый товар полный",
				"article":"ART-001",
				"nomenclature_type":"Товары",
				"unit":"шт",
				"is_service":false
			}`),
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Номенклатура создана") {
		t.Fatalf("unexpected text result:\n%s", text)
	}
	if !strings.Contains(text, "Тестовый товар") {
		t.Fatalf("missing item name in result:\n%s", text)
	}
}

func TestNomenclatureRoundTripByRef(t *testing.T) {
	const ref = "44444444-4444-4444-4444-444444444444"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/nomenclature":
			_, _ = w.Write([]byte(`{
				"success": true,
				"nomenclature": {
					"ref":"` + ref + `",
					"code":"00-000002",
					"name":"Тест Roundtrip Номенклатура",
					"full_name":"Тест Roundtrip Номенклатура Полная",
					"article":"RT-002",
					"is_service": false
				}
			}`))
		case "/nomenclatures":
			_, _ = w.Write([]byte(`{
				"nomenclature":[
					{
						"ref":"` + ref + `",
						"code":"00-000002",
						"name":"Тест Roundtrip Номенклатура",
						"full_name":"Тест Roundtrip Номенклатура Полная",
						"article":"RT-002",
						"is_service": false
					}
				],
				"total":1,
				"truncated":false
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := onec.NewClient(srv.URL, "", "")
	createHandler := NewCreateNomenclatureHandler(client)
	readHandler := NewReadNomenclatureHandler(client)

	_, err := createHandler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_nomenclature",
			Arguments: []byte(`{"name":"Тест Roundtrip Номенклатура","article":"RT-002","is_service":false}`),
		},
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	readResult, err := readHandler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "read_nomenclature",
			Arguments: []byte(`{"ref":"` + ref + `"}`),
		},
	})
	if err != nil {
		t.Fatalf("read by ref failed: %v", err)
	}

	text := readResult.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Тест Roundtrip Номенклатура") {
		t.Fatalf("roundtrip result missing entity:\n%s", text)
	}
}
