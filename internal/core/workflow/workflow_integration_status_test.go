package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeQueueCapacity observes real claim locks and unexpired process
// identity without acquiring capacity or pruning historical process ledgers.
func TestNativeQueueCapacity(t *testing.T) {
	store, _, request := workflowExecutionGateFixture(t, workflowPreparationGate("true"))
	run := workflowTestStart(t, store, request).Invocation
	claim := workflowTestClaim(t, store, run.ID)
	before := nativeQueueImages(t, store.layout.Root())

	capacity, err := store.QueryWorkflowIntegrationCapacity(t.Context())
	if err != nil || capacity.Available != 0 || len(capacity.UnexpiredOwners) != 1 || capacity.UnexpiredOwners[0].Role != "executor" {
		t.Fatal("query ignored live owner or actual reservation", capacity, err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("capacity query changed authority")
	}

	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker")); err != nil {
		t.Fatal(err)
	}

	capacity, err = store.QueryWorkflowIntegrationCapacity(t.Context())
	if err != nil || capacity.Available != capacity.Maximum || len(capacity.UnexpiredOwners) != 0 {
		t.Fatal("finished role retained capacity", capacity, err)
	}
}

// TestNativeQueueDeadProcess keeps dead process evidence on disk while the
// read-only observation exposes no occupied slot; cleanup remains a mutation.
func TestNativeQueueDeadProcess(t *testing.T) {
	store, _ := nativeEntryFixture(t)
	identity, stop := nativeSubmissionDispatcher(t)
	stop()

	process := workflowstate.Process{Identity: identity, GroupID: identity.PID, CapacitySlot: 1}

	common, err := projectcore.ResolveRepositoryDirectory(t.Context(), store.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	name, err := workflowCapacityProcessName(process)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(common, "chr-workflow-processes", name)

	data, err := json.Marshal(process)
	if err != nil {
		t.Fatal(err)
	}

	if err := opsfile.WriteAtomic(path, data); err != nil {
		t.Fatal(err)
	}

	capacity, err := store.QueryWorkflowIntegrationCapacity(t.Context())
	if err != nil || len(capacity.LiveProcesses) != 0 {
		t.Fatal(capacity, err)
	}

	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(data) {
		t.Fatal("query pruned dead process source", err)
	}
}

// TestNativeQueueClosure validates actual receipt fanout and a retained final
// gate without running a gate or rewriting the set merely to report status.
func TestNativeQueueClosure(t *testing.T) {
	f, ports := workflowClosingFixture(t)

	set := workflowClosingContract(t, f, workflowPreparationGate("true"))

	before := nativeQueueImages(t, f.store.layout.Root())

	status, err := f.store.QueryWorkflowIntegrationSetClosure(t.Context(), set.Path)
	if err != nil || !status.ReceiptsVerified || status.GateVerified || status.CanClose {
		t.Fatal("status invented final gate", status, err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, f.store.layout.Root())) {
		t.Fatal("closure query changed set")
	}

	nativeQueueClosedStatus(t, f, ports)
}

func nativeQueueClosedStatus(t *testing.T, f workflowIntegrationFixture, ports FrontierPorts) {
	t.Helper()

	closed, err := f.store.CloseWorkflowInvocationSet(t.Context(), f.run.ID, "dev", ports)
	if err != nil {
		t.Fatal(err)
	}

	before := nativeQueueImages(t, f.store.layout.Root())

	status, err := f.store.QueryWorkflowIntegrationSetClosure(t.Context(), closed.Set.Path)
	if err != nil || !status.CanClose || !status.GateVerified {
		t.Fatal("verified closing receipt was lost", status, err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, f.store.layout.Root())) {
		t.Fatal("status reran confirmed closure")
	}
}
