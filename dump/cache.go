package dump

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// cachePath returns the platform-specific cache directory for a dump index.
// Uses os.UserCacheDir():
//
//	macOS: ~/Library/Caches/mcp-1c/<hash>
//	Linux: ~/.cache/mcp-1c/<hash>  (or $XDG_CACHE_HOME)
//	Windows: %LocalAppData%/mcp-1c/<hash>
func cachePath(dumpDir string) (string, error) {
	absDir, err := filepath.Abs(dumpDir)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(absDir))
	hash := hex.EncodeToString(h[:8]) // first 16 hex chars

	cacheBase, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheBase, "mcp-1c", hash), nil
}
