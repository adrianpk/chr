package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// WorkReadinessOptions describes a package boundary value.
type WorkReadinessOptions struct {
	ProjectDir   string
	Policy       WorkBranchPolicy
	Workspace    workspacecore.ExecutionConfig
	TicketKey    string
	TicketBranch string
	Git          WorkReadinessGit
	WorktreeGit  WorktreeGit
}

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

// WorkReadinessGit defines a package boundary interface.
type WorkReadinessGit interface {
	CurrentBranch(ctx context.Context, projectDir string) (string, error)
	WorktreeStatus(ctx context.Context, projectDir string) ([]string, error)
	BranchExists(ctx context.Context, projectDir, branch string) (bool, error)
	RefExists(ctx context.Context, projectDir, ref string) (bool, error)
	BranchContainsRef(ctx context.Context, projectDir, branch, ref string) (bool, error)
	GitStateInProgress(ctx context.Context, projectDir string) (string, bool, error)
}

// WorkReadinessRepairGit defines a package boundary interface.
type WorkReadinessRepairGit interface {
	WorkReadinessGit
	ExcludeLocalPattern(ctx context.Context, projectDir, pattern string) (bool, error)
	FetchRemote(ctx context.Context, projectDir, remote string) error
	FastForwardBranch(ctx context.Context, projectDir, branch, ref string) error
	SwitchBranch(ctx context.Context, projectDir, branch string) error
	CreateBranch(ctx context.Context, projectDir, branch string) error
}

// OSWorkReadinessGit describes a package boundary value.
type OSWorkReadinessGit struct{}

// CheckWorkReadiness executes the corresponding package operation.
func CheckWorkReadiness(ctx context.Context, opts WorkReadinessOptions) (WorkReadinessReport, error) {
	settings, err := newWorkReadinessSettings(ctx, opts)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	report := WorkReadinessReport{
		ProjectDir:   settings.projectDir,
		Policy:       settings.policy,
		Workspace:    settings.workspace,
		Status:       WorkReadinessStatusReady,
		TicketKey:    settings.ticketKey,
		TicketBranch: settings.ticketBranch,
	}

	if err := ValidateWorkBranchPolicy(settings.policy); err != nil {
		report.Diagnostics = append(report.Diagnostics, policyInvalidDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	if err := workspacecore.ValidateExecutionConfig(settings.workspace); err != nil {
		report.Diagnostics = append(report.Diagnostics, policyInvalidDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	if settings.policy.Mode == WorkBranchPolicyModeNoGitEnforcement {
		return report, nil
	}

	return workReadinessGitStateCheck{
		projectDir:  settings.projectDir,
		git:         settings.git,
		worktreeGit: settings.worktreeGit,
		report:      report,
	}.run(ctx)
}

type workReadinessSettings struct {
	projectDir   string
	policy       WorkBranchPolicy
	workspace    workspacecore.ExecutionConfig
	ticketKey    string
	ticketBranch string
	git          WorkReadinessGit
	worktreeGit  WorktreeGit
}

func newWorkReadinessSettings(ctx context.Context, opts WorkReadinessOptions) (workReadinessSettings, error) {
	err := ctx.Err()
	if err != nil {
		return workReadinessSettings{}, err
	}

	projectDir, err := authoringcore.CleanContextProjectDir(opts.ProjectDir)
	if err != nil {
		return workReadinessSettings{}, err
	}

	policy := opts.Policy
	if policy == (WorkBranchPolicy{}) {
		policy = DefaultWorkBranchPolicy()
	}

	workspace := opts.Workspace
	if workspace == (workspacecore.ExecutionConfig{}) {
		workspace = workspacecore.DefaultExecutionConfig()
	} else {
		workspace = workspacecore.FillExecutionDefaults(workspace)
	}

	git := opts.Git
	if git == nil {
		git = OSWorkReadinessGit{}
	}

	worktreeGit := opts.WorktreeGit
	if worktreeGit == nil {
		worktreeGit = OSWorktreeGit{}
	}

	return workReadinessSettings{
		projectDir:   projectDir,
		policy:       policy,
		workspace:    workspace,
		ticketKey:    strings.TrimSpace(opts.TicketKey),
		ticketBranch: strings.TrimSpace(opts.TicketBranch),
		git:          git,
		worktreeGit:  worktreeGit,
	}, nil
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

		repair, ok, err := applySafeRepair(ctx, run, diagnostic)
		if err != nil {
			return nil, err
		}

		if ok {
			applied = append(applied, repair)
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

func applySafeRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (WorkReadinessAppliedRepair, bool, error) {
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
		return WorkReadinessAppliedRepair{}, false, nil
	}
}

func applyRuntimeArtifactsRepair(ctx context.Context, run repairRun) (WorkReadinessAppliedRepair, bool, error) {
	changed := false

	for _, pattern := range runtimeArtifactIgnorePatterns() {
		patternChanged, err := run.git.ExcludeLocalPattern(ctx, run.projectDir, pattern)
		if err != nil {
			return WorkReadinessAppliedRepair{}, false, fmt.Errorf("exclude runtime artifacts: %w", err)
		}

		changed = changed || patternChanged
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairIgnoreRuntimeArtifacts,
		Detail: "excluded runtime artifact roots in local Git info exclude",
	}, changed, nil
}

func applyFetchRemoteRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (WorkReadinessAppliedRepair, bool, error) {
	err := run.git.FetchRemote(ctx, run.projectDir, diagnostic.Remote)
	if err != nil {
		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("fetch remote: %w", err)
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairFetchRemote,
		Detail: fmt.Sprintf("fetched %s", diagnostic.Remote),
	}, true, nil
}

func applyFastForwardRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (WorkReadinessAppliedRepair, bool, error) {
	if !run.branchMutationSafe {
		return WorkReadinessAppliedRepair{}, false, nil
	}

	err := run.git.FastForwardBranch(ctx, run.projectDir, diagnostic.ExpectedBranch, diagnostic.RemoteRef)
	if err != nil {
		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("fast-forward branch: %w", err)
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairFastForwardBranch,
		Detail: fmt.Sprintf("fast-forwarded %s from %s", diagnostic.ExpectedBranch, diagnostic.RemoteRef),
	}, true, nil
}

func applySwitchBranchRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (WorkReadinessAppliedRepair, bool, error) {
	if !run.branchMutationSafe {
		return WorkReadinessAppliedRepair{}, false, nil
	}

	err := run.git.SwitchBranch(ctx, run.projectDir, diagnostic.ExpectedBranch)
	if err != nil {
		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("switch branch: %w", err)
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairSwitchBranch,
		Detail: fmt.Sprintf("switched to %s", diagnostic.ExpectedBranch),
	}, true, nil
}

func applyCreateBranchRepair(ctx context.Context, run repairRun, diagnostic WorkReadinessDiagnostic) (WorkReadinessAppliedRepair, bool, error) {
	if !run.branchMutationSafe {
		return WorkReadinessAppliedRepair{}, false, nil
	}

	err := run.git.CreateBranch(ctx, run.projectDir, diagnostic.ExpectedBranch)
	if err != nil {
		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("create branch: %w", err)
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairCreateBranch,
		Detail: fmt.Sprintf("created branch %s", diagnostic.ExpectedBranch),
	}, true, nil
}

func applyCreateWorktreeRootRepair(diagnostic WorkReadinessDiagnostic) (WorkReadinessAppliedRepair, bool, error) {
	if strings.TrimSpace(diagnostic.Path) == "" {
		return WorkReadinessAppliedRepair{}, false, nil
	}

	if err := os.MkdirAll(diagnostic.Path, 0o755); err != nil {
		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("create worktree root: %w", err)
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairCreateWorktreeRoot,
		Detail: fmt.Sprintf("created worktree root %s", diagnostic.Path),
	}, true, nil
}

func applyPruneWorktreeMetadataRepair(diagnostic WorkReadinessDiagnostic) (WorkReadinessAppliedRepair, bool, error) {
	if strings.TrimSpace(diagnostic.Path) == "" {
		return WorkReadinessAppliedRepair{}, false, nil
	}

	err := os.Remove(diagnostic.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return WorkReadinessAppliedRepair{}, false, nil
		}

		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("prune worktree metadata: %w", err)
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairPruneWorktreeMetadata,
		Detail: fmt.Sprintf("removed stale worktree metadata %s", diagnostic.Path),
	}, true, nil
}

func applyRemoveOrphanedWorktreeRepair(diagnostic WorkReadinessDiagnostic) (WorkReadinessAppliedRepair, bool, error) {
	if strings.TrimSpace(diagnostic.Path) == "" {
		return WorkReadinessAppliedRepair{}, false, nil
	}

	entries, err := os.ReadDir(diagnostic.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return WorkReadinessAppliedRepair{}, false, nil
		}

		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("inspect orphaned worktree path: %w", err)
	}

	if len(entries) > 0 {
		return WorkReadinessAppliedRepair{}, false, nil
	}

	if err := os.Remove(diagnostic.Path); err != nil {
		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("remove orphaned worktree path: %w", err)
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairRemoveOrphanedWorktree,
		Detail: fmt.Sprintf("removed empty orphaned worktree path %s", diagnostic.Path),
	}, true, nil
}

type workReadinessGitStateCheck struct {
	projectDir  string
	git         WorkReadinessGit
	worktreeGit WorktreeGit
	report      WorkReadinessReport
}

func (check workReadinessGitStateCheck) run(ctx context.Context) (WorkReadinessReport, error) {
	report := check.report

	branch, err := check.git.CurrentBranch(ctx, check.projectDir)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	report.CurrentBranch = branch

	state, inProgress, err := check.git.GitStateInProgress(ctx, check.projectDir)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	if inProgress {
		report.Diagnostics = append(report.Diagnostics, gitStateInProgressDiagnostic(state))
	}

	report, err = checkWorkReadinessWorktree(ctx, check.projectDir, check.git, report)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	report, err = check.baseReadiness(ctx, report)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	report, err = checkWorkReadinessTicketWorktree(ctx, workReadinessWorktreeCheck{
		projectDir: check.projectDir,
		git:        check.worktreeGit,
		report:     report,
	})
	if err != nil {
		return WorkReadinessReport{}, err
	}

	return finalizeWorkReadinessReport(report), nil
}

func (check workReadinessGitStateCheck) baseReadiness(ctx context.Context, report WorkReadinessReport) (WorkReadinessReport, error) {
	if report.Workspace.Mode == workspacecore.ExecutionWorktree {
		report = checkWorkReadinessWorktreeRoot(check.projectDir, report)

		return checkWorkReadinessRemoteBase(ctx, workReadinessWorktreeCheck{
			projectDir: check.projectDir,
			git:        check.worktreeGit,
			report:     report,
		})
	}

	report, err := checkWorkReadinessBranch(ctx, check.projectDir, check.git, report)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	return checkWorkReadinessRemote(ctx, check.projectDir, check.git, report)
}

func checkWorkReadinessWorktreeRoot(projectDir string, report WorkReadinessReport) WorkReadinessReport {
	root := projectcore.ResolvePath(projectDir, report.Workspace.WorktreeRoot)
	if pathExists(root) {
		return report
	}

	report.Diagnostics = append(report.Diagnostics, worktreeRootMissingDiagnostic(root))

	return report
}

func checkWorkReadinessBranch(ctx context.Context, projectDir string, git WorkReadinessGit, report WorkReadinessReport) (WorkReadinessReport, error) {
	policy := report.Policy
	if policy.Mode != WorkBranchPolicyModeIntegrationBranch {
		return report, nil
	}

	expectedBranch := strings.TrimSpace(policy.IntegrationBranch)
	if report.CurrentBranch == expectedBranch {
		return report, nil
	}

	exists, err := git.BranchExists(ctx, projectDir, expectedBranch)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return report, nil
	}

	if !exists {
		report.Diagnostics = append(report.Diagnostics, missingBranchDiagnostic(expectedBranch, report.CurrentBranch))

		return report, nil
	}

	report.Diagnostics = append(report.Diagnostics, wrongBranchDiagnostic(expectedBranch, report.CurrentBranch))

	return report, nil
}

func checkWorkReadinessWorktree(ctx context.Context, projectDir string, git WorkReadinessGit, report WorkReadinessReport) (WorkReadinessReport, error) {
	if !report.Policy.RequireCleanWorktree {
		return report, nil
	}

	status, err := git.WorktreeStatus(ctx, projectDir)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return report, nil
	}

	runtimeArtifacts, userChanges := splitWorktreeStatus(status)
	report.UserCheckout = WorkReadinessWorkspaceState{
		Path:        projectDir,
		State:       workReadinessStateForStatus(userChanges),
		Branch:      report.CurrentBranch,
		StatusLines: append([]string(nil), status...),
	}

	if len(runtimeArtifacts) > 0 {
		report.Diagnostics = append(report.Diagnostics, runtimeArtifactsVisibleDiagnostic(report.Workspace.Mode, len(runtimeArtifacts)))
	}

	if len(userChanges) > 0 {
		report.Diagnostics = append(report.Diagnostics, userCheckoutDirtyDiagnostic(report.Workspace.Mode, len(userChanges)))
	}

	return report, nil
}

func workReadinessStateForStatus(status []string) WorktreeState {
	if len(status) > 0 {
		return WorktreeStateDirty
	}

	return WorktreeStateClean
}

func checkWorkReadinessRemote(ctx context.Context, projectDir string, git WorkReadinessGit, report WorkReadinessReport) (WorkReadinessReport, error) {
	policy := report.Policy
	if policy.Mode != WorkBranchPolicyModeIntegrationBranch || !policy.RequireRemoteTracking {
		return report, nil
	}

	expectedBranch := strings.TrimSpace(policy.IntegrationBranch)
	remote := strings.TrimSpace(policy.CanonicalRemote)
	remoteRef := remote + "/" + expectedBranch

	exists, err := git.RefExists(ctx, projectDir, remoteRef)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return report, nil
	}

	if !exists {
		report.Diagnostics = append(report.Diagnostics, missingRemoteRefDiagnostic(remote, remoteRef, expectedBranch))

		return report, nil
	}

	contains, err := git.BranchContainsRef(ctx, projectDir, expectedBranch, remoteRef)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return report, nil
	}

	if !contains {
		report.Diagnostics = append(report.Diagnostics, branchBehindRemoteDiagnostic(remote, remoteRef, expectedBranch))
	}

	return report, nil
}

type workReadinessWorktreeCheck struct {
	projectDir string
	git        WorktreeGit
	report     WorkReadinessReport
}

func checkWorkReadinessRemoteBase(ctx context.Context, check workReadinessWorktreeCheck) (WorkReadinessReport, error) {
	report := check.report
	if report.Workspace.WorktreeBase != workspacecore.BaseRemoteIntegrationBranch {
		return report, nil
	}

	remote := strings.TrimSpace(report.Policy.CanonicalRemote)

	baseRef, err := worktreeBaseRef(report.Policy, report.Workspace)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, policyInvalidDiagnostic(err))

		return report, nil
	}

	if err := check.git.FetchRemote(ctx, check.projectDir, remote); err != nil {
		report.Diagnostics = append(report.Diagnostics, remoteBaseFetchFailedDiagnostic(remote, baseRef, err))

		return report, nil
	}

	baseSHA, resolved := check.resolveBase(ctx, baseRef)
	if !resolved {
		report.Diagnostics = append(
			report.Diagnostics,
			missingRemoteRefDiagnostic(remote, baseRef, report.Policy.IntegrationBranch),
		)

		return report, nil
	}

	report.TicketWorktree.BaseRef = baseRef
	report.TicketWorktree.BaseSHA = baseSHA

	return report, nil
}

func (check workReadinessWorktreeCheck) resolveBase(ctx context.Context, baseRef string) (string, bool) {
	baseSHA, err := check.git.ResolveRef(ctx, check.projectDir, baseRef)
	if err != nil {
		return "", false
	}

	return baseSHA, true
}

func checkWorkReadinessTicketWorktree(ctx context.Context, check workReadinessWorktreeCheck) (WorkReadinessReport, error) {
	report := check.report
	if report.Workspace.Mode != workspacecore.ExecutionWorktree {
		return report, nil
	}

	if strings.TrimSpace(report.TicketKey) == "" || strings.TrimSpace(report.TicketBranch) == "" {
		return report, nil
	}

	inspection, err := WorktreeService{Git: check.git}.Inspect(ctx, WorktreePlanOptions{
		ProjectDir:   check.projectDir,
		TicketKey:    report.TicketKey,
		Branch:       report.TicketBranch,
		BranchPolicy: report.Policy,
		Workspace:    report.Workspace,
	})
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return report, nil
	}

	report.TicketWorktree = workReadinessStateWithBaseFallback(inspection, report.TicketWorktree)

	return addTicketWorktreeDiagnostic(report, inspection), nil
}

func workReadinessStateWithBaseFallback(inspection WorktreeInspection, fallback WorkReadinessWorkspaceState) WorkReadinessWorkspaceState {
	state := workReadinessStateFromInspection(inspection)
	if state.BaseRef == "" {
		state.BaseRef = fallback.BaseRef
	}

	if state.BaseSHA == "" {
		state.BaseSHA = fallback.BaseSHA
	}

	return state
}

func addTicketWorktreeDiagnostic(report WorkReadinessReport, inspection WorktreeInspection) WorkReadinessReport {
	switch inspection.State {
	case WorktreeStateDirty:
		report.Diagnostics = append(report.Diagnostics, dirtyTicketWorktreeDiagnostic(inspection))
	case WorktreeStateMismatched, WorktreeStateBranchInUse:
		report.Diagnostics = append(report.Diagnostics, unsafeTicketWorktreeDiagnostic(inspection))
	case WorktreeStateOrphaned:
		report.Diagnostics = append(report.Diagnostics, orphanedTicketWorktreeDiagnostic(inspection))
	case WorktreeStateMissing:
		report = checkStaleWorktreeMetadata(report, inspection)
	}

	return report
}

func checkStaleWorktreeMetadata(report WorkReadinessReport, inspection WorktreeInspection) WorkReadinessReport {
	path, err := WorktreeMetadataPath(inspection.Plan)
	if err != nil || !pathExists(path) {
		return report
	}

	report.Diagnostics = append(report.Diagnostics, staleWorktreeMetadataDiagnostic(path))

	return report
}

func workReadinessStateFromInspection(inspection WorktreeInspection) WorkReadinessWorkspaceState {
	return WorkReadinessWorkspaceState{
		Path:        inspection.Plan.WorktreePath,
		State:       inspection.State,
		Branch:      inspection.Plan.Branch,
		BaseRef:     inspection.Plan.BaseRef,
		BaseSHA:     inspection.Plan.BaseSHA,
		StatusLines: append([]string(nil), inspection.StatusLines...),
	}
}

func finalizeWorkReadinessReport(report WorkReadinessReport) WorkReadinessReport {
	report = planWorkReadinessRepairs(report)

	report.Status = WorkReadinessStatusReady
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Severity == WorkReadinessSeverityBlocking {
			report.Status = WorkReadinessStatusBlocked

			return report
		}
	}

	return report
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

func splitWorktreeStatus(lines []string) ([]string, []string) {
	runtimeArtifacts := make([]string, 0)
	userChanges := make([]string, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if worktreeStatusLineIsRuntimeArtifact(line) {
			runtimeArtifacts = append(runtimeArtifacts, line)

			continue
		}

		userChanges = append(userChanges, line)
	}

	return runtimeArtifacts, userChanges
}

func worktreeStatusLineIsRuntimeArtifact(line string) bool {
	return strings.Contains(line, ".chr/default/runtime/") || strings.Contains(line, "ops/default/run/")
}

func runtimeArtifactIgnorePatterns() []string {
	return []string{".chr/default/runtime/", "ops/*/run/"}
}

func policyInvalidDiagnostic(err error) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticPolicyInvalid,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   err.Error(),
	}
}

func gitUnavailableDiagnostic(err error) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticGitUnavailable,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   fmt.Sprintf("git state is unavailable: %v", err),
	}
}

func gitStateInProgressDiagnostic(state string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticGitStateInProgress,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   fmt.Sprintf("git operation in progress: %s", state),
	}
}

func wrongBranchDiagnostic(expectedBranch, currentBranch string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticWrongBranch,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         fmt.Sprintf("expected branch %s but current branch is %s", expectedBranch, currentBranch),
		CurrentBranch:  currentBranch,
		ExpectedBranch: expectedBranch,
	}
}

func missingBranchDiagnostic(expectedBranch, currentBranch string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticMissingBranch,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         fmt.Sprintf("expected branch %s does not exist", expectedBranch),
		CurrentBranch:  currentBranch,
		ExpectedBranch: expectedBranch,
	}
}

func runtimeArtifactsVisibleDiagnostic(mode workspacecore.ExecutionMode, count int) WorkReadinessDiagnostic {
	severity := WorkReadinessSeverityBlocking
	if mode == workspacecore.ExecutionWorktree {
		severity = WorkReadinessSeverityInfo
	}

	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticRuntimeArtifactsVisible,
		Severity: severity,
		Detail:   fmt.Sprintf("chr runtime artifacts are visible to git: %d", count),
	}
}

func dirtyWorktreeDiagnostic(count int) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticDirtyWorktree,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   fmt.Sprintf("worktree has user-visible changes: %d", count),
	}
}

func userCheckoutDirtyDiagnostic(mode workspacecore.ExecutionMode, count int) WorkReadinessDiagnostic {
	if mode != workspacecore.ExecutionWorktree {
		return dirtyWorktreeDiagnostic(count)
	}

	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticDirtyUserCheckout,
		Severity: WorkReadinessSeverityInfo,
		Detail:   fmt.Sprintf("user checkout has visible changes: %d", count),
	}
}

func dirtyTicketWorktreeDiagnostic(inspection WorktreeInspection) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticDirtyTicketWorktree,
		Severity: WorkReadinessSeverityBlocking,
		Detail: fmt.Sprintf(
			"ticket worktree %s has visible changes: %d",
			inspection.Plan.WorktreePath,
			len(inspection.StatusLines),
		),
		ExpectedBranch: inspection.Plan.Branch,
		RemoteRef:      inspection.Plan.BaseRef,
	}
}

func unsafeTicketWorktreeDiagnostic(inspection WorktreeInspection) WorkReadinessDiagnostic {
	detail := fmt.Sprintf("ticket worktree %s is %s", inspection.Plan.WorktreePath, inspection.State)
	if len(inspection.Diagnostics) > 0 {
		detail = inspection.Diagnostics[0].Detail
	}

	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticUnsafeTicketWorktree,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         detail,
		ExpectedBranch: inspection.Plan.Branch,
		RemoteRef:      inspection.Plan.BaseRef,
	}
}

func orphanedTicketWorktreeDiagnostic(inspection WorktreeInspection) WorkReadinessDiagnostic {
	detail := fmt.Sprintf("ticket worktree %s is orphaned", inspection.Plan.WorktreePath)
	if len(inspection.Diagnostics) > 0 {
		detail = inspection.Diagnostics[0].Detail
	}

	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticOrphanedTicketWorktree,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         detail,
		ExpectedBranch: inspection.Plan.Branch,
		RemoteRef:      inspection.Plan.BaseRef,
		Path:           inspection.Plan.WorktreePath,
	}
}

func worktreeRootMissingDiagnostic(path string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticWorktreeRootMissing,
		Severity: WorkReadinessSeverityInfo,
		Detail:   fmt.Sprintf("worktree root %s is missing", path),
		Path:     path,
	}
}

func staleWorktreeMetadataDiagnostic(path string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticStaleWorktreeMetadata,
		Severity: WorkReadinessSeverityInfo,
		Detail:   fmt.Sprintf("worktree metadata %s points to a missing worktree", path),
		Path:     path,
	}
}

func missingRemoteRefDiagnostic(remote, remoteRef, expectedBranch string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticMissingRemoteRef,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         fmt.Sprintf("remote tracking ref %s is missing", remoteRef),
		ExpectedBranch: expectedBranch,
		Remote:         remote,
		RemoteRef:      remoteRef,
	}
}

func branchBehindRemoteDiagnostic(remote, remoteRef, expectedBranch string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticBranchBehindRemote,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         fmt.Sprintf("branch %s does not contain %s", expectedBranch, remoteRef),
		ExpectedBranch: expectedBranch,
		Remote:         remote,
		RemoteRef:      remoteRef,
	}
}

func remoteBaseFetchFailedDiagnostic(remote, remoteRef string, err error) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:      WorkReadinessDiagnosticRemoteBaseFetchFailed,
		Severity:  WorkReadinessSeverityBlocking,
		Detail:    fmt.Sprintf("fetch %s failed before resolving %s: %v", remote, remoteRef, err),
		Remote:    remote,
		RemoteRef: remoteRef,
	}
}

// CurrentBranch executes the corresponding package operation.
func (OSWorkReadinessGit) CurrentBranch(ctx context.Context, projectDir string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "branch", "--show-current").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// WorktreeStatus executes the corresponding package operation.
func (OSWorkReadinessGit) WorktreeStatus(ctx context.Context, projectDir string) ([]string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "status", "--porcelain").Output()
	if err != nil {
		return nil, err
	}

	return workReadinessNonEmptyLines(string(output)), nil
}

// BranchExists executes the corresponding package operation.
func (OSWorkReadinessGit) BranchExists(ctx context.Context, projectDir, branch string) (bool, error) {
	return gitRefExists(ctx, projectDir, "refs/heads/"+branch)
}

// RefExists executes the corresponding package operation.
func (OSWorkReadinessGit) RefExists(ctx context.Context, projectDir, ref string) (bool, error) {
	return gitRefExists(ctx, projectDir, ref)
}

// BranchContainsRef executes the corresponding package operation.
func (OSWorkReadinessGit) BranchContainsRef(ctx context.Context, projectDir, branch, ref string) (bool, error) {
	err := exec.CommandContext(ctx, "git", "-C", projectDir, "merge-base", "--is-ancestor", ref, branch).Run()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, err
}

// GitStateInProgress executes the corresponding package operation.
func (OSWorkReadinessGit) GitStateInProgress(ctx context.Context, projectDir string) (string, bool, error) {
	for _, state := range []string{"MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "rebase-merge", "rebase-apply"} {
		path, err := gitPath(ctx, projectDir, state)
		if err != nil {
			return "", false, err
		}

		_, err = os.Stat(path)
		if err == nil {
			return state, true, nil
		}

		if !os.IsNotExist(err) {
			return "", false, err
		}
	}

	return "", false, nil
}

// ExcludeLocalPattern executes the corresponding package operation.
func (OSWorkReadinessGit) ExcludeLocalPattern(ctx context.Context, projectDir, pattern string) (bool, error) {
	err := ctx.Err()
	if err != nil {
		return false, err
	}

	path, err := gitPath(ctx, projectDir, "info/exclude")
	if err != nil {
		return false, err
	}

	return appendLineIfMissing(path, pattern)
}

// FetchRemote executes the corresponding package operation.
func (OSWorkReadinessGit) FetchRemote(ctx context.Context, projectDir, remote string) error {
	return runWorkReadinessGit(ctx, projectDir, "fetch", remote)
}

// FastForwardBranch executes the corresponding package operation.
func (git OSWorkReadinessGit) FastForwardBranch(ctx context.Context, projectDir, branch, ref string) error {
	current, err := git.CurrentBranch(ctx, projectDir)
	if err != nil {
		return err
	}

	if current != branch {
		return fmt.Errorf("current branch is %s, expected %s", current, branch)
	}

	return runWorkReadinessGit(ctx, projectDir, "merge", "--ff-only", ref)
}

// SwitchBranch executes the corresponding package operation.
func (OSWorkReadinessGit) SwitchBranch(ctx context.Context, projectDir, branch string) error {
	return runWorkReadinessGit(ctx, projectDir, "switch", branch)
}

// CreateBranch executes the corresponding package operation.
func (OSWorkReadinessGit) CreateBranch(ctx context.Context, projectDir, branch string) error {
	return runWorkReadinessGit(ctx, projectDir, "switch", "-c", branch)
}

func gitRefExists(ctx context.Context, projectDir, ref string) (bool, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--verify", ref).CombinedOutput()
	if err == nil {
		return true, nil
	}

	if gitCommandOutputLooksUnavailable(output) {
		return false, fmt.Errorf("verify git ref %s: %w", ref, err)
	}

	return false, nil
}

func gitPath(ctx context.Context, projectDir, path string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--git-path", path).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve git path %s: %w", path, err)
	}

	resolvedPath := strings.TrimSpace(string(output))
	if filepath.IsAbs(resolvedPath) {
		return resolvedPath, nil
	}

	return filepath.Join(projectDir, resolvedPath), nil
}

func gitCommandOutputLooksUnavailable(output []byte) bool {
	return strings.Contains(string(output), "not a git repository")
}

func runWorkReadinessGit(ctx context.Context, projectDir string, args ...string) error {
	commandArgs := append([]string{"-C", projectDir}, args...)

	output, err := exec.CommandContext(ctx, "git", commandArgs...).CombinedOutput()
	if err == nil {
		return nil
	}

	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, detail)
}

func appendLineIfMissing(path, value string) (bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return false, nil
	}

	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read %s: %w", path, err)
	}

	if lineExists(string(content), value) {
		return false, nil
	}

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return false, fmt.Errorf("create %s parent: %w", path, err)
	}

	next := string(content)
	if next != "" && !strings.HasSuffix(next, "\n") {
		next += "\n"
	}

	next += value + "\n"

	err = os.WriteFile(path, []byte(next), 0o644)
	if err != nil {
		return false, fmt.Errorf("write %s: %w", path, err)
	}

	return true, nil
}

func lineExists(content, value string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == value {
			return true
		}
	}

	return false
}

func workReadinessNonEmptyLines(value string) []string {
	lines := make([]string, 0)

	for _, line := range strings.Split(value, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func pathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
