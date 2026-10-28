package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowSource pushes one retained validated source, retries idempotently
// and refuses a branch collision without changing the canonical target.
func TestWorkflowSource(t *testing.T) {
	store, request, run, candidate := workflowSourceCandidate(t)
	remote := filepath.Join(t.TempDir(), "canonical.git")
	workflowSelectionGit(t, store, "init", "--bare", remote)
	workflowSelectionGit(t, store, "push", remote, request.Selection.Target.Commit+":refs/heads/dev")

	publisher := NewWorkflowIntegrationGit(request.Selection.Target.Repository, remote)
	for range 2 {
		if err := store.PublishWorkflowSource(t.Context(), run.ID, publisher); err != nil {
			t.Fatal(err)
		}
	}

	target := checkWorkflowSourceTarget(t, store, candidate, remote, request.Selection.Target.Commit)
	workflowSelectionGit(t, store, "push", "--force", remote, target+":"+candidate.Lane.Branch)

	if err := store.PublishWorkflowSource(t.Context(), run.ID, publisher); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("source collision overwritten", err)
	}

	retained, err := store.ReadWorkflow(t.Context(), run.ID)
	if err != nil || retained.Status != workflowstate.InvocationWaitingIntegration || len(retained.Attempts) != 1 || retained.DeliveryRef != "" {
		t.Fatal("source push fabricated completion", err)
	}

	if _, err := store.StopWorkflow(t.Context(), run.ID, "pause"); err != nil {
		t.Fatal(err)
	}

	if err := store.PublishWorkflowSource(t.Context(), run.ID, publisher); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("paused source performed IO", err)
	}
}

// TestWorkflowSourceCorruption rejects changed immutable evidence before IO.
func TestWorkflowSourceCorruption(t *testing.T) {
	store, _, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("test -f result.txt"))
	run := workflowTestStart(t, store, request).Invocation

	claim, err := store.ClaimWorkflow(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	if _, err := store.PrepareWorkflowTicketCandidate(t.Context(), run.ID); err != nil {
		t.Fatal(err)
	}

	retained, err := store.ReadWorkflow(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.layout.Root(), retained.CandidateRef)

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.WriteString(" "); err != nil {
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	if err := store.PublishWorkflowSource(t.Context(), run.ID, NewWorkflowIntegrationGit(request.Selection.Target.Repository, filepath.Join(t.TempDir(), "absent.git"))); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("source IO preceded evidence guard", err)
	}
}

func checkWorkflowSourceTarget(t *testing.T, store *Store, candidate SliceCandidate, remote, expectedTarget string) string {
	t.Helper()

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: store.layout.ProjectDir()}}
	source, err := reader.workflowSourceHead(t.Context(), remote, candidate.Lane.Branch)

	target, targetErr := reader.remoteTarget(t.Context(), remote, "dev")
	if err != nil || targetErr != nil || source != candidate.Result.Commit || target != expectedTarget {
		t.Fatal("source push changed delivery target", err, targetErr)
	}

	return target
}

func workflowSourceCandidate(t *testing.T) (*Store, StartRequest, Invocation, SliceCandidate) {
	t.Helper()

	store, _, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("test -f result.txt"))
	run := workflowTestStart(t, store, request).Invocation

	claim, err := store.ClaimWorkflow(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	candidate, err := store.PrepareWorkflowTicketCandidate(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	return store, request, run, candidate
}
