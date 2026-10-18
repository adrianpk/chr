package core

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type workflowFileChange struct {
	Path   string `json:"path"`
	Before string `json:"before"`
	After  []byte `json:"after"`
}

func workflowHash(data []byte) string {
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

func readWorkflowFile(path string) ([]byte, error) {
	return readBoundedWorkflowFile(path, WorkflowFileLimit)
}

func readBoundedWorkflowFile(path string, limit int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		//nolint:errcheck // Read-only descriptor close cannot change stored state.
		_ = file.Close()
	}()

	data, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if err != nil {
		return nil, err
	}

	if len(data) > limit {
		return nil, fmt.Errorf("workflow file exceeds %d bytes: %s", limit, filepath.Base(path))
	}

	return data, nil
}

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

func atomicWorkflowFile(path string, content []byte) (err error) {
	return atomicBoundedWorkflowFile(path, content, WorkflowFileLimit)
}

func atomicBoundedWorkflowFile(path string, content []byte, limit int) (err error) {
	if len(content) > limit {
		return fmt.Errorf("workflow file exceeds %d bytes", limit)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".workflow-*")
	if err != nil {
		return err
	}

	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, removeErr)
		}
	}()

	if _, err := file.Write(content); err != nil {
		return errors.Join(err, file.Close())
	}

	if err := file.Sync(); err != nil {
		return errors.Join(err, file.Close())
	}

	if err := file.Close(); err != nil {
		return err
	}

	if err := os.Rename(file.Name(), path); err != nil {
		return err
	}

	return syncWorkflowDirectory(filepath.Dir(path))
}

func syncWorkflowDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}

	return errors.Join(dir.Sync(), dir.Close())
}

func (s *WorkflowStore) encodeChanges(r WorkflowRecord) ([]workflowFileChange, error) {
	content, err := encodeWorkflowRecord(r)
	if err != nil {
		return nil, err
	}

	changes := []workflowFileChange{{Path: r.Path, After: content}}
	if strings.HasSuffix(r.Path, "/metadata.json") {
		changes = append(changes, workflowFileChange{Path: filepath.ToSlash(filepath.Join(filepath.Dir(r.Path), "artifact.md")), After: r.Body})
	}

	for i := range changes {
		path, err := s.safePath(changes[i].Path)
		if err != nil {
			return nil, err
		}

		before, err := readWorkflowFile(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		if err == nil {
			changes[i].Before = workflowHash(before)
		}
	}

	return changes, nil
}
