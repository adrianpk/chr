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

	if err := guardWorkflowProcessCapacity(common); err != nil {
		return nil, errors.Join(err, file.Close())
	}

	return file, nil
}

func (s *WorkflowStore) reserveWorkflowClaim(ctx context.Context, run workflowInvocationFile) (claim *WorkflowClaim, err error) {
	if _, err := s.checkWorkflowCoordinator(ctx); err != nil {
		return nil, err
	}

	if err := s.checkWorkflowFrontierOwnership(ctx, run); err != nil {
		return nil, err
	}

	file, guard, err := s.acquireWorkflowWorkerSlot(ctx)
	if err != nil {
		return nil, err
	}

	claim = &WorkflowClaim{InvocationID: run.ID, AttemptID: fmt.Sprintf("%s-attempt-%d", run.ID, len(run.Attempts)+1), store: s, reservation: file, capacityGuard: guard, execution: &workflowClaimExecution{}}
	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.releaseReservation())
		}
	}()

	layouts, err := s.workflowOperationLayouts(ctx)
	if err != nil {
		return nil, err
	}

	for _, layout := range layouts {
		lock, err := acquireWorkflowOperationLock(layout, claim)
		if err != nil {
			return nil, err
		}

		claim.locks = append(claim.locks, lock)
	}

	return claim, nil
}

func (s *WorkflowStore) workflowOperationLayouts(ctx context.Context) ([]OpsLayout, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return nil, err
	}

	lanes := workflowRetainedLanePaths(runs)

	layouts, err := s.workflowRepositoryLayouts(ctx)
	if err != nil {
		return nil, err
	}

	result := []OpsLayout{}

	for _, layout := range layouts {
		// The inherited common-directory reservation also excludes legacy starts
		// in native lanes. Do not dirty their immutable candidate with copied locks.
		if !lanes[layout.ProjectDir()] {
			result = append(result, layout)
		}
	}

	return result, nil
}

func acquireWorkflowOperationLock(layout OpsLayout, claim *WorkflowClaim) (*AgentOperationLock, error) {
	path := filepath.Join(layout.Root(), "lock", "agent-operation.lock")

	file, err := createAgentOperationLockFile(path)
	if err != nil {
		return nil, err
	}

	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		return nil, errors.Join(err, file.Close(), os.Remove(path))
	}

	record := agentOperationLockRecord{Operation: "workflow_execution", Owner: claim.InvocationID, AttemptID: claim.AttemptID, Process: &identity, PID: os.Getpid(), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
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

	if claim.capacityGuard != nil {
		result = errors.Join(result, claim.capacityGuard.Close())
		claim.capacityGuard = nil
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

	return acquireLegacyWorkflowCapacity(ctx, projectDir)
}

func workflowRetainedLanePaths(runs []workflowInvocationFile) map[string]bool {
	lanes := map[string]bool{}

	for _, run := range runs {
		if run.Lane != nil {
			lanes[run.Lane.Worktree] = true
		}

		for _, attempt := range run.Attempts {
			if attempt.Lane != nil {
				lanes[attempt.Lane.Worktree] = true
			}
		}
	}

	return lanes
}
