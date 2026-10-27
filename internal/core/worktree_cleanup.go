package core

import (
	"context"
	"errors"
	"fmt"
	"os"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

const (
	// WorktreeCleanupStatusSkipped defines a package boundary constant.
	WorktreeCleanupStatusSkipped = "skipped"
	// WorktreeCleanupStatusRemoved defines a package boundary constant.
	WorktreeCleanupStatusRemoved = "removed"
	// WorktreeCleanupStatusBlocked defines a package boundary constant.
	WorktreeCleanupStatusBlocked = "blocked"
)

// WorktreeCleanupOptions describes a package boundary value.
type WorktreeCleanupOptions struct {
	ProjectDir   string
	TicketKey    string
	Branch       string
	BranchPolicy WorkBranchPolicy
	Workspace    workspacecore.WorkWorkspaceConfig
	Git          WorktreeGit
}

// WorktreeCleanupResult describes a package boundary value.
type WorktreeCleanupResult struct {
	Plan            WorktreePlan
	Inspection      WorktreeInspection
	Status          string
	Detail          string
	RemovedWorktree bool
	RemovedMetadata bool
}

// CleanupMergedTicketWorktree executes the corresponding package operation.
func CleanupMergedTicketWorktree(ctx context.Context, opts WorktreeCleanupOptions) (WorktreeCleanupResult, error) {
	if err := ctx.Err(); err != nil {
		return WorktreeCleanupResult{}, err
	}

	settings, err := newWorktreeServiceSettings(WorktreePlanOptions{
		ProjectDir:   opts.ProjectDir,
		TicketKey:    opts.TicketKey,
		Branch:       opts.Branch,
		BranchPolicy: opts.BranchPolicy,
		Workspace:    opts.Workspace,
	}, opts.Git)
	if err != nil {
		return WorktreeCleanupResult{}, err
	}

	result := WorktreeCleanupResult{Plan: settings.plan, Status: WorktreeCleanupStatusSkipped}
	if settings.workspace.Mode != workspacecore.WorkWorkspaceModeWorktree {
		result.Detail = "workspace mode is project"

		return result, nil
	}

	if settings.workspace.WorktreeCleanup != workspacecore.WorktreeCleanupPolicyAfterMerge {
		result.Detail = fmt.Sprintf("worktree cleanup policy is %s", settings.workspace.WorktreeCleanup)

		return result, nil
	}

	return cleanupMergedTicketWorktree(ctx, settings, result)
}

func cleanupMergedTicketWorktree(ctx context.Context, settings worktreeServiceSettings, result WorktreeCleanupResult) (WorktreeCleanupResult, error) {
	inspection, err := inspectWorktree(ctx, settings)
	if err != nil {
		return WorktreeCleanupResult{}, err
	}

	result.Inspection = inspection

	switch inspection.State {
	case WorktreeStateMissing:
		result.Detail = "ticket worktree is already missing"

		return removeWorktreeCleanupMetadata(result)
	case WorktreeStateClean:
		return removeCleanMergedTicketWorktree(ctx, settings, result)
	case WorktreeStateDirty, WorktreeStateMismatched, WorktreeStateBranchInUse, WorktreeStateOrphaned:
		result.Status = WorktreeCleanupStatusBlocked
		result.Detail = worktreeCleanupBlockerDetail(inspection)

		return result, nil
	default:
		result.Status = WorktreeCleanupStatusBlocked
		result.Detail = fmt.Sprintf("ticket worktree state is %s", inspection.State)

		return result, nil
	}
}

func removeCleanMergedTicketWorktree(ctx context.Context, settings worktreeServiceSettings, result WorktreeCleanupResult) (WorktreeCleanupResult, error) {
	if err := settings.git.RemoveWorktree(ctx, settings.plan.ProjectDir, settings.plan.WorktreePath); err != nil {
		result.Status = WorktreeCleanupStatusBlocked
		result.Detail = fmt.Sprintf("remove clean worktree: %v", err)

		return result, nil
	}

	result.Status = WorktreeCleanupStatusRemoved
	result.Detail = fmt.Sprintf("removed clean ticket worktree %s", settings.plan.WorktreePath)
	result.RemovedWorktree = true

	return removeWorktreeCleanupMetadata(result)
}

func removeWorktreeCleanupMetadata(result WorktreeCleanupResult) (WorktreeCleanupResult, error) {
	path, err := WorktreeMetadataPath(result.Plan)
	if err != nil {
		return WorktreeCleanupResult{}, err
	}

	err = os.Remove(path)
	if err == nil {
		result.RemovedMetadata = true

		return result, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}

	return WorktreeCleanupResult{}, fmt.Errorf("remove worktree metadata: %w", err)
}

func worktreeCleanupBlockerDetail(inspection WorktreeInspection) string {
	if len(inspection.Diagnostics) > 0 {
		return inspection.Diagnostics[0].Detail
	}

	return fmt.Sprintf("ticket worktree %s is %s", inspection.Plan.WorktreePath, inspection.State)
}
