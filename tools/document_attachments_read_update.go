package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"path/filepath"
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
				"document_ref":{"type":"string","description":"Типизированная ссылка документа (<metadataFullName>:<uuid>)"},
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
				"attachment_ref":{"type":"string","description":"Типизированная ссылка вложения (attachmentCatalog:<uuid>)"}
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
				"attachment_ref":{"type":"string","description":"Типизированная ссылка вложения (attachmentCatalog:<uuid>)"},
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
	content := r.Content
	content.MimeType = normalizeMimeType(content.MimeType, content.Name)
	if strings.TrimSpace(content.ContractVersion) == "" {
		content.ContractVersion = "1.0"
	}
	if strings.TrimSpace(content.Encoding) == "" {
		content.Encoding = "base64"
	}
	if strings.TrimSpace(content.Injection.Mode) == "" {
		content.Injection.Mode = detectInjectionMode(content.MimeType)
	}
	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Sprintf(`{"id":"%s","name":"%s","mime_type":"%s","size_bytes":%d,"encoding":"%s","content":"%s","injection":{"mode":"%s"},"contract_version":"%s"}`,
			content.ID,
			content.Name,
			content.MimeType,
			content.SizeBytes,
			content.Encoding,
			content.Content,
			content.Injection.Mode,
			content.ContractVersion,
		)
	}
	return string(data)
}

func normalizeMimeType(rawMimeType, fileName string) string {
	mimeType := strings.TrimSpace(rawMimeType)
	if mimeType != "" {
		return mimeType
	}
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	if ext == "" {
		return ""
	}
	detected := mime.TypeByExtension(ext)
	return strings.TrimSpace(strings.Split(detected, ";")[0])
}

func detectInjectionMode(mimeType string) string {
	lower := strings.ToLower(strings.TrimSpace(mimeType))
	switch {
	case strings.HasPrefix(lower, "image/"):
		return "multimodal_image"
	case strings.HasPrefix(lower, "text/"),
		lower == "application/json",
		lower == "application/xml",
		lower == "text/xml",
		lower == "application/csv":
		return "inline_text"
	default:
		return "file_reference"
	}
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
