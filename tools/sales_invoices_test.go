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

func TestReadSalesInvoicesTool(t *testing.T) {
	tool := ReadSalesInvoicesTool()
	if tool == nil || tool.Name != "read_sales_invoices" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestCreateSalesInvoiceTool(t *testing.T) {
	tool := CreateSalesInvoiceTool()
	if tool == nil || tool.Name != "create_sales_invoice" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestUpdateSalesInvoiceTool(t *testing.T) {
	tool := UpdateSalesInvoiceTool()
	if tool == nil || tool.Name != "update_sales_invoice" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestReadSalesInvoicesHandler(t *testing.T) {
	const resp = `{
		"documents":[{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","number":"0001","date":"2026-04-20","organization":"ООО Наша","counterparty":"ООО Тест","amount":"1000","posted":false}],
		"total":1,
		"truncated":false
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales-invoices" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewReadSalesInvoicesHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Name: "read_sales_invoices", Arguments: []byte(`{"limit":10}`)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "0001") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestCreateSalesInvoiceHandler(t *testing.T) {
	const resp = `{
		"success":true,
		"document":{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","number":"0001","date":"2026-04-20","posted":true}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales-invoice" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"date":"2026-04-20T00:00:00"`) {
			http.Error(w, "date not passed", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewCreateSalesInvoiceHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_sales_invoice",
			Arguments: []byte(`{"organization_ref":"a","counterparty_ref":"b","date":"2026-04-20T00:00:00","post":true,"items":[{"nomenclature_ref":"n1","quantity":1,"price":100}]}`),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Счет покупателю создан") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestCreateSalesInvoiceHandler_PostWithoutItemsFails(t *testing.T) {
	handler := NewCreateSalesInvoiceHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_sales_invoice",
			Arguments: []byte(`{"organization_ref":"a","counterparty_ref":"b","post":true}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "items are required when post=true") {
		t.Fatalf("expected items validation error, got: %v", err)
	}
}

func TestCreateSalesInvoiceHandler_PostWithInvalidItemFails(t *testing.T) {
	handler := NewCreateSalesInvoiceHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_sales_invoice",
			Arguments: []byte(`{"organization_ref":"a","counterparty_ref":"b","post":true,"items":[{"nomenclature_ref":"n1","quantity":0,"price":100}]}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "quantity must be > 0 when post=true") {
		t.Fatalf("expected quantity validation error, got: %v", err)
	}
}

func TestCreateSalesInvoiceHandler_DraftWithoutItemsAllowed(t *testing.T) {
	const resp = `{
		"success":true,
		"document":{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","number":"0002","date":"2026-04-20","posted":false}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales-invoice" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewCreateSalesInvoiceHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "create_sales_invoice",
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

func TestUpdateSalesInvoiceHandler(t *testing.T) {
	const resp = `{
		"success":true,
		"document":{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","number":"0001","date":"2026-04-20","posted":true}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales-invoice" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		bodyText := string(body)
		if !strings.Contains(bodyText, `"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc"`) || !strings.Contains(bodyText, `"comment":"updated"`) {
			http.Error(w, "required fields not passed", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewUpdateSalesInvoiceHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "update_sales_invoice",
			Arguments: []byte(`{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","comment":"updated","post":true,"items":[{"nomenclature_ref":"n1","quantity":1,"price":100}]}`),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Счет покупателю обновлен") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestUpdateSalesInvoiceHandler_RequiresRef(t *testing.T) {
	handler := NewUpdateSalesInvoiceHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "update_sales_invoice",
			Arguments: []byte(`{"comment":"updated"}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "ref is required") {
		t.Fatalf("expected ref validation error, got: %v", err)
	}
}

func TestUpdateSalesInvoiceHandler_PostWithInvalidItemFails(t *testing.T) {
	handler := NewUpdateSalesInvoiceHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "update_sales_invoice",
			Arguments: []byte(`{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","post":true,"items":[{"nomenclature_ref":"n1","quantity":0,"price":100}]}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "quantity must be > 0 when post=true") {
		t.Fatalf("expected quantity validation error, got: %v", err)
	}
}
