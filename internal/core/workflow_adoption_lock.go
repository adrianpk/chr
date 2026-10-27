package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) acquireWorkflowAdoptionLocks(ctx context.Context) ([]*AgentOperationLock, error) {
	worktrees, err := (OSWorktreeGit{}).WorktreeList(ctx, s.layout.ProjectDir())
	if err != nil {
		return nil, err
	}

	if len(worktrees) == 0 || len(worktrees) > workflowstate.WorkflowPageLimit {
		return nil, fmt.Errorf("legacy worktree ownership scan bound exceeded")
	}

	sort.Slice(worktrees, func(i, j int) bool { return worktrees[i].Path < worktrees[j].Path })

	locks := []*AgentOperationLock{}

	for _, worktree := range worktrees {
		if worktree.Bare {
			continue
		}

		layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: worktree.Path, Namespace: s.layout.Namespace()})
		if err != nil {
			return nil, errors.Join(err, releaseWorkflowAdoptionLocks(locks))
		}

		lock, err := acquireWorkflowAdoptionLock(layout)
		if err != nil {
			return nil, errors.Join(err, releaseWorkflowAdoptionLocks(locks))
		}

		locks = append(locks, lock)
	}

	return locks, nil
}

func acquireWorkflowAdoptionLock(layout OpsLayout) (*AgentOperationLock, error) {
	path := filepath.Join(layout.Root(), "lock", "agent-operation.lock")

	if err := reclaimWorkflowAdoptionLock(path); err != nil {
		return nil, err
	}

	file, err := createAgentOperationLockFile(path)
	if err != nil {
		return nil, err
	}

	record := agentOperationLockRecord{Operation: "workflow_adoption", Owner: "chr", PID: os.Getpid(), CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := writeAgentOperationLockFile(file, path, record); err != nil {
		return nil, err
	}

	return &AgentOperationLock{path: path}, nil
}

func releaseWorkflowAdoptionLocks(locks []*AgentOperationLock) error {
	var result error
	for _, lock := range locks {
		result = errors.Join(result, lock.Release())
	}

	return result
}

func reclaimWorkflowAdoptionLock(path string) error {
	data, err := readWorkflowFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	var owner agentOperationLockRecord
	if err := json.Unmarshal(data, &owner); err != nil {
		return fmt.Errorf("legacy operation lock requires reconciliation")
	}

	if owner.Operation != "workflow_adoption" || owner.Owner != "chr" || owner.PID < 1 {
		return nil
	}

	if err := syscall.Kill(owner.PID, 0); !errors.Is(err, syscall.ESRCH) {
		return nil
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return syncWorkflowDirectory(filepath.Dir(path))
}
