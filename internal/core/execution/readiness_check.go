package execution

import (
	"context"
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

	operation, err := check.git.GitStateInProgress(ctx, check.projectDir)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, gitUnavailableDiagnostic(err))

		return finalizeWorkReadinessReport(report), nil
	}

	if operation.InProgress {
		report.Diagnostics = append(report.Diagnostics, gitStateInProgressDiagnostic(operation.Name))
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

	changes := splitWorktreeStatus(status)
	report.UserCheckout = WorkReadinessWorkspaceState{
		Path:        projectDir,
		State:       workReadinessStateForStatus(changes.userChanges),
		Branch:      report.CurrentBranch,
		StatusLines: append([]string(nil), status...),
	}

	if len(changes.runtimeArtifacts) > 0 {
		report.Diagnostics = append(report.Diagnostics, runtimeArtifactsVisibleDiagnostic(report.Workspace.Mode, len(changes.runtimeArtifacts)))
	}

	if len(changes.userChanges) > 0 {
		report.Diagnostics = append(report.Diagnostics, userCheckoutDirtyDiagnostic(report.Workspace.Mode, len(changes.userChanges)))
	}

	return report, nil
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
