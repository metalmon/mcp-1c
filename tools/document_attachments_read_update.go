package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultDocumentAttachmentsLimit = 50
	maxDocumentAttachmentsLimit     = 500
)

type readDocumentAttachmentsInput struct {
	DocumentRef string `json:"document_ref"`
	Limit       int    `json:"limit,omitempty"`
	Search      string `json:"search,omitempty"`
}

type getDocumentAttachmentContentInput struct {
	AttachmentRef string `json:"attachment_ref"`
}

type updateDocumentAttachmentMetadataInput struct {
	AttachmentRef string `json:"attachment_ref"`
	FileName      string `json:"file_name,omitempty"`
	Description   string `json:"description,omitempty"`
}

func ReadDocumentAttachmentsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "read_document_attachments",
		Title: "Чтение вложений документа",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Description: "Инструмент 1С: получить список вложений документа по document_ref.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"document_ref":{"type":"string","description":"Ссылка документа (UUID)"},
				"limit":{"type":"integer","description":"Максимум строк (по умолчанию 50, максимум 500)"},
				"search":{"type":"string","description":"Поиск по имени или описанию вложения"}
			},
			"required":["document_ref"]
		}`),
	}
}

func NewReadDocumentAttachmentsHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input readDocumentAttachmentsInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if strings.TrimSpace(input.DocumentRef) == "" {
			return nil, fmt.Errorf("document_ref is required")
		}
		body := onec.ReadDocumentAttachmentsRequest{
			DocumentRef: strings.TrimSpace(input.DocumentRef),
			Limit:       clampLimit(input.Limit, defaultDocumentAttachmentsLimit, maxDocumentAttachmentsLimit),
			Search:      strings.TrimSpace(input.Search),
		}
		var result onec.ReadDocumentAttachmentsResult
		if err := client.Post(ctx, "/document-attachments", body, &result); err != nil {
			return nil, fmt.Errorf("reading document attachments from 1C: %w", err)
		}
		return textResult(formatReadDocumentAttachmentsResult(&result)), nil
	}
}

func GetDocumentAttachmentContentTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "get_document_attachment_content",
		Title: "Чтение содержимого вложения",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Description: "Инструмент 1С: получить содержимое вложения по attachment_ref в base64.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"attachment_ref":{"type":"string","description":"Ссылка вложения (UUID)"}
			},
			"required":["attachment_ref"]
		}`),
	}
}

func NewGetDocumentAttachmentContentHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input getDocumentAttachmentContentInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if strings.TrimSpace(input.AttachmentRef) == "" {
			return nil, fmt.Errorf("attachment_ref is required")
		}
		body := onec.GetDocumentAttachmentContentRequest{
			AttachmentRef: strings.TrimSpace(input.AttachmentRef),
		}
		var result onec.GetDocumentAttachmentContentResult
		if err := client.Post(ctx, "/document-attachment-content", body, &result); err != nil {
			return nil, fmt.Errorf("reading document attachment content from 1C: %w", err)
		}
		return textResult(formatGetDocumentAttachmentContentResult(&result)), nil
	}
}

func UpdateDocumentAttachmentMetadataTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "update_document_attachment_metadata",
		Title: "Обновление метаданных вложения",
		Description: "Инструмент 1С: обновить имя и/или описание вложения по attachment_ref.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"attachment_ref":{"type":"string","description":"Ссылка вложения (UUID)"},
				"file_name":{"type":"string","description":"Новое имя файла"},
				"description":{"type":"string","description":"Новое описание вложения"}
			},
			"required":["attachment_ref"]
		}`),
	}
}

func NewUpdateDocumentAttachmentMetadataHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input updateDocumentAttachmentMetadataInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if strings.TrimSpace(input.AttachmentRef) == "" {
			return nil, fmt.Errorf("attachment_ref is required")
		}
		if strings.TrimSpace(input.FileName) == "" && strings.TrimSpace(input.Description) == "" {
			return nil, fmt.Errorf("at least one field is required: file_name or description")
		}
		body := onec.UpdateDocumentAttachmentMetadataRequest{
			AttachmentRef: strings.TrimSpace(input.AttachmentRef),
			FileName:      strings.TrimSpace(input.FileName),
			Description:   strings.TrimSpace(input.Description),
		}
		var result onec.UpdateDocumentAttachmentMetadataResult
		if err := client.Post(ctx, "/document-attachment-metadata", body, &result); err != nil {
			return nil, fmt.Errorf("updating document attachment metadata in 1C: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("1C returned unsuccessful update result")
		}
		return textResult(formatUpdateDocumentAttachmentMetadataResult(&result)), nil
	}
}

func formatReadDocumentAttachmentsResult(r *onec.ReadDocumentAttachmentsResult) string {
	var b strings.Builder
	b.WriteString("# Вложения документа\n\n")
	if len(r.Attachments) == 0 {
		b.WriteString("Вложения не найдены.\n")
		return b.String()
	}
	b.WriteString("| Ref | File Name | Description | Mime Type | Size (bytes) | Created At |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, att := range r.Attachments {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %d | %s |\n",
			att.Ref, att.FileName, att.Description, att.MimeType, att.SizeBytes, att.CreatedAt)
	}
	if r.Truncated {
		b.WriteString("\n> Показаны не все записи. Увеличьте limit.\n")
	}
	return b.String()
}

func formatGetDocumentAttachmentContentResult(r *onec.GetDocumentAttachmentContentResult) string {
	att := r.Attachment
	var b strings.Builder
	b.WriteString("# Содержимое вложения\n\n")
	fmt.Fprintf(&b, "- Ref: %s\n", att.Ref)
	fmt.Fprintf(&b, "- Document Ref: %s\n", att.DocumentRef)
	fmt.Fprintf(&b, "- File Name: %s\n", att.FileName)
	if att.MimeType != "" {
		fmt.Fprintf(&b, "- Mime Type: %s\n", att.MimeType)
	}
	if att.SizeBytes > 0 {
		fmt.Fprintf(&b, "- Size (bytes): %d\n", att.SizeBytes)
	}
	if att.ContentBase64 != "" {
		fmt.Fprintf(&b, "- Content Base64: %s\n", att.ContentBase64)
	}
	return b.String()
}

func formatUpdateDocumentAttachmentMetadataResult(r *onec.UpdateDocumentAttachmentMetadataResult) string {
	att := r.Attachment
	var b strings.Builder
	b.WriteString("# Метаданные вложения обновлены\n\n")
	fmt.Fprintf(&b, "- Ref: %s\n", att.Ref)
	fmt.Fprintf(&b, "- File Name: %s\n", att.FileName)
	fmt.Fprintf(&b, "- Description: %s\n", att.Description)
	return b.String()
}
