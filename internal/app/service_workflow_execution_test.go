package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestServiceWorkflowClaim uses the shared service for file-owned admission and checkpoints.
func TestServiceWorkflowClaim(t *testing.T) {
	project := newTestWorkspace(t)
	if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\n")
	writeServiceWorkspaceFile(t, filepath.Join(project, "ops", "default", "plan", "approved.md"), "approved\n")

	operations := newTestService(t, project)

	hierarchy, err := operations.MaterializeWorkflowPlan(t.Context(), serviceWorkflowPlanFixture())
	if err != nil {
		t.Fatal(err)
	}

	if output, err := exec.Command("git", "-C", project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-qm", "Fixture").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v: %s", err, output)
	}

	output, err := exec.Command("git", "-C", project, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}

	request := core.WorkflowStartRequest{RequestKey: "manual-claim", Selection: core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: core.WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: strings.TrimSpace(string(output))}}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	checkServiceWorkflowClaim(t, operations, run.Invocation.ID)
}

func checkServiceWorkflowClaim(t *testing.T, operations service, id string) {
	t.Helper()

	claim, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if err := operations.CheckWorkflowClaim(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	if err := operations.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	read, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || read.Status != "paused" || read.Cursor != 0 || len(read.Attempts) != 1 {
		t.Fatalf("checkpoint = %+v, %v", read, err)
	}
}
