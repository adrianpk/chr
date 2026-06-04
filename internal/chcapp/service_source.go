package chcapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SourceOpenResult describes a source file selected for editor opening.
type SourceOpenResult struct {
	EditorCommand *DumpEditorCommand
}

func (s service) OpenSource(relativePath string) (SourceOpenResult, error) {
	path, err := s.resolveProjectSourcePath(relativePath)
	if err != nil {
		return SourceOpenResult{}, err
	}

	command, err := s.dumpEditor.EditorCommand(s.ctx, path)
	if err != nil {
		return SourceOpenResult{}, err
	}

	return SourceOpenResult{EditorCommand: &command}, nil
}

func (s service) resolveProjectSourcePath(relativePath string) (string, error) {
	relativePath = filepath.ToSlash(strings.TrimSpace(relativePath))
	if relativePath == "" {
		return "", fmt.Errorf("source path is required")
	}

	cleanPath := filepath.Clean(filepath.FromSlash(relativePath))
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || cleanPath == ".." {
		return "", fmt.Errorf("source path must be project-relative")
	}

	path := filepath.Join(s.projectContext.ProjectDir, cleanPath)

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat source: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("source path is a directory")
	}

	return path, nil
}
