package core

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAgentOperationLockBlocksConcurrentOperation(t *testing.T) {
	projectDir := t.TempDir()

	lock, err := AcquireAgentOperationLock(context.Background(), AgentOperationLockOptions{
		ProjectDir: projectDir,
		Operation:  AgentOperationDevelopment,
		Owner:      "test",
	})
	if err != nil {
		t.Fatalf("AcquireAgentOperationLock returned error: %v", err)
	}

	defer func() {
		if err := lock.Release(); err != nil {
			t.Fatalf("Release returned error: %v", err)
		}
	}()

	_, err = AcquireAgentOperationLock(context.Background(), AgentOperationLockOptions{
		ProjectDir: projectDir,
		Operation:  AgentOperationReview,
		Owner:      "test",
	})
	if err == nil || !strings.Contains(err.Error(), "agent operation lock is already held") {
		t.Fatalf("second AcquireAgentOperationLock error = %v, want held lock", err)
	}
}

func TestWorkSelectionPreflightPrioritizesReviewRepair(t *testing.T) {
	tickets := []IndexedTicket{
		{Key: "CHR-2", Status: TicketStatusReadyForDev, ReviewState: "", ReviewResult: ""},
		{Key: "CHR-1", Status: TicketStatusReadyForDev, ReviewState: "changes_requested", ReviewResult: "needs_changes"},
	}

	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets:    tickets,
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatalf("Allowed = false, want true")
	}

	if len(result.SelectedTickets) != 1 || result.SelectedTickets[0].Key != "CHR-1" {
		t.Fatalf("SelectedTickets = %#v, want CHR-1 repair", result.SelectedTickets)
	}
}

func TestWorkSelectionPreflightBlocksLaterTicketWhenRepairPending(t *testing.T) {
	tickets := []IndexedTicket{
		{Key: "CHR-1", Status: TicketStatusReadyForDev, ReviewState: "changes_requested"},
		{Key: "CHR-2", Status: TicketStatusReadyForDev},
	}

	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir:        t.TempDir(),
		Tickets:           tickets,
		SelectedTicketKey: "CHR-2",
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || result.BlockReasons[0].Code != WorkSelectionBlockReviewRepairPending {
		t.Fatalf("BlockReasons = %#v, want review repair pending", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightBlocksWrongIntegrationBranch(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir:          t.TempDir(),
		Tickets:             []IndexedTicket{{Key: "CHR-1", Status: TicketStatusReadyForDev}},
		RequireBranchPolicy: true,
		Git: fakeWorkSelectionGit{
			branch:      "feat/slice-7-serialized-work-preflight",
			clean:       true,
			containsRef: true,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || result.BlockReasons[0].Code != WorkSelectionBlockBranchPolicy {
		t.Fatalf("BlockReasons = %#v, want branch policy blocked", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightBlocksIntegrationBranchBehindRemote(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir:          t.TempDir(),
		Tickets:             []IndexedTicket{{Key: "CHR-1", Status: TicketStatusReadyForDev}},
		RequireBranchPolicy: true,
		Git: fakeWorkSelectionGit{
			branch:      "dev",
			clean:       true,
			containsRef: false,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || result.BlockReasons[0].Code != WorkSelectionBlockBranchPolicy {
		t.Fatalf("BlockReasons = %#v, want branch policy blocked", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightAllowsConfiguredIntegrationBranch(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets:    []IndexedTicket{{Key: "CHR-1", Status: TicketStatusReadyForDev}},
		BranchPolicy: WorkBranchPolicy{
			Mode:                  WorkBranchPolicyModeIntegrationBranch,
			IntegrationBranch:     "develop",
			CanonicalRemote:       "upstream",
			RequireCleanWorktree:  true,
			RequireRemoteTracking: true,
		},
		Git: fakeWorkSelectionGit{
			branch:      "develop",
			clean:       true,
			containsRef: true,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatalf("Allowed = false, want true: %#v", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightBlocksConfiguredIntegrationBranchBehindRemote(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets:    []IndexedTicket{{Key: "CHR-1", Status: TicketStatusReadyForDev}},
		BranchPolicy: WorkBranchPolicy{
			Mode:                  WorkBranchPolicyModeIntegrationBranch,
			IntegrationBranch:     "develop",
			CanonicalRemote:       "upstream",
			RequireCleanWorktree:  true,
			RequireRemoteTracking: true,
		},
		Git: fakeWorkSelectionGit{
			branch:      "develop",
			clean:       true,
			containsRef: false,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || !strings.Contains(result.BlockReasons[0].Detail, "upstream/develop") {
		t.Fatalf("BlockReasons = %#v, want upstream/develop detail", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightCurrentBranchModeAllowsAnyCleanBranch(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets:    []IndexedTicket{{Key: "CHR-1", Status: TicketStatusReadyForDev}},
		BranchPolicy: WorkBranchPolicy{
			Mode:                 WorkBranchPolicyModeCurrentBranch,
			RequireCleanWorktree: true,
		},
		Git: fakeWorkSelectionGit{
			branch:      "feature/current-work",
			clean:       true,
			containsRef: false,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatalf("Allowed = false, want true: %#v", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightCurrentBranchModeRequiresCleanWorktree(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets:    []IndexedTicket{{Key: "CHR-1", Status: TicketStatusReadyForDev}},
		BranchPolicy: WorkBranchPolicy{
			Mode:                 WorkBranchPolicyModeCurrentBranch,
			RequireCleanWorktree: true,
		},
		Git: fakeWorkSelectionGit{
			branch: "feature/current-work",
			clean:  false,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || result.BlockReasons[0].Detail != "work selection requires a clean local worktree" {
		t.Fatalf("BlockReasons = %#v, want generic clean worktree detail", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightWorktreeModeAllowsDirtyUserCheckout(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets:    []IndexedTicket{{Key: "CHR-1", Status: TicketStatusReadyForDev}},
		BranchPolicy: WorkBranchPolicy{
			Mode:                  WorkBranchPolicyModeIntegrationBranch,
			IntegrationBranch:     "dev",
			CanonicalRemote:       "origin",
			RequireCleanWorktree:  true,
			RequireRemoteTracking: true,
		},
		Workspace: WorkWorkspaceConfig{
			Mode:            WorkWorkspaceModeWorktree,
			WorktreeRoot:    DefaultWorktreeRoot,
			WorktreeBase:    WorktreeBasePolicyRemoteIntegrationBranch,
			WorktreeCleanup: WorktreeCleanupPolicyAfterMerge,
		},
		Git: fakeWorkSelectionGit{
			branch:      "feature/user-work",
			clean:       false,
			containsRef: false,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatalf("Allowed = false, want true: %#v", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightNoGitEnforcementSkipsGitChecks(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets:    []IndexedTicket{{Key: "CHR-1", Status: TicketStatusReadyForDev}},
		BranchPolicy: WorkBranchPolicy{
			Mode: WorkBranchPolicyModeNoGitEnforcement,
		},
		Git: failingWorkSelectionGit{},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatalf("Allowed = false, want true: %#v", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightNoGitEnforcementKeepsReviewRepairOrdering(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets: []IndexedTicket{
			{Key: "CHR-1", Status: TicketStatusReadyForDev, ReviewState: "changes_requested"},
			{Key: "CHR-2", Status: TicketStatusReadyForDev},
		},
		SelectedTicketKey: "CHR-2",
		BranchPolicy: WorkBranchPolicy{
			Mode: WorkBranchPolicyModeNoGitEnforcement,
		},
		Git: failingWorkSelectionGit{},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || result.BlockReasons[0].Code != WorkSelectionBlockReviewRepairPending {
		t.Fatalf("BlockReasons = %#v, want review repair pending", result.BlockReasons)
	}
}

type fakeWorkSelectionGit struct {
	branch      string
	clean       bool
	containsRef bool
}

func (git fakeWorkSelectionGit) CurrentBranch(context.Context, string) (string, error) {
	return git.branch, nil
}

func (git fakeWorkSelectionGit) WorktreeClean(context.Context, string) (bool, error) {
	return git.clean, nil
}

func (git fakeWorkSelectionGit) BranchContainsRef(context.Context, string, string, string) (bool, error) {
	return git.containsRef, nil
}

type failingWorkSelectionGit struct{}

func (failingWorkSelectionGit) CurrentBranch(context.Context, string) (string, error) {
	return "", errors.New("current branch should not be read")
}

func (failingWorkSelectionGit) WorktreeClean(context.Context, string) (bool, error) {
	return false, errors.New("worktree should not be read")
}

func (failingWorkSelectionGit) BranchContainsRef(context.Context, string, string, string) (bool, error) {
	return false, errors.New("remote ref should not be read")
}
