package dump

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

// SearchMode determines the search strategy.
type SearchMode string

const (
	SearchModeSmart SearchMode = "smart"
	SearchModeRegex SearchMode = "regex"
	SearchModeExact SearchMode = "exact"
)

// SearchParams holds all parameters for a search query.
type SearchParams struct {
	Query    string
	Category string     // filter by metadata type, empty = all
	Module   string     // filter by module type, empty = all
	Mode     SearchMode // default: SearchModeSmart
	Limit    int        // default: 50, max: 500
}

// bslDocument is the struct indexed by Bleve. Field names must match mapping keys.
// Implements mapping.Classifier so Bleve routes it to the "module" document mapping.
type bslDocument struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Module   string `json:"module"`
	Content  string `json:"content"`
}

func (bslDocument) Type() string { return "module" }

// Index provides full-text search over BSL modules using Bleve.
type Index struct {
	dir           string
	index         bleve.Index
	names         []string            // module names in load order
	contentByName map[string]string   // single source of truth
}

// NewIndex walks the dump directory, loads all .bsl files and builds an in-memory
// Bleve index. Progress is printed to stderr.
func NewIndex(dir string) (*Index, error) {
	idx := &Index{
		dir:           dir,
		contentByName: make(map[string]string),
	}

	// Collect all .bsl files.
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

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("relative path for %s: %w", path, err)
		}
		name := bslPathToModuleName(rel)
		idx.names = append(idx.names, name)
		idx.contentByName[name] = string(data)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking dump directory: %w", err)
	}

	// Build Bleve in-memory index.
	bslMapping := buildBSLMapping()
	blevIdx, err := bleve.NewMemOnly(bslMapping)
	if err != nil {
		return nil, fmt.Errorf("creating bleve index: %w", err)
	}
	idx.index = blevIdx

	total := len(idx.names)
	for i, name := range idx.names {
		parts := parseModuleName(name)

		doc := bslDocument{
			Name:     parts.name,
			Category: parts.category,
			Module:   parts.module,
			Content:  idx.contentByName[name],
		}

		if err := blevIdx.Index(name, doc); err != nil {
			blevIdx.Close()
			return nil, fmt.Errorf("indexing module %q: %w", name, err)
		}

		// Print progress every 500 modules.
		if (i+1)%500 == 0 || i+1 == total {
			fmt.Fprintf(os.Stderr, "\rIndexing BSL modules... %d/%d", i+1, total)
		}
	}
	if total > 0 {
		fmt.Fprintln(os.Stderr, " done")
	}

	return idx, nil
}

// moduleNameParts holds the parsed components of a human-readable module name.
type moduleNameParts struct {
	category string // e.g. "Справочник"
	name     string // e.g. "Номенклатура"
	module   string // e.g. "МодульОбъекта"
}

// parseModuleName splits "Справочник.Номенклатура.МодульОбъекта" into parts.
// For form paths like "Документ.Док.Форма.ФормаДок.МодульФормы", the module type
// is the last dot-separated segment ("МодульФормы"), not the third segment.
func parseModuleName(fullName string) moduleNameParts {
	parts := strings.Split(fullName, ".")
	switch {
	case len(parts) >= 3:
		return moduleNameParts{
			category: parts[0],
			name:     parts[1],
			module:   parts[len(parts)-1],
		}
	case len(parts) == 2:
		return moduleNameParts{category: parts[0], name: parts[1]}
	default:
		return moduleNameParts{name: fullName}
	}
}

// Search finds matches in indexed BSL modules. Dispatches by mode.
func (idx *Index) Search(params SearchParams) ([]Match, int, error) {
	if params.Mode == "" {
		params.Mode = SearchModeSmart
	}
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 500 {
		params.Limit = 500
	}

	switch params.Mode {
	case SearchModeSmart:
		return idx.searchSmart(params)
	case SearchModeRegex:
		re, err := regexp.Compile(params.Query)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid regex %q: %w", params.Query, err)
		}
		return idx.searchLineByLine(params, func(line, _ string) bool {
			return re.MatchString(line)
		}, params.Query)
	case SearchModeExact:
		lower := strings.ToLower(params.Query)
		return idx.searchLineByLine(params, func(line, q string) bool {
			return strings.Contains(strings.ToLower(line), q)
		}, lower)
	default:
		return nil, 0, fmt.Errorf("unknown search mode: %q", params.Mode)
	}
}

// searchSmart performs full-text BM25 search via Bleve.
func (idx *Index) searchSmart(params SearchParams) ([]Match, int, error) {
	mq := bleve.NewMatchQuery(params.Query)
	mq.SetField("content")
	mq.Analyzer = analyzerBSL

	var q query.Query = mq

	// Apply category/module filters as conjunction.
	if params.Category != "" || params.Module != "" {
		queries := []query.Query{mq}
		if params.Category != "" {
			tq := bleve.NewTermQuery(params.Category)
			tq.SetField("category")
			queries = append(queries, tq)
		}
		if params.Module != "" {
			tq := bleve.NewTermQuery(params.Module)
			tq.SetField("module")
			queries = append(queries, tq)
		}
		q = bleve.NewConjunctionQuery(queries...)
	}

	req := bleve.NewSearchRequestOptions(q, params.Limit, 0, false)
	result, err := idx.index.Search(req)
	if err != nil {
		return nil, 0, fmt.Errorf("bleve search: %w", err)
	}

	var matches []Match
	for _, hit := range result.Hits {
		content, ok := idx.contentByName[hit.ID]
		if !ok {
			continue
		}
		// Find the first line containing any query term for context.
		lines := strings.Split(content, "\n")
		lower := strings.ToLower(params.Query)
		tokens := strings.Fields(lower)

		lineNum := 0
		for i, line := range lines {
			ll := strings.ToLower(line)
			for _, tok := range tokens {
				if strings.Contains(ll, tok) {
					lineNum = i + 1
					break
				}
			}
			if lineNum > 0 {
				break
			}
		}
		if lineNum == 0 {
			lineNum = 1
		}

		ctx := extractContext(lines, lineNum-1, 2)
		matches = append(matches, Match{
			Module:  hit.ID,
			Line:    lineNum,
			Context: ctx,
			Score:   hit.Score,
		})
	}

	return matches, int(result.Total), nil
}

// searchLineByLine performs line-by-line search using a matcher function.
// Used for regex and exact modes. Optionally pre-filters modules via Bleve.
func (idx *Index) searchLineByLine(params SearchParams, match func(line, q string) bool, q string) ([]Match, int, error) {
	candidates, err := idx.filterModules(params.Category, params.Module)
	if err != nil {
		return nil, 0, err
	}

	var matches []Match
	total := 0

	for _, name := range candidates {
		content := idx.contentByName[name]
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if match(line, q) {
				total++
				if len(matches) < params.Limit {
					ctx := extractContext(lines, i, 2)
					matches = append(matches, Match{
						Module:  name,
						Line:    i + 1,
						Context: ctx,
					})
				}
			}
		}
	}

	return matches, total, nil
}

// filterModules returns the subset of module names matching category/module filters.
// If no filters are set, returns all names. Uses Bleve for efficient filtering.
func (idx *Index) filterModules(category, moduleType string) ([]string, error) {
	if category == "" && moduleType == "" {
		return idx.names, nil
	}

	// Use Bleve to filter by category/module.
	queries := []query.Query{bleve.NewMatchAllQuery()}
	if category != "" {
		tq := bleve.NewTermQuery(category)
		tq.SetField("category")
		queries = append(queries, tq)
	}
	if moduleType != "" {
		tq := bleve.NewTermQuery(moduleType)
		tq.SetField("module")
		queries = append(queries, tq)
	}

	q := bleve.NewConjunctionQuery(queries...)
	req := bleve.NewSearchRequestOptions(q, len(idx.names), 0, false)
	result, err := idx.index.Search(req)
	if err != nil {
		return nil, fmt.Errorf("bleve filter: %w", err)
	}

	names := make([]string, 0, len(result.Hits))
	for _, hit := range result.Hits {
		if _, ok := idx.contentByName[hit.ID]; ok {
			names = append(names, hit.ID)
		}
	}
	return names, nil
}

// ModuleCount returns the number of indexed BSL modules.
func (idx *Index) ModuleCount() int {
	return len(idx.names)
}

// Dir returns the dump directory path.
func (idx *Index) Dir() string {
	return idx.dir
}

// Close releases the Bleve index resources.
func (idx *Index) Close() error {
	return idx.index.Close()
}
