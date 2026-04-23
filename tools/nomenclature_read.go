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
	defaultNomenclatureLimit = 50
	maxNomenclatureLimit     = 500
)

type readNomenclatureInput struct {
	Search  string `json:"search,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Code    string `json:"code,omitempty"`
	Ref     string `json:"ref,omitempty"`
	Article string `json:"article,omitempty"`
}

// ReadNomenclatureTool returns the MCP tool definition for read_nomenclature.
func ReadNomenclatureTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "read_nomenclature",
		Title: "Чтение номенклатуры",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Description: "Инструмент 1С: получить номенклатуру: список с поиском по наименованию/артикулу " +
			"или один элемент по code/ref/article.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"search":{"type":"string","description":"Поиск по наименованию или артикулу"},
				"limit":{"type":"integer","description":"Максимум строк (по умолчанию 50, максимум 500)"},
				"code":{"type":"string","description":"Код номенклатуры для точечного чтения"},
				"ref":{"type":"string","description":"Ссылка номенклатуры (строковый UUID) для точечного чтения"},
				"article":{"type":"string","description":"Артикул номенклатуры для точечного чтения"}
			}
		}`),
	}
}

// NewReadNomenclatureHandler returns a ToolHandler that reads nomenclature.
func NewReadNomenclatureHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input readNomenclatureInput
		if req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
				return nil, fmt.Errorf("parsing input: %w", err)
			}
		}

		limit := clampLimit(input.Limit, defaultNomenclatureLimit, maxNomenclatureLimit)
		body := onec.ReadNomenclatureRequest{
			Search:  strings.TrimSpace(input.Search),
			Limit:   limit,
			Code:    strings.TrimSpace(input.Code),
			Ref:     strings.TrimSpace(input.Ref),
			Article: strings.TrimSpace(input.Article),
		}

		var result onec.ReadNomenclatureResult
		if err := client.Post(ctx, "/nomenclatures", body, &result); err != nil {
			return nil, fmt.Errorf("reading nomenclature from 1C: %w", err)
		}

		return textResult(formatNomenclatureReadResult(&result)), nil
	}
}

func formatNomenclatureReadResult(r *onec.ReadNomenclatureResult) string {
	var b strings.Builder
	b.WriteString("# Номенклатура\n\n")
	if len(r.Nomenclature) == 0 {
		b.WriteString("Ничего не найдено.\n")
		return b.String()
	}

	b.WriteString("| Ref | Code | Name | Article | Type | Unit | Is service |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, item := range r.Nomenclature {
		fmt.Fprintf(
			&b,
			"| %s | %s | %s | %s | %s | %s | %t |\n",
			item.Ref,
			item.Code,
			item.Name,
			item.Article,
			item.NomenclatureType,
			item.Unit,
			item.IsService,
		)
	}
	if r.Truncated {
		b.WriteString("\n> Показаны не все записи. Увеличьте limit.\n")
	}
	return b.String()
}
