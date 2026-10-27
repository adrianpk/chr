package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) safePath(relative string) (string, error) {
	if !filepath.IsLocal(relative) || filepath.Clean(relative) != relative {
		return "", fmt.Errorf("invalid workflow relative path")
	}

	path := filepath.Join(s.layout.Root(), relative)
	for parent := path; parent != s.layout.ProjectDir(); parent = filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("workflow record path cannot traverse a symlink")
		}
	}

	return path, nil
}

func (s *WorkflowStore) encodeChanges(r WorkflowRecord) ([]workflowstate.FileChange, error) {
	content, err := encodeWorkflowRecord(r)
	if err != nil {
		return nil, err
	}

	changes := []workflowstate.FileChange{{Path: r.Path, After: content}}
	if strings.HasSuffix(r.Path, "/metadata.json") {
		changes = append(changes, workflowstate.FileChange{Path: filepath.ToSlash(filepath.Join(filepath.Dir(r.Path), "artifact.md")), After: r.Body})
	}

	for i := range changes {
		path, err := s.safePath(changes[i].Path)
		if err != nil {
			return nil, err
		}

		before, err := opsfile.Read(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		if err == nil {
			changes[i].Before = contenthash.Sum(before)
		}
	}

	return changes, nil
}
