package generic

import (
	"github.com/feenlace/mcp-1c/onec"
	businesscommon "github.com/feenlace/mcp-1c/tools/business/common"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterTools(s *mcp.Server, onecClient *onec.Client, readOnly bool) {
	businesscommon.RegisterReadTools(s, onecClient)
	if readOnly {
		return
	}
	businesscommon.RegisterWriteTools(s, onecClient)
}
