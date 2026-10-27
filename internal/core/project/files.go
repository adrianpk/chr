package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomically preserves the existing adjacent temporary-file replacement.
// Callers retain mutation permission, locking and transaction ownership.
func WriteFileAtomically(path string, content []byte) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	tmpPath := path + ".tmp"

	err = os.WriteFile(tmpPath, content, 0o644)
	if err != nil {
		return fmt.Errorf("write temporary file: %w", err)
	}

	err = os.Rename(tmpPath, path)
	if err != nil {
		return fmt.Errorf("rename temporary file: %w", err)
	}

	return nil
}
