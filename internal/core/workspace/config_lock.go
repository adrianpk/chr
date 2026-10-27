package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
)

// WithConfigLock preserves the existing guarded configuration filesystem contract.
// Configuration mutations share this owner lock, including legacy upgrades.
// Readers observe a complete old or new image at the atomic config rename.
func WithConfigLock(ctx context.Context, project string, action func() error) (err error) {
	path, err := ConfigPath(project, "tmp/config.lock")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	if err := opsfile.Lock(ctx, file); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)) }()

	return action()
}
