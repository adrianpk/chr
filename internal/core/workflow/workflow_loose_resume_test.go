package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowLooseResumeTarget rejects provider disagreement and rewritten
// canonical history without clearing the stop or changing its admitted image.
func TestWorkflowLooseResumeTarget(t *testing.T) {
	for _, reason := range []string{"provider moved", "non descendant"} {
		t.Run(reason, func(t *testing.T) { testWorkflowLooseResumeTarget(t, reason) })
	}
}

func testWorkflowLooseResumeTarget(t *testing.T, reason string) {
	t.Helper()

	f := workflowLooseBlockedFixture(t)

	before := workflowLoosePauseDependency(t, f)
	if reason == "non descendant" {
		workflowIntegrationCommand(t, f.actor, "checkout", "--orphan", "replacement")
		workflowIntegrationCommand(t, f.actor, "rm", "-rf", ".")
	}

	target := workflowIntegrationCommit(t, f.actor, "dependency.txt", "Outside prerequisite")
	if reason == "non descendant" {
		workflowIntegrationCommand(t, f.actor, "push", "--force", "origin", "HEAD:refs/heads/dev")
	} else {
		workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")
	}

	workflowLooseConfirmDependency(t, f, target)

	provider := PullRequestProvider(workflowFrontierProvider{f.provider, f.remote})
	if reason == "provider moved" {
		provider = &workflowPRProviderFixture{pr: workflowstate.PullRequestSnapshot{HeadCommit: f.target}}
	}

	ports := ResumePorts{Provider: provider, Target: f.git, TargetBranch: "dev"}
	if _, err := f.store.ResumeWorkflow(t.Context(), before.ID, ports); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("unaccepted canonical target cleared a prerequisite stop", err)
	}

	if retained := workflowTestRead(t, f.store, before.ID); !reflect.DeepEqual(retained, before) {
		t.Fatal("rejected target changed the invocation")
	}
}

func workflowLooseConfirmDependency(t *testing.T, f workflowIntegrationFixture, commit string) Record {
	t.Helper()

	record := workflowLooseRecord(t, f.store, "PRJ-89")
	tree := workflowIntegrationCommand(t, f.actor, "rev-parse", commit+"^{tree}")
	receipt := selectionReceipt(t, f.store, workflowstate.Ref{Kind: "ticket", ID: record.Task.Key}, commit, tree)
	record.Task.Phase, record.Task.Status, record.Task.DeliveryRef = "integrated", string(ticketcore.StatusDone), receipt.Path

	return workflowTestWrite(t, f.store, record)
}

func workflowLoosePauseDependency(t *testing.T, f workflowIntegrationFixture) Invocation {
	t.Helper()

	if _, err := f.store.ClaimWorkflow(t.Context(), f.run.ID); err == nil {
		t.Fatal("unverified dependency acquired a worker")
	}

	return workflowTestRead(t, f.store, f.run.ID)
}

// TestWorkflowLooseResume observes real prerequisite history again. A receipt
// on the retained target needs no refresh; a newer target needs canonical IO.
func TestWorkflowLooseResume(t *testing.T) {
	for _, movement := range []string{"retained", "descendant"} {
		t.Run(movement, func(t *testing.T) { testWorkflowLooseResume(t, movement) })
	}
}

func testWorkflowLooseResume(t *testing.T, movement string) {
	t.Helper()

	f := workflowLooseBlockedFixture(t)
	before := workflowLoosePauseDependency(t, f)

	target := f.target
	if movement == "descendant" {
		target = workflowIntegrationCommit(t, f.actor, "dependency.txt", "Verified outside prerequisite")
		workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")
	}

	workflowLooseConfirmDependency(t, f, target)

	ports := []ResumePorts{}

	if movement == "descendant" {
		if _, err := f.store.ResumeWorkflow(t.Context(), before.ID); err == nil {
			t.Fatal("new target was trusted without canonical observation")
		}

		ports = append(ports, ResumePorts{Provider: workflowFrontierProvider{f.provider, f.remote}, Target: f.git, TargetBranch: "dev"})
	}

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	run, err := restarted.ResumeWorkflow(t.Context(), before.ID, ports...)
	checkWorkflowLooseResumed(t, before, run, target, err)

	claim := workflowTestClaim(t, restarted, before.ID)
	if _, err := restarted.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	next := workflowTestRead(t, restarted, before.ID)
	if next.Lane.Unit.ID != "PRJ-91" || next.Lane.BaseCommit != target || len(next.Attempts) != len(before.Attempts)+1 || next.Cursor != 2 || next.ConfirmedCursor != 1 {
		t.Fatal("resumed execution added outside work or reused its old base")
	}

	checkWorkflowLooseResumeDelivery(t, f, target, next)
}

func checkWorkflowLooseResumeDelivery(t *testing.T, f workflowIntegrationFixture, target string, before Invocation) {
	t.Helper()

	f.target = target
	workflowLoosePublishValidated(t, &f, "PRJ-91")

	run := workflowAdvanceFrontier(t, f)
	if run.ConfirmedCursor != 2 || len(run.Deliveries) != 2 || run.Deliveries[1].Target.Commit != target || len(run.DependencyTargets) != len(before.DependencyTargets) {
		t.Fatal("delivery lost its dependency target checkpoint")
	}

	claim := workflowTestClaim(t, f.store, run.ID)
	if _, err := f.store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal("later work could not verify retained dependency-target history", err)
	}
}

func checkWorkflowLooseResumed(t *testing.T, before, run Invocation, target string, failure error) {
	t.Helper()

	if failure != nil || run.Status != "pending" || run.Blocker != nil || run.ExecutionSelection().Target.Commit != target || run.Selection.Target != before.Selection.Target || run.Cursor != before.Cursor {
		t.Fatal("resume changed scope or failed to verify the prerequisite", failure)
	}

	if !reflect.DeepEqual(run.Members, before.Members) || run.Policy != before.Policy {
		t.Fatal("resume changed admitted members or frozen policy")
	}
}

// TestWorkflowLooseResumeCorruption rejects changed resume/receipt evidence
// before another attempt, including after the fresh target was hydrated.
func TestWorkflowLooseResumeCorruption(t *testing.T) {
	for _, evidence := range []string{"resume", "receipt"} {
		t.Run(evidence, func(t *testing.T) {
			testWorkflowLooseResumeCorruption(t, evidence)
		})
	}
}

func testWorkflowLooseResumeCorruption(t *testing.T, evidence string) {
	t.Helper()

	f := workflowLooseBlockedFixture(t)
	workflowLoosePauseDependency(t, f)

	target := workflowIntegrationCommit(t, f.actor, "dependency.txt", "Verified outside prerequisite")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	record := workflowLooseConfirmDependency(t, f, target)
	ports := ResumePorts{Provider: workflowFrontierProvider{f.provider, f.remote}, Target: f.git, TargetBranch: "dev"}

	run, err := f.store.ResumeWorkflow(t.Context(), f.run.ID, ports)
	if err != nil {
		t.Fatal(err)
	}

	path := run.DependencyTargets[0].EvidenceRef
	if evidence == "receipt" {
		path = record.Task.DeliveryRef
	}

	absolute := filepath.Join(f.store.layout.Root(), path)

	data, err := os.ReadFile(absolute)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(absolute, append(data, ' '), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.ClaimWorkflow(t.Context(), run.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("corrupted resume evidence admitted another worker", err)
	}

	if retained := workflowTestRead(t, f.store, run.ID); retained.Revision != run.Revision || len(retained.Attempts) != len(run.Attempts) {
		t.Fatal("rejected evidence changed the invocation")
	}
}
