package app

import (
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestServiceWorkflowClosing executes an explicit empty-set gate through the
// shared service on a real bare target, retaining the exact target and receipt.
func TestServiceWorkflowClosing(t *testing.T) {
	operations, _, target := serviceWorkflowSliceFixture(t)

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	plan := serviceWorkflowPlanFixture()

	set, err := store.Create(t.Context(), "Empty service set", core.WorkflowRecord{Kind: "plan", Set: &core.WorkflowSet{Schema: core.WorkflowSchema, Status: "ready", PlanRef: plan.PlanRef, PlanHash: plan.PlanHash, ClosingValidation: workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"git", "rev-parse", "HEAD"}}}}})
	if err != nil {
		t.Fatal(err)
	}

	project := operations.projectContext.ProjectDir
	remote := filepath.Join(t.TempDir(), "canonical.git")
	repository := "https://forge.example/test/repo"

	serviceWorkflowDeliveryGit(t, project, "init", "--bare", remote)
	serviceWorkflowDeliveryGit(t, project, "push", remote, target+":refs/heads/dev")

	operations.workflowIntegrationGit = core.NewWorkflowIntegrationGit(repository, remote)
	operations.workflowPullRequests = &serviceWorkflowPRFixture{pr: core.WorkflowPullRequestSnapshot{HeadCommit: target}}
	request := core.WorkflowSetClosureRequest{SetPath: set.Path, ExpectedRevision: set.Revision, ExpectedSourceHash: set.SourceHash, Repository: repository, Branch: "dev"}

	closed, err := operations.CloseWorkflowSet(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	if closed.Set.Set.ClosureRef != closed.ReceiptRef || closed.Outcome != "completed" || closed.Receipt.Target.Commit != target {
		t.Fatal("shared closing operation lost canonical confirmation")
	}

	if len(closed.Receipt.Validation) != 1 || closed.Receipt.Validation[0].Stdout != target+"\n" {
		t.Fatal("shared closing operation executed another target")
	}
}
