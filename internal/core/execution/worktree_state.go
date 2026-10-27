package execution

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// WorktreeState defines a package boundary type.
type WorktreeState string

// WorktreeDiagnosticCode defines a package boundary type.
type WorktreeDiagnosticCode string

const (
	// WorktreeStateMissing defines a package boundary constant.
	WorktreeStateMissing WorktreeState = "missing"
	// WorktreeStateClean defines a package boundary constant.
	WorktreeStateClean WorktreeState = "clean"
	// WorktreeStateDirty defines a package boundary constant.
	WorktreeStateDirty WorktreeState = "dirty"
	// WorktreeStateMismatched defines a package boundary constant.
	WorktreeStateMismatched WorktreeState = "mismatched"
	// WorktreeStateBranchInUse defines a package boundary constant.
	WorktreeStateBranchInUse WorktreeState = "branch_in_use"
	// WorktreeStateOrphaned defines a package boundary constant.
	WorktreeStateOrphaned WorktreeState = "orphaned"

	// WorktreeDiagnosticMissing defines a package boundary constant.
	WorktreeDiagnosticMissing WorktreeDiagnosticCode = "missing"
	// WorktreeDiagnosticClean defines a package boundary constant.
	WorktreeDiagnosticClean WorktreeDiagnosticCode = "clean"
	// WorktreeDiagnosticDirty defines a package boundary constant.
	WorktreeDiagnosticDirty WorktreeDiagnosticCode = "dirty"
	// WorktreeDiagnosticMismatched defines a package boundary constant.
	WorktreeDiagnosticMismatched WorktreeDiagnosticCode = "mismatched"
	// WorktreeDiagnosticBranchInUse defines a package boundary constant.
	WorktreeDiagnosticBranchInUse WorktreeDiagnosticCode = "branch_in_use"
	// WorktreeDiagnosticOrphaned defines a package boundary constant.
	WorktreeDiagnosticOrphaned WorktreeDiagnosticCode = "orphaned"
)

// WorktreeInspection describes a package boundary value.
type WorktreeInspection struct {
	Plan        WorktreePlan
	State       WorktreeState
	Diagnostics []WorktreeDiagnostic
	Entry       WorktreeGitEntry
	StatusLines []string
}

// WorktreeDiagnostic describes a package boundary value.
type WorktreeDiagnostic struct {
	Code           WorktreeDiagnosticCode
	Detail         string
	Path           string
	Branch         string
	ExpectedPath   string
	ExpectedBranch string
}

// Inspect executes the corresponding package operation.
func (s WorktreeService) Inspect(ctx context.Context, opts WorktreePlanOptions) (WorktreeInspection, error) {
	settings, err := newWorktreeServiceSettings(opts, s.Git)
	if err != nil {
		return WorktreeInspection{}, err
	}

	if err := ctx.Err(); err != nil {
		return WorktreeInspection{}, err
	}

	return inspectWorktree(ctx, settings)
}

func inspectWorktree(ctx context.Context, settings worktreeServiceSettings) (WorktreeInspection, error) {
	entries, err := settings.git.WorktreeList(ctx, settings.plan.ProjectDir)
	if err != nil {
		return WorktreeInspection{}, fmt.Errorf("list worktrees: %w", err)
	}

	return classifyWorktree(ctx, worktreeInspectionRun{settings: settings, entries: entries})
}

type worktreeInspectionRun struct {
	settings worktreeServiceSettings
	entries  []WorktreeGitEntry
}

func classifyWorktree(ctx context.Context, run worktreeInspectionRun) (WorktreeInspection, error) {
	if entry, ok := worktreeEntryByPath(run.entries, run.settings.plan.WorktreePath); ok {
		return run.target(ctx, entry)
	}

	if entry, ok := worktreeEntryByBranch(run.entries, run.settings.plan.Branch); ok {
		return branchInUseWorktreeInspection(run.settings.plan, entry), nil
	}

	if worktreePathExists(run.settings.plan.WorktreePath) {
		return orphanedWorktreeInspection(run.settings.plan, WorktreeGitEntry{}), nil
	}

	return missingWorktreeInspection(run.settings.plan), nil
}

func (run worktreeInspectionRun) target(ctx context.Context, entry WorktreeGitEntry) (WorktreeInspection, error) {
	settings := run.settings
	if entry.Prunable {
		return orphanedWorktreeInspection(settings.plan, entry), nil
	}

	if entry.Branch != settings.plan.Branch {
		return mismatchedWorktreeInspection(settings.plan, entry), nil
	}

	status, err := settings.git.WorktreeStatus(ctx, settings.plan.WorktreePath)
	if err != nil {
		return WorktreeInspection{}, fmt.Errorf("inspect worktree status: %w", err)
	}

	if len(status) > 0 {
		return dirtyWorktreeInspection(settings.plan, entry, status), nil
	}

	return cleanWorktreeInspection(settings.plan, entry), nil
}

func worktreeEntryByPath(entries []WorktreeGitEntry, path string) (WorktreeGitEntry, bool) {
	want := filepath.Clean(path)
	for _, entry := range entries {
		if filepath.Clean(entry.Path) == want {
			return entry, true
		}
	}

	return WorktreeGitEntry{}, false
}

func worktreeEntryByBranch(entries []WorktreeGitEntry, branch string) (WorktreeGitEntry, bool) {
	for _, entry := range entries {
		if entry.Branch == branch {
			return entry, true
		}
	}

	return WorktreeGitEntry{}, false
}

func worktreePathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func missingWorktreeInspection(plan WorktreePlan) WorktreeInspection {
	return worktreeInspection(worktreeInspectionInput{plan: plan, state: WorktreeStateMissing, diagnostic: WorktreeDiagnostic{
		Code:           WorktreeDiagnosticMissing,
		Detail:         fmt.Sprintf("worktree path %s is missing", plan.WorktreePath),
		ExpectedPath:   plan.WorktreePath,
		ExpectedBranch: plan.Branch,
	}})
}

func cleanWorktreeInspection(plan WorktreePlan, entry WorktreeGitEntry) WorktreeInspection {
	return worktreeInspection(worktreeInspectionInput{plan: plan, state: WorktreeStateClean, entry: entry, diagnostic: WorktreeDiagnostic{
		Code:           WorktreeDiagnosticClean,
		Detail:         fmt.Sprintf("worktree %s is clean", plan.WorktreePath),
		Path:           entry.Path,
		Branch:         entry.Branch,
		ExpectedPath:   plan.WorktreePath,
		ExpectedBranch: plan.Branch,
	}})
}

func dirtyWorktreeInspection(plan WorktreePlan, entry WorktreeGitEntry, status []string) WorktreeInspection {
	return worktreeInspection(worktreeInspectionInput{plan: plan, state: WorktreeStateDirty, entry: entry, status: status, diagnostic: WorktreeDiagnostic{
		Code:           WorktreeDiagnosticDirty,
		Detail:         fmt.Sprintf("worktree %s has visible changes: %d", plan.WorktreePath, len(status)),
		Path:           entry.Path,
		Branch:         entry.Branch,
		ExpectedPath:   plan.WorktreePath,
		ExpectedBranch: plan.Branch,
	}})
}

func mismatchedWorktreeInspection(plan WorktreePlan, entry WorktreeGitEntry) WorktreeInspection {
	return worktreeInspection(worktreeInspectionInput{plan: plan, state: WorktreeStateMismatched, entry: entry, diagnostic: WorktreeDiagnostic{
		Code:           WorktreeDiagnosticMismatched,
		Detail:         fmt.Sprintf("worktree path %s is assigned to branch %s", plan.WorktreePath, entry.Branch),
		Path:           entry.Path,
		Branch:         entry.Branch,
		ExpectedPath:   plan.WorktreePath,
		ExpectedBranch: plan.Branch,
	}})
}

func branchInUseWorktreeInspection(plan WorktreePlan, entry WorktreeGitEntry) WorktreeInspection {
	return worktreeInspection(worktreeInspectionInput{plan: plan, state: WorktreeStateBranchInUse, entry: entry, diagnostic: WorktreeDiagnostic{
		Code:           WorktreeDiagnosticBranchInUse,
		Detail:         fmt.Sprintf("branch %s is already checked out at %s", plan.Branch, entry.Path),
		Path:           entry.Path,
		Branch:         entry.Branch,
		ExpectedPath:   plan.WorktreePath,
		ExpectedBranch: plan.Branch,
	}})
}

func orphanedWorktreeInspection(plan WorktreePlan, entry WorktreeGitEntry) WorktreeInspection {
	return worktreeInspection(worktreeInspectionInput{plan: plan, state: WorktreeStateOrphaned, entry: entry, diagnostic: WorktreeDiagnostic{
		Code:           WorktreeDiagnosticOrphaned,
		Detail:         fmt.Sprintf("worktree path %s exists without a usable Git worktree entry", plan.WorktreePath),
		Path:           entry.Path,
		Branch:         entry.Branch,
		ExpectedPath:   plan.WorktreePath,
		ExpectedBranch: plan.Branch,
	}})
}

type worktreeInspectionInput struct {
	plan       WorktreePlan
	state      WorktreeState
	entry      WorktreeGitEntry
	status     []string
	diagnostic WorktreeDiagnostic
}

func worktreeInspection(input worktreeInspectionInput) WorktreeInspection {
	return WorktreeInspection{
		Plan:        input.plan,
		State:       input.state,
		Diagnostics: []WorktreeDiagnostic{input.diagnostic},
		Entry:       input.entry,
		StatusLines: append([]string(nil), input.status...),
	}
}
