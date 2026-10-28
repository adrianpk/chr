package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeMergerCheckpointGuards rejects altered checkpoint images and actual
// Git parents before a capacity claim, model dispatch or new budget charge.
func TestNativeMergerCheckpointGuards(t *testing.T) {
	for _, name := range []string{"image", "Git parent"} {
		t.Run(name, func(t *testing.T) {
			f, request, harness, state := nativeMergerAcceptanceFixture(t, "merger-checkpoint-guard", "merge")
			if name == "image" {
				path := filepath.Join(f.store.layout.Root(), request.Input.RepairCheckpointRef.Path)
				if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			} else {
				workflowIntegrationCommand(t, state.Query.Lane.Worktree, "checkout", "--force", "--detach", request.Input.Source)
			}

			before := nativeQueueImages(t, f.store.layout.Root())
			if _, err := ExecuteMergerRepair(t.Context(), f.store, request, harness, MergerRepairPorts{Provider: f.provider, Git: f.git}); err == nil {
				t.Fatal("unverified checkpoint dispatched")
			}

			if !workflowNativeJSONEqual(before, nativeQueueImages(t, f.store.layout.Root())) {
				t.Fatal("checkpoint rejection charged or mutated the queue")
			}
		})
	}
}

// TestNativeMergerCheckpointBinding prevents a valid checkpoint from another
// queue generation from being substituted into an otherwise sealed request.
func TestNativeMergerCheckpointBinding(t *testing.T) {
	f, request, harness, _ := nativeMergerAcceptanceFixture(t, "merger-checkpoint-binding", "merge")
	request.Input.CandidateRevision++

	input, err := workflowmerge.SealInput(request.Input, request.Input.AdmittedAt)
	if err != nil {
		t.Fatal(err)
	}

	request.Input = input
	if _, err := ExecuteMergerRepair(t.Context(), f.store, request, harness, MergerRepairPorts{Provider: f.provider, Git: f.git}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale checkpoint binding accepted", err)
	}
}
