package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func serviceWorkflowTicketFixture(t *testing.T, plans ...generationcore.GeneratedWorkflowPlan) (service, workflowcore.Record, workflowcore.StartRequest) {
	t.Helper()

	operations, hierarchy, target := serviceWorkflowTaskFixture(t, plans...)
	grouped := hierarchy.Tasks[0]
	task := *grouped.Task
	task.Key, task.SetID, task.SliceID, task.UseCase, task.Order = "PRJ-90", "", "", "", 0

	var contract generationcore.GeneratedWorkflowTask
	if err := json.Unmarshal(grouped.Domain["execution_contract"], &contract); err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	hash := sha256.Sum256(append([]byte("\x00\x00"), data...))
	task.ContractHash = hex.EncodeToString(hash[:])

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	record, err := store.Create(t.Context(), "Independent task", workflowcore.Record{Kind: "ticket", Task: &task, Domain: grouped.Domain, Body: grouped.Body})
	if err != nil {
		t.Fatal(err)
	}

	request := workflowcore.StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "independent",
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: task.Key}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}},
	}}

	return operations, record, request
}

// TestServiceWorkflowTicket drives an independent worker through shared service
// candidate/PR operations and proves that a bound PR cannot mark its task done.
func TestServiceWorkflowTicket(t *testing.T) {
	operations, record, request := serviceWorkflowTicketFixture(t)

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	worker := []string{"sh", "-c", "printf 'Delivered\n' > result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task result'"}
	if _, err := operations.ExecuteWorkflowTask(t.Context(), claim, worker); err != nil {
		t.Fatal(err)
	}

	candidate, err := operations.PrepareWorkflowTicketCandidate(t.Context(), run.Invocation.ID)
	if err != nil || candidate.Ticket == nil || candidate.Ticket.Unit.ID != record.Task.Key {
		t.Fatal("service lost independent candidate", err)
	}

	checkServiceWorkflowTicketPR(t, operations, run.Invocation.ID, request, candidate)
}

func checkServiceWorkflowTicketPR(t *testing.T, operations service, id string, request workflowcore.StartRequest, candidate workflowcore.SliceCandidate) {
	t.Helper()

	provider := &serviceWorkflowPRFixture{pr: workflowstate.PullRequestSnapshot{URL: request.Selection.Target.Repository + "/pulls/7", Repository: request.Selection.Target.Repository, HeadRepository: request.Selection.Target.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: request.Selection.Target.Commit, State: "open"}}
	operations.workflowPullRequests = provider

	content := workflowcore.SlicePullRequestRequest{BaseBranch: "dev", Title: "Deliver independent task", Body: "Deliver its host-validated candidate."}
	for range 2 {
		if _, err := operations.EnsureWorkflowSlicePullRequest(t.Context(), id, content); err != nil {
			t.Fatal(err)
		}
	}

	retained, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || retained.Status != "waiting_review" || provider.creates != 1 || retained.PullRequest.Unit.Kind != "ticket" || len(retained.Attempts) != 1 {
		t.Fatal("service completed without delivery or duplicated candidate work", err)
	}
}
