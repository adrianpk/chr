package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func workflowVerifiedDelivery(t *testing.T, method string) workflowIntegrationFixture {
	t.Helper()

	fixture := workflowDeliveryFixture(t, method)
	if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, method, fixture.provider, fixture.git); err != nil {
		t.Fatal(err)
	}

	return fixture
}

// TestWorkflowFinalization shares one receipt across the entire slice and keeps
// immutable task validation, the sole PR and exact finalization revisions on retry.
func TestWorkflowFinalization(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			fixture := workflowVerifiedDelivery(t, method)

			receipt, err := fixture.store.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, fixture.provider, fixture.git)
			if err != nil {
				t.Fatal(err)
			}

			checkWorkflowFinalizedSlice(t, fixture, receipt)

			before := workflowTestRead(t, fixture.store, fixture.run.ID)
			reads := fixture.provider.reads
			restarted := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

			again, err := restarted.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, nil, nil)
			if err != nil || again.SourceHash != receipt.SourceHash {
				t.Fatalf("finalized receipt was not reused: %v", err)
			}

			after := workflowTestRead(t, fixture.store, fixture.run.ID)
			if after.Revision != before.Revision || fixture.provider.reads != reads {
				t.Fatal("idempotent finalization rewrote state or repeated provider IO")
			}
		})
	}
}

func checkWorkflowFinalizedSlice(t *testing.T, fixture workflowIntegrationFixture, receipt Record) {
	t.Helper()

	run := workflowTestRead(t, fixture.store, fixture.run.ID)
	if run.Status != "completed" || run.Outcome != "slice_integrated" || run.DeliveryRef != receipt.Path || len(run.Attempts) != 4 || fixture.provider.creates != 1 {
		t.Fatal("finalization changed PR/worker count or did not complete the selected slice")
	}

	if receipt.Receipt.SourceCommit != fixture.run.Lane.HeadCommit || receipt.Receipt.BaseCommit != fixture.base || receipt.Receipt.ResultCommit != fixture.result || len(receipt.Receipt.ValidationRefs) != 4 {
		t.Fatal("receipt lacks common delivery/validation facts")
	}

	checkWorkflowFinalizedTasks(t, fixture, receipt)
	checkWorkflowFinalizedSet(t, fixture, receipt)
}

func checkWorkflowFinalizedTasks(t *testing.T, fixture workflowIntegrationFixture, receipt Record) {
	t.Helper()

	for _, task := range fixture.hierarchy.Tasks[:3] {
		record, err := fixture.store.Read(t.Context(), task.Path)
		if err != nil || !workflowFinalizedTaskMatches(record, fixture.run.Lane.Branch, receipt) || len(record.Task.Evidence) == 0 {
			t.Fatal("task finalization did not share the receipt and retained proof")
		}
	}
}

func checkWorkflowFinalizedSet(t *testing.T, fixture workflowIntegrationFixture, receipt Record) {
	t.Helper()

	set, err := fixture.store.Read(t.Context(), fixture.hierarchy.Set.Path)
	if err != nil || set.Set.Slices[0].Phase != "integrated" || set.Set.Slices[0].DeliveryRef != receipt.Path || set.Set.Slices[0].PullRequest != fixture.provider.pr.URL {
		t.Fatal("set slice does not own its verified common delivery")
	}

	if set.Set.Status == "closed" || set.Set.Slices[1].DeliveryRef != "" {
		t.Fatal("slice finalization closed its owning set or finalized another slice")
	}

	checkWorkflowFinalizedReceiptCount(t, fixture.store)
}

func checkWorkflowFinalizedReceiptCount(t *testing.T, store *Store) {
	t.Helper()

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	count := 0

	for _, record := range snapshot.Records {
		if record.Receipt != nil {
			count++
		}
	}

	if count != 1 {
		t.Fatal("task fan-out created multiple delivery receipts")
	}
}

// TestWorkflowFinalizationReject never turns status claims, an unverified merge,
// changed canonical target or corrupted proof into task completion.
func TestWorkflowFinalizationReject(t *testing.T) {
	for _, stage := range []string{"unverified", "target changed", "proof changed"} {
		t.Run(stage, func(t *testing.T) {
			fixture := workflowDeliveryFixture(t, "rebase")
			if stage != "unverified" {
				if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, "rebase", fixture.provider, fixture.git); err != nil {
					t.Fatal(err)
				}
			}

			invalidateWorkflowFinalization(t, fixture, stage)

			if _, err := fixture.store.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, fixture.provider, fixture.git); err == nil {
				t.Fatal("unverified or changed delivery finalized tasks")
			}

			checkWorkflowPRTasks(t, fixture.store, fixture.hierarchy.Tasks[:3])
		})
	}
}

func invalidateWorkflowFinalization(t *testing.T, fixture workflowIntegrationFixture, stage string) {
	t.Helper()

	switch stage {
	case "target changed":
		workflowIntegrationCommand(t, fixture.actor, "push", "--force", "origin", fixture.base+":refs/heads/dev")
	case "proof changed":
		run := workflowTestRead(t, fixture.store, fixture.run.ID)
		if err := os.WriteFile(filepath.Join(fixture.store.layout.Root(), run.DeliveryProofRef), []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
