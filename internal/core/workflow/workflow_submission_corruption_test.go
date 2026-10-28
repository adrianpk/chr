package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowSubmissionCorruption rejects deleted, unknown, symlinked and
// syntactically forged resolution images without resetting the budget or
// clearing an invocation's previously recorded unknown exclusion.
func TestWorkflowSubmissionCorruption(t *testing.T) {
	for _, changed := range []string{"missing", "schema", "symlink", "sealed"} {
		t.Run(changed, func(t *testing.T) { testWorkflowSubmissionCorruption(t, changed) })
	}
}

func testWorkflowSubmissionCorruption(t *testing.T, changed string) {
	t.Helper()

	f, provider := newWorkflowSubmissionFixture(t, "rebase")
	provider.unknown = true

	first, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}

	journal := corruptWorkflowSubmission(t, f, first, changed)
	if _, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f)); err == nil || provider.calls != 1 {
		t.Fatal("journal corruption permitted another provider mutation", err)
	}

	alias := workflowSubmissionAlias(t, f)

	err = alias.withLock(t.Context(), func() error {
		return alias.guardWorkflowSubmissionTarget(t.Context(), contenthash.Sum([]byte("another-unit")), journal.TargetKey)
	})
	if err == nil {
		t.Fatal("journal corruption released exclusion through an alias")
	}
}

func corruptWorkflowSubmission(t *testing.T, f workflowPreparationFixture, first workflowstate.Submission, changed string) workflowSubmissionJournal {
	t.Helper()

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	journal, err := f.store.readWorkflowSubmissionJournal(t.Context(), run)
	if err != nil {
		t.Fatal(err)
	}

	path, err := workflowSubmissionPath(t.Context(), f.store.layout.ProjectDir(), journal.Key)
	if err != nil {
		t.Fatal(err)
	}

	editWorkflowSubmissionJournal(t, path, &journal, first, changed)

	return journal
}

func editWorkflowSubmissionJournal(t *testing.T, path string, journal *workflowSubmissionJournal, first workflowstate.Submission, changed string) {
	t.Helper()

	switch changed {
	case "missing", "symlink":
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}

		if changed == "symlink" {
			if err := os.Symlink(path+".outside", path); err != nil {
				t.Fatal(err)
			}
		}

		return
	case "schema":
		journal.Schema = "chr-integration-journal/v99"
	case "sealed":
		journal.Attempts[0].State = workflowstate.SubmissionUnpublished
		journal.Attempts[0].Unpublished = &workflowstate.SubmissionObservation{SubmissionID: first.ID, Unpublished: true}
	}

	data, err := json.Marshal(journal)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
