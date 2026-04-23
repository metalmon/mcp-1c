package server

import (
	"github.com/feenlace/mcp-1c/dump"
	"github.com/feenlace/mcp-1c/onec"
	"github.com/feenlace/mcp-1c/prompts"
	"github.com/feenlace/mcp-1c/tools"
	businessbuh30 "github.com/feenlace/mcp-1c/tools/business/buh30"
	businessgeneric "github.com/feenlace/mcp-1c/tools/business/generic"
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
		if options[0].Access != "" {
			cfg.Access = options[0].Access
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
		prompts.RegisterDeveloper(s)
	case ToolsetBusiness:
		registerBusinessTools(s, onecClient, cfg.Profile, cfg.Access)
		prompts.RegisterBusiness(s, cfg.Profile)
	default:
		registerDeveloperTools(s, onecClient, dumpIndex)
		registerBusinessTools(s, onecClient, cfg.Profile, cfg.Access)
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

func registerBusinessTools(s *mcp.Server, onecClient *onec.Client, resolvedProfile string, access AccessMode) {
	switch resolvedProfile {
	case "buh_3_0":
		businessbuh30.RegisterTools(s, onecClient, access == AccessReadOnly)
	case "generic":
		businessgeneric.RegisterTools(s, onecClient, access == AccessReadOnly)
	}
}
