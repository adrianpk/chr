package app

import (
	"os/exec"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestServiceWorkflowManual keeps preparation, harness execution and explicit
// evidence reconciliation behind the same adapter-neutral service operation.
func TestServiceWorkflowManual(t *testing.T) {
	operations, hierarchy, target := serviceWorkflowTaskFixture(t)
	request := core.WorkflowStartRequest{RequestKey: "manual-task", Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	worker := []string{"sh", "-c", "printf 'Delivered\\n' > result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Manual result'"}

	task, err := operations.PrepareWorkflowManualTask(t.Context(), claim, worker)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ExecuteWorkflowManualTask(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	checkServiceWorkflowManualResult(t, operations, claim, run.Invocation.ID, task, target)
}

func checkServiceWorkflowManualResult(t *testing.T, operations service, claim *core.WorkflowClaim, id string, task core.WorkflowManualTask, target string) {
	t.Helper()

	candidate := serviceWorkflowManualCandidate(t, task)

	proof, err := operations.ReconcileWorkflowTask(t.Context(), claim, candidate)
	if err != nil || !proof.Successful || proof.ExecutionMode != "manual" || proof.Commit == target {
		t.Fatalf("manual service evidence: %+v, %v", proof, err)
	}

	retained, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || retained.Status != "completed" || retained.Attempts[0].ResultRef == "" {
		t.Fatal("manual service lost retained evidence")
	}
}

func serviceWorkflowManualCandidate(t *testing.T, task core.WorkflowManualTask) core.WorkflowManualCandidate {
	t.Helper()

	lane := task.Input.Lane.Worktree

	output, err := exec.CommandContext(t.Context(), "git", "-C", lane, "rev-parse", "HEAD", "HEAD^{tree}").Output()
	if err != nil {
		t.Fatal(err)
	}

	identities := strings.Fields(string(output))
	if len(identities) != 2 {
		t.Fatal("candidate commit/tree missing")
	}

	return core.WorkflowManualCandidate{InputHash: task.InputHash, Branch: task.Input.Lane.Branch, Commit: identities[0], Tree: identities[1]}
}
