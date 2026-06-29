package tui

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

func TestNewWorkReadinessViewProjectsWorkspaceState(t *testing.T) {
	report := core.WorkReadinessReport{
		Status: core.WorkReadinessStatusReady,
		Policy: core.WorkBranchPolicy{
			Mode:              core.WorkBranchPolicyModeIntegrationBranch,
			IntegrationBranch: "dev",
			CanonicalRemote:   "origin",
		},
		Workspace: core.WorkWorkspaceConfig{
			Mode: core.WorkWorkspaceModeWorktree,
		},
		CurrentBranch: "feature/user-work",
		TicketKey:     "CHR-1",
		TicketBranch:  "feat/chr-1",
		UserCheckout: core.WorkReadinessWorkspaceState{
			Path:  "/repo",
			State: core.WorktreeStateDirty,
		},
		TicketWorktree: core.WorkReadinessWorkspaceState{
			Path:    "/repo/.chr/user/runtime/worktrees/chr-1",
			State:   core.WorktreeStateClean,
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
