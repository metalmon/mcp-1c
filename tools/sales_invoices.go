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
	defaultSalesInvoicesLimit = 50
	maxSalesInvoicesLimit     = 500
)

type readSalesInvoicesInput struct {
	Search          string `json:"search,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Ref             string `json:"ref,omitempty"`
	Number          string `json:"number,omitempty"`
	CounterpartyRef string `json:"counterparty_ref,omitempty"`
	OrganizationRef string `json:"organization_ref,omitempty"`
}

type createSalesInvoiceInput struct {
	OrganizationRef string           `json:"organization_ref"`
	CounterpartyRef string           `json:"counterparty_ref"`
	ContractRef     string           `json:"contract_ref,omitempty"`
	Date            string           `json:"date,omitempty"`
	Comment         string           `json:"comment,omitempty"`
	Post            bool             `json:"post"`
	Items           []onec.SalesItem `json:"items,omitempty"`
}

func ReadSalesInvoicesTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "read_sales_invoices",
		Title: "Чтение счетов покупателю",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Description: "Инструмент 1С: получить документы СчетНаОплатуПокупателю: список, поиск или точечное чтение по ref/number.",
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

func NewReadSalesInvoicesHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input readSalesInvoicesInput
		if req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
				return nil, fmt.Errorf("parsing input: %w", err)
			}
		}
		limit := clampLimit(input.Limit, defaultSalesInvoicesLimit, maxSalesInvoicesLimit)
		body := onec.ReadSalesInvoicesRequest{
			Search:          strings.TrimSpace(input.Search),
			Limit:           limit,
			Ref:             strings.TrimSpace(input.Ref),
			Number:          strings.TrimSpace(input.Number),
			CounterpartyRef: strings.TrimSpace(input.CounterpartyRef),
			OrganizationRef: strings.TrimSpace(input.OrganizationRef),
		}
		var result onec.ReadSalesInvoicesResult
		if err := client.Post(ctx, "/sales-invoices", body, &result); err != nil {
			return nil, fmt.Errorf("reading sales invoices from 1C: %w", err)
		}
		return textResult(formatSalesInvoicesReadResult(&result)), nil
	}
}

func CreateSalesInvoiceTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "create_sales_invoice",
		Title: "Создание счета покупателю",
		Description: "Инструмент 1С: создать документ СчетНаОплатуПокупателю. " +
			"Обязательные поля: organization_ref, counterparty_ref. " +
			"Для post=true обязательна хотя бы одна товарная позиция items с quantity>0 и price>0. " +
			"Проведение управляется флагом post.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"organization_ref":{"type":"string","description":"Ссылка организации (UUID)"},
				"counterparty_ref":{"type":"string","description":"Ссылка контрагента (UUID)"},
				"contract_ref":{"type":"string","description":"Ссылка договора (UUID)"},
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

func NewCreateSalesInvoiceHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input createSalesInvoiceInput
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
		body := onec.CreateSalesInvoiceRequest{
			OrganizationRef: strings.TrimSpace(input.OrganizationRef),
			CounterpartyRef: strings.TrimSpace(input.CounterpartyRef),
			ContractRef:     strings.TrimSpace(input.ContractRef),
			Date:            strings.TrimSpace(input.Date),
			Comment:         strings.TrimSpace(input.Comment),
			Post:            input.Post,
			Items:           input.Items,
		}
		var result onec.CreateSalesInvoiceResult
		if err := client.Post(ctx, "/sales-invoice", body, &result); err != nil {
			return nil, fmt.Errorf("creating sales invoice in 1C: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("1C returned unsuccessful create result")
		}
		return textResult(formatCreateSalesInvoiceResult(&result)), nil
	}
}

func validatePostItems(items []onec.SalesItem) error {
	if len(items) == 0 {
		return fmt.Errorf("items are required when post=true")
	}
	for idx, item := range items {
		line := idx + 1
		if strings.TrimSpace(item.NomenclatureRef) == "" {
			return fmt.Errorf("items[%d].nomenclature_ref is required when post=true", line)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("items[%d].quantity must be > 0 when post=true", line)
		}
		if item.Price <= 0 {
			return fmt.Errorf("items[%d].price must be > 0 when post=true", line)
		}
	}
	return nil
}

func formatSalesInvoicesReadResult(r *onec.ReadSalesInvoicesResult) string {
	var b strings.Builder
	b.WriteString("# Счета покупателю\n\n")
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

func formatCreateSalesInvoiceResult(r *onec.CreateSalesInvoiceResult) string {
	doc := r.Document
	var b strings.Builder
	b.WriteString("# Счет покупателю создан\n\n")
	fmt.Fprintf(&b, "- Ref: %s\n", doc.Ref)
	fmt.Fprintf(&b, "- Number: %s\n", doc.Number)
	if doc.Date != "" {
		fmt.Fprintf(&b, "- Date: %s\n", doc.Date)
	}
	fmt.Fprintf(&b, "- Posted: %t\n", doc.Posted)
	return b.String()
}
