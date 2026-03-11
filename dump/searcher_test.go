package dump

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSearcher(t *testing.T) {
	dir := t.TempDir()
	mkBSL(t, dir, "Catalogs/Номенклатура/Ext/ObjectModule.bsl",
		"Процедура ПередЗаписью(Отказ)\n\t// проверка\nКонецПроцедуры\n")
	mkBSL(t, dir, "Documents/Реализация/Ext/ObjectModule.bsl",
		"Процедура ОбработкаПроведения(Отказ)\n\t// проведение\nКонецПроцедуры\n")

	s, err := NewSearcher(dir)
	if err != nil {
		t.Fatalf("NewSearcher: %v", err)
	}

	if s.ModuleCount() != 2 {
		t.Errorf("expected 2 modules, got %d", s.ModuleCount())
	}

	if s.Dir() != dir {
		t.Errorf("expected dir %q, got %q", dir, s.Dir())
	}
}

func TestSearcher_Search(t *testing.T) {
	dir := t.TempDir()
	mkBSL(t, dir, "Catalogs/Номенклатура/Ext/ObjectModule.bsl",
		"Строка1\nПроцедура ОбновитьЦены()\n\t// обновление цен\nКонецПроцедуры\nСтрока5\n")

	s, err := NewSearcher(dir)
	if err != nil {
		t.Fatalf("NewSearcher: %v", err)
	}

	matches, total := s.Search("ОбновитьЦены", 50)

	if total != 1 {
		t.Errorf("expected 1 match, got %d", total)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match result, got %d", len(matches))
	}
	if matches[0].Line != 2 {
		t.Errorf("expected line 2, got %d", matches[0].Line)
	}
	if !strings.Contains(matches[0].Module, "Справочник.Номенклатура.МодульОбъекта") {
		t.Errorf("expected module name to contain 'Справочник.Номенклатура.МодульОбъекта', got %q", matches[0].Module)
	}
}

func TestSearcher_SearchCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	mkBSL(t, dir, "Catalogs/Тест/Ext/ObjectModule.bsl",
		"ПРОЦЕДУРА Тестирование()\nКонецПроцедуры\n")

	s, err := NewSearcher(dir)
	if err != nil {
		t.Fatalf("NewSearcher: %v", err)
	}

	matches, total := s.Search("процедура", 50)
	if total != 1 {
		t.Errorf("expected 1 case-insensitive match, got %d", total)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestSearcher_SearchLimit(t *testing.T) {
	dir := t.TempDir()
	mkBSL(t, dir, "Catalogs/Тест/Ext/ObjectModule.bsl",
		"Строка1\nСтрока2\nСтрока3\nСтрока4\nСтрока5\n")

	s, err := NewSearcher(dir)
	if err != nil {
		t.Fatalf("NewSearcher: %v", err)
	}

	matches, total := s.Search("Строка", 2)
	if total != 5 {
		t.Errorf("expected 5 total matches, got %d", total)
	}
	if len(matches) != 2 {
		t.Errorf("expected 2 matches (limited), got %d", len(matches))
	}
}

func TestBslPathToModuleName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"Catalogs/Номенклатура/Ext/ObjectModule.bsl", "Справочник.Номенклатура.МодульОбъекта"},
		{"Documents/Реализация/Ext/ObjectModule.bsl", "Документ.Реализация.МодульОбъекта"},
		{"DataProcessors/Обработка1/Ext/ObjectModule.bsl", "Обработка.Обработка1.МодульОбъекта"},
		{"Documents/Док/Forms/ФормаДок/Ext/Module.bsl", "Документ.Док.Форма.ФормаДок.МодульФормы"},
		{"CommonModules/ОбщийМодуль1/Ext/Module.bsl", "ОбщийМодуль.ОбщийМодуль1.Модуль"},
	}

	for _, tt := range tests {
		got := bslPathToModuleName(tt.path)
		if got != tt.want {
			t.Errorf("bslPathToModuleName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func mkBSL(t *testing.T, base, relPath, content string) {
	t.Helper()
	full := filepath.Join(base, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
