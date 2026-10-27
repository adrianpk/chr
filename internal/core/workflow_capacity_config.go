package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowWorkerSlotLimit = 8

type workflowCapacityLimits struct {
	Workers  int
	Queue    int
	Prepared int
}

func setWorkCapacityConfigValue(config *WorkConfig, key, value string) bool {
	var field *int

	switch key {
	case "max_active_workers":
		field = &config.MaxActiveWorkers
	case "ready_queue_capacity":
		field = &config.ReadyQueueCapacity
	default:
		return false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		parsed = -1
	}

	*field = parsed

	return true
}

func workflowCapacityConfig(projectDir string) (workflowCapacityLimits, error) {
	config, err := LoadWorkspaceConfig(projectDir)
	if err != nil {
		return workflowCapacityLimits{}, err
	}

	limits := workflowCapacityLimits{Workers: config.Work.MaxActiveWorkers, Queue: config.Work.ReadyQueueCapacity, Prepared: min(config.Work.Integration.MaxPreparedPerTarget, config.Work.ReadyQueueCapacity)}
	if limits.Workers < 1 || limits.Workers > workflowWorkerSlotLimit || limits.Queue < 1 || limits.Queue > 1024 {
		return limits, fmt.Errorf("work.max_active_workers must be 1..8 and work.ready_queue_capacity must be 1..1024")
	}

	return limits, nil
}

// Linked host worktrees must resolve one policy. Lane worktrees retain immutable
// admission snapshots; their copied config does not define another worker pool.
func (s *WorkflowStore) workflowRepositoryCapacity(ctx context.Context, runs []workflowInvocationFile) (workflowCapacityLimits, error) {
	limits, err := workflowCapacityConfig(s.layout.ProjectDir())
	if err != nil {
		return limits, err
	}

	worktrees, err := (OSWorktreeGit{}).WorktreeList(ctx, s.layout.ProjectDir())
	if err != nil {
		return limits, err
	}

	if len(worktrees) == 0 || len(worktrees) > workflowstate.PageLimit {
		return limits, fmt.Errorf("workflow worktree capacity observation bound exceeded")
	}

	return limits, checkWorkflowCapacityRoots(worktrees, workflowRetainedLanePaths(runs), limits)
}

func checkWorkflowCapacityRoots(worktrees []WorktreeGitEntry, lanes map[string]bool, limits workflowCapacityLimits) error {
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
	if workflowInvocationTerminal(run.Status) {
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
	if len(lanes) == 0 && workflowInvocationNeedsClosing(run.WorkflowInvocation) {
		return 1
	}

	return len(lanes)
}
