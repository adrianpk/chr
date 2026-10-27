package core

import (
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// TestNativeMergerWorkspace exercises both Git object formats. The private
// checkout must import the exact source without changing the source repository.
func TestNativeMergerWorkspace(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			workflowIntegrationCommand(t, dir, "init", "--quiet", "--object-format="+format)

			layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: dir, Namespace: "default"})
			if err != nil {
				t.Fatal(err)
			}

			store := NewWorkflowStore(layout, "merger-workspace-fixture")
			seed, _ := workflowSelectionCommit(t, store)
			input := WorkflowMergerInput{WorkflowMergerIdentity: WorkflowMergerIdentity{InvocationID: "fixture", EntryID: "entry", ID: "format-" + format}}
			input.OwnedWorkspaceID = workflowMergerWorkspaceID(input)
			before := workflowIntegrationCommand(t, dir, "status", "--porcelain")

			lane, _, err := store.workflowMergerWorkspace(t.Context(), input, seed)
			if err != nil {
				t.Fatal(err)
			}

			if workflowIntegrationCommand(t, lane.Worktree, "rev-parse", "HEAD") != seed || workflowIntegrationCommand(t, lane.Worktree, "remote") != "" {
				t.Fatal("private checkout changed object format or acquired a remote")
			}

			if workflowIntegrationCommand(t, dir, "status", "--porcelain") != before || workflowIntegrationCommand(t, dir, "rev-parse", "HEAD") != seed {
				t.Fatal("private checkout changed the source repository")
			}
		})
	}
}
