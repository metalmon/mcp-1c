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
	defaultCounterpartiesLimit = 50
	maxCounterpartiesLimit     = 500
)

type readCounterpartiesInput struct {
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Code   string `json:"code,omitempty"`
	Ref    string `json:"ref,omitempty"`
	INN    string `json:"inn,omitempty"`
	KPP    string `json:"kpp,omitempty"`
}

type createCounterpartyInput struct {
	Name             string `json:"name"`
	INN              string `json:"inn"`
	KPP              string `json:"kpp"`
	CounterpartyType string `json:"counterparty_type"`
}

// ReadCounterpartiesTool returns the MCP tool definition for read_counterparties.
func ReadCounterpartiesTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "read_counterparties",
		Title: "Чтение контрагентов",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Description: "Получить контрагентов из справочника: список с поиском по наименованию/ИНН " +
			"или один элемент по code/ref.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"search":{"type":"string","description":"Поиск по наименованию или ИНН"},
				"limit":{"type":"integer","description":"Максимум строк (по умолчанию 50, максимум 500)"},
				"code":{"type":"string","description":"Код контрагента для точечного чтения"},
				"ref":{"type":"string","description":"Ссылка контрагента (строковый UUID) для точечного чтения"},
				"inn":{"type":"string","description":"ИНН для точечного чтения (в паре с kpp)"},
				"kpp":{"type":"string","description":"КПП для точечного чтения (в паре с inn)"}
			}
		}`),
	}
}

// NewReadCounterpartiesHandler returns a ToolHandler that reads counterparties via query endpoint.
func NewReadCounterpartiesHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input readCounterpartiesInput
		if req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
				return nil, fmt.Errorf("parsing input: %w", err)
			}
		}

		limit := clampLimit(input.Limit, defaultCounterpartiesLimit, maxCounterpartiesLimit)

		body := onec.ReadCounterpartiesRequest{
			Search: strings.TrimSpace(input.Search),
			Limit:  limit,
			Code:   strings.TrimSpace(input.Code),
			Ref:    strings.TrimSpace(input.Ref),
			INN:    strings.TrimSpace(input.INN),
			KPP:    strings.TrimSpace(input.KPP),
		}
		var result onec.ReadCounterpartiesResult
		if err := client.Post(ctx, "/counterparties", body, &result); err != nil {
			return nil, fmt.Errorf("reading counterparties from 1C: %w", err)
		}

		return textResult(formatCounterpartiesReadResult(&result)), nil
	}
}

// CreateCounterpartyTool returns the MCP tool definition for create_counterparty.
func CreateCounterpartyTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "create_counterparty",
		Title: "Создание контрагента",
		Description: "Создать контрагента с обязательными полями: Наименование, ИНН, КПП, Вид контрагента. " +
			"counterparty_type: legal|individual.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"name":{"type":"string","description":"Наименование"},
				"inn":{"type":"string","description":"ИНН"},
				"kpp":{"type":"string","description":"КПП"},
				"counterparty_type":{"type":"string","description":"Вид: legal или individual"}
			},
			"required":["name","inn","kpp","counterparty_type"]
		}`),
	}
}

// NewCreateCounterpartyHandler returns a ToolHandler that creates a counterparty.
func NewCreateCounterpartyHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input createCounterpartyInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if strings.TrimSpace(input.Name) == "" ||
			strings.TrimSpace(input.INN) == "" ||
			strings.TrimSpace(input.KPP) == "" ||
			strings.TrimSpace(input.CounterpartyType) == "" {
			return nil, fmt.Errorf("name, inn, kpp, counterparty_type are required")
		}

		body := onec.CreateCounterpartyRequest{
			Name:             strings.TrimSpace(input.Name),
			INN:              strings.TrimSpace(input.INN),
			KPP:              strings.TrimSpace(input.KPP),
			CounterpartyType: strings.TrimSpace(input.CounterpartyType),
		}
		var result onec.CreateCounterpartyResult
		if err := client.Post(ctx, "/counterparty", body, &result); err != nil {
			return nil, fmt.Errorf("creating counterparty in 1C: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("1C returned unsuccessful create result")
		}

		return textResult(formatCreateCounterpartyResult(&result)), nil
	}
}

func formatCounterpartiesReadResult(r *onec.ReadCounterpartiesResult) string {
	var b strings.Builder
	b.WriteString("# Контрагенты\n\n")
	if len(r.Counterparties) == 0 {
		b.WriteString("Ничего не найдено.\n")
		return b.String()
	}

	b.WriteString("| Ref | Code | Name | INN | KPP | Type |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, cp := range r.Counterparties {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
			cp.Ref, cp.Code, cp.Name, cp.INN, cp.KPP, cp.CounterpartyType)
	}
	if r.Truncated {
		b.WriteString("\n> Показаны не все записи. Увеличьте limit.\n")
	}
	return b.String()
}

func formatCreateCounterpartyResult(r *onec.CreateCounterpartyResult) string {
	cp := r.Counterparty
	var b strings.Builder
	b.WriteString("# Контрагент создан\n\n")
	fmt.Fprintf(&b, "- Ref: %s\n", cp.Ref)
	fmt.Fprintf(&b, "- Code: %s\n", cp.Code)
	fmt.Fprintf(&b, "- Name: %s\n", cp.Name)
	fmt.Fprintf(&b, "- INN: %s\n", cp.INN)
	fmt.Fprintf(&b, "- KPP: %s\n", cp.KPP)
	if cp.CounterpartyType != "" {
		fmt.Fprintf(&b, "- Type: %s\n", cp.CounterpartyType)
	}
	return b.String()
}
