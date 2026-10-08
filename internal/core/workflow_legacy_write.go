package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Legacy writers remain available for unadopted tickets. Adopted records must
// pass the repository lock, image and revision guards through WorkflowStore.
func readLegacyTicketFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	front, _, err := splitWorkflowMarkdown(content)
	if err == nil && strings.Contains("\n"+front, "\nworkflow_schema:") {
		return nil, fmt.Errorf("adopted workflow ticket requires a guarded record write")
	}

	return content, nil
}

// An optimistic legacy reader may race explicit adoption. Take the same lock
// before its replacement and refuse any changed image, including new adoption.
func writeLegacyTicketFile(ctx context.Context, path string, observed, replacement []byte) error {
	root, err := legacyTicketGitRoot(path)
	if err != nil {
		return err
	}

	action := func() error {
		current, err := readLegacyTicketFile(path)
		if err != nil {
			return err
		}

		if !bytes.Equal(current, observed) {
			return ErrWorkflowConflict
		}

		return os.WriteFile(path, replacement, 0o644)
	}
	if root == "" {
		// WorkflowStore requires Git coordination, so non-Git legacy tickets
		// cannot race an executable workflow adoption.
		return action()
	}

	layout := OpsLayout{projectDir: root, namespace: DefaultOpsNamespace}

	relative, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}

	parts := strings.Split(filepath.ToSlash(relative), "/")
	if len(parts) > 2 && parts[0] == OpsDirectoryName && IsValidOpsNamespace(parts[1]) {
		layout.namespace = parts[1]
	}

	return NewWorkflowStore(layout, "").withLock(ctx, action)
}

func legacyTicketGitRoot(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	for dir := filepath.Dir(absolute); ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if dir == filepath.Dir(dir) {
			return "", nil
		}
	}
}
