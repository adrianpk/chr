package core

import (
	"context"
	"errors"
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
