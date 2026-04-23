package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createNomenclatureInput struct {
	Name             string `json:"name"`
	FullName         string `json:"full_name,omitempty"`
	Article          string `json:"article,omitempty"`
	NomenclatureType string `json:"nomenclature_type,omitempty"`
	Unit             string `json:"unit,omitempty"`
	IsService        *bool  `json:"is_service,omitempty"`
}

// CreateNomenclatureTool returns the MCP tool definition for create_nomenclature.
func CreateNomenclatureTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "create_nomenclature",
		Title: "Создание номенклатуры",
		Description: "Инструмент 1С: создать элемент номенклатуры. Обязательно: name. " +
			"Опционально: full_name, article, nomenclature_type, unit, is_service.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"name":{"type":"string","description":"Наименование номенклатуры"},
				"full_name":{"type":"string","description":"Полное наименование"},
				"article":{"type":"string","description":"Артикул"},
				"nomenclature_type":{"type":"string","description":"Вид номенклатуры (код или наименование)"},
				"unit":{"type":"string","description":"Единица измерения (код или наименование)"},
				"is_service":{"type":"boolean","description":"Признак услуги"}
			},
			"required":["name"]
		}`),
	}
}

// NewCreateNomenclatureHandler returns a ToolHandler that creates nomenclature.
func NewCreateNomenclatureHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input createNomenclatureInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if strings.TrimSpace(input.Name) == "" {
			return nil, fmt.Errorf("name is required")
		}

		body := onec.CreateNomenclatureRequest{
			Name:             strings.TrimSpace(input.Name),
			FullName:         strings.TrimSpace(input.FullName),
			Article:          strings.TrimSpace(input.Article),
			NomenclatureType: strings.TrimSpace(input.NomenclatureType),
			Unit:             strings.TrimSpace(input.Unit),
			IsService:        input.IsService,
		}
		var result onec.CreateNomenclatureResult
		if err := client.Post(ctx, "/nomenclature", body, &result); err != nil {
			return nil, fmt.Errorf("creating nomenclature in 1C: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("1C returned unsuccessful create result")
		}

		return textResult(formatCreateNomenclatureResult(&result)), nil
	}
}

func formatCreateNomenclatureResult(r *onec.CreateNomenclatureResult) string {
	n := r.Nomenclature
	var b strings.Builder
	b.WriteString("# Номенклатура создана\n\n")
	fmt.Fprintf(&b, "- Ref: %s\n", n.Ref)
	fmt.Fprintf(&b, "- Code: %s\n", n.Code)
	fmt.Fprintf(&b, "- Name: %s\n", n.Name)
	if n.FullName != "" {
		fmt.Fprintf(&b, "- Full name: %s\n", n.FullName)
	}
	if n.Article != "" {
		fmt.Fprintf(&b, "- Article: %s\n", n.Article)
	}
	if n.NomenclatureType != "" {
		fmt.Fprintf(&b, "- Type: %s\n", n.NomenclatureType)
	}
	if n.Unit != "" {
		fmt.Fprintf(&b, "- Unit: %s\n", n.Unit)
	}
	fmt.Fprintf(&b, "- Is service: %t\n", n.IsService)
	return b.String()
}
