package extension

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/feenlace/mcp-1c/internal/profile"
)

//go:embed src/ConfigDumpInfo.xml
//go:embed src/Configuration.xml
//go:embed src/Languages/Русский.xml
//go:embed src/Roles/MCP_ОсновнаяРоль.xml
//go:embed src/Roles/MCP_ОсновнаяРоль/Ext/Rights.xml
//go:embed src/HTTPServices/MCPService.xml
//go:embed src/HTTPServices/MCPService/Ext/Module.bsl
//go:embed src/CommonModules/MCPBusinessFacade.xml
//go:embed src/CommonModules/MCPBusinessFacade/Ext/Module.bsl
//go:embed profiles/generic/src/CommonModules/MCPBusinessFacade/Ext/Module.bsl
//go:embed profiles/buh_3_0/src/CommonModules/MCPBusinessFacade/Ext/Module.bsl
var Source embed.FS

func SourceForInstallProfile(installProfile string) (embed.FS, string, string, error) {
	normalized, err := profile.NormalizeInstallProfile(installProfile)
	if err != nil {
		return embed.FS{}, "", "", err
	}

	switch normalized {
	case profile.Generic, profile.Buh30:
		overlayRoot := "profiles/" + normalized + "/src"
		if _, statErr := fs.Stat(Source, overlayRoot); statErr == nil {
			return Source, "src", overlayRoot, nil
		}
		return Source, "src", "", nil
	default:
		return embed.FS{}, "", "", fmt.Errorf("unsupported install profile %q", installProfile)
	}
}
