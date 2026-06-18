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
	ProjectDir string
	Policy     WorkBranchPolicy
	Git        WorkReadinessGit
}

// WorkReadinessRepairOptions describes a package boundary value.
type WorkReadinessRepairOptions struct {
	ProjectDir string
	Policy     WorkBranchPolicy
	Git        WorkReadinessRepairGit
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

	report := WorkReadinessReport{
		ProjectDir: projectDir,
		Policy:     policy,
		Status:     WorkReadinessStatusReady,
	}

	err = ValidateWorkBranchPolicy(policy)
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

	return checkWorkReadinessGitState(ctx, projectDir, git, report)
}

// RepairWorkReadiness executes the corresponding package operation.
func RepairWorkReadiness(ctx context.Context, opts WorkReadinessRepairOptions) (WorkReadinessRepairResult, error) {
	err := ctx.Err()
	if err != nil {
		return WorkReadinessRepairResult{}, err
	}

	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return WorkReadinessRepairResult{}, err
	}

	policy := opts.Policy
	if policy == (WorkBranchPolicy{}) {
		policy = DefaultWorkBranchPolicy()
	}

	git := opts.Git
	if git == nil {
		git = OSWorkReadinessGit{}
	}

	before, err := CheckWorkReadiness(ctx, WorkReadinessOptions{
		ProjectDir: projectDir,
		Policy:     policy,
		Git:        git,
	})
	if err != nil {
		return WorkReadinessRepairResult{}, err
	}

	result := WorkReadinessRepairResult{Before: before, After: before}
	if !policy.AllowSafeRepair {
		return result, nil
	}

	result.Applied, err = applySafeWorkReadinessRepairs(ctx, projectDir, git, before.Diagnostics)
	if err != nil {
		return WorkReadinessRepairResult{}, err
	}

	result.After, err = CheckWorkReadiness(ctx, WorkReadinessOptions{
		ProjectDir: projectDir,
		Policy:     policy,
		Git:        git,
	})
	if err != nil {
		return WorkReadinessRepairResult{}, err
	}

	return result, nil
}

func applySafeWorkReadinessRepairs(
	ctx context.Context,
	projectDir string,
	git WorkReadinessRepairGit,
	diagnostics []WorkReadinessDiagnostic,
) ([]WorkReadinessAppliedRepair, error) {
	applied := make([]WorkReadinessAppliedRepair, 0)
	for _, diagnostic := range diagnostics {
		if !diagnostic.Repair.Available || !diagnostic.Repair.Safe {
			continue
		}

		repair, ok, err := applySafeWorkReadinessRepair(ctx, projectDir, git, diagnostic)
		if err != nil {
			return nil, err
		}

		if ok {
			applied = append(applied, repair)
		}
	}

	return applied, nil
}

func applySafeWorkReadinessRepair(
	ctx context.Context,
	projectDir string,
	git WorkReadinessRepairGit,
	diagnostic WorkReadinessDiagnostic,
) (WorkReadinessAppliedRepair, bool, error) {
	switch diagnostic.Repair.Code {
	case WorkReadinessRepairIgnoreRuntimeArtifacts:
		changed, err := git.ExcludeLocalPattern(ctx, projectDir, ".chc/user/runtime/")
		if err != nil {
			return WorkReadinessAppliedRepair{}, false, fmt.Errorf("exclude runtime artifacts: %w", err)
		}

		return WorkReadinessAppliedRepair{
			Code:   WorkReadinessRepairIgnoreRuntimeArtifacts,
			Detail: "excluded .chc/user/runtime/ in local Git info exclude",
		}, changed, nil
	default:
		return WorkReadinessAppliedRepair{}, false, nil
	}
}

func checkWorkReadinessGitState(ctx context.Context, projectDir string, git WorkReadinessGit, report WorkReadinessReport) (WorkReadinessReport, error) {
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
	if len(runtimeArtifacts) > 0 {
		report.Diagnostics = append(report.Diagnostics, runtimeArtifactsVisibleDiagnostic(len(runtimeArtifacts)))
	}

	if len(userChanges) > 0 {
		report.Diagnostics = append(report.Diagnostics, dirtyWorktreeDiagnostic(len(userChanges)))
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

func runtimeArtifactsVisibleDiagnostic(count int) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticRuntimeArtifactsVisible,
		Severity: WorkReadinessSeverityBlocking,
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
