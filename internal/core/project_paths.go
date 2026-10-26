package core

import (
	"path/filepath"
	"strings"
)

func pathIsInsideProject(projectDir, targetPath string) bool {
	rel, err := filepath.Rel(projectDir, targetPath)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func resolveProjectPath(projectDir, configuredPath string) string {
	path := filepath.Clean(filepath.FromSlash(configuredPath))
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(projectDir, path)
}
