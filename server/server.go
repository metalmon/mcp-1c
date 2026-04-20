package server

import (
	"github.com/feenlace/mcp-1c/dump"
	"github.com/feenlace/mcp-1c/internal/profile"
	"github.com/feenlace/mcp-1c/onec"
	"github.com/feenlace/mcp-1c/prompts"
	"github.com/feenlace/mcp-1c/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New creates an MCP server with basic configuration and registers tools.
// If dumpIndex is provided, the search_code tool will be registered.
func New(version string, onecClient *onec.Client, dumpIndex *dump.Index, options ...Options) *mcp.Server {
	cfg := defaultOptions()
	if len(options) > 0 {
		if options[0].Toolset != "" {
			cfg.Toolset = options[0].Toolset
		}
		if options[0].Profile != "" {
			cfg.Profile = options[0].Profile
		}
	}

	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-1c",
			Version: version,
		},
		nil,
	)

	switch cfg.Toolset {
	case ToolsetDeveloper:
		registerDeveloperTools(s, onecClient, dumpIndex)
		prompts.RegisterAll(s)
	case ToolsetBusiness:
		registerBusinessTools(s, onecClient, cfg.Profile)
	default:
		registerDeveloperTools(s, onecClient, dumpIndex)
		registerBusinessTools(s, onecClient, cfg.Profile)
		prompts.RegisterAll(s)
	}
	return s
}

func registerDeveloperTools(s *mcp.Server, onecClient *onec.Client, dumpIndex *dump.Index) {
	s.AddTool(tools.MetadataTool(), tools.NewMetadataHandler(onecClient))
	s.AddTool(tools.ObjectStructureTool(), tools.NewObjectStructureHandler(onecClient))
	s.AddTool(tools.QueryTool(), tools.NewQueryHandler(onecClient))
	if dumpIndex != nil {
		s.AddTool(tools.SearchCodeTool(), tools.NewSearchCodeHandler(dumpIndex))
	}

	// Pass dump directory to form handler so it can enrich the HTTP response
	// with data from Form.xml files parsed from the dump.
	var dumpDir string
	if dumpIndex != nil {
		dumpDir = dumpIndex.Dir()
	}
	s.AddTool(tools.FormStructureTool(), tools.NewFormStructureHandler(onecClient, dumpDir))
	s.AddTool(tools.ValidateQueryTool(), tools.NewValidateQueryHandler(onecClient))
	s.AddTool(tools.EventLogTool(), tools.NewEventLogHandler(onecClient))
	s.AddTool(tools.ConfigurationInfoTool(), tools.NewConfigurationInfoHandler(onecClient))
	tools.RegisterBSLHelp(s)
}

func registerBusinessTools(s *mcp.Server, onecClient *onec.Client, resolvedProfile string) {
	if !isBusinessToolSupported("read_counterparties", resolvedProfile) {
		return
	}
	s.AddTool(tools.ReadCounterpartiesTool(), tools.NewReadCounterpartiesHandler(onecClient))
	s.AddTool(tools.CreateCounterpartyTool(), tools.NewCreateCounterpartyHandler(onecClient))
}

func isBusinessToolSupported(toolName, resolvedProfile string) bool {
	supported := map[string]map[string]struct{}{
		"read_counterparties": {
			profile.Buh30:   {},
			profile.Generic: {},
		},
		"create_counterparty": {
			profile.Buh30:   {},
			profile.Generic: {},
		},
	}

	profiles, ok := supported[toolName]
	if !ok {
		return false
	}
	_, ok = profiles[resolvedProfile]
	return ok
}
