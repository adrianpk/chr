package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeQueueMigration adds only a missing ledger under its observed
// revision. Reads cannot perform migration; source entry and legacy usage remain.
func TestNativeQueueMigration(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)

	run := nativeQueueRemoveBudget(t, store, entry.InvocationID)

	before := nativeQueueImages(t, store.layout.Root())
	if _, err := store.QueryWorkflowIntegrationQueue(t.Context(), IntegrationQueueQuery{Limit: 10}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("read manufactured ledger", err)
	}

	if _, err := store.AdmitWorkflowIntegration(t.Context(), admission); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("retry manufactured ledger", err)
	}

	if _, err := store.MigrateWorkflowIntegration(t.Context(), admission); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale migration changed owner", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("failed migration changed authority")
	}

	admission.ExpectedRevision = run.Revision

	migrated, err := store.MigrateWorkflowIntegration(t.Context(), admission)
	if err != nil || !reflect.DeepEqual(entry, migrated) {
		t.Fatal("migration changed entry allocation", err)
	}

	page := nativeQueuePage(t, store, IntegrationQueueQuery{Limit: 10})
	if len(page.Entries) != 1 || page.Entries[0].Budget.Revision != 1 {
		t.Fatal("migration has no sealed budget")
	}
}

// TestNativeQueueMigrationBytes rejects malformed or future unbound ledger
// bytes instead of replacing them with zero counters during explicit migration.
func TestNativeQueueMigrationBytes(t *testing.T) {
	for _, image := range []string{"{malformed", `{"schema_version":99,"consumed":123}`} {
		t.Run(image, func(t *testing.T) {
			store, admission := nativeEntryFixture(t)

			path := filepath.Join(store.layout.Root(), workflowIntegrationBudgetPath(admission.InvocationID))
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(path, []byte(image), 0o600); err != nil {
				t.Fatal(err)
			}

			before := nativeQueueImages(t, store.layout.Root())
			if _, err := store.MigrateWorkflowIntegration(t.Context(), admission); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("unbound source became runnable", err)
			}

			if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
				t.Fatal("migration erased malformed bytes")
			}
		})
	}
}

// TestNativeQueueJournal keeps a common intent visible even when its namespace
// projection was interrupted. Reconciliation attaches without another dispatch.
func TestNativeQueueJournal(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	crash := errors.New("common intent retained")
	f.store.afterReplace = func(at int) error {
		if at == -1 {
			return crash
		}

		return nil
	}
	guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)

	first, dispatch, err := f.store.BeginWorkflowNativeSubmission(t.Context(), guard, f.provider)
	if !errors.Is(err, crash) || dispatch {
		t.Fatal("intent fault did not retain common barrier", err)
	}

	f.store.afterReplace = nil

	nativeQueueInterruptedIntent(t, f, first)

	again, dispatch, err := f.store.BeginWorkflowNativeSubmission(t.Context(), guard, f.provider)
	if err != nil || dispatch || again.ID != first.ID {
		t.Fatal("retry invented another dispatch", err)
	}

	nativeQueueRetainedSubmission(t, f, first)
}

func nativeQueueRetainedSubmission(t *testing.T, f nativeSubmissionFixture, first Submission) {
	t.Helper()

	before := nativeQueueImages(t, f.store.layout.Root())

	retained, err := f.store.QueryWorkflowIntegrationSubmissions(t.Context(), SubmissionQuery{SubmissionID: first.ID, Limit: 1})
	if err != nil || len(retained) != 1 || retained[0].Submission.Native.CandidateHash != first.Native.CandidateHash {
		t.Fatal("submission lookup lost native identity", err)
	}

	page := nativeQueuePage(t, f.store, IntegrationQueueQuery{Limit: 10, EligibleOnly: true})
	if len(page.Entries) != 0 {
		t.Fatal("uncertain target returned eligible work")
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, f.store.layout.Root())) {
		t.Fatal("observation repaired projection")
	}
}

func nativeQueueRemoveBudget(t *testing.T, store *Store, id string) workflowInvocationFile {
	t.Helper()

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	budgetPath := filepath.Join(store.layout.Root(), run.IntegrationBudget.Ref.Path)
	if err := os.Remove(budgetPath); err != nil {
		t.Fatal(err)
	}

	run.IntegrationBudget = nil

	change, err := workflowInvocationChange(run)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.commitFiles([]workflowstate.FileChange{change}); err != nil {
		t.Fatal(err)
	}

	return run
}

func nativeQueueInterruptedIntent(t *testing.T, f nativeSubmissionFixture, first Submission) {
	t.Helper()

	barriers, err := f.store.QueryWorkflowPublicationBarriers(t.Context(), "")
	if err != nil || len(barriers) != 1 || barriers[0].Submission.ID != first.ID {
		t.Fatal("common intent lost exclusion", err)
	}

	if _, err := f.store.QueryWorkflowIntegrationQueue(t.Context(), IntegrationQueueQuery{Limit: 10}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale projection looked runnable", err)
	}
}
