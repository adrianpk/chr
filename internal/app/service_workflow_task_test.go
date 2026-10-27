package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestServiceWorkflowTask executes an isolated worker through the shared service.
func TestServiceWorkflowTask(t *testing.T) {
	operations, hierarchy, target := serviceWorkflowTaskFixture(t)
	request := core.WorkflowStartRequest{RequestKey: "execute-task", Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	worker := []string{"sh", "-c", "printf 'Delivered\\n' > result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task result'"}

	proof, err := operations.ExecuteWorkflowTask(t.Context(), claim, worker)
	if err != nil {
		t.Fatal(err)
	}

	if !proof.Successful || proof.Commit == target || len(proof.Validation) != 1 {
		t.Fatal("service did not return host validation")
	}

	retained, err := operations.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || retained.Status != "completed" || retained.Attempts[0].ResultRef == "" {
		t.Fatal("service did not retain task evidence")
	}
}

func serviceWorkflowTaskFixture(t *testing.T, plans ...generationcore.GeneratedWorkflowPlan) (service, core.WorkflowMaterializationResult, string) {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git workflow integration runs in the nightly suite")
	}

	project := newTestWorkspace(t)
	if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\n")
	writeServiceWorkspaceFile(t, filepath.Join(project, "ops", "default", "plan", "approved.md"), "approved\n")

	operations := newTestService(t, project)

	plan := serviceWorkflowPlanFixture()
	if len(plans) == 1 {
		plan = plans[0]
	}

	plan.Slices[0].Tasks[0].Validation = workflowstate.Validation{Required: true, Commands: [][]string{{"git", "show", "HEAD:result.txt"}}, TimeoutSeconds: 30}

	hierarchy, err := operations.MaterializeWorkflowPlan(t.Context(), plan)
	if err != nil {
		t.Fatal(err)
	}

	serviceWorkflowPublishCoordinator(t, project)

	if output, err := exec.Command("git", "-C", project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-qm", "Fixture").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v: %s", err, output)
	}

	output, err := exec.Command("git", "-C", project, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}

	return operations, hierarchy, strings.TrimSpace(string(output))
}
