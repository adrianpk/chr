package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func serviceWorkflowDeliveryFixture(t *testing.T) (service, string, *serviceWorkflowPRFixture) {
	t.Helper()

	return serviceWorkflowDeliveryExtentFixture(t, "slice")
}

func serviceWorkflowDeliveryExtentFixture(t *testing.T, extent string) (service, string, *serviceWorkflowPRFixture) {
	t.Helper()

	operations, hierarchy, target := serviceWorkflowSliceFixture(t)
	repository := "https://forge.example/test/repo"

	identity := hierarchy.Set.Set.Slices[0].ID
	if extent == "slice_set" {
		identity = hierarchy.Set.ID
	}

	request := core.WorkflowStartRequest{RequestKey: "external-delivery", Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: extent, Identity: identity}, Target: workflowstate.PreviewTarget{Repository: repository, Commit: target}}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	worker := []string{"sh", "-c", "printf 'Task\\n' >> result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task'"}

	candidate, err := operations.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, worker)
	if err != nil {
		t.Fatal(err)
	}

	provider := &serviceWorkflowPRFixture{pr: core.WorkflowPullRequestSnapshot{URL: repository + "/pulls/7", Repository: repository, HeadRepository: repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: target, State: "open"}}
	operations.workflowPullRequests = provider

	content := core.WorkflowSlicePullRequestRequest{BaseBranch: "dev", Title: "Deliver slice", Body: "Deliver the aggregate candidate."}
	if _, err := operations.EnsureWorkflowSlicePullRequest(t.Context(), run.Invocation.ID, content); err != nil {
		t.Fatal(err)
	}

	remote := filepath.Join(t.TempDir(), "canonical.git")
	serviceWorkflowDeliveryGit(t, candidate.Lane.Worktree, "init", "--bare", remote)
	serviceWorkflowDeliveryGit(t, candidate.Lane.Worktree, "push", remote, candidate.Result.Commit+":refs/heads/dev")

	operations.workflowIntegrationGit = core.NewWorkflowIntegrationGit(repository, remote)
	provider.pr.State, provider.pr.Merged, provider.pr.ResultCommit = "closed", true, candidate.Result.Commit
	provider.pr.BaseCommit = candidate.Result.Commit

	return operations, run.Invocation.ID, provider
}

func serviceWorkflowDeliveryGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git workflow integration runs in the nightly suite")
	}

	command := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("delivery Git fixture: %v: %s", err, output)
	}
}

// TestServiceWorkflowIntegration proves a real bare-remote fast-forward rebase
// through the shared service without completing tasks or submitting another PR.
func TestServiceWorkflowIntegration(t *testing.T) {
	operations, id, provider := serviceWorkflowDeliveryFixture(t)

	proof, err := operations.VerifyWorkflowSliceIntegration(t.Context(), id, "rebase")
	if err != nil || proof.Git.ResultCommit != provider.pr.ResultCommit || len(proof.Git.History) != 3 {
		t.Fatalf("service external delivery proof: %+v: %v", proof, err)
	}

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || run.Status != "waiting_integration" || run.Outcome != "slice_delivery_verified" || provider.creates != 1 {
		t.Fatal("service skipped finalization or created another PR")
	}
}
