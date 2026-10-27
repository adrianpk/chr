package tui

import (
	"testing"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestNewWorkReadinessViewProjectsWorkspaceState(t *testing.T) {
	report := executioncore.WorkReadinessReport{
		Status: executioncore.WorkReadinessStatusReady,
		Policy: executioncore.WorkBranchPolicy{
			Mode:              executioncore.WorkBranchPolicyModeIntegrationBranch,
			IntegrationBranch: "dev",
			CanonicalRemote:   "origin",
		},
		Workspace: workspacecore.ExecutionConfig{
			Mode: workspacecore.ExecutionWorktree,
		},
		CurrentBranch: "feature/user-work",
		TicketKey:     "CHR-1",
		TicketBranch:  "feat/chr-1",
		UserCheckout: executioncore.WorkReadinessWorkspaceState{
			Path:  "/repo",
			State: executioncore.WorktreeStateDirty,
		},
		TicketWorktree: executioncore.WorkReadinessWorkspaceState{
			Path:    "/repo/.chr/default/runtime/worktrees/chr-1",
			State:   executioncore.WorktreeStateClean,
			Branch:  "feat/chr-1",
			BaseRef: "origin/dev",
			BaseSHA: "base-sha",
		},
	}

	view := NewWorkReadinessView(report)

	if view.WorkspaceMode != "worktree" {
		t.Fatalf("WorkspaceMode = %q, want worktree", view.WorkspaceMode)
	}

	if view.UserCheckoutState != "dirty" || view.TicketWorktree.State != "clean" {
		t.Fatalf("view = %#v, want projected workspace states", view)
	}

	if view.TicketKey != "CHR-1" || view.TicketWorktree.BaseRef != "origin/dev" {
		t.Fatalf("view = %#v, want ticket and base metadata", view)
	}
}
