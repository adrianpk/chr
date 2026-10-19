package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// Configuration mutations share this owner lock, including legacy upgrades.
// Readers observe a complete old or new image at the atomic config rename.
func withWorkspaceConfigLock(ctx context.Context, project string, action func() error) (err error) {
	path, err := workflowUpgradePath(project, "tmp/config.lock")
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

	if err := acquireWorkflowLock(ctx, file); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)) }()

	return action()
}
