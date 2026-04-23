package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type attachFileToDocumentInput struct {
	DocumentRef   string `json:"document_ref"`
	FileName      string `json:"file_name"`
	MimeType      string `json:"mime_type,omitempty"`
	ContentBase64 string `json:"content_base64"`
	Comment       string `json:"comment,omitempty"`
}

func AttachFileToDocumentTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "attach_file_to_document",
		Title: "Прикрепление файла к документу",
		Description: "Инструмент 1С: прикрепить файл к документу по document_ref. " +
			"Файл передается в base64 без изменения формата.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"document_ref":{"type":"string","description":"Ссылка документа (UUID)"},
				"file_name":{"type":"string","description":"Имя файла с расширением"},
				"mime_type":{"type":"string","description":"MIME-тип файла, например application/pdf"},
				"content_base64":{"type":"string","description":"Содержимое файла в формате base64"},
				"comment":{"type":"string","description":"Комментарий к вложению"}
			},
			"required":["document_ref","file_name","content_base64"]
		}`),
	}
}

func NewAttachFileToDocumentHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input attachFileToDocumentInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if strings.TrimSpace(input.DocumentRef) == "" {
			return nil, fmt.Errorf("document_ref is required")
		}
		if strings.TrimSpace(input.FileName) == "" {
			return nil, fmt.Errorf("file_name is required")
		}
		if strings.TrimSpace(input.ContentBase64) == "" {
			return nil, fmt.Errorf("content_base64 is required")
		}
		body := onec.AttachFileToDocumentRequest{
			DocumentRef:   strings.TrimSpace(input.DocumentRef),
			FileName:      strings.TrimSpace(input.FileName),
			MimeType:      strings.TrimSpace(input.MimeType),
			ContentBase64: strings.TrimSpace(input.ContentBase64),
			Comment:       strings.TrimSpace(input.Comment),
		}
		var result onec.AttachFileToDocumentResult
		if err := client.Post(ctx, "/document-attachment", body, &result); err != nil {
			return nil, fmt.Errorf("attaching file to document in 1C: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("1C returned unsuccessful attach result")
		}
		return textResult(formatAttachFileToDocumentResult(&result)), nil
	}
}

func formatAttachFileToDocumentResult(r *onec.AttachFileToDocumentResult) string {
	att := r.Attachment
	var b strings.Builder
	b.WriteString("# Файл прикреплен к документу\n\n")
	fmt.Fprintf(&b, "- File Ref: %s\n", att.Ref)
	fmt.Fprintf(&b, "- Document Ref: %s\n", att.DocumentRef)
	fmt.Fprintf(&b, "- File Name: %s\n", att.FileName)
	if att.MimeType != "" {
		fmt.Fprintf(&b, "- Mime Type: %s\n", att.MimeType)
	}
	if att.SizeBytes > 0 {
		fmt.Fprintf(&b, "- Size (bytes): %d\n", att.SizeBytes)
	}
	return b.String()
}
