package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Read retains the default workflow file bound.
func Read(path string) ([]byte, error) {
	return ReadBounded(path, workflowstate.FileLimit)
}

// ReadBounded reads at most limit bytes and detects overflow.
func ReadBounded(path string, limit int) ([]byte, error) {
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

// WriteAtomic retains the default bounded, synced replacement.
func WriteAtomic(path string, content []byte) (err error) {
	return WriteAtomicBounded(path, content, workflowstate.FileLimit)
}

// WriteAtomicBounded syncs an adjacent temporary image before rename and directory sync.
func WriteAtomicBounded(path string, content []byte, limit int) (err error) {
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

	return SyncDirectory(filepath.Dir(path))
}

// SyncDirectory makes the directory update durable and preserves close errors.
func SyncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}

	return errors.Join(dir.Sync(), dir.Close())
}
