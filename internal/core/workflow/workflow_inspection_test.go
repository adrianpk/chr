package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowInspectionJournal leaves interrupted transitions byte-identical
// instead of silently recovering them while an adapter displays status.
func TestWorkflowInspectionJournal(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)

	store.afterReplace = func(int) error { return errors.New("interrupted replacement") }
	if _, err := store.StopWorkflow(t.Context(), run.Invocation.ID, "pause"); err == nil {
		t.Fatal("transition interruption not injected")
	}

	journal := filepath.Join(store.layout.Root(), workflowstate.JournalPath)

	before, err := os.ReadFile(journal)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.InspectWorkflow(t.Context(), run.Invocation.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("status replayed a pending transition", err)
	}

	if _, err := store.InspectWorkflowPolicy(t.Context()); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("policy inspection replayed a pending transition", err)
	}

	after, err := os.ReadFile(journal)
	if err != nil || string(before) != string(after) {
		t.Fatal("read-only inspection changed the recovery journal", err)
	}
}

// TestWorkflowClaimClose fences a running execution and releases only host
// descriptors once quiescent, without accepting a checkpoint or task result.
func TestWorkflowClaimClose(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	if err := beginWorkflowClaimExecution(claim); err != nil {
		t.Fatal(err)
	}

	if err := claim.Close(); !errors.Is(err, workflowstate.ErrOwned) || claim.reservation == nil {
		t.Fatal("close released a running execution", err)
	}

	endWorkflowClaimExecution(claim)
	checkWorkflowClaimClosed(t, store, claim, run.Invocation.ID)
}

func checkWorkflowClaimClosed(t *testing.T, store *Store, claim *Claim, id string) {
	t.Helper()

	before, err := store.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if err := claim.Close(); err != nil {
		t.Fatal(err)
	}

	after, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || after.Revision != before.Revision || after.Cursor != 0 || after.Status != workflowstate.InvocationRunning || claim.reservation != nil || claim.capacityGuard != nil {
		t.Fatal("close changed authority or retained owned descriptors", err)
	}
}

// TestWorkflowInspectionLinked blocks a journal in another namespace/worktree
// before repository discovery can replay it during a status or dashboard read.
func TestWorkflowInspectionLinked(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)
	linked := filepath.Join(t.TempDir(), "linked")
	workflowSelectionGit(t, store, "worktree", "add", "-qb", "inspection-linked", linked, "HEAD")

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: linked, Namespace: "other"})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(layout.Root(), "lock"), 0o755); err != nil {
		t.Fatal(err)
	}

	journal := filepath.Join(layout.Root(), workflowstate.JournalPath)

	before := []byte("interrupted linked transition\n")
	if err := os.WriteFile(journal, before, 0o600); err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		name string
		read func() error
	}{
		{"invocation", func() error { _, err := store.InspectWorkflow(t.Context(), run.Invocation.ID); return err }},
		{"page", func() error { _, err := store.InspectWorkflowPage(t.Context(), 0, 8); return err }},
		{"records", func() error { _, err := store.InspectWorkflowRecords(t.Context()); return err }},
		{"board", func() error {
			_, err := ReadTicketBoard(t.Context(), TicketBoardReadOptions{Layout: store.layout, ProjectID: store.projectID, NativeWorkflow: true})

			return err
		}},
		{"policy", func() error { _, err := store.InspectWorkflowPolicy(t.Context()); return err }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.read(); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("linked journal was recovered by inspection", err)
			}

			after, err := os.ReadFile(journal)
			if err != nil || string(before) != string(after) {
				t.Fatal("inspection changed the linked journal", err)
			}
		})
	}
}
