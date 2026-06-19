package core

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckWorkReadinessReportsMissingRemoteRef(t *testing.T) {
	report, err := CheckWorkReadiness(context.Background(), WorkReadinessOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Git: fakeWorkReadinessGit{
			branch:            "dev",
			cleanStatus:       nil,
			branchExists:      true,
			remoteRefExists:   false,
			branchContainsRef: true,
		},
	})
	if err != nil {
		t.Fatalf("CheckWorkReadiness returned error: %v", err)
	}

	assertWorkReadinessDiagnostic(t, report, WorkReadinessDiagnosticMissingRemoteRef)
}

func TestCheckWorkReadinessReportsBranchBehindRemote(t *testing.T) {
	report, err := CheckWorkReadiness(context.Background(), WorkReadinessOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Git: fakeWorkReadinessGit{
			branch:            "dev",
			cleanStatus:       nil,
			branchExists:      true,
			remoteRefExists:   true,
			branchContainsRef: false,
		},
	})
	if err != nil {
		t.Fatalf("CheckWorkReadiness returned error: %v", err)
	}

	assertWorkReadinessDiagnostic(t, report, WorkReadinessDiagnosticBranchBehindRemote)
}

func TestCheckWorkReadinessReportsWrongBranch(t *testing.T) {
	report, err := CheckWorkReadiness(context.Background(), WorkReadinessOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Git: fakeWorkReadinessGit{
			branch:            "main",
			branchExists:      true,
			remoteRefExists:   true,
			branchContainsRef: true,
		},
	})
	if err != nil {
		t.Fatalf("CheckWorkReadiness returned error: %v", err)
	}

	assertWorkReadinessDiagnostic(t, report, WorkReadinessDiagnosticWrongBranch)
}

func TestCheckWorkReadinessReportsRuntimeArtifactsSeparately(t *testing.T) {
	report, err := CheckWorkReadiness(context.Background(), WorkReadinessOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Git: fakeWorkReadinessGit{
			branch:            "dev",
			cleanStatus:       []string{"?? .chc/user/runtime/", " M main.go"},
			branchExists:      true,
			remoteRefExists:   true,
			branchContainsRef: true,
		},
	})
	if err != nil {
		t.Fatalf("CheckWorkReadiness returned error: %v", err)
	}

	assertWorkReadinessDiagnostic(t, report, WorkReadinessDiagnosticRuntimeArtifactsVisible)
	assertWorkReadinessDiagnostic(t, report, WorkReadinessDiagnosticDirtyWorktree)
}

func TestCheckWorkReadinessReportsDirtyUserCheckoutInWorktreeMode(t *testing.T) {
	report, err := CheckWorkReadiness(context.Background(), WorkReadinessOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Workspace:  testWorktreeWorkspaceConfig(),
		Git: fakeWorkReadinessGit{
			branch:            "dev",
			cleanStatus:       []string{" M main.go"},
			branchExists:      true,
			remoteRefExists:   true,
			branchContainsRef: true,
		},
	})
	if err != nil {
		t.Fatalf("CheckWorkReadiness returned error: %v", err)
	}

	assertWorkReadinessDiagnosticSeverity(t, workReadinessDiagnosticAssertion{
		report:   report,
		code:     WorkReadinessDiagnosticDirtyUserCheckout,
		severity: WorkReadinessSeverityInfo,
	})

	if report.Status != WorkReadinessStatusReady {
		t.Fatalf("Status = %s, want ready", report.Status)
	}

	if report.UserCheckout.State != WorktreeStateDirty {
		t.Fatalf("UserCheckout.State = %s, want dirty", report.UserCheckout.State)
	}
}

func TestCheckWorkReadinessBlocksDirtyTicketWorktree(t *testing.T) {
	projectDir := t.TempDir()
	workspace := testWorktreeWorkspaceConfig()
	plan, err := NewWorktreePlan(WorktreePlanOptions{
		ProjectDir:   projectDir,
		TicketKey:    "CHC-1",
		Branch:       "feat/chc-1",
		BranchPolicy: DefaultWorkBranchPolicy(),
		Workspace:    workspace,
	})
	if err != nil {
		t.Fatalf("NewWorktreePlan returned error: %v", err)
	}

	worktreeGit := newFakeWorktreeGit()
	worktreeGit.entries = []WorktreeGitEntry{{Path: plan.WorktreePath, Branch: plan.Branch}}
	worktreeGit.statuses[plan.WorktreePath] = []string{" M main.go"}

	report, err := CheckWorkReadiness(context.Background(), WorkReadinessOptions{
		ProjectDir:   projectDir,
		Policy:       DefaultWorkBranchPolicy(),
		Workspace:    workspace,
		TicketKey:    "CHC-1",
		TicketBranch: "feat/chc-1",
		WorktreeGit:  worktreeGit,
		Git: fakeWorkReadinessGit{
			branch:            "dev",
			branchExists:      true,
			remoteRefExists:   true,
			branchContainsRef: true,
		},
	})
	if err != nil {
		t.Fatalf("CheckWorkReadiness returned error: %v", err)
	}

	assertWorkReadinessDiagnostic(t, report, WorkReadinessDiagnosticDirtyTicketWorktree)

	if report.TicketWorktree.State != WorktreeStateDirty {
		t.Fatalf("TicketWorktree.State = %s, want dirty", report.TicketWorktree.State)
	}
}

func TestCheckWorkReadinessNoGitEnforcementSkipsGit(t *testing.T) {
	report, err := CheckWorkReadiness(context.Background(), WorkReadinessOptions{
		ProjectDir: t.TempDir(),
		Policy: WorkBranchPolicy{
			Mode: WorkBranchPolicyModeNoGitEnforcement,
		},
		Git: fakeWorkReadinessGit{err: errors.New("git should not be called")},
	})
	if err != nil {
		t.Fatalf("CheckWorkReadiness returned error: %v", err)
	}

	if report.Status != WorkReadinessStatusReady {
		t.Fatalf("Status = %s, want ready", report.Status)
	}

	if len(report.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %#v, want none", report.Diagnostics)
	}
}

func TestRepairWorkReadinessIgnoresRuntimeArtifacts(t *testing.T) {
	git := &fakeWorkReadinessRepairGit{
		branch:            "dev",
		cleanStatus:       []string{"?? .chc/user/runtime/runs/"},
		branchExists:      true,
		remoteRefExists:   true,
		branchContainsRef: true,
	}

	result, err := RepairWorkReadiness(context.Background(), WorkReadinessRepairOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Git:        git,
	})
	if err != nil {
		t.Fatalf("RepairWorkReadiness returned error: %v", err)
	}

	assertAppliedWorkReadinessRepair(t, result, WorkReadinessRepairIgnoreRuntimeArtifacts)

	if result.After.Status != WorkReadinessStatusReady {
		t.Fatalf("After.Status = %s, want ready", result.After.Status)
	}

	result, err = RepairWorkReadiness(context.Background(), WorkReadinessRepairOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Git:        git,
	})
	if err != nil {
		t.Fatalf("RepairWorkReadiness second run returned error: %v", err)
	}

	if len(result.Applied) != 0 {
		t.Fatalf("Applied = %#v, want no idempotent repairs", result.Applied)
	}
}

func TestRepairWorkReadinessFetchesThenFastForwardsBranch(t *testing.T) {
	git := &fakeWorkReadinessRepairGit{
		branch:            "dev",
		branchExists:      true,
		remoteRefExists:   false,
		branchContainsRef: false,
	}

	result, err := RepairWorkReadiness(context.Background(), WorkReadinessRepairOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Git:        git,
	})
	if err != nil {
		t.Fatalf("RepairWorkReadiness returned error: %v", err)
	}

	assertAppliedWorkReadinessRepair(t, result, WorkReadinessRepairFetchRemote)
	assertAppliedWorkReadinessRepair(t, result, WorkReadinessRepairFastForwardBranch)

	if result.After.Status != WorkReadinessStatusReady {
		t.Fatalf("After.Status = %s, want ready", result.After.Status)
	}
}

func TestRepairWorkReadinessCreatesAllowedMissingBranch(t *testing.T) {
	policy := DefaultWorkBranchPolicy()
	policy.RequireRemoteTracking = false
	policy.AllowBranchCreation = true
	git := &fakeWorkReadinessRepairGit{
		branch:       "main",
		branchExists: false,
	}

	result, err := RepairWorkReadiness(context.Background(), WorkReadinessRepairOptions{
		ProjectDir: t.TempDir(),
		Policy:     policy,
		Git:        git,
	})
	if err != nil {
		t.Fatalf("RepairWorkReadiness returned error: %v", err)
	}

	assertAppliedWorkReadinessRepair(t, result, WorkReadinessRepairCreateBranch)

	if result.After.Status != WorkReadinessStatusReady {
		t.Fatalf("After.Status = %s, want ready", result.After.Status)
	}
}

func TestRepairWorkReadinessRefusesBranchMutationWithUserChanges(t *testing.T) {
	git := &fakeWorkReadinessRepairGit{
		branch:            "main",
		cleanStatus:       []string{" M main.go"},
		branchExists:      true,
		remoteRefExists:   true,
		branchContainsRef: true,
	}

	result, err := RepairWorkReadiness(context.Background(), WorkReadinessRepairOptions{
		ProjectDir: t.TempDir(),
		Policy:     DefaultWorkBranchPolicy(),
		Git:        git,
	})
	if err != nil {
		t.Fatalf("RepairWorkReadiness returned error: %v", err)
	}

	if len(result.Applied) != 0 {
		t.Fatalf("Applied = %#v, want no branch repair with user changes", result.Applied)
	}

	if git.branch != "main" {
		t.Fatalf("branch = %s, want main", git.branch)
	}
}

func TestOSWorkReadinessGitExcludeLocalPatternIsIdempotent(t *testing.T) {
	projectDir := t.TempDir()
	runWorkReadinessTestGit(t, projectDir, "init")

	git := OSWorkReadinessGit{}
	for range 2 {
		_, err := git.ExcludeLocalPattern(context.Background(), projectDir, ".chc/user/runtime/")
		if err != nil {
			t.Fatalf("ExcludeLocalPattern returned error: %v", err)
		}
	}

	content, err := os.ReadFile(filepath.Join(projectDir, ".git", "info", "exclude"))
	if err != nil {
		t.Fatalf("read local exclude: %v", err)
	}

	if count := strings.Count(string(content), ".chc/user/runtime/"); count != 1 {
		t.Fatalf("exclude content = %q, want one runtime pattern", string(content))
	}
}

func assertWorkReadinessDiagnostic(t *testing.T, report WorkReadinessReport, code WorkReadinessDiagnosticCode) {
	t.Helper()

	if report.Status != WorkReadinessStatusBlocked {
		t.Fatalf("Status = %s, want blocked", report.Status)
	}

	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Code == code {
			return
		}
	}

	t.Fatalf("Diagnostics = %#v, want code %s", report.Diagnostics, code)
}

type workReadinessDiagnosticAssertion struct {
	report   WorkReadinessReport
	code     WorkReadinessDiagnosticCode
	severity WorkReadinessSeverity
}

func assertWorkReadinessDiagnosticSeverity(t *testing.T, assertion workReadinessDiagnosticAssertion) {
	t.Helper()

	for _, diagnostic := range assertion.report.Diagnostics {
		if diagnostic.Code == assertion.code {
			if diagnostic.Severity != assertion.severity {
				t.Fatalf("%s severity = %s, want %s", assertion.code, diagnostic.Severity, assertion.severity)
			}

			return
		}
	}

	t.Fatalf("Diagnostics = %#v, want code %s", assertion.report.Diagnostics, assertion.code)
}

func assertAppliedWorkReadinessRepair(t *testing.T, result WorkReadinessRepairResult, code WorkReadinessRepairCode) {
	t.Helper()

	for _, repair := range result.Applied {
		if repair.Code == code {
			return
		}
	}

	t.Fatalf("Applied = %#v, want code %s", result.Applied, code)
}

func testWorktreeWorkspaceConfig() WorkWorkspaceConfig {
	return WorkWorkspaceConfig{
		Mode:            WorkWorkspaceModeWorktree,
		WorktreeRoot:    DefaultWorktreeRoot,
		WorktreeBase:    WorktreeBasePolicyRemoteIntegrationBranch,
		WorktreeCleanup: WorktreeCleanupPolicyAfterMerge,
	}
}

func runWorkReadinessTestGit(t *testing.T, projectDir string, args ...string) {
	t.Helper()

	commandArgs := append([]string{
		"-C", projectDir,
		"-c", "user.name=ChrysoCoord Test",
		"-c", "user.email=chc@example.test",
	}, args...)
	cmd := exec.Command("git", commandArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

type fakeWorkReadinessGit struct {
	branch            string
	cleanStatus       []string
	branchExists      bool
	remoteRefExists   bool
	branchContainsRef bool
	gitState          string
	gitStateActive    bool
	err               error
}

func (git fakeWorkReadinessGit) CurrentBranch(context.Context, string) (string, error) {
	if git.err != nil {
		return "", git.err
	}

	return git.branch, nil
}

func (git fakeWorkReadinessGit) WorktreeStatus(context.Context, string) ([]string, error) {
	if git.err != nil {
		return nil, git.err
	}

	return git.cleanStatus, nil
}

func (git fakeWorkReadinessGit) BranchExists(context.Context, string, string) (bool, error) {
	if git.err != nil {
		return false, git.err
	}

	return git.branchExists, nil
}

func (git fakeWorkReadinessGit) RefExists(context.Context, string, string) (bool, error) {
	if git.err != nil {
		return false, git.err
	}

	return git.remoteRefExists, nil
}

func (git fakeWorkReadinessGit) BranchContainsRef(context.Context, string, string, string) (bool, error) {
	if git.err != nil {
		return false, git.err
	}

	return git.branchContainsRef, nil
}

func (git fakeWorkReadinessGit) GitStateInProgress(context.Context, string) (string, bool, error) {
	if git.err != nil {
		return "", false, git.err
	}

	return git.gitState, git.gitStateActive, nil
}

type fakeWorkReadinessRepairGit struct {
	branch            string
	cleanStatus       []string
	branchExists      bool
	remoteRefExists   bool
	branchContainsRef bool
	gitState          string
	gitStateActive    bool
	excludedPatterns  []string
}

func (git *fakeWorkReadinessRepairGit) CurrentBranch(context.Context, string) (string, error) {
	return git.branch, nil
}

func (git *fakeWorkReadinessRepairGit) WorktreeStatus(context.Context, string) ([]string, error) {
	return git.cleanStatus, nil
}

func (git *fakeWorkReadinessRepairGit) BranchExists(context.Context, string, string) (bool, error) {
	return git.branchExists, nil
}

func (git *fakeWorkReadinessRepairGit) RefExists(context.Context, string, string) (bool, error) {
	return git.remoteRefExists, nil
}

func (git *fakeWorkReadinessRepairGit) BranchContainsRef(context.Context, string, string, string) (bool, error) {
	return git.branchContainsRef, nil
}

func (git *fakeWorkReadinessRepairGit) GitStateInProgress(context.Context, string) (string, bool, error) {
	return git.gitState, git.gitStateActive, nil
}

func (git *fakeWorkReadinessRepairGit) ExcludeLocalPattern(_ context.Context, _, pattern string) (bool, error) {
	for _, existing := range git.excludedPatterns {
		if existing == pattern {
			return false, nil
		}
	}

	_, userChanges := splitWorktreeStatus(git.cleanStatus)
	git.cleanStatus = userChanges
	git.excludedPatterns = append(git.excludedPatterns, pattern)

	return true, nil
}

func (git *fakeWorkReadinessRepairGit) FetchRemote(context.Context, string, string) error {
	git.remoteRefExists = true

	return nil
}

func (git *fakeWorkReadinessRepairGit) FastForwardBranch(context.Context, string, string, string) error {
	git.branchContainsRef = true

	return nil
}

func (git *fakeWorkReadinessRepairGit) SwitchBranch(_ context.Context, _, branch string) error {
	git.branch = branch

	return nil
}

func (git *fakeWorkReadinessRepairGit) CreateBranch(_ context.Context, _, branch string) error {
	git.branch = branch
	git.branchExists = true

	return nil
}
