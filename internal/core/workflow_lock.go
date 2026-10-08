package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func (s *WorkflowStore) withLock(ctx context.Context, action func() error) (err error) {
	return s.withRepositoryLock(ctx, s.recoverTransition, action)
}

func (s *WorkflowStore) withRepositoryLock(ctx context.Context, prepare, action func() error) (err error) {
	command := exec.CommandContext(ctx, "git", "-C", s.layout.ProjectDir(), "rev-parse", "--path-format=absolute", "--git-common-dir")

	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("resolve workflow repository coordinator: %w", err)
	}

	commonDir, err := filepath.EvalSymlinks(strings.TrimSpace(string(output)))
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filepath.Join(commonDir, "chr-workflow.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, file.Close()) }()

	if err := acquireWorkflowLock(ctx, file); err != nil {
		return err
	}

	defer func() { err = errors.Join(err, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)) }()

	if err := prepare(); err != nil {
		return err
	}

	return action()
}

func acquireWorkflowLock(ctx context.Context, file *os.File) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}

		if !errors.Is(err, syscall.EWOULDBLOCK) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func (s *WorkflowStore) requireNoWorkflowTransition() error {
	path, err := s.safePath(workflowJournalPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: reconcile pending transition before preview", ErrWorkflowConflict)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
