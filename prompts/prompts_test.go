package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegisterAll(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	RegisterAll(s)

	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	_, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	result, err := session.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts error: %v", err)
	}

	// Derive expected names from the source of truth.
	expected := make(map[string]bool, len(developerPrompts))
	for _, p := range developerPrompts {
		expected[p.prompt.Name] = false
	}

	if len(result.Prompts) != len(expected) {
		names := make([]string, len(result.Prompts))
		for i, p := range result.Prompts {
			names[i] = p.Name
		}
		t.Fatalf("expected %d prompts, got %d: %v", len(expected), len(result.Prompts), names)
	}
	for _, p := range result.Prompts {
		if _, ok := expected[p.Name]; !ok {
			t.Errorf("unexpected prompt name: %s", p.Name)
		}
		expected[p.Name] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("prompt %q not found in ListPrompts result", name)
		}
	}
}

func TestRegisterBusiness_Buh30(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	RegisterBusiness(s, "buh_3_0")

	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	_, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	result, err := session.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts error: %v", err)
	}

	if len(result.Prompts) != len(buh30BusinessPrompts) {
		t.Fatalf("expected %d prompts, got %d", len(buh30BusinessPrompts), len(result.Prompts))
	}

	got := make(map[string]bool, len(result.Prompts))
	for _, p := range result.Prompts {
		got[p.Name] = true
	}
	for _, p := range buh30BusinessPrompts {
		if !got[p.prompt.Name] {
			t.Errorf("prompt %q not found in business buh_3_0 registration", p.prompt.Name)
		}
	}
}

func TestRegisterBusiness_UnknownProfile(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	RegisterBusiness(s, "generic")

	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	_, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	result, err := session.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts error: %v", err)
	}

	if len(result.Prompts) != 0 {
		t.Fatalf("expected 0 prompts for unknown business profile, got %d", len(result.Prompts))
	}
}

func TestRequiredArg_Missing(t *testing.T) {
	tests := []struct {
		name string
		req  *mcp.GetPromptRequest
	}{
		{
			name: "nil params",
			req:  &mcp.GetPromptRequest{},
		},
		{
			name: "nil arguments",
			req:  &mcp.GetPromptRequest{Params: &mcp.GetPromptParams{}},
		},
		{
			name: "empty value",
			req: &mcp.GetPromptRequest{Params: &mcp.GetPromptParams{
				Arguments: map[string]string{"other": "value"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := requiredArg(tt.req, "object_type")
			if err == nil {
				t.Fatal("expected error for missing argument")
			}
			if !strings.Contains(err.Error(), "object_type") {
				t.Errorf("error should mention argument name, got: %v", err)
			}
		})
	}
}

func TestPromptHandlers(t *testing.T) {
	tests := []struct {
		name        string
		handler     mcp.PromptHandler
		arguments   map[string]string
		wantKeyword string
	}{
		{
			name:    "review_module",
			handler: handleReviewModule,
			arguments: map[string]string{
				"object_type": "Document",
				"object_name": "ПриходнаяНакладная",
			},
			wantKeyword: "get_object_structure",
		},
		{
			name:    "write_posting",
			handler: handleWritePosting,
			arguments: map[string]string{
				"document_name": "РеализацияТоваровУслуг",
			},
			wantKeyword: "ОбработкаПроведения",
		},
		{
			name:    "optimize_query",
			handler: handleOptimizeQuery,
			arguments: map[string]string{
				"query": "ВЫБРАТЬ * ИЗ Справочник.Контрагенты",
			},
			wantKeyword: "execute_query",
		},
		{
			name:        "explain_config",
			handler:     handleExplainConfig,
			arguments:   map[string]string{},
			wantKeyword: "get_metadata_tree",
		},
		{
			name:    "analyze_error",
			handler: handleAnalyzeError,
			arguments: map[string]string{
				"error_text": "Поле не найдено \"Номенклатура\"",
			},
			wantKeyword: "search_code",
		},
		{
			name:    "find_duplicates",
			handler: handleFindDuplicates,
			arguments: map[string]string{
				"object_type": "Catalog",
				"object_name": "Контрагенты",
			},
			wantKeyword: "search_code",
		},
		{
			name:    "write_report",
			handler: handleWriteReport,
			arguments: map[string]string{
				"description": "Отчёт по продажам за период",
			},
			wantKeyword: "execute_query",
		},
		{
			name:    "explain_object",
			handler: handleExplainObject,
			arguments: map[string]string{
				"object_type": "AccumulationRegister",
				"object_name": "ТоварыНаСкладах",
			},
			wantKeyword: "get_object_structure",
		},
		{
			name:        "buh30_sales_health_audit",
			handler:     handleBuh30SalesHealthAudit,
			arguments:   map[string]string{},
			wantKeyword: "read_sales_invoices",
		},
		{
			name:        "buh30_cfo_cashflow_risk_snapshot",
			handler:     handleBuh30CFOCashflowRiskSnapshot,
			arguments:   map[string]string{},
			wantKeyword: "Резюме для финдиректора",
		},
		{
			name:        "buh30_revenue_leakage_watch",
			handler:     handleBuh30RevenueLeakageWatch,
			arguments:   map[string]string{},
			wantKeyword: "Каналы утечки",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &mcp.GetPromptRequest{
				Params: &mcp.GetPromptParams{
					Name:      tt.name,
					Arguments: tt.arguments,
				},
			}

			result, err := tt.handler(context.Background(), req)
			if err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if result.Description == "" {
				t.Error("expected non-empty Description")
			}

			if len(result.Messages) != 1 {
				t.Fatalf("expected 1 message, got %d", len(result.Messages))
			}

			msg := result.Messages[0]
			if msg.Role != "user" {
				t.Errorf("expected role \"user\", got %q", msg.Role)
			}

			tc, ok := msg.Content.(*mcp.TextContent)
			if !ok {
				t.Fatalf("expected *mcp.TextContent, got %T", msg.Content)
			}

			if tc.Text == "" {
				t.Error("expected non-empty text content")
			}

			if len(tc.Text) < 50 {
				t.Errorf("text content too short (%d chars), expected detailed instructions", len(tc.Text))
			}

			if !strings.Contains(tc.Text, tt.wantKeyword) {
				t.Errorf("text content does not contain expected keyword %q:\n%s", tt.wantKeyword, tc.Text)
			}
		})
	}
}
