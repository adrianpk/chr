package tui

import (
	"testing"

	"forge.adrianpk.com/chc/internal/core"
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
		TicketKey:     "CHC-1",
		TicketBranch:  "feat/chc-1",
		UserCheckout: core.WorkReadinessWorkspaceState{
			Path:  "/repo",
			State: core.WorktreeStateDirty,
		},
		TicketWorktree: core.WorkReadinessWorkspaceState{
			Path:    "/repo/.chc/user/runtime/worktrees/chc-1",
			State:   core.WorktreeStateClean,
			Branch:  "feat/chc-1",
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

	if view.TicketKey != "CHC-1" || view.TicketWorktree.BaseRef != "origin/dev" {
		t.Fatalf("view = %#v, want ticket and base metadata", view)
	}
}
