package app

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestServiceWorkRepair publishes a revised exact source onto the same bound PR
// with its retained previous-head lease. It neither creates a second PR nor
// replays the successful original worker or publishes the canonical target.
func TestServiceWorkRepair(t *testing.T) {
	operations, ticket, request, provider := serviceWorkNativeFixture(t)

	first := serviceWorkApprovalRun(t, operations)

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.RepairWorkflowSlice(t.Context(), first.ID, ticket.Task.Key); err != nil {
		t.Fatal(err)
	}

	operations.workflowWorker = []string{"sh", "-c", "printf 'Repair\\n' >> result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Repair task'"}

	second, err := operations.StartAllReadyAgentWork()
	checkServiceWorkRepaired(t, first, second, provider.creates, err)

	source, err := provider.ReadWorkflowBranch(t.Context(), provider.repository, provider.pr.HeadBranch)

	target, targetErr := provider.ReadWorkflowBranch(t.Context(), provider.repository, "dev")
	if err != nil || targetErr != nil || source != second.Workflow.PullRequest.SourceCommit || target != request.Selection.Target.Commit {
		t.Fatal("repair changed the target or lost exact source publication", err, targetErr)
	}
}

func serviceWorkApprovalRun(t *testing.T, operations service) core.WorkflowInvocation {
	t.Helper()

	first, err := operations.StartAllReadyAgentWork()
	if err != nil || first.Workflow == nil || first.Workflow.Review == nil || first.Workflow.Review.State != "waiting_approval" {
		t.Fatal("initial candidate did not wait for approval", err)
	}

	return *first.Workflow
}

func checkServiceWorkRepaired(t *testing.T, first core.WorkflowInvocation, second AgentWorkResult, creates int, failure error) {
	t.Helper()

	if failure != nil || second.Workflow == nil || second.Workflow.ID != first.ID || second.Workflow.Review == nil || second.Workflow.Review.State != "waiting_approval" || creates != 1 || second.Workflow.PullRequest.SourceCommit == first.PullRequest.SourceCommit {
		t.Fatal("repair failed to reuse bound PR source", failure)
	}
}
