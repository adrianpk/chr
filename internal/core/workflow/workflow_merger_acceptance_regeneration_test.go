package workflow

import (
	"slices"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeMergerRegenerationFixture(t *testing.T, failing bool) (workflowPreparationFixture, MergerRequest, MergerHarness) {
	t.Helper()

	f, request, harness, _ := nativeMergerAcceptanceFixture(t, "merger-regenerate", "merge")
	input := &request.Input
	input.AllowedPaths = []string{"result.txt", "manual.txt", "lock.txt"}
	conflict := &input.Conflicts[0]
	conflict.Kind = "generated"
	conflict.RegenerationCommand = []string{"sh", "-c", "test -f manual.txt && printf 'Preserved source and target behavior\\n' > result.txt"}
	conflict.ValidationCommand = slices.Clone(input.ValidationCommands[0])
	second := *conflict
	second.ID, second.Path, second.Kind = "conflict-2", "lock.txt", "lockfile"

	second.RegenerationCommand = []string{"sh", "-c", "test -f result.txt && printf 'Regenerated lock state\\n' > lock.txt"}
	if failing {
		second.RegenerationCommand = []string{"sh", "-c", "exit 7"}
	}

	input.Conflicts = append(input.Conflicts, second)

	var err error

	*input, err = SealWorkflowMergerInput(*input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	return f, request, harness
}

// TestNativeMergerHostRegeneration records real owned subprocesses and the exact
// manual-input/generated-tree chain; model-reported green checks cannot supply it.
func TestNativeMergerHostRegeneration(t *testing.T) {
	f, request, harness := nativeMergerRegenerationFixture(t, false)

	attempt, err := f.store.ExecuteWorkflowMergerRepair(t.Context(), request, harness, f.provider, f.git)
	if err != nil || attempt.Acceptance == nil || attempt.Acceptance.Outcome != "accepted_pending_validation" {
		t.Fatal("host regeneration rejected", attempt.Failure, err)
	}

	checkNativeMergerRegenerationChain(t, *attempt.Acceptance.Repair)

	run := workflowTestRead(t, f.store, f.run.ID)
	if len(run.Attempts[len(run.Attempts)-1].Processes) != 5 {
		t.Fatal("procedures bypassed owned process supervision")
	}

	if len(attempt.Checks) != 0 {
		t.Fatal("model check observations acquired host evidence")
	}

	next := attempt.Acceptance.Next
	if next.Result == nil || len(next.Result.Provenance.Repairs[0].Regeneration) != 2 {
		t.Fatal("method proof dropped host regeneration")
	}

	checkNativeMergerAcceptanceRetainedProcedures(t, f)
}

func checkNativeMergerAcceptanceRetainedProcedures(t *testing.T, f workflowPreparationFixture) {
	t.Helper()

	run := workflowTestRead(t, f.store, f.run.ID)
	for _, process := range run.Attempts[len(run.Attempts)-1].Processes {
		if process.State != ProcessStopped || process.StdoutHash == "" || process.StderrHash == "" {
			t.Fatal("host procedure has unresolved ownership")
		}
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.run.ID)
	if err != nil || budget.UnitCounters[0].Merger != 1 {
		t.Fatal("host procedures consumed a second agent call", err)
	}

	if _, err := f.store.CheckWorkflowPublicationGuard(t.Context(), f.run.ID, f.provider); err == nil {
		t.Fatal("declared local validation replaced candidate gate/review")
	}
}

// TestNativeMergerHostRegenerationFailure retains a failed real command and
// its consumed call without creating a prepared result or replacing the gate.
func TestNativeMergerHostRegenerationFailure(t *testing.T) {
	f, request, harness := nativeMergerRegenerationFixture(t, true)

	attempt, err := f.store.ExecuteWorkflowMergerRepair(t.Context(), request, harness, f.provider, f.git)
	if err == nil || attempt.Acceptance == nil || attempt.Acceptance.Outcome != "rejected" || attempt.Acceptance.Next != nil {
		t.Fatal("failed procedure accepted", err)
	}

	evidence := attempt.Acceptance.Repair.Regeneration
	if len(evidence) != 2 || evidence[1].Command.ExitCode != 7 {
		t.Fatal("failed host procedure observation was dropped")
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	latest := run.Attempts[len(run.Attempts)-1]

	data, readErr := f.store.readWorkflowReviewImage(latest.ResultRef, latest.ResultHash)
	if readErr != nil || !strings.Contains(string(data), "\"exit_code\": 7") {
		t.Fatal("failure image did not retain exact command evidence", readErr)
	}

	checkNativeMergerAcceptanceRetainedProcedures(t, f)
}

func checkNativeMergerRegenerationChain(t *testing.T, repair workflowstate.IntegrationTreeRepair) {
	t.Helper()

	if len(repair.Regeneration) != 2 || repair.ProposalTree == repair.BaselineTree || repair.Regeneration[0].InputTree != repair.ProposalTree || repair.Regeneration[0].OutputTree != repair.Regeneration[1].InputTree || repair.Regeneration[1].OutputTree != repair.ResolvedTree {
		t.Fatal("host procedure tree chain was not retained")
	}
}
