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
	defaultOrganizationsLimit = 50
	maxOrganizationsLimit     = 500
)

type readOrganizationsInput struct {
	Search    string `json:"search,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Code      string `json:"code,omitempty"`
	Ref       string `json:"ref,omitempty"`
	INN       string `json:"inn,omitempty"`
	KPP       string `json:"kpp,omitempty"`
	AfterCode string `json:"after_code,omitempty"`
}

// ReadOrganizationsTool returns the MCP tool definition for read_organizations.
func ReadOrganizationsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "read_organizations",
		Title: "Чтение организаций",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Description: "Инструмент 1С: получить организации: список с поиском по наименованию/ИНН " +
			"(в списке возвращаются ИНН и КПП, если заполнены) или один элемент по code/ref/inn+kpp.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"search":{"type":"string","description":"Подстрочный поиск по краткому и полному наименованию или ИНН (регистр не важен; «» в наименовании игнорируются при сравнении)"},
				"limit":{"type":"integer","description":"Максимум строк (по умолчанию 50, максимум 500)"},
				"code":{"type":"string","description":"Код организации для точечного чтения"},
				"ref":{"type":"string","description":"Типизированная ссылка организации (Справочник.Организации:<uuid>) для точечного чтения"},
				"inn":{"type":"string","description":"ИНН для точечного чтения в паре с kpp (точное совпадение; подстрочный поиск по номеру — через search)"},
				"kpp":{"type":"string","description":"КПП для точечного чтения в паре с inn (точное совпадение)"},
				"after_code":{"type":"string","description":"Курсор постраничного обхода: код последней организации предыдущей страницы (next_cursor). Для следующей страницы повторите запрос с тем же limit и этим значением."}
			}
		}`),
	}
}

// NewReadOrganizationsHandler returns a ToolHandler that reads organizations.
func NewReadOrganizationsHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input readOrganizationsInput
		if req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
				return nil, fmt.Errorf("parsing input: %w", err)
			}
		}

		limit := clampLimit(input.Limit, defaultOrganizationsLimit, maxOrganizationsLimit)
		body := onec.ReadOrganizationsRequest{
			Search:    strings.TrimSpace(input.Search),
			Limit:     limit,
			Code:      strings.TrimSpace(input.Code),
			Ref:       strings.TrimSpace(input.Ref),
			INN:       strings.TrimSpace(input.INN),
			KPP:       strings.TrimSpace(input.KPP),
			AfterCode: strings.TrimSpace(input.AfterCode),
		}

		var result onec.ReadOrganizationsResult
		if err := client.Post(ctx, "/organizations", body, &result); err != nil {
			return nil, fmt.Errorf("reading organizations from 1C: %w", err)
		}
		return textResult(formatOrganizationsReadResult(&result)), nil
	}
}

func formatOrganizationsReadResult(r *onec.ReadOrganizationsResult) string {
	var b strings.Builder
	b.WriteString("# Организации\n\n")
	if len(r.Organizations) == 0 {
		b.WriteString("Ничего не найдено.\n")
		return b.String()
	}

	b.WriteString("| Ref | Code | Name | INN | KPP |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, org := range r.Organizations {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
			org.Ref, org.Code, org.Name, org.INN, org.KPP)
	}
	if r.Truncated && r.NextCursor != "" {
		fmt.Fprintf(&b, "\n> Показаны не все записи. Следующая страница: after_code=%q (с тем же limit).\n", r.NextCursor)
	} else if r.Truncated {
		b.WriteString("\n> Показаны не все записи. Увеличьте limit.\n")
	}
	return b.String()
}

