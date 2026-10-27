package core

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTicketFixture(t *testing.T, extent string, gate workflowstate.Validation) (*WorkflowStore, WorkflowRecord, WorkflowStartRequest) {
	t.Helper()

	store := workflowTestStore(t)
	contract := workflowPlanFixture().Slices[0].Tasks[0]
	contract.ID, contract.Label, contract.Validation = "independent", "Independent task", gate

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	record := WorkflowRecord{Kind: "ticket", Task: &WorkflowTask{Schema: WorkflowSchema, Key: "PRJ-90", Status: TicketStatusReadyForDev, Phase: "pending", ContractHash: contenthash.Sum(append([]byte("\x00\x00"), data...))}, Body: []byte(contract.Body), Domain: map[string]json.RawMessage{"execution_contract": data}}

	record, err = store.Create(t.Context(), "Independent task", record)
	if err != nil {
		t.Fatal(err)
	}

	commit, _ := workflowSelectionCommit(t, store)

	selector := workflowstate.Selector{Extent: extent}
	if extent == "ticket" {
		selector.Identity = record.Task.Key
	}

	request := WorkflowStartRequest{RequestKey: "independent", Selection: workflowstate.PreviewRequest{Selector: selector, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}}}

	return store, record, request
}

// TestWorkflowTicketCandidate proves that one real independent task is its own
// integration unit. Source validation is reused, never replaced by completion.
func TestWorkflowTicketCandidate(t *testing.T) {
	for _, extent := range []string{"ticket", "loose"} {
		t.Run(extent, func(t *testing.T) {
			store, record, request := workflowTicketFixture(t, extent, workflowPreparationGate("test -f result.txt"))

			run := workflowTestStart(t, store, request).Invocation
			if _, err := store.PrepareWorkflowTicketCandidate(t.Context(), run.ID); err == nil {
				t.Fatal("unvalidated task became an integration candidate")
			}

			claim := workflowTestClaim(t, store, run.ID)
			if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
				t.Fatal(err)
			}

			candidate, err := store.PrepareWorkflowTicketCandidate(t.Context(), run.ID)
			if err != nil {
				t.Fatal(err)
			}

			checkWorkflowTicketCandidate(t, store, record, run.ID, candidate)
		})
	}
}

func checkWorkflowTicketCandidate(t *testing.T, store *WorkflowStore, record WorkflowRecord, id string, candidate WorkflowSliceCandidate) {
	t.Helper()

	data, err := json.Marshal(candidate)
	if err != nil || strings.Contains(string(data), "\"slice\"") || strings.Contains(string(data), "\"set_id\"") || candidate.Ticket == nil || candidate.Ticket.Unit != (workflowstate.Ref{Kind: "ticket", ID: record.Task.Key}) {
		t.Fatal("candidate fabricated hierarchy", err)
	}

	checkWorkflowTicketCandidateRetry(t, store, id, candidate)
	checkWorkflowTicketNoHierarchy(t, store)
}

func checkWorkflowTicketCandidateRetry(t *testing.T, store *WorkflowStore, id string, candidate WorkflowSliceCandidate) {
	t.Helper()

	before := workflowTestRead(t, store, id)
	restarted := NewWorkflowStore(store.layout, store.projectID)

	again, err := restarted.PrepareWorkflowTicketCandidate(t.Context(), id)
	if err != nil || again.ID != candidate.ID || workflowTestRead(t, restarted, id).Revision != before.Revision || len(before.Attempts) != 1 || before.Status != "waiting_integration" {
		t.Fatal("candidate retry executed a gate or completed the invocation", err)
	}
}

func checkWorkflowTicketNoHierarchy(t *testing.T, store *WorkflowStore) {
	t.Helper()

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Set != nil || record.Task != nil && record.Task.Status == TicketStatusDone {
			t.Fatal("candidate preparation fabricated a set or marked a task done")
		}
	}
}

func workflowTicketPreparationFixture(t *testing.T, method, extent string) (workflowPreparationFixture, WorkflowStartRequest) {
	return workflowTicketConfiguredFixture(t, method, extent, nil)
}

func workflowTicketConfiguredFixture(t *testing.T, method, extent string, setup func(*WorkflowStore, WorkflowRecord, *WorkflowStartRequest)) (workflowPreparationFixture, WorkflowStartRequest) {
	t.Helper()

	gate := workflowPreparationGate("test -f result.txt && if test -f target.txt; then printf combined; else printf source; fi")

	store, record, request := workflowTicketFixture(t, extent, gate)
	if setup != nil {
		setup(store, record, &request)
	}

	run := workflowTestStart(t, store, request).Invocation

	claim := workflowTestClaim(t, store, run.ID)
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	candidate, err := store.PrepareWorkflowTicketCandidate(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	workflowTicketReviewConfig(t, store, method, workflowstate.ReviewTrustedAutoMerge)

	pr := WorkflowPullRequestSnapshot{URL: request.Selection.Target.Repository + "/pulls/7", Repository: request.Selection.Target.Repository, HeadRepository: request.Selection.Target.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: request.Selection.Target.Commit, State: "open"}
	provider := &workflowPreparationProviderFixture{workflowReviewProviderFixture: &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}, target: pr.BaseCommit}, guarded: true, revision: "guard-v1"}
	workflowPREnsure(t, store, run.ID, provider)

	root := t.TempDir()
	remote, actor := filepath.Join(root, "canonical.git"), filepath.Join(root, "target-writer")
	workflowIntegrationCommand(t, root, "init", "--bare", remote)
	workflowIntegrationCommand(t, candidate.Lane.Worktree, "push", remote, pr.BaseCommit+":refs/heads/dev", candidate.Result.Commit+":"+candidate.Lane.Branch)
	workflowIntegrationCommand(t, root, "clone", "--no-checkout", remote, actor)
	workflowIntegrationCommand(t, actor, "config", "user.name", "Test")
	workflowIntegrationCommand(t, actor, "config", "user.email", "test@example.com")
	workflowIntegrationCommand(t, actor, "checkout", "-b", "dev", "origin/dev")

	base := workflowIntegrationCommit(t, actor, "target.txt", "Concurrent target content")
	workflowIntegrationCommand(t, actor, "push", "origin", "dev")

	provider.pr.BaseCommit, provider.target = base, base

	f := workflowPreparationFixture{store: store, hierarchy: WorkflowMaterializationResult{Tasks: []WorkflowRecord{record}}, run: workflowTestRead(t, store, run.ID), provider: provider, git: NewWorkflowPreparationGit(pr.Repository, remote), actor: actor, remote: remote, base: base}

	return f, request
}

// TestWorkflowTicketDelivery applies the existing policies and real method
// histories to a single ticket. Lost responses require canonical observation;
// receipt finalization and terminal retries cannot run implementation again.
func TestWorkflowTicketDelivery(t *testing.T) {
	cases := []struct {
		method, extent string
		mode           workflowstate.ReviewMode
	}{
		{"merge", "ticket", workflowstate.ReviewTrustedAutoMerge},
		{"rebase", "ticket", workflowstate.ReviewTrustedAutoMerge},
		{"squash", "loose", workflowstate.ReviewTrustedAutoMerge},
		{"rebase", "ticket", workflowstate.ReviewHuman},
		{"rebase", "ticket", workflowstate.ReviewAgentic},
	}
	for _, tc := range cases {
		t.Run(tc.method+"/"+string(tc.mode), func(t *testing.T) { testWorkflowTicketDelivery(t, tc.method, tc.extent, tc.mode) })
	}
}

func testWorkflowTicketDelivery(t *testing.T, method, extent string, mode workflowstate.ReviewMode) {
	t.Helper()

	f, request := workflowTicketPreparationFixture(t, method, extent)
	p := &workflowSubmissionProviderFixture{workflowPreparationProviderFixture: f.provider, remote: f.remote, sealed: map[string]bool{}, lost: true}
	workflowTicketReviewConfig(t, f.store, method, mode)

	ports := workflowPublicationPorts(f, p)

	progress, err := workflowPublicationModeStart(t, f, p, ports, mode)
	if !errors.Is(err, context.DeadlineExceeded) || progress.Receipt != nil || p.calls != 1 {
		t.Fatal("unobserved response completed or repeated publication", err)
	}

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	final, err := restarted.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, ports)
	if err != nil || final.Receipt == nil || !workflowTicketReceiptMatches(final.Receipt, f.run.Lane.Unit, method) || final.Invocation.ConfirmedCursor != 1 || final.Invocation.Outcome != "ticket_integrated" {
		t.Fatal("ticket did not finalize from actual canonical delivery", err)
	}

	checkWorkflowTicketDelivery(t, restarted, f, final, extent)

	checkWorkflowTicketTerminalRetry(t, restarted, f, request, final, p)
}

func checkWorkflowTicketTerminalRetry(t *testing.T, restarted *WorkflowStore, f workflowPreparationFixture, request WorkflowStartRequest, final WorkflowIntegrationProgress, p *workflowSubmissionProviderFixture) {
	t.Helper()

	again := workflowTestStart(t, restarted, request)
	if !again.Attached || again.Invocation.ID != f.run.ID || again.Invocation.Revision != final.Invocation.Revision || p.calls != 1 || f.provider.creates != 1 {
		t.Fatal("retry changed invocation or duplicated delivery")
	}
}

func checkWorkflowTicketDelivery(t *testing.T, store *WorkflowStore, f workflowPreparationFixture, progress WorkflowIntegrationProgress, extent string) {
	t.Helper()

	record, err := store.Read(t.Context(), f.hierarchy.Tasks[0].Path)
	if err != nil || record.Task.Status != TicketStatusDone || record.Task.Phase != "integrated" || record.Task.DeliveryRef != progress.Receipt.Path || record.Task.SetID != "" || record.Task.SliceID != "" {
		t.Fatal("independent delivery projection differs from its receipt", err)
	}

	checkWorkflowTicketCombinedGate(t, store, f, progress, extent)
}

func checkWorkflowTicketCombinedGate(t *testing.T, store *WorkflowStore, f workflowPreparationFixture, progress WorkflowIntegrationProgress, extent string) {
	t.Helper()

	expectedStatus := "completed"
	if extent == "loose" {
		expectedStatus = "waiting_integration"
	}

	if progress.Invocation.Status != expectedStatus || progress.Invocation.Preparation == nil {
		t.Fatal("ticket delivery bypassed its frontier or combined gate")
	}

	run, err := store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	prepared, err := store.readWorkflowPreparation(run)
	if err != nil || prepared.Validation.Validation[0].Stdout != "combined" || prepared.Input.Source.Unit.Kind != "ticket" {
		t.Fatal("combined-result gate was not bound to the independent ticket", err)
	}
}

func workflowTicketReviewConfig(t *testing.T, store *WorkflowStore, method string, mode workflowstate.ReviewMode) {
	t.Helper()

	writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "review", "mode", string(mode))
	writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "work", "merge_strategy", method)
}

func workflowTicketReceiptMatches(record *WorkflowRecord, unit workflowstate.Ref, method string) bool {
	return record != nil && record.Receipt != nil && record.Receipt.Unit == unit && record.Receipt.Method == method
}
