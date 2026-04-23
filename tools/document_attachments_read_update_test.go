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

func TestReadDocumentAttachmentsTool(t *testing.T) {
	tool := ReadDocumentAttachmentsTool()
	if tool == nil || tool.Name != "read_document_attachments" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestGetDocumentAttachmentContentTool(t *testing.T) {
	tool := GetDocumentAttachmentContentTool()
	if tool == nil || tool.Name != "get_document_attachment_content" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestUpdateDocumentAttachmentMetadataTool(t *testing.T) {
	tool := UpdateDocumentAttachmentMetadataTool()
	if tool == nil || tool.Name != "update_document_attachment_metadata" {
		t.Fatalf("unexpected tool: %+v", tool)
	}
}

func TestReadDocumentAttachmentsHandler(t *testing.T) {
	const resp = `{
		"attachments":[
			{
				"ref":"attachmentCatalog:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"document_ref":"Документ.РеализацияТоваровУслуг:bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"file_name":"scan.pdf",
				"description":"УПД",
				"mime_type":"application/pdf",
				"size_bytes":128
			}
		],
		"total":1,
		"truncated":false
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/document-attachments" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewReadDocumentAttachmentsHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Name: "read_document_attachments", Arguments: []byte(`{"document_ref":"Документ.РеализацияТоваровУслуг:bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}`)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "scan.pdf") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestGetDocumentAttachmentContentHandler(t *testing.T) {
	const resp = `{
		"content":{
			"id":"attachmentCatalog:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"name":"scan.pdf",
			"mime_type":"application/pdf",
			"size_bytes":128,
			"encoding":"base64",
			"content":"QUJDRA==",
			"injection":{"mode":"file_reference"},
			"contract_version":"1.0"
		}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/document-attachment-content" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewGetDocumentAttachmentContentHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Name: "get_document_attachment_content", Arguments: []byte(`{"attachment_ref":"attachmentCatalog:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, `"content":"QUJDRA=="`) || !strings.Contains(text, `"contract_version":"1.0"`) {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestUpdateDocumentAttachmentMetadataHandler(t *testing.T) {
	const resp = `{
		"success":true,
		"attachment":{
			"ref":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"file_name":"renamed.pdf",
			"description":"Новое описание"
		}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/document-attachment-metadata" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"file_name":"renamed.pdf"`) {
			http.Error(w, "required fields not passed", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	handler := NewUpdateDocumentAttachmentMetadataHandler(onec.NewClient(srv.URL, "", ""))
	result, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "update_document_attachment_metadata",
			Arguments: []byte(`{"attachment_ref":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","file_name":"renamed.pdf"}`),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "renamed.pdf") {
		t.Fatalf("unexpected result:\n%s", text)
	}
}

func TestUpdateDocumentAttachmentMetadataHandler_RequiresField(t *testing.T) {
	handler := NewUpdateDocumentAttachmentMetadataHandler(onec.NewClient("http://127.0.0.1:1", "", ""))
	_, err := handler(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "update_document_attachment_metadata",
			Arguments: []byte(`{"attachment_ref":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "at least one field is required") {
		t.Fatalf("expected metadata validation error, got: %v", err)
	}
}
