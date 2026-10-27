package execution

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// WorkReadinessRepairOptions describes a package boundary value.
type WorkReadinessRepairOptions struct {
	ProjectDir   string
	Policy       WorkBranchPolicy
	Workspace    workspacecore.ExecutionConfig
	TicketKey    string
	TicketBranch string
	Git          WorkReadinessRepairGit
	WorktreeGit  WorktreeGit
}

// RepairWorkReadiness executes the corresponding package operation.
func RepairWorkReadiness(ctx context.Context, opts WorkReadinessRepairOptions) (WorkReadinessRepairResult, error) {
	settings, err := newWorkReadinessRepairSettings(ctx, opts)
	if err != nil {
		return WorkReadinessRepairResult{}, err
	}

	before, err := settings.check(ctx)
	if err != nil {
		return WorkReadinessRepairResult{}, err
	}

	result := WorkReadinessRepairResult{Before: before, After: before}
	if !settings.policy.AllowSafeRepair {
		return result, nil
	}

	return runRepairLoop(ctx, settings, result)
}

type workReadinessRepairSettings struct {
	workReadinessSettings
	git WorkReadinessRepairGit
}

func newWorkReadinessRepairSettings(ctx context.Context, opts WorkReadinessRepairOptions) (workReadinessRepairSettings, error) {
	git := opts.Git
	if git == nil {
		git = OSWorkReadinessGit{}
	}

	settings, err := newWorkReadinessSettings(ctx, WorkReadinessOptions{
		ProjectDir:   opts.ProjectDir,
		Policy:       opts.Policy,
		Workspace:    opts.Workspace,
		TicketKey:    opts.TicketKey,
		TicketBranch: opts.TicketBranch,
		Git:          git,
		WorktreeGit:  opts.WorktreeGit,
	})
	if err != nil {
		return workReadinessRepairSettings{}, err
	}

	return workReadinessRepairSettings{workReadinessSettings: settings, git: git}, nil
}

func (settings workReadinessRepairSettings) check(ctx context.Context) (WorkReadinessReport, error) {
	return CheckWorkReadiness(ctx, WorkReadinessOptions{
		ProjectDir:   settings.projectDir,
		Policy:       settings.policy,
		Workspace:    settings.workspace,
		TicketKey:    settings.ticketKey,
		TicketBranch: settings.ticketBranch,
		Git:          settings.git,
		WorktreeGit:  settings.worktreeGit,
	})
}

func runRepairLoop(ctx context.Context, settings workReadinessRepairSettings, result WorkReadinessRepairResult) (WorkReadinessRepairResult, error) {
	current := result.Before
	for range 3 {
		applied, err := applySafeRepairs(ctx, settings.projectDir, settings.git, current.Diagnostics)
		if err != nil {
			return WorkReadinessRepairResult{}, err
		}

		if len(applied) == 0 {
			break
		}

		result.Applied = append(result.Applied, applied...)

		current, err = settings.check(ctx)
		if err != nil {
			return WorkReadinessRepairResult{}, err
		}

		result.After = current
		if current.Status == WorkReadinessStatusReady {
			break
		}
	}

	return result, nil
}

func applySafeRepairs(ctx context.Context, projectDir string, git WorkReadinessRepairGit, diagnostics []WorkReadinessDiagnostic) ([]WorkReadinessAppliedRepair, error) {
	applied := make([]WorkReadinessAppliedRepair, 0)
	branchMutationSafe := branchMutationRepairSafe(diagnostics)
	run := repairRun{
		projectDir:         projectDir,
		git:                git,
		branchMutationSafe: branchMutationSafe,
	}

	for _, diagnostic := range diagnostics {
		if !diagnostic.Repair.Available || !diagnostic.Repair.Safe {
			continue
		}

		outcome, err := applySafeRepair(ctx, run, diagnostic)
		if err != nil {
			return nil, err
		}

		if outcome.applied {
			applied = append(applied, outcome.repair)
		}
	}

	return applied, nil
}

type repairRun struct {
	projectDir         string
	git                WorkReadinessRepairGit
	branchMutationSafe bool
}

func branchMutationRepairSafe(diagnostics []WorkReadinessDiagnostic) bool {
	return !diagnosticsContain(
		diagnostics,
		WorkReadinessDiagnosticDirtyWorktree,
		WorkReadinessDiagnosticGitStateInProgress,
	)
}

func diagnosticsContain(diagnostics []WorkReadinessDiagnostic, codes ...WorkReadinessDiagnosticCode) bool {
	for _, diagnostic := range diagnostics {
		for _, code := range codes {
			if diagnostic.Code == code {
				return true
			}
		}
	}

	return false
}

type repairOutcome struct {
	repair  WorkReadinessAppliedRepair
	applied bool
}

func applySafeRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (repairOutcome, error) {
	switch diagnostic.Repair.Code {
	case WorkReadinessRepairIgnoreRuntimeArtifacts:
		return applyRuntimeArtifactsRepair(ctx, run)
	case WorkReadinessRepairFetchRemote:
		return applyFetchRemoteRepair(ctx, run, diagnostic)
	case WorkReadinessRepairFastForwardBranch:
		return applyFastForwardRepair(ctx, run, diagnostic)
	case WorkReadinessRepairSwitchBranch:
		return applySwitchBranchRepair(ctx, run, diagnostic)
	case WorkReadinessRepairCreateBranch:
		return applyCreateBranchRepair(ctx, run, diagnostic)
	case WorkReadinessRepairCreateWorktreeRoot:
		return applyCreateWorktreeRootRepair(diagnostic)
	case WorkReadinessRepairPruneWorktreeMetadata:
		return applyPruneWorktreeMetadataRepair(diagnostic)
	case WorkReadinessRepairRemoveOrphanedWorktree:
		return applyRemoveOrphanedWorktreeRepair(diagnostic)
	default:
		return repairOutcome{}, nil
	}
}

func applyRuntimeArtifactsRepair(ctx context.Context, run repairRun) (repairOutcome, error) {
	changed := false

	for _, pattern := range runtimeArtifactIgnorePatterns() {
		patternChanged, err := run.git.ExcludeLocalPattern(ctx, run.projectDir, pattern)
		if err != nil {
			return repairOutcome{}, fmt.Errorf("exclude runtime artifacts: %w", err)
		}

		changed = changed || patternChanged
	}

	return repairOutcome{repair: WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairIgnoreRuntimeArtifacts,
		Detail: "excluded runtime artifact roots in local Git info exclude",
	}, applied: changed}, nil
}

func applyFetchRemoteRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (repairOutcome, error) {
	err := run.git.FetchRemote(ctx, run.projectDir, diagnostic.Remote)
	if err != nil {
		return repairOutcome{}, fmt.Errorf("fetch remote: %w", err)
	}

	return repairOutcome{repair: WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairFetchRemote,
		Detail: fmt.Sprintf("fetched %s", diagnostic.Remote),
	}, applied: true}, nil
}

func applyFastForwardRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (repairOutcome, error) {
	if !run.branchMutationSafe {
		return repairOutcome{}, nil
	}

	err := run.git.FastForwardBranch(ctx, run.projectDir, diagnostic.ExpectedBranch, diagnostic.RemoteRef)
	if err != nil {
		return repairOutcome{}, fmt.Errorf("fast-forward branch: %w", err)
	}

	return repairOutcome{repair: WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairFastForwardBranch,
		Detail: fmt.Sprintf("fast-forwarded %s from %s", diagnostic.ExpectedBranch, diagnostic.RemoteRef),
	}, applied: true}, nil
}

func applySwitchBranchRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (repairOutcome, error) {
	if !run.branchMutationSafe {
		return repairOutcome{}, nil
	}

	err := run.git.SwitchBranch(ctx, run.projectDir, diagnostic.ExpectedBranch)
	if err != nil {
		return repairOutcome{}, fmt.Errorf("switch branch: %w", err)
	}

	return repairOutcome{repair: WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairSwitchBranch,
		Detail: fmt.Sprintf("switched to %s", diagnostic.ExpectedBranch),
	}, applied: true}, nil
}

func applyCreateBranchRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (repairOutcome, error) {
	if !run.branchMutationSafe {
		return repairOutcome{}, nil
	}

	err := run.git.CreateBranch(ctx, run.projectDir, diagnostic.ExpectedBranch)
	if err != nil {
		return repairOutcome{}, fmt.Errorf("create branch: %w", err)
	}

	return repairOutcome{repair: WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairCreateBranch,
		Detail: fmt.Sprintf("created branch %s", diagnostic.ExpectedBranch),
	}, applied: true}, nil
}

func applyCreateWorktreeRootRepair(diagnostic WorkReadinessDiagnostic) (repairOutcome, error) {
	if strings.TrimSpace(diagnostic.Path) == "" {
		return repairOutcome{}, nil
	}

	if err := os.MkdirAll(diagnostic.Path, 0o755); err != nil {
		return repairOutcome{}, fmt.Errorf("create worktree root: %w", err)
	}

	return repairOutcome{repair: WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairCreateWorktreeRoot,
		Detail: fmt.Sprintf("created worktree root %s", diagnostic.Path),
	}, applied: true}, nil
}

func applyPruneWorktreeMetadataRepair(diagnostic WorkReadinessDiagnostic) (repairOutcome, error) {
	if strings.TrimSpace(diagnostic.Path) == "" {
		return repairOutcome{}, nil
	}

	err := os.Remove(diagnostic.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repairOutcome{}, nil
		}

		return repairOutcome{}, fmt.Errorf("prune worktree metadata: %w", err)
	}

	return repairOutcome{repair: WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairPruneWorktreeMetadata,
		Detail: fmt.Sprintf("removed stale worktree metadata %s", diagnostic.Path),
	}, applied: true}, nil
}

func applyRemoveOrphanedWorktreeRepair(diagnostic WorkReadinessDiagnostic) (repairOutcome, error) {
	if strings.TrimSpace(diagnostic.Path) == "" {
		return repairOutcome{}, nil
	}

	entries, err := os.ReadDir(diagnostic.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repairOutcome{}, nil
		}

		return repairOutcome{}, fmt.Errorf("inspect orphaned worktree path: %w", err)
	}

	if len(entries) > 0 {
		return repairOutcome{}, nil
	}

	if err := os.Remove(diagnostic.Path); err != nil {
		return repairOutcome{}, fmt.Errorf("remove orphaned worktree path: %w", err)
	}

	return repairOutcome{repair: WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairRemoveOrphanedWorktree,
		Detail: fmt.Sprintf("removed empty orphaned worktree path %s", diagnostic.Path),
	}, applied: true}, nil
}

func planWorkReadinessRepairs(report WorkReadinessReport) WorkReadinessReport {
	for index := range report.Diagnostics {
		report.Diagnostics[index].Repair = plannedWorkReadinessRepair(report.Policy, report.Diagnostics[index])
	}

	return report
}

func plannedWorkReadinessRepair(policy WorkBranchPolicy, diagnostic WorkReadinessDiagnostic) WorkReadinessRepair {
	if !policy.AllowSafeRepair {
		return manualWorkReadinessRepair("safe repair is disabled by work policy")
	}

	if repair, ok := branchWorkReadinessRepair(policy, diagnostic); ok {
		return repair
	}

	if repair, ok := workspaceWorkReadinessRepair(diagnostic); ok {
		return repair
	}

	return manualWorkReadinessRepair("manual intervention required")
}

func branchWorkReadinessRepair(policy WorkBranchPolicy, diagnostic WorkReadinessDiagnostic) (WorkReadinessRepair, bool) {
	switch diagnostic.Code {
	case WorkReadinessDiagnosticWrongBranch:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairSwitchBranch,
			Safe:      true,
			Detail:    fmt.Sprintf("switch to existing branch %s", diagnostic.ExpectedBranch),
		}, true
	case WorkReadinessDiagnosticMissingBranch:
		if policy.AllowBranchCreation {
			return WorkReadinessRepair{
				Available: true,
				Code:      WorkReadinessRepairCreateBranch,
				Safe:      true,
				Detail:    fmt.Sprintf("create branch %s from the current branch", diagnostic.ExpectedBranch),
			}, true
		}

		return manualWorkReadinessRepair("branch creation is disabled by work policy"), true
	case WorkReadinessDiagnosticMissingRemoteRef:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairFetchRemote,
			Safe:      true,
			Detail:    fmt.Sprintf("fetch configured remote %s", diagnostic.Remote),
		}, true
	case WorkReadinessDiagnosticRemoteBaseFetchFailed:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairFetchRemote,
			Safe:      true,
			Detail:    fmt.Sprintf("fetch configured remote %s", diagnostic.Remote),
		}, true
	case WorkReadinessDiagnosticBranchBehindRemote:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairFastForwardBranch,
			Safe:      true,
			Detail:    fmt.Sprintf("fast-forward %s from %s", diagnostic.ExpectedBranch, diagnostic.RemoteRef),
		}, true
	default:
		return WorkReadinessRepair{}, false
	}
}

func workspaceWorkReadinessRepair(diagnostic WorkReadinessDiagnostic) (WorkReadinessRepair, bool) {
	switch diagnostic.Code {
	case WorkReadinessDiagnosticRuntimeArtifactsVisible:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairIgnoreRuntimeArtifacts,
			Safe:      true,
			Detail:    "exclude chr runtime artifacts from local Git status",
		}, true
	case WorkReadinessDiagnosticWorktreeRootMissing:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairCreateWorktreeRoot,
			Safe:      true,
			Detail:    fmt.Sprintf("create worktree root %s", diagnostic.Path),
		}, true
	case WorkReadinessDiagnosticStaleWorktreeMetadata:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairPruneWorktreeMetadata,
			Safe:      true,
			Detail:    fmt.Sprintf("remove stale worktree metadata %s", diagnostic.Path),
		}, true
	case WorkReadinessDiagnosticOrphanedTicketWorktree:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairRemoveOrphanedWorktree,
			Safe:      true,
			Detail:    fmt.Sprintf("remove empty orphaned worktree path %s", diagnostic.Path),
		}, true
	default:
		return WorkReadinessRepair{}, false
	}
}

func manualWorkReadinessRepair(detail string) WorkReadinessRepair {
	return WorkReadinessRepair{
		Available: false,
		Code:      WorkReadinessRepairManual,
		Safe:      false,
		Detail:    detail,
	}
}
