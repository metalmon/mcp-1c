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

func TestReadCounterpartiesTool(t *testing.T) {
	tool := ReadCounterpartiesTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "read_counterparties" {
		t.Fatalf("unexpected tool name: %s", tool.Name)
	}
}

func TestCreateCounterpartyTool(t *testing.T) {
	tool := CreateCounterpartyTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "create_counterparty" {
		t.Fatalf("unexpected tool name: %s", tool.Name)
	}
}

func TestReadCounterpartiesHandler(t *testing.T) {
	const resp = `{
		"counterparties":[{"ref":"11111111-1111-1111-1111-111111111111","code":"0001","name":"ООО Ромашка","inn":"7701000001","kpp":"770101001","counterparty_type":"ЮридическоеЛицо"}],
		"total":1,
		"truncated":false
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/counterparties" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewReadCounterpartiesHandler(onec.NewClient(srv.URL, "", ""))
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "read_counterparties",
			Arguments: []byte(`{"search":"Ромашка","limit":10,"inn":"7701000001","kpp":"770101001"}`),
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "ООО Ромашка") {
		t.Fatalf("expected row in result, got:\n%s", text)
	}
}

func TestCreateCounterpartyHandler(t *testing.T) {
	const resp = `{
		"success": true,
		"counterparty": {
			"ref":"11111111-1111-1111-1111-111111111111",
			"code":"0001",
			"name":"ООО Ромашка",
			"inn":"7701000001",
			"kpp":"770101001",
			"counterparty_type":"ЮридическоеЛицо"
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/counterparty" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewCreateCounterpartyHandler(onec.NewClient(srv.URL, "", ""))
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_counterparty",
			Arguments: []byte(`{"name":"ООО Ромашка","inn":"7701000001","kpp":"770101001","counterparty_type":"legal"}`),
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Контрагент создан") {
		t.Fatalf("unexpected text result:\n%s", text)
	}
}

func TestCounterpartyRoundTripByRef(t *testing.T) {
	const ref = "33333333-3333-3333-3333-333333333333"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/counterparty":
			_, _ = w.Write([]byte(`{
				"success": true,
				"counterparty": {
					"ref":"` + ref + `",
					"code":"0002",
					"name":"ООО Тест Roundtrip",
					"inn":"7702000002",
					"kpp":"770201001",
					"counterparty_type":"ЮридическоеЛицо"
				}
			}`))
		case "/counterparties":
			_, _ = w.Write([]byte(`{
				"counterparties":[{"ref":"` + ref + `","code":"0002","name":"ООО Тест Roundtrip","inn":"7702000002","kpp":"770201001","counterparty_type":"ЮридическоеЛицо"}],
				"total":1,
				"truncated":false
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := onec.NewClient(srv.URL, "", "")
	createHandler := NewCreateCounterpartyHandler(client)
	readHandler := NewReadCounterpartiesHandler(client)

	_, err := createHandler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_counterparty",
			Arguments: []byte(`{"name":"ООО Тест Roundtrip","inn":"7702000002","kpp":"770201001","counterparty_type":"legal"}`),
		},
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	readResult, err := readHandler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "read_counterparties",
			Arguments: []byte(`{"ref":"` + ref + `"}`),
		},
	})
	if err != nil {
		t.Fatalf("read by ref failed: %v", err)
	}

	text := readResult.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "ООО Тест Roundtrip") {
		t.Fatalf("roundtrip result missing entity:\n%s", text)
	}
}

func TestReadCounterpartiesHandler_InvalidRefError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/counterparties" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"ref must be valid UUID"}`))
	}))
	defer srv.Close()

	handler := NewReadCounterpartiesHandler(onec.NewClient(srv.URL, "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "read_counterparties",
			Arguments: []byte(`{"ref":"not-a-uuid"}`),
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}
