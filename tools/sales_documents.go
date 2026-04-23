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
	defaultSalesDocumentsLimit = 50
	maxSalesDocumentsLimit     = 500
)

type readSalesDocumentsInput struct {
	Search          string `json:"search,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Ref             string `json:"ref,omitempty"`
	Number          string `json:"number,omitempty"`
	CounterpartyRef string `json:"counterparty_ref,omitempty"`
	OrganizationRef string `json:"organization_ref,omitempty"`
}

type createSalesDocumentInput struct {
	OrganizationRef string           `json:"organization_ref"`
	CounterpartyRef string           `json:"counterparty_ref"`
	ContractRef     string           `json:"contract_ref,omitempty"`
	InvoiceRef      string           `json:"invoice_ref,omitempty"`
	Date            string           `json:"date,omitempty"`
	Comment         string           `json:"comment,omitempty"`
	Post            bool             `json:"post"`
	Items           []onec.SalesItem `json:"items,omitempty"`
}

func ReadSalesDocumentsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "read_sales_documents",
		Title: "Чтение реализаций",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Description: "Инструмент 1С: получить документы РеализацияТоваровУслуг: список, поиск или точечное чтение по ref/number.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"search":{"type":"string","description":"Поиск по номеру или комментарию"},
				"limit":{"type":"integer","description":"Максимум строк (по умолчанию 50, максимум 500)"},
				"ref":{"type":"string","description":"Ссылка документа (UUID)"},
				"number":{"type":"string","description":"Номер документа"},
				"counterparty_ref":{"type":"string","description":"Фильтр по контрагенту (UUID)"},
				"organization_ref":{"type":"string","description":"Фильтр по организации (UUID)"}
			}
		}`),
	}
}

func NewReadSalesDocumentsHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input readSalesDocumentsInput
		if req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
				return nil, fmt.Errorf("parsing input: %w", err)
			}
		}
		limit := clampLimit(input.Limit, defaultSalesDocumentsLimit, maxSalesDocumentsLimit)
		body := onec.ReadSalesDocumentsRequest{
			Search:          strings.TrimSpace(input.Search),
			Limit:           limit,
			Ref:             strings.TrimSpace(input.Ref),
			Number:          strings.TrimSpace(input.Number),
			CounterpartyRef: strings.TrimSpace(input.CounterpartyRef),
			OrganizationRef: strings.TrimSpace(input.OrganizationRef),
		}
		var result onec.ReadSalesDocumentsResult
		if err := client.Post(ctx, "/sales-documents", body, &result); err != nil {
			return nil, fmt.Errorf("reading sales documents from 1C: %w", err)
		}
		return textResult(formatSalesDocumentsReadResult(&result)), nil
	}
}

func CreateSalesDocumentTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "create_sales_document",
		Title: "Создание реализации",
		Description: "Инструмент 1С: создать документ РеализацияТоваровУслуг. " +
			"Обязательные поля: organization_ref, counterparty_ref. " +
			"Для post=true обязательна хотя бы одна товарная позиция items с quantity>0 и price>0. " +
			"Проведение управляется флагом post.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"organization_ref":{"type":"string","description":"Ссылка организации (UUID)"},
				"counterparty_ref":{"type":"string","description":"Ссылка контрагента (UUID)"},
				"contract_ref":{"type":"string","description":"Ссылка договора (UUID)"},
				"invoice_ref":{"type":"string","description":"Ссылка счета покупателю (UUID)"},
				"date":{"type":"string","description":"Дата документа в формате XML даты 1С (например, 2026-04-20T00:00:00)"},
				"comment":{"type":"string","description":"Комментарий"},
				"post":{"type":"boolean","description":"Провести документ после записи"},
				"items":{
					"type":"array",
					"description":"Строки табличной части Товары",
					"items":{
						"type":"object",
						"properties":{
							"nomenclature_ref":{"type":"string","description":"Ссылка номенклатуры (UUID)"},
							"quantity":{"type":"number","description":"Количество"},
							"price":{"type":"number","description":"Цена"}
						},
						"required":["nomenclature_ref","quantity","price"]
					}
				}
			},
			"required":["organization_ref","counterparty_ref"]
		}`),
	}
}

func NewCreateSalesDocumentHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input createSalesDocumentInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if strings.TrimSpace(input.OrganizationRef) == "" || strings.TrimSpace(input.CounterpartyRef) == "" {
			return nil, fmt.Errorf("organization_ref and counterparty_ref are required")
		}
		if input.Post {
			if err := validatePostItems(input.Items); err != nil {
				return nil, err
			}
		}
		body := onec.CreateSalesDocumentRequest{
			OrganizationRef: strings.TrimSpace(input.OrganizationRef),
			CounterpartyRef: strings.TrimSpace(input.CounterpartyRef),
			ContractRef:     strings.TrimSpace(input.ContractRef),
			InvoiceRef:      strings.TrimSpace(input.InvoiceRef),
			Date:            strings.TrimSpace(input.Date),
			Comment:         strings.TrimSpace(input.Comment),
			Post:            input.Post,
			Items:           input.Items,
		}
		var result onec.CreateSalesDocumentResult
		if err := client.Post(ctx, "/sales-document", body, &result); err != nil {
			return nil, fmt.Errorf("creating sales document in 1C: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("1C returned unsuccessful create result")
		}
		return textResult(formatCreateSalesDocumentResult(&result)), nil
	}
}

func formatSalesDocumentsReadResult(r *onec.ReadSalesDocumentsResult) string {
	var b strings.Builder
	b.WriteString("# Реализации\n\n")
	if len(r.Documents) == 0 {
		b.WriteString("Ничего не найдено.\n")
		return b.String()
	}
	b.WriteString("| Ref | Number | Date | Organization | Counterparty | Amount | Posted |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, doc := range r.Documents {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %t |\n",
			doc.Ref, doc.Number, doc.Date, doc.Organization, doc.Counterparty, doc.Amount, doc.Posted)
	}
	if r.Truncated {
		b.WriteString("\n> Показаны не все записи. Увеличьте limit.\n")
	}
	return b.String()
}

func formatCreateSalesDocumentResult(r *onec.CreateSalesDocumentResult) string {
	doc := r.Document
	var b strings.Builder
	b.WriteString("# Реализация создана\n\n")
	fmt.Fprintf(&b, "- Ref: %s\n", doc.Ref)
	fmt.Fprintf(&b, "- Number: %s\n", doc.Number)
	if doc.Date != "" {
		fmt.Fprintf(&b, "- Date: %s\n", doc.Date)
	}
	fmt.Fprintf(&b, "- Posted: %t\n", doc.Posted)
	return b.String()
}
