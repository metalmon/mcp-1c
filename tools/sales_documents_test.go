package tools

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestReadSalesDocumentsTool(t *testing.T) {
	tool := ReadSalesDocumentsTool()
	if tool == nil || tool.Name != "read_sales_documents" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestCreateSalesDocumentTool(t *testing.T) {
	tool := CreateSalesDocumentTool()
	if tool == nil || tool.Name != "create_sales_document" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestUpdateSalesDocumentTool(t *testing.T) {
	tool := UpdateSalesDocumentTool()
	if tool == nil || tool.Name != "update_sales_document" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestReadSalesDocumentsHandler(t *testing.T) {
	const resp = `{
		"documents":[{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","number":"0001","date":"2026-04-20","organization":"ООО Наша","counterparty":"ООО Тест","amount":"1000","posted":false}],
		"total":1,
		"truncated":false
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales-documents" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewReadSalesDocumentsHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Name: "read_sales_documents", Arguments: []byte(`{"limit":10}`)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "0001") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestCreateSalesDocumentHandler(t *testing.T) {
	const resp = `{
		"success":true,
		"document":{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","number":"0001","date":"2026-04-20","posted":true}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales-document" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		bodyText := string(body)
		if !strings.Contains(bodyText, `"date":"2026-04-20T00:00:00"`) || !strings.Contains(bodyText, `"invoice_ref":"inv-1"`) {
			http.Error(w, "required fields not passed", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewCreateSalesDocumentHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_sales_document",
			Arguments: []byte(`{"organization_ref":"a","counterparty_ref":"b","invoice_ref":"inv-1","date":"2026-04-20T00:00:00","post":true,"items":[{"nomenclature_ref":"n1","quantity":1,"price":100}]}`),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Реализация создана") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestCreateSalesDocumentHandler_PostWithoutItemsFails(t *testing.T) {
	handler := NewCreateSalesDocumentHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_sales_document",
			Arguments: []byte(`{"organization_ref":"a","counterparty_ref":"b","post":true}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "items are required when post=true") {
		t.Fatalf("expected items validation error, got: %v", err)
	}
}

func TestCreateSalesDocumentHandler_PostWithInvalidPriceFails(t *testing.T) {
	handler := NewCreateSalesDocumentHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_sales_document",
			Arguments: []byte(`{"organization_ref":"a","counterparty_ref":"b","post":true,"items":[{"nomenclature_ref":"n1","quantity":1,"price":0}]}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "price must be > 0 when post=true") {
		t.Fatalf("expected price validation error, got: %v", err)
	}
}

func TestCreateSalesDocumentHandler_DraftWithoutItemsAllowed(t *testing.T) {
	const resp = `{
		"success":true,
		"document":{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","number":"0002","date":"2026-04-20","posted":false}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales-document" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewCreateSalesDocumentHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_sales_document",
			Arguments: []byte(`{"organization_ref":"a","counterparty_ref":"b","post":false}`),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Posted: false") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestUpdateSalesDocumentHandler(t *testing.T) {
	const resp = `{
		"success":true,
		"document":{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","number":"0001","date":"2026-04-20","posted":true}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales-document" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		bodyText := string(body)
		if !strings.Contains(bodyText, `"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd"`) || !strings.Contains(bodyText, `"invoice_ref":"inv-2"`) {
			http.Error(w, "required fields not passed", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewUpdateSalesDocumentHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "update_sales_document",
			Arguments: []byte(`{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","invoice_ref":"inv-2","post":true,"items":[{"nomenclature_ref":"n1","quantity":1,"price":100}]}`),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Реализация обновлена") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestUpdateSalesDocumentHandler_RequiresRef(t *testing.T) {
	handler := NewUpdateSalesDocumentHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "update_sales_document",
			Arguments: []byte(`{"comment":"updated"}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "ref is required") {
		t.Fatalf("expected ref validation error, got: %v", err)
	}
}

func TestUpdateSalesDocumentHandler_PostWithInvalidPriceFails(t *testing.T) {
	handler := NewUpdateSalesDocumentHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "update_sales_document",
			Arguments: []byte(`{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","post":true,"items":[{"nomenclature_ref":"n1","quantity":1,"price":0}]}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "price must be > 0 when post=true") {
		t.Fatalf("expected price validation error, got: %v", err)
	}
}
