package workflow

import (
	"encoding/json"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestWorkflowTaskIndependent keeps validation distinct from required integration.
func TestWorkflowTaskIndependent(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	record := workflowIndependentTask(t, hierarchy.Tasks[0])

	created, err := store.Create(t.Context(), "Independent", record)
	if err != nil {
		t.Fatal(err)
	}

	request.Selection.Selector.Identity = created.Task.Key
	run := workflowTestStart(t, store, request)
	proof := workflowTestExecute(t, store, run.Invocation.ID, "worker")

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	if retained.Status != "waiting_integration" || retained.Outcome != "task_validated" || !proof.IntegrationPending {
		t.Fatal("independent ticket completed without integration")
	}

	if workflowInvocationTerminal(retained.Status) {
		t.Fatal("independent ticket released delivery ownership")
	}
}

func workflowIndependentTask(t *testing.T, grouped Record) Record {
	t.Helper()

	var contract generationcore.GeneratedWorkflowTask
	if err := json.Unmarshal(grouped.Domain["execution_contract"], &contract); err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	task := *grouped.Task
	task.Key, task.SetID, task.SliceID, task.UseCase, task.Order = "PRJ-90", "", "", "", 0
	task.ContractHash = contenthash.Sum(append([]byte("\x00\x00"), data...))

	return Record{Kind: "ticket", Task: &task, Body: grouped.Body, Domain: map[string]json.RawMessage{"execution_contract": data}}
}
