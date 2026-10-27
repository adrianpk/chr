package project

import (
	"path/filepath"
	"strings"
)

// ContainsPath compares project-relative spelling without resolving symlinks.
func ContainsPath(projectDir, targetPath string) bool {
	rel, err := filepath.Rel(projectDir, targetPath)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// ResolvePath resolves a configured absolute or project-relative path.
func ResolvePath(projectDir, configuredPath string) string {
	path := filepath.Clean(filepath.FromSlash(configuredPath))
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(projectDir, path)
}
