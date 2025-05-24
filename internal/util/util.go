package util

import (
	"path/filepath"
	"strings"
)

func FilePath(parts ...string) string {
	if len(parts) == 1 {
		return filepath.FromSlash(parts[0])
	}
	return filepath.FromSlash(strings.Join(parts, ``))
}
