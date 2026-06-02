package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Input for create_incoming_goods_document. Mirrors
// onec.CreateIncomingGoodsRequest but uses pointer bools for
// is_universal_document / vat_included so the BSL side can apply its
// own defaults (true) when the AI omits them.
type createIncomingGoodsInput struct {
	OrganizationRef     string              `json:"organization_ref"`
	CounterpartyRef     string              `json:"counterparty_ref"`
	ContractRef         string              `json:"contract_ref,omitempty"`
	DocNumberIn         string              `json:"doc_number_in,omitempty"`
	DocDateIn           string              `json:"doc_date_in,omitempty"`
	Amount              string              `json:"amount,omitempty"`
	VATIncluded         *bool               `json:"vat_included,omitempty"`
	IsUniversalDocument *bool               `json:"is_universal_document,omitempty"`
	OriginalReceived    *bool               `json:"original_received,omitempty"`
	Comment             string              `json:"comment,omitempty"`
	Post                bool                `json:"post"`
	Items               []onec.IncomingGoodsItem `json:"items"`
}

// CreateIncomingGoodsDocumentTool exposes "Поступление товаров и услуг"
// creation to AI clients. The OCR plugin uses the same /incoming-goods
// HTTP endpoint directly — this tool is the AI-facing parity wrapper.
func CreateIncomingGoodsDocumentTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "create_incoming_goods_document",
		Title: "Создание поступления",
		Description: "Инструмент 1С: создать документ ПоступлениеТоваровУслуг " +
			"(акт, накладная, УПД). Все справочники (Организация, Контрагент, " +
			"Договор, Номенклатура) задаются типизированными ссылками — " +
			"матчинг по наименованию/ИНН выполняйте через read_* tools заранее. " +
			"Идемпотентно по (Контрагент, doc_number_in, doc_date_in, amount) — " +
			"повторный вызов возвращает existing=true. Проведение управляется " +
			"флагом post (по умолчанию документ записывается без проведения).",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"organization_ref":{"type":"string","description":"Ссылка организации (Справочник.Организации:<uuid>)"},
				"counterparty_ref":{"type":"string","description":"Ссылка контрагента (Справочник.Контрагенты:<uuid>)"},
				"contract_ref":{"type":"string","description":"Ссылка договора (Справочник.ДоговорыКонтрагентов:<uuid>). Обязательный реквизит документа — если опущен и у контрагента нет единственного договора с поставщиком, 1С вернёт ошибку."},
				"doc_number_in":{"type":"string","description":"Номер документа сторонней организации (УПД/ТОРГ-12)"},
				"doc_date_in":{"type":"string","description":"Дата документа сторонней организации, YYYY-MM-DD"},
				"amount":{"type":"string","description":"СуммаДокумента (сумма с НДС или без, согласно vat_included)"},
				"vat_included":{"type":"boolean","description":"СуммаВключаетНДС. По умолчанию true."},
				"is_universal_document":{"type":"boolean","description":"ЭтоУниверсальныйДокумент: true для УПД (по умолчанию), false для ТОРГ-12 (с отдельным счёт-фактурой)."},
				"original_received":{"type":"boolean","description":"Отметить «Оригинал получен» (подсистема учёта оригиналов первичных документов). По умолчанию true — документ создан из скана."},
				"comment":{"type":"string","description":"Комментарий"},
				"post":{"type":"boolean","description":"Провести документ после записи"},
				"items":{
					"type":"array",
					"description":"Строки табличной части Товары",
					"items":{
						"type":"object",
						"properties":{
							"nomenclature_ref":{"type":"string","description":"Ссылка номенклатуры (Справочник.Номенклатура:<uuid>)"},
							"quantity":{"type":"string","description":"Количество"},
							"price":{"type":"string","description":"Цена"},
							"amount":{"type":"string","description":"Сумма без/с НДС, согласно vat_included"},
							"vat_rate":{"type":"string","description":"СтавкаНДС: 20, 10, 0, 18, 20/120, 10/110, 18/118, БезНДС"},
							"vat_amount":{"type":"string","description":"СуммаНДС. Если опущена — посчитается типовым хелпером по ставке и сумме."}
						},
						"required":["nomenclature_ref"]
					}
				}
			},
			"required":["organization_ref","counterparty_ref","items"]
		}`),
	}
}

func NewCreateIncomingGoodsDocumentHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input createIncomingGoodsInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if strings.TrimSpace(input.OrganizationRef) == "" || strings.TrimSpace(input.CounterpartyRef) == "" {
			return nil, fmt.Errorf("organization_ref and counterparty_ref are required")
		}
		if len(input.Items) == 0 {
			return nil, fmt.Errorf("items must contain at least one row")
		}
		body := onec.CreateIncomingGoodsRequest{
			OrganizationRef:     strings.TrimSpace(input.OrganizationRef),
			CounterpartyRef:     strings.TrimSpace(input.CounterpartyRef),
			ContractRef:         strings.TrimSpace(input.ContractRef),
			DocNumberIn:         strings.TrimSpace(input.DocNumberIn),
			DocDateIn:           strings.TrimSpace(input.DocDateIn),
			Amount:              strings.TrimSpace(input.Amount),
			VATIncluded:         input.VATIncluded,
			IsUniversalDocument: input.IsUniversalDocument,
			OriginalReceived:    input.OriginalReceived,
			Comment:             strings.TrimSpace(input.Comment),
			Post:                input.Post,
			Items:               input.Items,
		}
		var result onec.CreateIncomingGoodsResult
		if err := client.Post(ctx, "/incoming-goods", body, &result); err != nil {
			return nil, fmt.Errorf("creating incoming goods document in 1C: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("1C returned unsuccessful create result")
		}
		return textResult(formatCreateIncomingGoodsResult(&result)), nil
	}
}

func formatCreateIncomingGoodsResult(r *onec.CreateIncomingGoodsResult) string {
	doc := r.Document
	var b strings.Builder
	if r.Existing {
		b.WriteString("# Поступление уже было создано ранее (existing=true)\n\n")
	} else {
		b.WriteString("# Поступление создано\n\n")
	}
	fmt.Fprintf(&b, "- Ref: %s\n", doc.Ref)
	fmt.Fprintf(&b, "- Number: %s\n", doc.Number)
	if doc.Date != "" {
		fmt.Fprintf(&b, "- Date: %s\n", doc.Date)
	}
	if doc.Organization != "" {
		fmt.Fprintf(&b, "- Organization: %s\n", doc.Organization)
	}
	if doc.Counterparty != "" {
		fmt.Fprintf(&b, "- Counterparty: %s\n", doc.Counterparty)
	}
	if doc.Contract != "" {
		fmt.Fprintf(&b, "- Contract: %s\n", doc.Contract)
	}
	if doc.Amount != "" {
		fmt.Fprintf(&b, "- Amount: %s\n", doc.Amount)
	}
	if doc.DocNumberIn != "" {
		fmt.Fprintf(&b, "- Doc number (in): %s\n", doc.DocNumberIn)
	}
	if doc.DocDateIn != "" {
		fmt.Fprintf(&b, "- Doc date (in): %s\n", doc.DocDateIn)
	}
	fmt.Fprintf(&b, "- Posted: %t\n", doc.Posted)
	return b.String()
}
