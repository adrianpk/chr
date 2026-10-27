package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveDirectory validates an existing project directory and returns its absolute path.
func ResolveDirectory(projectDir string) (string, error) {
	if projectDir == "" {
		projectDir = "."
	}

	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}

	info, err := os.Stat(absProjectDir)
	if err != nil {
		return "", fmt.Errorf("inspect project directory: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("project directory is not a directory: %s", absProjectDir)
	}

	return absProjectDir, nil
}
