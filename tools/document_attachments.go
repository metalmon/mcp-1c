package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// stripWhitespaceFromBase64 removes all Unicode whitespace from a base64 payload.
// PEM-style line breaks or accidental spaces/newlines inside JSON strings break 1C Base64Значение
// or decode to corrupted binaries (PDF then fails to open).
func stripWhitespaceFromBase64(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

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
		Description: "Прикрепляет файл к документу 1С (HTTP POST, JSON). " +
			"Сигнатура для модели: обязательно знать document_ref целевого документа (типизированная ссылка Документ.<Имя>:<uuid>). " +
			"Поля file_name и content_base64 обязательны для запроса к 1С: base64 — сырое содержимое файла без префикса data:..., без переносов строк внутри строки. " +
			"Интегрированный клиент (например okta-chat): если пользователь приложил ровно один файл к последнему сообщению, клиент может сам подставить file_name, mime_type и content_base64 до вызова сервера — модели не нужно копировать base64 из «видения» картинки или из текста; достаточно document_ref и при необходимости comment. " +
			"Прямой вызов MCP без такого хоста: передайте полный набор document_ref + file_name + content_base64 (+ mime_type по желанию).",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"document_ref":{"type":"string","description":"Типизированная ссылка документа: Документ.<ИмяМетаданных>:<uuid>"},
				"file_name":{"type":"string","description":"Имя с расширением. При прямом MCP обязательно; в okta-chat с одним вложением у пользователя может подставить клиент."},
				"mime_type":{"type":"string","description":"MIME, например application/pdf или image/png. Опционально; клиент может вывести из вложения."},
				"content_base64":{"type":"string","description":"Сырой base64 тела файла. При прямом MCP обязательно целиком, не обрезать. Не восстанавливать из памяти модели — в okta-chat с одним вложением подставляет клиент из файла пользователя."},
				"comment":{"type":"string","description":"Комментарий к вложению в 1С (опционально)"}
			},
			"required":["document_ref"]
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
		if strings.TrimSpace(input.FileName) == "" || strings.TrimSpace(input.ContentBase64) == "" {
			return nil, fmt.Errorf(
				"file_name and content_base64 are required (integrated MCP hosts may inject them from the user attachment before this call)",
			)
		}
		body := onec.AttachFileToDocumentRequest{
			DocumentRef:   strings.TrimSpace(input.DocumentRef),
			FileName:      strings.TrimSpace(input.FileName),
			MimeType:      strings.TrimSpace(input.MimeType),
			ContentBase64: stripWhitespaceFromBase64(strings.TrimSpace(input.ContentBase64)),
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
