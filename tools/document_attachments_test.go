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

func TestAttachFileToDocumentTool(t *testing.T) {
	tool := AttachFileToDocumentTool()
	if tool == nil || tool.Name != "attach_file_to_document" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestAttachFileToDocumentHandler(t *testing.T) {
	const resp = `{
		"success": true,
		"attachment": {
			"ref": "attachmentCatalog:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"document_ref": "Документ.РеализацияТоваровУслуг:bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			"file_name": "Счет_0001_20260420_scan.pdf",
			"mime_type": "application/pdf",
			"size_bytes": 128
		}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/document-attachment" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		bodyText := string(body)
		if !strings.Contains(bodyText, `"document_ref":"Документ.РеализацияТоваровУслуг:bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"`) {
			http.Error(w, "document_ref not passed", http.StatusBadRequest)
			return
		}
		if !strings.Contains(bodyText, `"content_base64":"QUJDRA=="`) {
			http.Error(w, "content_base64 not passed", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewAttachFileToDocumentHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name: "attach_file_to_document",
			Arguments: []byte(`{
				"document_ref":"Документ.РеализацияТоваровУслуг:bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"file_name":"scan.pdf",
				"mime_type":"application/pdf",
				"content_base64":"QUJDRA=="
			}`),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Файл прикреплен к документу") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestAttachFileToDocumentHandler_Validation(t *testing.T) {
	handler := NewAttachFileToDocumentHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "attach_file_to_document",
			Arguments: []byte(`{"file_name":"x.pdf","content_base64":"AA=="}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "document_ref is required") {
		t.Fatalf("expected validation error, got: %v", err)
	}
}
