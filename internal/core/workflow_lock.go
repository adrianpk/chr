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

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) withLock(ctx context.Context, action func() error) (err error) {
	return s.withRepositoryLock(ctx, func() error {
		if err := s.guardWorkflowCoordinatorMutation(ctx); err != nil {
			return err
		}

		return s.recoverTransition()
	}, action)
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

	if err := opsfile.Lock(ctx, file); err != nil {
		return err
	}

	defer func() { err = errors.Join(err, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)) }()

	if err := prepare(); err != nil {
		return err
	}

	return action()
}

func (s *WorkflowStore) requireNoWorkflowTransition() error {
	path, err := s.safePath(workflowJournalPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: reconcile pending transition before preview", workflowstate.ErrConflict)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
