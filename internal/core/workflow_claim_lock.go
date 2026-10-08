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

func workflowRepositoryDirectory(ctx context.Context, projectDir string) (string, error) {
	command := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--path-format=absolute", "--git-common-dir")

	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("resolve workflow repository coordinator: %w", err)
	}

	return filepath.EvalSymlinks(strings.TrimSpace(string(output)))
}

func acquireWorkflowCapacity(ctx context.Context, projectDir string) (*os.File, error) {
	common, err := workflowRepositoryDirectory(ctx, projectDir)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(common, "chr-workflow-worker.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		closeErr := file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, errors.Join(ErrWorkflowCapacity, closeErr)
		}

		return nil, errors.Join(err, closeErr)
	}

	return file, nil
}

func (s *WorkflowStore) reserveWorkflowClaim(ctx context.Context, id string) (claim *WorkflowClaim, err error) {
	file, err := acquireWorkflowCapacity(ctx, s.layout.ProjectDir())
	if err != nil {
		return nil, err
	}

	claim = &WorkflowClaim{InvocationID: id, store: s, reservation: file}
	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.releaseReservation())
		}
	}()

	layouts, err := s.workflowRepositoryLayouts(ctx)
	if err != nil {
		return nil, err
	}

	for _, layout := range layouts {
		lock, err := acquireWorkflowOperationLock(layout, id)
		if err != nil {
			return nil, err
		}

		claim.locks = append(claim.locks, lock)
	}

	return claim, nil
}

func acquireWorkflowOperationLock(layout OpsLayout, id string) (*AgentOperationLock, error) {
	path := filepath.Join(layout.Root(), "lock", "agent-operation.lock")

	file, err := createAgentOperationLockFile(path)
	if err != nil {
		return nil, err
	}

	record := agentOperationLockRecord{Operation: "workflow_execution", Owner: id, PID: os.Getpid(), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if err := writeAgentOperationLockFile(file, path, record); err != nil {
		return nil, err
	}

	data, err := readWorkflowFile(path)
	if err != nil {
		return nil, err
	}

	return &AgentOperationLock{path: path, guard: workflowHash(data)}, nil
}

func (claim *WorkflowClaim) releaseReservation() error {
	if claim == nil {
		return nil
	}

	result := releaseWorkflowAdoptionLocks(claim.locks)
	if claim.reservation != nil {
		// Close, never LOCK_UN: an inherited descriptor keeps a live child counted.
		result = errors.Join(result, claim.reservation.Close())
		claim.reservation = nil
	}

	return result
}

func guardedAgentOperationRelease(lock *AgentOperationLock) error {
	if lock.guard != "" {
		data, err := readWorkflowFile(lock.path)
		if err != nil {
			return err
		}

		if workflowHash(data) != lock.guard {
			return fmt.Errorf("%w: operation lock owner changed", ErrWorkflowConflict)
		}
	}

	if err := os.Remove(lock.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("release agent operation lock: %w", err)
	}

	lock.path = ""

	return nil
}

func acquireAgentRepositoryCapacity(ctx context.Context, projectDir string) (*os.File, error) {
	// Legacy standalone workspaces retain their file lock without a Git pool.
	probe := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--git-common-dir")

	output, err := probe.CombinedOutput()
	if err != nil && strings.Contains(string(output), "not a git repository") {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return acquireWorkflowCapacity(ctx, projectDir)
}
