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
	defaultContractsLimit = 50
	maxContractsLimit     = 500
)

type readContractsInput struct {
	Search          string `json:"search,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Code            string `json:"code,omitempty"`
	Ref             string `json:"ref,omitempty"`
	Number          string `json:"number,omitempty"`
	CounterpartyRef string `json:"counterparty_ref,omitempty"`
	OrganizationRef string `json:"organization_ref,omitempty"`
}

// ReadContractsTool returns the MCP tool definition for read_contracts.
func ReadContractsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "read_contracts",
		Title: "Чтение договоров",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Description: "Инструмент 1С: получить договоры контрагентов: список с поиском по наименованию " +
			"или один элемент по code/ref/number.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"search":{"type":"string","description":"Поиск по наименованию"},
				"limit":{"type":"integer","description":"Максимум строк (по умолчанию 50, максимум 500)"},
				"code":{"type":"string","description":"Код договора для точечного чтения"},
				"ref":{"type":"string","description":"Ссылка договора (UUID) для точечного чтения"},
				"number":{"type":"string","description":"Номер договора для точечного чтения"},
				"counterparty_ref":{"type":"string","description":"Фильтр по ссылке контрагента (UUID)"},
				"organization_ref":{"type":"string","description":"Фильтр по ссылке организации (UUID)"}
			}
		}`),
	}
}

// NewReadContractsHandler returns a ToolHandler that reads contracts.
func NewReadContractsHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input readContractsInput
		if req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
				return nil, fmt.Errorf("parsing input: %w", err)
			}
		}

		limit := clampLimit(input.Limit, defaultContractsLimit, maxContractsLimit)
		body := onec.ReadContractsRequest{
			Search:          strings.TrimSpace(input.Search),
			Limit:           limit,
			Code:            strings.TrimSpace(input.Code),
			Ref:             strings.TrimSpace(input.Ref),
			Number:          strings.TrimSpace(input.Number),
			CounterpartyRef: strings.TrimSpace(input.CounterpartyRef),
			OrganizationRef: strings.TrimSpace(input.OrganizationRef),
		}

		var result onec.ReadContractsResult
		if err := client.Post(ctx, "/contracts", body, &result); err != nil {
			return nil, fmt.Errorf("reading contracts from 1C: %w", err)
		}
		return textResult(formatContractsReadResult(&result)), nil
	}
}

func formatContractsReadResult(r *onec.ReadContractsResult) string {
	var b strings.Builder
	b.WriteString("# Договоры\n\n")
	if len(r.Contracts) == 0 {
		b.WriteString("Ничего не найдено.\n")
		return b.String()
	}

	b.WriteString("| Ref | Code | Number | Name | Counterparty | Organization | Currency |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, c := range r.Contracts {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n",
			c.Ref, c.Code, c.Number, c.Name, c.Counterparty, c.Organization, c.Currency)
	}
	if r.Truncated {
		b.WriteString("\n> Показаны не все записи. Увеличьте limit.\n")
	}
	return b.String()
}

