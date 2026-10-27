package core

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowLooseQueueFixture(t *testing.T, includeLater bool) workflowIntegrationFixture {
	return workflowLooseConfiguredQueue(t, includeLater, nil)
}

func workflowLooseConfiguredQueue(t *testing.T, includeLater bool, setup func(*WorkflowStore, *WorkflowStartRequest)) workflowIntegrationFixture {
	t.Helper()

	f, _ := workflowTicketConfiguredFixture(t, "rebase", "loose", func(store *WorkflowStore, first WorkflowRecord, request *WorkflowStartRequest) {
		request.Selection.Order = workflowstate.TicketLoopOrderPriority
		request.Policy.IncludeLaterTickets = includeLater

		workflowLooseClassify(t, store, first, "critical", TicketStatusReadyForDev)
		workflowLooseCreate(t, store, first, "PRJ-91", "high", nil)
		workflowLooseCreate(t, store, first, "PRJ-92", "normal", []workflowstate.Ref{{Kind: "ticket", ID: first.Task.Key}})

		if setup != nil {
			setup(store, request)
		}
	})
	workflowIntegrationCommand(t, f.actor, "branch", "source", "origin/"+f.provider.pr.HeadBranch)
	workflowFixtureIntegrate(t, f.actor, "rebase")

	result := workflowIntegrationCommand(t, f.actor, "rev-parse", "HEAD")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	f.provider.pr.State, f.provider.pr.Merged, f.provider.pr.ResultCommit = "closed", true, result
	f.provider.pr.BaseCommit, f.provider.target = result, result

	return workflowIntegrationFixture{store: f.store, run: workflowTestRead(t, f.store, f.run.ID), provider: f.provider.workflowPRProviderFixture, git: NewWorkflowIntegrationGit(f.provider.pr.Repository, f.remote), actor: f.actor, remote: f.remote, base: f.base, result: result, target: result}
}

func workflowLooseCreate(t *testing.T, store *WorkflowStore, original WorkflowRecord, key, priority string, dependencies []workflowstate.Ref) WorkflowRecord {
	t.Helper()

	var contract generationcore.GeneratedWorkflowTask
	if err := json.Unmarshal(original.Domain["execution_contract"], &contract); err != nil {
		t.Fatal(err)
	}

	contract.ID, contract.Label, contract.DependsOn = key, key, dependencies

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	record := WorkflowRecord{Kind: "ticket", Task: &WorkflowTask{Schema: WorkflowSchema, Key: key, Status: TicketStatusReadyForDev, Phase: "pending", DependsOn: dependencies, ContractHash: contenthash.Sum(append([]byte("\x00\x00"), data...))}, Body: workflowLooseBody(key, priority), Domain: map[string]json.RawMessage{"execution_contract": data}}

	created, err := store.Create(t.Context(), key, record)
	if err != nil {
		t.Fatal(err)
	}

	return created
}

func workflowLooseBody(key, priority string) []byte {
	return []byte(fmt.Sprintf("---\nkey: %s\ntitle: Independent task\nstatus: ready_for_dev\ncreated_at: 2026-01-02T03:04:05Z\nupdated_at: 2026-01-02T03:04:05Z\npriority: %s\nseverity: high\nproject_area: frontend\ntype: task\ntags: [delivery, ui]\n---\nFrozen description\n", key, priority))
}

func workflowLooseClassify(t *testing.T, store *WorkflowStore, record WorkflowRecord, priority, status string) WorkflowRecord {
	t.Helper()

	record.Body, record.Task.Status = workflowLooseBody(record.Task.Key, priority), status

	changed, err := store.Write(t.Context(), record, record.Revision)
	if err != nil {
		t.Fatal(err)
	}

	return changed
}

func workflowLooseRecord(t *testing.T, store *WorkflowStore, key string) WorkflowRecord {
	t.Helper()

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Task != nil && record.Task.Key == key {
			return record
		}
	}

	t.Fatal("missing queue ticket", key)

	return WorkflowRecord{}
}

// TestWorkflowLooseReselection changes pending rank and status while the previous
// ticket owns its lane. Only verified delivery releases the next eligible unit.
func TestWorkflowLooseReselection(t *testing.T) {
	f := workflowLooseQueueFixture(t, false)
	first := workflowTestRead(t, f.store, f.run.ID)
	workflowLooseClassify(t, f.store, workflowLooseRecord(t, f.store, "PRJ-91"), "high", TicketStatusBacklog)
	workflowLooseClassify(t, f.store, workflowLooseRecord(t, f.store, "PRJ-92"), "critical", TicketStatusReadyForDev)

	before, err := f.store.ReconcileWorkflowMembership(t.Context(), first.ID)
	if err != nil || before.Members[0].Admission.Key != "PRJ-90" || before.Members[1].Admission.Key != "PRJ-91" || before.Lane.Unit.ID != "PRJ-90" {
		t.Fatal("pending classification preempted a retained candidate", err)
	}

	workflowFinalizeFrontier(t, f)

	advanced := workflowAdvanceFrontier(t, f)

	run, err := f.store.ReconcileWorkflowMembership(t.Context(), first.ID)
	checkWorkflowLooseReselected(t, first, run, err)
	checkWorkflowLooseNext(t, f, advanced, run)
}

func checkWorkflowLooseReselected(t *testing.T, first, run WorkflowInvocation, failure error) {
	t.Helper()

	if failure != nil || run.Cursor != 1 || run.ConfirmedCursor != 1 || run.Members[1].Admission.Key != "PRJ-92" || run.LooseQueue.Excluded["PRJ-91"] != "status_mismatch" || !slices.Equal(run.LooseQueue.Admitted, first.LooseQueue.Admitted) {
		t.Fatal("fresh queue lost frozen scope or selected unready work", failure)
	}
}

func checkWorkflowLooseNext(t *testing.T, f workflowIntegrationFixture, advanced, run WorkflowInvocation) {
	t.Helper()

	claim := workflowTestClaim(t, f.store, run.ID)
	if _, err := f.store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal("canonical prerequisite delivery was not observed", err)
	}

	candidate, err := f.store.PrepareWorkflowTicketCandidate(t.Context(), run.ID)
	if err != nil || candidate.Ticket.Unit.ID != "PRJ-92" || candidate.Lane.BaseCommit != f.target || candidate.Lane.Branch == advanced.Deliveries[0].Lane.Branch {
		t.Fatal("successor reused a lane or stale target", err)
	}
}
