package core

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowPublicationExternal verifies actual externally rewritten results
// and moved bases for all configured methods. An uncertain native intent is
// resolved as external delivery; stops survive and no second command is sent.
func TestWorkflowPublicationExternal(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		for _, intent := range []string{"", "pause", "cancel"} {
			t.Run(method+"/"+intent, func(t *testing.T) { testWorkflowPublicationExternal(t, method, intent) })
		}
	}
}

func testWorkflowPublicationExternal(t *testing.T, method, intent string) {
	t.Helper()

	f, p := newWorkflowSubmissionFixture(t, method)

	p.unknown = true
	if _, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, p, workflowSubmissionObserver(f)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}

	if intent != "" {
		if _, err := f.store.StopWorkflow(t.Context(), f.run.ID, intent); err != nil {
			t.Fatal(err)
		}
	}

	before := workflowTestRead(t, f.store, f.run.ID)
	result := workflowPublicationExternalResult(t, f, p, method)

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))
	if err != nil || progress.Receipt == nil || progress.Receipt.Receipt.ResultCommit != result || len(progress.Invocation.Attempts) != len(before.Attempts) || p.calls != 1 {
		t.Fatal("external delivery was replaced by new work or publication", err)
	}

	checkWorkflowPublicationExternalFacts(t, before, progress, intent)

	checkWorkflowPublicationFinalized(t, f, progress)
	assertWorkflowPublicationStopped(t, f, p, progress, intent)
}

func workflowPublicationExternalResult(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture, method string) string {
	t.Helper()

	workflowIntegrationCommand(t, f.actor, "fetch", "origin")
	workflowIntegrationCommit(t, f.actor, "external-base.txt", "External integration uses a moved base")
	workflowIntegrationCommand(t, f.actor, "branch", "source", "origin/"+p.pr.HeadBranch)
	workflowFixtureIntegrate(t, f.actor, method)

	result := workflowIntegrationCommand(t, f.actor, "rev-parse", "HEAD")
	workflowIntegrationCommit(t, f.actor, "later.txt", "Later reachable target content")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	pr := p.pr
	pr.State, pr.Merged, pr.ResultCommit = "closed", true, result

	data, err := json.Marshal(pr)
	if err != nil {
		t.Fatal(err)
	}

	if err := opsfile.WriteAtomic(filepath.Join(f.remote, "workflow-provider-pr.json"), data); err != nil {
		t.Fatal(err)
	}

	return result
}

func assertWorkflowPublicationStopped(t *testing.T, f workflowPreparationFixture, p *workflowSubmissionProviderFixture, progress WorkflowIntegrationProgress, intent string) {
	t.Helper()

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	again, err := restarted.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))
	if err != nil || again.Invocation.Revision != progress.Invocation.Revision || again.Receipt.SourceHash != progress.Receipt.SourceHash || p.calls != 1 {
		t.Fatal("restart repeated external finalization", err)
	}

	if intent == "pause" {
		resumed, err := restarted.ResumeWorkflow(t.Context(), f.run.ID)
		if err != nil || resumed.Status != "waiting_integration" || resumed.DeliveryRef != progress.Receipt.Path {
			t.Fatal("pause resume returned to old task execution", err)
		}
	}
}

// TestWorkflowPublicationReadOnly proves external truth can finalize through a
// provider with no automatic mutation contract, even after cancellation. Missing
// atomic support does not prevent read-only canonical verification.
func TestWorkflowPublicationReadOnly(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "squash", workflowstate.Validation{WaiverReason: "External publication fixture observes actual Git"})
	p := &workflowSubmissionProviderFixture{workflowPreparationProviderFixture: f.provider, remote: f.remote}
	workflowPublicationExternalResult(t, f, p, "squash")

	if _, err := f.store.StopWorkflow(t.Context(), f.run.ID, "cancel"); err != nil {
		t.Fatal(err)
	}

	ports := workflowPublicationPorts(f, workflowReadOnlyPublicationProvider{p})
	ports.Preparer, ports.Reviewer = nil, nil

	progress, err := f.store.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if err != nil || progress.Receipt == nil || progress.Invocation.Status != "cancelled" || progress.Invocation.Submission != nil || p.calls != 0 {
		t.Fatal("read-only external delivery required execution authority", err)
	}
}

type workflowReadOnlyPublicationProvider struct{ WorkflowPullRequestProvider }

func checkWorkflowPublicationExternalFacts(t *testing.T, before WorkflowInvocation, progress WorkflowIntegrationProgress, intent string) {
	t.Helper()

	if !progress.Invocation.Submission.ExternalDelivery || progress.Invocation.Submission.State != "confirmed" || progress.Receipt.Receipt.SubmissionID != before.PullRequest.ID {
		t.Fatal("external history was misreported as the exact guarded submission")
	}

	if progress.Invocation.StopIntent != intent || intent != "" && progress.Invocation.Status != workflowStopStatus(intent) {
		t.Fatal("delivery observation resumed stopped execution")
	}
}
