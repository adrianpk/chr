package core

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowAssuranceOrdered composes six real worker processes, two automatic
// guarded publications, receipt barriers and set closure across store restarts.
// An edited projection stays untouched until explicit repair after delivery.
func TestWorkflowAssuranceOrdered(t *testing.T) {
	store, options := workflowMaterializationFixture(t)
	gate := workflowPreparationGate("ls PRJ-*.txt >/dev/null")
	options.Plan.ClosingValidation = workflowPreparationGate("test \"$(ls PRJ-*.txt | wc -l)\" -eq 6")

	for i := range options.Plan.Slices {
		options.Plan.Slices[i].Validation = gate
		for j := range options.Plan.Slices[i].Tasks {
			options.Plan.Slices[i].Tasks[j].Validation = gate
		}
	}

	hierarchy, err := store.MaterializePlan(t.Context(), options)
	if err != nil {
		t.Fatal(err)
	}

	again, err := store.MaterializePlan(t.Context(), options)
	if err != nil || again.Set.ID != hierarchy.Set.ID || len(again.Tasks) != 6 {
		t.Fatal("composed materialization retry replaced the hierarchy", err)
	}

	projection, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.layout.Root(), projection.Path)
	workflowTrackerWrite(t, path, []byte("# Operator says all tasks are done\n"))

	target, _ := workflowSelectionCommit(t, store)
	ports, provider := workflowAssurancePorts(t, store, target)
	request := WorkflowStartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "composed-ordered",
		Policy:     workflowstate.ExecutionPolicy{CheckpointStepLimit: 1},
		Selection:  workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, "slice_set"), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}},
	}}
	run := workflowTestStart(t, store, request).Invocation
	completed := workflowAssuranceRun(t, store, run.ID, ports)
	checkWorkflowAssuranceDelivery(t, store, completed, provider, []string{"PRJ-1", "PRJ-2", "PRJ-3", "PRJ-4", "PRJ-5", "PRJ-6"}, 2)
	checkWorkflowAssuranceClosure(t, store, hierarchy, completed, path)
	checkWorkflowAssuranceBudgets(t, store, completed.ID)
}

func checkWorkflowAssuranceBudgets(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	run, err := store.readWorkflowInvocation(id)
	if err != nil || run.ReviewCount != 2 || run.PreparationCount != 2 || run.Deliveries[0].Review.Round != 1 || run.Deliveries[1].Review.Round != 2 {
		t.Fatal("cross-slice continuation reset cumulative review or preparation rounds", err)
	}

	cases := []struct {
		name   string
		mutate func(*workflowInvocationFile)
	}{
		{"review rewind", func(r *workflowInvocationFile) { r.ReviewCount-- }},
		{"review without evidence", func(r *workflowInvocationFile) { r.ReviewCount++ }},
		{"preparation rewind", func(r *workflowInvocationFile) { r.PreparationCount-- }},
		{"preparation without evidence", func(r *workflowInvocationFile) { r.PreparationCount++ }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := run
			tc.mutate(&changed)

			if !errors.Is(validateWorkflowInvocation(changed), workflowstate.ErrConflict) {
				t.Fatal("unbound cumulative round accepted")
			}
		})
	}
}

func workflowAssuranceRun(t *testing.T, store *WorkflowStore, id string, ports WorkflowContinuationPorts) WorkflowInvocation {
	t.Helper()

	for range 80 {
		progress := workflowAssuranceContinue(t, store, id, ports)
		run := progress.Invocation

		if run.Status == "completed" {
			return run
		}

		if run.Cursor > run.ConfirmedCursor+3 {
			t.Fatal("a successor ran before its previous slice receipt")
		}

		if progress.Reason == "blocked" || progress.Reason == "waiting_capacity" {
			t.Fatalf("unexpected composed barrier: %s", progress.Reason)
		}
	}

	t.Fatal("composed workflow exceeded eighty bounded calls", workflowTestRead(t, store, id))

	return WorkflowInvocation{}
}

func checkWorkflowAssuranceDelivery(t *testing.T, store *WorkflowStore, run WorkflowInvocation, provider *workflowAssuranceProvider, keys []string, publications int) {
	t.Helper()

	var actual []string

	for _, attempt := range run.Attempts {
		if attempt.Purpose == "" {
			actual = append(actual, attempt.TaskKey)
			if attempt.Outcome != "validated" {
				t.Fatal("composed worker has no accepted host proof")
			}
		}
	}

	if !reflect.DeepEqual(actual, keys) || run.ConfirmedCursor != len(keys) || len(run.Deliveries) != publications || provider.creates != publications || provider.calls != publications {
		t.Fatal("composed execution order, PR count or verified deliveries differ", actual, run.ConfirmedCursor, len(run.Deliveries), provider.creates, provider.calls)
	}

	checkWorkflowAssuranceMembers(t, store, run, provider)
}

func checkWorkflowAssuranceMembers(t *testing.T, store *WorkflowStore, run WorkflowInvocation, provider *workflowAssuranceProvider) {
	t.Helper()

	for _, member := range run.Members {
		record, err := store.Read(t.Context(), member.Admission.Path)
		if err != nil || record.Task.Status != TicketStatusDone || record.Task.DeliveryRef == "" {
			t.Fatal("composed member lacks finalized delivery", err)
		}

		if content := workflowIntegrationCommand(t, provider.remote, "show", "dev:"+record.Task.Key+".txt"); content != record.Task.Key {
			t.Fatal("canonical target lacks actual worker content", content)
		}
	}
}

func checkWorkflowAssuranceClosure(t *testing.T, store *WorkflowStore, hierarchy WorkflowMaterializationResult, run WorkflowInvocation, path string) {
	t.Helper()

	set, err := store.Read(t.Context(), hierarchy.Set.Path)
	if err != nil || set.Set.ClosureRef == "" || run.Outcome != "sets_closed" {
		t.Fatal("composed deliveries omitted the exact set closure", run.Outcome, err)
	}

	if string(workflowTrackerRead(t, path)) != "# Operator says all tasks are done\n" {
		t.Fatal("native execution overwrote operator projection edits")
	}

	stale, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
	if !errors.Is(err, ErrWorkflowProjectionStale) || stale.Status != "stale" {
		t.Fatal("projection failure did not report stale state after successful closure", err)
	}

	before := workflowTestRead(t, store, run.ID)
	_, err = store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{ReplaceHash: contenthash.Sum(workflowTrackerRead(t, path))})

	if err != nil || workflowTestRead(t, store, run.ID).Revision != before.Revision {
		t.Fatal("projection repair ran work or replaced closure", err)
	}
}

// TestWorkflowAssuranceLoose runs a dependent priority queue from admission to
// exhaustion through actual workers, one canonical PR per ticket and fresh bases.
func TestWorkflowAssuranceLoose(t *testing.T) {
	store, request, ports, provider := workflowAssuranceLooseFixture(t)
	run := workflowTestStart(t, store, request).Invocation
	completed := workflowAssuranceRun(t, store, run.ID, ports)
	checkWorkflowAssuranceDelivery(t, store, completed, provider, []string{"PRJ-90", "PRJ-91", "PRJ-92"}, 3)

	if completed.Outcome != "loose_exhausted" {
		t.Fatal("loose exhaustion fabricated a set closure")
	}
}

func workflowAssuranceLooseFixture(t *testing.T) (*WorkflowStore, WorkflowStartRequest, WorkflowContinuationPorts, *workflowAssuranceProvider) {
	t.Helper()

	store, first, request := workflowTicketFixture(t, "loose", workflowPreparationGate("ls PRJ-*.txt >/dev/null"))
	workflowLooseClassify(t, store, first, "critical", TicketStatusReadyForDev)
	workflowLooseCreate(t, store, first, "PRJ-91", "high", nil)
	workflowLooseCreate(t, store, first, "PRJ-92", "normal", []workflowstate.Ref{{Kind: "ticket", ID: first.Task.Key}})

	request.Selection.Order = workflowstate.TicketOrderPriority
	request.Policy.CheckpointStepLimit = 1
	request.Selection.Target.Commit, _ = workflowSelectionCommit(t, store)
	ports, provider := workflowAssurancePorts(t, store, request.Selection.Target.Commit)

	return store, request, ports, provider
}
