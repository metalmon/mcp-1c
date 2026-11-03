package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestObjectStructureHandler(t *testing.T) {
	const mockResponse = `{
		"name": "РеализацияТоваровУслуг",
		"synonym": "Реализация товаров и услуг",
		"attributes": [
			{"name": "Контрагент", "synonym": "Контрагент", "type": "СправочникСсылка.Контрагенты"},
			{"name": "Сумма", "synonym": "Сумма документа", "type": "Число"}
		],
		"tabularParts": [
			{
				"name": "Товары",
				"attributes": [
					{"name": "Номенклатура", "synonym": "Номенклатура", "type": "СправочникСсылка.Номенклатура"},
					{"name": "Количество", "synonym": "Количество", "type": "Число"}
				]
			}
		]
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/object/Document/РеализацияТоваровУслуг" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewObjectStructureHandler(client)

	args, _ := json.Marshal(map[string]string{
		"object_type": "Document",
		"object_name": "РеализацияТоваровУслуг",
	})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "get_object_structure",
			Arguments: args,
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if tc.Text == "" {
		t.Fatal("expected non-empty text")
	}

	for _, want := range []string{
		"РеализацияТоваровУслуг",
		"Реализация товаров и услуг",
		"Контрагент",
		"Сумма",
		"Товары",
		"Номенклатура",
		"Количество",
	} {
		if !strings.Contains(tc.Text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, tc.Text)
		}
	}
}

func TestObjectStructureHandler_Register(t *testing.T) {
	const mockResponse = `{
		"name": "ТоварыНаСкладах",
		"synonym": "Товары на складах",
		"dimensions": [
			{"name": "Номенклатура", "synonym": "Номенклатура", "type": "СправочникСсылка.Номенклатура"},
			{"name": "Склад", "synonym": "Склад", "type": "СправочникСсылка.Склады"}
		],
		"resources": [
			{"name": "Количество", "synonym": "Количество", "type": "Число"}
		],
		"attributes": []
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/object/AccumulationRegister/ТоварыНаСкладах" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewObjectStructureHandler(client)

	args, _ := json.Marshal(map[string]string{
		"object_type": "AccumulationRegister",
		"object_name": "ТоварыНаСкладах",
	})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "get_object_structure",
			Arguments: args,
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tc := result.Content[0].(*mcp.TextContent)
	for _, want := range []string{
		"ТоварыНаСкладах", "Измерения", "Номенклатура", "Склад",
		"Ресурсы", "Количество",
	} {
		if !strings.Contains(tc.Text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, tc.Text)
		}
	}
}

func TestObjectStructureHandler_MissingArgs(t *testing.T) {
	client := onec.NewClient("http://localhost:0", "", "")
	handler := NewObjectStructureHandler(client)

	args, _ := json.Marshal(map[string]string{})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "get_object_structure",
			Arguments: args,
		},
	}

	_, err := handler(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

func TestObjectStructureTool(t *testing.T) {
	tool := ObjectStructureTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "get_object_structure" {
		t.Errorf("expected tool name %q, got %q", "get_object_structure", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

// Путь к точке 1С склеивался из того, что прислал агент, без проверки: тип
// `../event-log` уводил запрос на другую точку той же публикации — к тем GET-ам,
// которых бизнес-набор нарочно не даёт. Файловая сторона этого же инструмента
// (`dump.FindFormFiles`) так защищена с самого начала, сетевая — не была.
// Находка ревью 3.11.2026.
func TestObjectStructureHandlerRefusesPathCharacters(t *testing.T) {
	var hits int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewObjectStructureHandler(client)

	for _, bad := range []struct{ objectType, objectName string }{
		{"../event-log", "X"},
		{"Document", "../../etc/passwd"},
		{"Document", "Реализация/Товаров"},
		{"Document", "X%2fY"},
	} {
		args, err := json.Marshal(map[string]string{"object_type": bad.objectType, "object_name": bad.objectName})
		if err != nil {
			t.Fatal(err)
		}
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: args}}
		if _, err := handler(context.Background(), req); err == nil {
			t.Errorf("тип %q и имя %q обязаны быть отвергнуты", bad.objectType, bad.objectName)
		}
	}
	if hits != 0 {
		t.Errorf("отвергнутый запрос не должен уходить на сервер, ушло %d", hits)
	}
}
