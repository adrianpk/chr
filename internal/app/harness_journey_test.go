package app

import (
	"context"
	"fmt"
	"testing"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

type harnessJourneyChoice struct{ transport, worker, generation string }

// One production-service journey switches every role combination through all
// four adapters while a claimed worker and an actual generation call remain
// active. Owned protocol peers replace provider IO; Docker is metadata-only.
func TestHarnessJourney(t *testing.T) {
	f := newHarnessJourney(t)
	worker, claim := f.startWorker(t)
	active := f.startGeneration(t)
	f.enablePi(t)

	for index, choice := range []harnessJourneyChoice{
		{"web", "code-worker", "code-gen"},
		{"tui", "pi-worker", "code-gen"},
		{"telegram", "pi-worker", "pi-gen"},
		{"cli", "code-worker", "pi-gen"},
	} {
		t.Run(choice.transport, func(t *testing.T) {
			f.change(t, choice)
			f.checkViews(t, choice)
			f.checkAdmission(t, index, choice)
			f.generate(t.Context(), t, "fresh-"+choice.transport, choice.generation)
			webHarnessCheckRetained(t, f.operations, worker)
		})
	}

	active.finish(t)
	f.resumeWorker(t, worker, claim)
	f.resumeGeneration(t)
	f.checkLegacyAPI(t)
	webHarnessCheckProcess(t, f.operations, webHarnessSnapshot(t, f.operations))
}

func (f *harnessJourney) startWorker(t *testing.T) (workflowcore.Invocation, *workflowcore.Claim) {
	t.Helper()

	started, err := f.operations.StartWorkflow(t.Context(), f.start)
	if err != nil {
		t.Fatal(err)
	}

	claim, err := f.operations.ClaimWorkflow(t.Context(), started.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	})

	retained, err := f.operations.ReadWorkflow(t.Context(), started.Invocation.ID)
	if err != nil || retained.HarnessBinding == nil || retained.HarnessBinding.Profile != "code-worker" {
		t.Fatal("Codex-only admission did not retain its worker", err)
	}

	return retained, claim
}

func (f *harnessJourney) checkAdmission(t *testing.T, index int, choice harnessJourneyChoice) {
	t.Helper()

	store, err := f.operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	task := *f.record.Task

	task.Key = fmt.Sprintf("PRJ-%d", 91+index)
	if _, err := store.Create(t.Context(), "Independent role combination", workflowcore.Record{Kind: "ticket", Task: &task, Domain: f.record.Domain, Body: f.record.Body}); err != nil {
		t.Fatal(err)
	}

	request := f.start
	request.RequestKey, request.Selection.Selector.Identity = "combination-"+choice.transport, task.Key

	started, err := f.operations.StartWorkflow(t.Context(), request)
	if err != nil || started.Invocation.HarnessBinding == nil {
		t.Fatal("new role combination was not admitted", err)
	}

	binding := started.Invocation.HarnessBinding
	if binding.Profile != choice.worker || binding.SelectionRevision != webHarnessSnapshot(t, f.operations).Revision {
		t.Fatal("admission mixed configuration images", binding)
	}

	if choice.worker == "pi-worker" && binding.RuntimeRef != f.image {
		t.Fatal("Pi admission lost its selected immutable runtime", binding.RuntimeRef)
	}
}

func (f *harnessJourney) resumeWorker(t *testing.T, before workflowcore.Invocation, claim *workflowcore.Claim) {
	t.Helper()

	if _, err := f.operations.StopWorkflow(t.Context(), before.ID, "pause", before.Revision); err != nil {
		t.Fatal(err)
	}

	if err := f.operations.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	restarted := newService(t.Context(), f.operations.env, f.operations.projectContext, testAssetsFS())

	resumed, err := restarted.ResumeWorkflow(t.Context(), before.ID)
	if err != nil || resumed.HarnessBinding == nil || *resumed.HarnessBinding != *before.HarnessBinding || resumed.Policy != before.Policy {
		t.Fatal("resume replaced the admitted worker or its policy", err)
	}

	harnessJourneyExecuteWorker(t, restarted, before.ID, f.start.Selection.Target.Commit)
}

func harnessJourneyExecuteWorker(t *testing.T, operations service, id, target string) {
	t.Helper()

	reclaimed, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reclaimed.Close(); err != nil {
			t.Error(err)
		}
	}()

	result, err := operations.ExecuteWorkflowTask(t.Context(), reclaimed, nil)
	if err != nil || !result.Successful || result.Commit == target {
		t.Fatal("resumed original worker did not produce host-validated Git output", result.Failure, err)
	}
}

func (f *harnessJourney) generate(ctx context.Context, t *testing.T, id, profile string) {
	t.Helper()

	bound, err := f.operations.bindGeneration(ctx, id, "spec_generation")
	ctx = bound.context(ctx)

	if err != nil {
		t.Fatal(err)
	}

	generation, ok := harnesshost.GenerationSelection(ctx)
	if !ok || generation.Binding.Selection.Profile != profile {
		t.Fatal("generation resolved the wrong role", generation.Binding.Selection.Profile)
	}

	if err := harnessJourneyInvoke(ctx, bound.service, id+"-call"); err != nil {
		t.Fatal("selected structured transport failed", err)
	}
}
