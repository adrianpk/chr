package core

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

func workflowLooseBlockedFixture(t *testing.T) workflowIntegrationFixture {
	t.Helper()

	f := workflowLooseConfiguredQueue(t, false, func(store *WorkflowStore, _ *WorkflowStartRequest) {
		first := workflowLooseRecord(t, store, "PRJ-90")
		outside := workflowLooseCreate(t, store, first, "PRJ-89", "low", nil)
		workflowLooseClassify(t, store, outside, "low", TicketStatusBacklog)
		workflowLooseDependency(t, store, "PRJ-91", "PRJ-89")
	})
	workflowFinalizeFrontier(t, f)
	workflowAdvanceFrontier(t, f)

	return f
}

func workflowLooseDependency(t *testing.T, store *WorkflowStore, key, dependency string) {
	t.Helper()

	record := workflowLooseRecord(t, store, key)

	var contract GeneratedWorkflowTask
	if err := json.Unmarshal(record.Domain["execution_contract"], &contract); err != nil {
		t.Fatal(err)
	}

	contract.DependsOn = []WorkflowRef{{Kind: "ticket", ID: dependency}}

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	record.Task.DependsOn, record.Domain["execution_contract"] = contract.DependsOn, data
	record.Task.ContractHash = contenthash.Sum(append([]byte("\x00\x00"), data...))
	workflowTestWrite(t, store, record)
}

// TestWorkflowLooseBlocker retains the first dependency refusal before any
// successor or outside-scope task can acquire a worker or displace the cursor.
func TestWorkflowLooseBlocker(t *testing.T) {
	f := workflowLooseBlockedFixture(t)

	before := workflowTestRead(t, f.store, f.run.ID)
	if _, err := f.store.ClaimWorkflow(t.Context(), before.ID); err == nil {
		t.Fatal("unmet dependency acquired a worker")
	}

	blocked := workflowTestRead(t, f.store, before.ID)
	if blocked.Status != "paused" || blocked.Blocker == nil || blocked.Blocker.Dependency.ID != "PRJ-89" || blocked.Cursor != 1 || len(blocked.Attempts) != len(before.Attempts) {
		t.Fatal("dependency refusal did not retain an invocation stop")
	}

	workflowLooseClassify(t, f.store, workflowLooseRecord(t, f.store, "PRJ-92"), "critical", TicketStatusReadyForDev)

	retained, err := f.store.ReconcileWorkflowMembership(t.Context(), before.ID)
	if err != nil || retained.Members[1].Admission.Key != "PRJ-91" || !reflect.DeepEqual(retained.LooseQueue.Admitted, before.LooseQueue.Admitted) {
		t.Fatal("rank change bypassed a retained dependency stop", err)
	}

	checkWorkflowLooseBlockedRetry(t, f, before.ID)
}

func checkWorkflowLooseBlockedRetry(t *testing.T, f workflowIntegrationFixture, id string) {
	t.Helper()

	if _, err := f.store.ClaimWorkflow(t.Context(), id); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("dependency stop accepted a direct claim", err)
	}

	if _, err := f.store.ResumeWorkflow(t.Context(), id); err == nil {
		t.Fatal("resume cleared an unverified prerequisite")
	}
}
