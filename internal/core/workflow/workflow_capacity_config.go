package workflow

import (
	"context"
	"errors"
	"fmt"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type workflowCapacityLimits struct {
	Workers  int
	Queue    int
	Prepared int
}

func workflowCapacityConfig(projectDir string) (workflowCapacityLimits, error) {
	config, err := workspaceconfig.Load(projectDir)
	if err != nil {
		return workflowCapacityLimits{}, err
	}

	limits := workflowCapacityLimits{Workers: config.Work.MaxActiveWorkers, Queue: config.Work.ReadyQueueCapacity, Prepared: min(config.Work.Integration.MaxPreparedPerTarget, config.Work.ReadyQueueCapacity)}
	if limits.Workers < 1 || limits.Workers > workflowstate.WorkerSlotLimit || limits.Queue < 1 || limits.Queue > 1024 {
		return limits, fmt.Errorf("work.max_active_workers must be 1..8 and work.ready_queue_capacity must be 1..1024")
	}

	return limits, nil
}

// Linked host worktrees must resolve one policy. Lane worktrees retain immutable
// admission snapshots; their copied config does not define another worker pool.
func (host workflowCapacityObserver) repository(ctx context.Context, runs []workflowInvocationFile) (workflowCapacityLimits, error) {
	limits, err := workflowCapacityConfig(host.store.layout.ProjectDir())
	if err != nil {
		return limits, err
	}

	worktrees, err := (executioncore.OSWorktreeGit{}).WorktreeList(ctx, host.store.layout.ProjectDir())
	if err != nil {
		return limits, err
	}

	if len(worktrees) == 0 || len(worktrees) > workflowstate.PageLimit {
		return limits, fmt.Errorf("workflow worktree capacity observation bound exceeded")
	}

	return limits, checkWorkflowCapacityRoots(worktrees, workflowRetainedLanePaths(runs), limits)
}

func checkWorkflowCapacityRoots(worktrees []executioncore.WorktreeGitEntry, lanes map[string]bool, limits workflowCapacityLimits) error {
	for _, worktree := range worktrees {
		if worktree.Bare || worktree.Prunable || lanes[worktree.Path] {
			continue
		}

		observed, err := workflowCapacityConfig(worktree.Path)
		if err != nil {
			return err
		}

		if observed != limits {
			return fmt.Errorf("%w: linked worktrees resolve conflicting repository capacity", workflowstate.ErrConflict)
		}
	}

	return nil
}

// ErrWorkflowQueueFull rejects new lanes while preserving all accepted work.
var ErrWorkflowQueueFull = errors.New("workflow ready queue capacity exceeded")

func guardWorkflowReadyQueue(runs []workflowInvocationFile, next workflowInvocationFile, limits workflowCapacityLimits) error {
	count := workflowUnfinishedLaneCount(next)
	for _, run := range runs {
		count += workflowUnfinishedLaneCount(run)
	}

	if count > limits.Queue {
		return ErrWorkflowQueueFull
	}

	return nil
}

func workflowUnfinishedLaneCount(run workflowInvocationFile) int {
	if workflowstate.InvocationTerminal(run.Status) {
		return 0
	}

	lanes := map[string]bool{}

	for _, member := range run.Members[run.ConfirmedCursor:] {
		identity := "ticket:" + member.Admission.Key
		if member.Admission.SliceID != "" {
			identity = "slice:" + member.Admission.SliceID
		}

		lanes[identity] = true
	}
	// A retained set closure is itself unfinished admission even with no tasks.
	if len(lanes) == 0 && workflowcloseout.NeedsClosing(run.Invocation) {
		return 1
	}

	return len(lanes)
}
