package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowPublicationUnknown blocks preparation, review and repair until an
// uncertain operation is canonically observed. Sealing permits a bounded retry
// with fresh base validation; no local expiry or status reset releases intent.
func TestWorkflowPublicationUnknown(t *testing.T) {
	f, p := newWorkflowSubmissionFixture(t, "merge")
	p.unknown = true

	submission, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, p, workflowSubmissionObserver(f))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}

	before := workflowTestRead(t, f.store, f.run.ID)

	ports := workflowPublicationPorts(f, p)
	if _, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports); !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("unknown result allowed another publication phase", err)
	}

	checkWorkflowPublicationUnknownInputs(t, f, p)

	if len(workflowTestRead(t, f.store, f.run.ID).Attempts) != len(before.Attempts) || p.calls != 1 {
		t.Fatal("uncertainty started more host work")
	}

	p.seal(submission.ID)

	p.unknown = false

	workflowIntegrationCommit(t, f.actor, "new-base.txt", "Fresh target after sealed attempt")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if err != nil || progress.Invocation.Submission.Attempt != 2 || progress.Invocation.Preparation.Round != before.Preparation.Round+1 || p.calls != 2 {
		t.Fatal("sealed retry lost its budget or reused a stale base", err)
	}
}

// TestWorkflowPublicationCorrupt prevents actual delivery from completing tasks
// when its retained source, task, prepared validation or review image changed.
// Canonical success cannot repair corrupt local evidence by running more work.
func TestWorkflowPublicationCorrupt(t *testing.T) {
	cases := []struct {
		name string
		path func(Invocation) string
	}{
		{"candidate", func(run Invocation) string { return run.CandidateRef }},
		{"task", func(run Invocation) string { return run.Members[0].Admission.Path }},
		{"preparation", func(run Invocation) string { return run.Preparation.ResultRef }},
		{"review", func(run Invocation) string { return run.Review.InputRef }},
	}
	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			f, p := newWorkflowSubmissionFixture(t, "rebase")
			workflowPublicationExternalResult(t, f, p, "rebase")

			before := workflowTestRead(t, f.store, f.run.ID)
			if err := os.WriteFile(filepath.Join(f.store.layout.Root(), check.path(before)), []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}

			progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))
			if err == nil || progress.Receipt != nil || progress.Invocation.DeliveryRef != "" || len(progress.Invocation.Attempts) != len(before.Attempts) || p.calls != 0 {
				t.Fatal("corrupt delivery evidence completed tasks or started work")
			}
		})
	}
}

// TestWorkflowPublicationUnsupported retains a durable blocker for source-only
// provider capability and proves neither a prepared gate nor mutation occurs.
func TestWorkflowPublicationUnsupported(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "rebase", workflowstate.Validation{WaiverReason: "Unsupported guard fixture"})
	f.provider.guarded = false
	p := &workflowSubmissionProviderFixture{workflowPreparationProviderFixture: f.provider, remote: f.remote}
	before := workflowTestRead(t, f.store, f.run.ID)

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))
	if !errors.Is(err, ErrWorkflowAtomicGuard) || progress.Invocation.Status != workflowstate.InvocationBlocked || progress.Invocation.Blocker == nil || len(progress.Invocation.Attempts) != len(before.Attempts) || p.calls != 0 {
		t.Fatal("unsupported provider bypassed the guard contract", err)
	}
}

func checkWorkflowPublicationUnknownInputs(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture) {
	t.Helper()

	checks := []struct {
		name string
		run  func() error
	}{
		{"preparation", func() error {
			_, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, p, f.git)

			return err
		}},
		{"review", func() error { _, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, p, nil); return err }},
		{"repair", func() error {
			_, err := f.store.RepairWorkflowSlice(t.Context(), f.run.ID, f.run.Members[0].Admission.Key)

			return err
		}},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.run(); !errors.Is(err, ErrWorkflowObservationRequired) {
				t.Fatal("pending publication inputs could change", err)
			}
		})
	}
}
