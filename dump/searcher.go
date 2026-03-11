package dump

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Match represents a single search hit in a BSL module.
type Match struct {
	Module  string  // Human-readable module path (e.g. "Документ.РеализацияТоваров.МодульОбъекта")
	Line    int     // 1-based line number of the match
	Context string  // Surrounding lines for context
	Score   float64 // BM25 relevance score (smart mode only)
}

// module holds the content and display name of a loaded BSL module.
type module struct {
	name    string // e.g. "Справочник.Номенклатура.МодульОбъекта"
	content string
}

// Searcher provides full-text search over BSL modules loaded from a DumpConfigToFiles directory.
type Searcher struct {
	dir     string
	modules []module
}

// NewSearcher walks the dump directory, loads all .bsl files and returns a Searcher.
func NewSearcher(dir string) (*Searcher, error) {
	s := &Searcher{dir: dir}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".bsl") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		rel, _ := filepath.Rel(dir, path)
		name := bslPathToModuleName(rel)
		s.modules = append(s.modules, module{name: name, content: string(data)})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

// ModuleCount returns the number of loaded BSL modules.
func (s *Searcher) ModuleCount() int {
	return len(s.modules)
}

// Dir returns the dump directory path.
func (s *Searcher) Dir() string {
	return s.dir
}

// Search finds all occurrences of query (case-insensitive) across loaded modules.
// It returns up to limit matches and the total count.
func (s *Searcher) Search(query string, limit int) ([]Match, int) {
	lower := strings.ToLower(query)
	var matches []Match
	total := 0

	for _, m := range s.modules {
		contentLower := strings.ToLower(m.content)
		scanner := bufio.NewScanner(strings.NewReader(m.content))
		scannerLower := bufio.NewScanner(strings.NewReader(contentLower))
		lineNum := 0

		lines := strings.Split(m.content, "\n")

		for scanner.Scan() {
			scannerLower.Scan()
			lineNum++
			if strings.Contains(scannerLower.Text(), lower) {
				total++
				if len(matches) < limit {
					ctx := extractContext(lines, lineNum-1, 2)
					matches = append(matches, Match{
						Module:  m.name,
						Line:    lineNum,
						Context: ctx,
					})
				}
			}
		}
	}

	return matches, total
}

// extractContext returns lines around the given index with a context window.
func extractContext(lines []string, idx, window int) string {
	start := idx - window
	if start < 0 {
		start = 0
	}
	end := idx + window + 1
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n")
}

// dumpDirNames maps 1C metadata type names (English) used in DumpConfigToFiles
// to their Russian display name prefixes.
var dumpDirNames = map[string]string{
	"Catalogs":               "Справочник",
	"Documents":              "Документ",
	"DataProcessors":         "Обработка",
	"Reports":                "Отчет",
	"InformationRegisters":   "РегистрСведений",
	"AccumulationRegisters":  "РегистрНакопления",
	"AccountingRegisters":    "РегистрБухгалтерии",
	"CalculationRegisters":   "РегистрРасчета",
	"ChartsOfAccounts":       "ПланСчетов",
	"ChartsOfCharacteristicTypes": "ПланВидовХарактеристик",
	"ChartsOfCalculationTypes":    "ПланВидовРасчета",
	"ExchangePlans":          "ПланОбмена",
	"BusinessProcesses":      "БизнесПроцесс",
	"Tasks":                  "Задача",
	"CommonModules":          "ОбщийМодуль",
	"Enums":                  "Перечисление",
	"Constants":              "Константа",
}

// moduleNameSuffixes maps BSL file names to their module type suffix.
var moduleNameSuffixes = map[string]string{
	"ObjectModule.bsl":    "МодульОбъекта",
	"ManagerModule.bsl":   "МодульМенеджера",
	"Module.bsl":          "МодульФормы",
	"RecordSetModule.bsl": "МодульНабораЗаписей",
	"CommandModule.bsl":   "МодульКоманды",
	"Ext.Module.bsl":      "МодульФормы",
}

// bslPathToModuleName converts a relative file path from the dump to a human-readable module name.
// Example: "Documents/РеализацияТоваров/Ext/ObjectModule.bsl" -> "Документ.РеализацияТоваров.МодульОбъекта"
func bslPathToModuleName(relPath string) string {
	// Normalise separators.
	relPath = filepath.ToSlash(relPath)
	parts := strings.Split(relPath, "/")

	if len(parts) < 2 {
		return relPath
	}

	// First part is the category directory.
	category := parts[0]
	prefix, ok := dumpDirNames[category]
	if !ok {
		prefix = category
	}

	objectName := parts[1]

	// Determine suffix from the file name.
	fileName := parts[len(parts)-1]
	suffix, ok := moduleNameSuffixes[fileName]
	if !ok {
		suffix = strings.TrimSuffix(fileName, ".bsl")
	}

	// Fix: CommonModules use "Модуль", not "МодульФормы" for Module.bsl.
	if category == "CommonModules" && fileName == "Module.bsl" {
		inForms := false
		for _, p := range parts {
			if p == "Forms" {
				inForms = true
				break
			}
		}
		if !inForms {
			suffix = "Модуль"
		}
	}

	// If the path has a Forms subdirectory, include form name.
	for i, p := range parts {
		if p == "Forms" && i+1 < len(parts) {
			formName := parts[i+1]
			return prefix + "." + objectName + ".Форма." + formName + "." + suffix
		}
	}

	return prefix + "." + objectName + "." + suffix
}
