package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func serviceWorkflowSliceFixture(t *testing.T) (service, workflowcore.MaterializationResult, string) {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git workflow integration runs in the nightly suite")
	}

	project := newTestWorkspace(t)
	if output, err := exec.CommandContext(t.Context(), "git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\n")
	writeServiceWorkspaceFile(t, filepath.Join(project, "ops", "default", "plan", "approved.md"), "approved\n")

	operations := newTestService(t, project)
	plan := serviceWorkflowPlanFixture()

	plan.Slices[0].Validation = workflowstate.Validation{Required: true, Commands: [][]string{{"sh", "-c", "test \"$(wc -l < result.txt)\" -eq 3"}}, TimeoutSeconds: 30}
	for i := range plan.Slices[0].Tasks {
		plan.Slices[0].Tasks[i].Validation = workflowstate.Validation{Required: true, Commands: [][]string{{"git", "show", "HEAD:result.txt"}}, TimeoutSeconds: 30}
	}

	hierarchy, err := operations.MaterializeWorkflowPlan(t.Context(), plan)
	if err != nil {
		t.Fatal(err)
	}

	serviceWorkflowPublishCoordinator(t, project)

	if output, err := exec.CommandContext(t.Context(), "git", "-C", project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-qm", "Fixture").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v: %s", err, output)
	}

	output, err := exec.CommandContext(t.Context(), "git", "-C", project, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}

	return operations, hierarchy, strings.TrimSpace(string(output))
}

// TestServiceWorkflowSlice runs cumulative workers and a distinct aggregate gate
// through the shared service, reuses host proof and explicitly invalidates repair.
func TestServiceWorkflowSlice(t *testing.T) {
	operations, hierarchy, target := serviceWorkflowSliceFixture(t)
	request := workflowcore.StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "slice",
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice", Identity: hierarchy.Set.Set.Slices[0].ID}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}},
	}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	worker := []string{"sh", "-c", "printf 'Delivered\\n' >> result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task result'"}

	candidate, err := operations.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, worker)
	if err != nil || !candidate.Result.Successful || len(candidate.Slice.Tasks) != 3 || candidate.Result.Commit == target {
		t.Fatalf("service aggregate candidate: %+v, %v", candidate, err)
	}

	checkServiceWorkflowSlice(t, operations, run.Invocation.ID, hierarchy, candidate)
}

func checkServiceWorkflowSlice(t *testing.T, operations service, id string, hierarchy workflowcore.MaterializationResult, candidate workflowcore.SliceCandidate) {
	t.Helper()

	read, err := operations.ReadWorkflowSliceCandidate(t.Context(), id)
	if err != nil || read.ID != candidate.ID {
		t.Fatal("service lost retained aggregate proof")
	}

	replayed, err := operations.ValidateWorkflowSlice(t.Context(), id)
	if err != nil || replayed.ID != candidate.ID {
		t.Fatal("service repeated accepted aggregate validation")
	}

	retained, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || retained.Status != workflowstate.InvocationWaitingIntegration || len(retained.Attempts) != 4 {
		t.Fatal("service skipped required integration or repeated tasks")
	}

	checkServiceWorkflowSliceRepair(t, operations, id, hierarchy.Tasks[2].Task.Key)
}

func checkServiceWorkflowSliceRepair(t *testing.T, operations service, id, key string) {
	t.Helper()

	repaired, err := operations.RepairWorkflowSlice(t.Context(), id, key)
	if err != nil || repaired.Cursor != 2 || repaired.CandidateRef != "" || repaired.Status != workflowstate.InvocationPending {
		t.Fatal("service repair retained affected aggregate acceptance")
	}
}
