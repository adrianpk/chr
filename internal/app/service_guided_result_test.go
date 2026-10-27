package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

func guidedResultFixture(t *testing.T) (service, core.AuthoringSession, *guidedWorkHost, *serviceWorkEntryProvider) {
	t.Helper()

	operations, prepared, _ := guidedAdmissionFixture(t)
	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	operations.guidedWorkHost = host

	session, err := operations.AdmitGuidedWork(t.Context(), guidedQuizCommand(prepared, "admit", "admit", ""))
	if err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()

	provider, ok := operations.workflowPullRequests.(*serviceWorkEntryProvider)
	if !ok {
		t.Fatal("missing local canonical provider")
	}

	return operations, session, host, provider
}

func finishGuidedResultFixture(t *testing.T, operations service, session core.AuthoringSession, host *guidedWorkHost, provider *serviceWorkEntryProvider) core.AuthoringSession {
	t.Helper()

	run := assertGuidedNativeAdmission(t, operations, session)
	if _, err := operations.ApproveWorkflowSlice(t.Context(), run.ID, run.Review.InputHash); err != nil {
		t.Fatal(err)
	}

	serviceWorkflowDeliveryGit(t, operations.projectContext.ProjectDir, "push", provider.remote, run.PullRequest.SourceCommit+":refs/heads/dev")

	provider.pr.State, provider.pr.Merged, provider.pr.ResultCommit = "closed", true, run.PullRequest.SourceCommit
	if err := host.activate(operations, run.ID); err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()

	current := readWebGuidedSession(t, operations, session.ID)
	if current.Guided.Stage != "results" || current.Guided.Status != "completed" || current.Guided.Result == nil {
		t.Fatalf("native evidence did not complete the guided run: stage=%s status=%s blocker=%s", current.Guided.Stage, current.Guided.Status, current.Guided.Blocker)
	}

	return current
}

// TestGuidedUseCaseResultEvidence requires real native validation, canonical
// delivery and set closure, retaining later ideas outside the completed extent.
func TestGuidedUseCaseResultEvidence(t *testing.T) {
	operations, session, host, provider := guidedResultFixture(t)

	assertGuidedResultUnconfirmed(t, operations, session)

	idea := guidedQuizCommand(session, "later-export", "idea", "Export a separate audit view later.")

	saved, err := operations.ChangeGuidedUseCase(t.Context(), idea)
	if err != nil {
		t.Fatal(err)
	}

	completed := finishGuidedResultFixture(t, operations, saved, host, provider)

	after, err := operations.InspectGuidedUseCase(t.Context(), completed.ID, completed.Namespace)
	if err != nil || !after.Result.Complete || len(after.Result.Receipts) != 1 || len(after.Result.Closures) != 1 || len(after.Ideas) != 1 {
		t.Fatal("final result lost original native evidence or later ideas", err)
	}

	if !reflect.DeepEqual(completed.Guided.Work, session.Guided.Work) || !reflect.DeepEqual(completed.Guided.Checkpoints, session.Guided.Checkpoints) {
		t.Fatal("later idea changed the admitted scope")
	}

	assertGuidedResultProjection(t, operations, completed)
	assertGuidedResultReadOnly(t, operations, completed)
}

func assertGuidedResultUnconfirmed(t *testing.T, operations service, session core.AuthoringSession) {
	t.Helper()

	before, err := operations.InspectGuidedUseCase(t.Context(), session.ID, session.Namespace)
	if err != nil || before.Result == nil || before.Result.Complete || len(before.Result.Receipts) != 0 || before.Work.ConfirmedTickets != 0 {
		t.Fatal("validated worker or open PR was presented as delivery", err)
	}
}

func assertGuidedResultProjection(t *testing.T, operations service, session core.AuthoringSession) {
	t.Helper()

	provider := webViewProvider{projectContext: operations.projectContext, serviceFactory: func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }}

	view, err := provider.GuidedUseCaseView(t.Context(), session.ID, session.Namespace)
	if err != nil || view.Result == nil || !view.Result.Complete || len(view.Result.Deliveries) != 1 {
		t.Fatal("missing evidence-backed Web result", err)
	}

	delivery := view.Result.Deliveries[0]
	if delivery.Repository == "" || delivery.Branch != "dev" || delivery.Commit == "" || delivery.PullRequest == "" || len(view.Result.Validation) < 2 {
		t.Fatal("result omitted actual code and validation locations")
	}
}

func assertGuidedResultReadOnly(t *testing.T, operations service, session core.AuthoringSession) {
	t.Helper()

	before, err := core.ReadGuidedResult(t.Context(), operations.projectContext.ProjectDir, session)
	if err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if _, err := operations.InspectGuidedUseCase(t.Context(), session.ID, session.Namespace); err != nil {
			t.Fatal(err)
		}
	}

	after, err := core.ReadGuidedResult(t.Context(), operations.projectContext.ProjectDir, session)

	current := readWebGuidedSession(t, operations, session.ID)
	if err != nil || !reflect.DeepEqual(before, after) || current.Revision != session.Revision {
		t.Fatal("result inspection mutated an owner", err)
	}
}

// TestGuidedUseCaseResultCorrupt rejects missing or changed receipt bytes even
// after native counters and the retained guided stage claim completion.
func TestGuidedUseCaseResultCorrupt(t *testing.T) {
	operations, session, host, provider := guidedResultFixture(t)
	completed := finishGuidedResultFixture(t, operations, session, host, provider)

	result, err := core.ReadGuidedResult(t.Context(), operations.projectContext.ProjectDir, completed)
	if err != nil {
		t.Fatal(err)
	}

	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: operations.projectContext.ProjectDir, Namespace: session.Namespace})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(layout.Root(), result.Receipts[0].Reference.Path)
	if err := os.WriteFile(path, []byte("Changed receipt"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := core.ReadGuidedResult(t.Context(), operations.projectContext.ProjectDir, completed); err == nil {
		t.Fatal("corrupt delivery was accepted")
	}

	viewProvider := webViewProvider{projectContext: operations.projectContext, serviceFactory: func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }}

	view, err := viewProvider.GuidedUseCaseView(t.Context(), completed.ID, completed.Namespace)
	if err != nil || view.Result != nil || view.Status == "completed" || !strings.Contains(view.Blocker, "evidence") {
		t.Fatal("corrupt evidence still claimed completion", err)
	}
}
