package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WorkReadinessOptions describes a package boundary value.
type WorkReadinessOptions struct {
	ProjectDir   string
	Policy       WorkBranchPolicy
	Workspace    WorkWorkspaceConfig
	TicketKey    string
	TicketBranch string
	Git          WorkReadinessGit
	WorktreeGit  WorktreeGit
}

// WorkReadinessRepairOptions describes a package boundary value.
type WorkReadinessRepairOptions struct {
	ProjectDir   string
	Policy       WorkBranchPolicy
	Workspace    WorkWorkspaceConfig
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
	err := ctx.Err()
	if err != nil {
		return WorkReadinessReport{}, err
	}

	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	policy := opts.Policy
	if policy == (WorkBranchPolicy{}) {
		policy = DefaultWorkBranchPolicy()
	}

	workspace := opts.Workspace
	if workspace == (WorkWorkspaceConfig{}) {
		workspace = DefaultWorkWorkspaceConfig()
	} else {
		workspace = fillWorkWorkspaceConfigDefaults(workspace)
	}

	report := WorkReadinessReport{
		ProjectDir:   projectDir,
		Policy:       policy,
		Workspace:    workspace,
		Status:       WorkReadinessStatusReady,
		TicketKey:    strings.TrimSpace(opts.TicketKey),
		TicketBranch: strings.TrimSpace(opts.TicketBranch),
	}

	err = ValidateWorkBranchPolicy(policy)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, policyInvalidDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	err = ValidateWorkWorkspaceConfig(workspace)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, policyInvalidDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	if policy.Mode == WorkBranchPolicyModeNoGitEnforcement {
		return report, nil
	}

	git := opts.Git
	if git == nil {
		git = OSWorkReadinessGit{}
	}

	worktreeGit := opts.WorktreeGit
	if worktreeGit == nil {
		worktreeGit = OSWorktreeGit{}
	}

	return checkWorkReadinessGitState(ctx, projectDir, git, worktreeGit, report)
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
	projectDir   string
	policy       WorkBranchPolicy
	workspace    WorkWorkspaceConfig
	ticketKey    string
	ticketBranch string
	git          WorkReadinessRepairGit
	worktreeGit  WorktreeGit
}

func newWorkReadinessRepairSettings(ctx context.Context, opts WorkReadinessRepairOptions) (workReadinessRepairSettings, error) {
	err := ctx.Err()
	if err != nil {
		return workReadinessRepairSettings{}, err
	}

	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return workReadinessRepairSettings{}, err
	}

	policy := opts.Policy
	if policy == (WorkBranchPolicy{}) {
		policy = DefaultWorkBranchPolicy()
	}

	workspace := opts.Workspace
	if workspace == (WorkWorkspaceConfig{}) {
		workspace = DefaultWorkWorkspaceConfig()
	} else {
		workspace = fillWorkWorkspaceConfigDefaults(workspace)
	}

	git := opts.Git
	if git == nil {
		git = OSWorkReadinessGit{}
	}

	worktreeGit := opts.WorktreeGit
	if worktreeGit == nil {
		worktreeGit = OSWorktreeGit{}
	}

	return workReadinessRepairSettings{
		projectDir:   projectDir,
		policy:       policy,
		workspace:    workspace,
		ticketKey:    strings.TrimSpace(opts.TicketKey),
		ticketBranch: strings.TrimSpace(opts.TicketBranch),
		git:          git,
		worktreeGit:  worktreeGit,
	}, nil
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
	default:
		return WorkReadinessAppliedRepair{}, false, nil
	}
}

func applyRuntimeArtifactsRepair(ctx context.Context, run repairRun) (WorkReadinessAppliedRepair, bool, error) {
	changed, err := run.git.ExcludeLocalPattern(ctx, run.projectDir, ".chc/user/runtime/")
	if err != nil {
		return WorkReadinessAppliedRepair{}, false, fmt.Errorf("exclude runtime artifacts: %w", err)
	}

	return WorkReadinessAppliedRepair{
		Code:   WorkReadinessRepairIgnoreRuntimeArtifacts,
		Detail: "excluded .chc/user/runtime/ in local Git info exclude",
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

func checkWorkReadinessGitState(
	ctx context.Context,
	projectDir string,
	git WorkReadinessGit,
	worktreeGit WorktreeGit,
	report WorkReadinessReport,
) (WorkReadinessReport, error) {
	branch, err := git.CurrentBranch(ctx, projectDir)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	report.CurrentBranch = branch

	state, inProgress, err := git.GitStateInProgress(ctx, projectDir)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	if inProgress {
		report.Diagnostics = append(report.Diagnostics, gitStateInProgressDiagnostic(state))
	}

	report, err = checkWorkReadinessBranch(ctx, projectDir, git, report)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	report, err = checkWorkReadinessWorktree(ctx, projectDir, git, report)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	report, err = checkWorkReadinessRemote(ctx, projectDir, git, report)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	report, err = checkWorkReadinessTicketWorktree(ctx, projectDir, worktreeGit, report)
	if err != nil {
		return WorkReadinessReport{}, err
	}

	return finalizeWorkReadinessReport(report), nil
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

func checkWorkReadinessTicketWorktree(
	ctx context.Context,
	projectDir string,
	git WorktreeGit,
	report WorkReadinessReport,
) (WorkReadinessReport, error) {
	if report.Workspace.Mode != WorkWorkspaceModeWorktree {
		return report, nil
	}

	if strings.TrimSpace(report.TicketKey) == "" || strings.TrimSpace(report.TicketBranch) == "" {
		return report, nil
	}

	inspection, err := WorktreeService{Git: git}.Inspect(ctx, WorktreePlanOptions{
		ProjectDir:   projectDir,
		TicketKey:    report.TicketKey,
		Branch:       report.TicketBranch,
		BranchPolicy: report.Policy,
		Workspace:    report.Workspace,
	})
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return report, nil
	}

	report.TicketWorktree = workReadinessStateFromInspection(inspection)
	switch inspection.State {
	case WorktreeStateDirty:
		report.Diagnostics = append(report.Diagnostics, dirtyTicketWorktreeDiagnostic(inspection))
	case WorktreeStateMismatched, WorktreeStateBranchInUse, WorktreeStateOrphaned:
		report.Diagnostics = append(report.Diagnostics, unsafeTicketWorktreeDiagnostic(inspection))
	}

	return report, nil
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

	switch diagnostic.Code {
	case WorkReadinessDiagnosticWrongBranch:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairSwitchBranch,
			Safe:      true,
			Detail:    fmt.Sprintf("switch to existing branch %s", diagnostic.ExpectedBranch),
		}
	case WorkReadinessDiagnosticMissingBranch:
		if policy.AllowBranchCreation {
			return WorkReadinessRepair{
				Available: true,
				Code:      WorkReadinessRepairCreateBranch,
				Safe:      true,
				Detail:    fmt.Sprintf("create branch %s from the current branch", diagnostic.ExpectedBranch),
			}
		}

		return manualWorkReadinessRepair("branch creation is disabled by work policy")
	case WorkReadinessDiagnosticRuntimeArtifactsVisible:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairIgnoreRuntimeArtifacts,
			Safe:      true,
			Detail:    "exclude ChrysoCoord runtime artifacts from local Git status",
		}
	case WorkReadinessDiagnosticMissingRemoteRef:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairFetchRemote,
			Safe:      true,
			Detail:    fmt.Sprintf("fetch configured remote %s", diagnostic.Remote),
		}
	case WorkReadinessDiagnosticBranchBehindRemote:
		return WorkReadinessRepair{
			Available: true,
			Code:      WorkReadinessRepairFastForwardBranch,
			Safe:      true,
			Detail:    fmt.Sprintf("fast-forward %s from %s", diagnostic.ExpectedBranch, diagnostic.RemoteRef),
		}
	default:
		return manualWorkReadinessRepair("manual intervention required")
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
	return strings.Contains(line, ".chc/user/runtime/")
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

func runtimeArtifactsVisibleDiagnostic(mode WorkWorkspaceMode, count int) WorkReadinessDiagnostic {
	severity := WorkReadinessSeverityBlocking
	if mode == WorkWorkspaceModeWorktree {
		severity = WorkReadinessSeverityInfo
	}

	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticRuntimeArtifactsVisible,
		Severity: severity,
		Detail:   fmt.Sprintf("chrysocoord runtime artifacts are visible to git: %d", count),
	}
}

func dirtyWorktreeDiagnostic(count int) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticDirtyWorktree,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   fmt.Sprintf("worktree has user-visible changes: %d", count),
	}
}

func userCheckoutDirtyDiagnostic(mode WorkWorkspaceMode, count int) WorkReadinessDiagnostic {
	if mode != WorkWorkspaceModeWorktree {
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
