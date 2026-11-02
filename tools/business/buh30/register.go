package buh30

import (
	"github.com/feenlace/mcp-1c/onec"
	businesscommon "github.com/feenlace/mcp-1c/tools/business/common"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterTools(s *mcp.Server, onecClient *onec.Client, readOnly bool) {
	// Инструменты языка запросов и метаданных — профиля не касаются, работают
	// против любой конфигурации (см. `RegisterQueryTools`).
	businesscommon.RegisterQueryTools(s, onecClient)
	// Именованные читающие — только здесь: за ними фасад, и он реализован
	// только для «Бухгалтерии 3.0».
	businesscommon.RegisterReadTools(s, onecClient)
	if readOnly {
		return
	}
	businesscommon.RegisterWriteTools(s, onecClient)
}
