package app

import (
	"context"
	"strings"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type serviceWorkflowReviewFixture struct{ *serviceWorkflowPRFixture }

func (p *serviceWorkflowReviewFixture) ReadWorkflowBranch(_ context.Context, _, branch string) (string, error) {
	if branch == p.pr.BaseBranch {
		return p.pr.BaseCommit, nil
	}

	return p.pr.HeadCommit, nil
}

// TestServiceWorkflowReview exercises current policy, explicit approval and
// revalidation through the same central service used by future adapter controls.
func TestServiceWorkflowReview(t *testing.T) {
	operations, id, provider := serviceWorkflowReviewReady(t)

	binding, err := operations.ReviewWorkflowSlice(t.Context(), id, nil)
	if err != nil || binding.State != workflowstate.ReviewWaitingApproval {
		t.Fatalf("service review: %+v, %v", binding, err)
	}

	if _, err := operations.ApproveWorkflowSlice(t.Context(), id, binding.InputHash); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.CheckWorkflowReviewGrant(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "review:\n  mode: agentic_review\n")

	if _, err := operations.CheckWorkflowReviewGrant(t.Context(), id); err == nil {
		t.Fatal("shared service reused approval after policy change")
	}

	if provider.creates != 1 {
		t.Fatal("service review created another PR")
	}
}

func serviceWorkflowReviewReady(t *testing.T) (service, string, *serviceWorkflowReviewFixture) {
	t.Helper()

	operations, hierarchy, target := serviceWorkflowSliceFixture(t)
	request := workflowcore.StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "slice-review",
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice", Identity: hierarchy.Set.Set.Slices[0].ID}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}},
	}}

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	worker := []string{"sh", "-c", "printf 'Task\\n' >> result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task'"}

	candidate, err := operations.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, worker)
	if err != nil {
		t.Fatal(err)
	}

	repo := request.Selection.Target.Repository
	provider := &serviceWorkflowReviewFixture{serviceWorkflowPRFixture: &serviceWorkflowPRFixture{pr: workflowstate.PullRequestSnapshot{URL: repo + "/pulls/7", Repository: repo, HeadRepository: repo, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: target, State: "open"}}}
	operations.workflowPullRequests = provider

	content := workflowstate.PullRequestContent{BaseBranch: "dev", Title: "Deliver slice", Body: "Deliver the aggregate candidate."}
	if _, err := operations.EnsureWorkflowSlicePullRequest(t.Context(), run.Invocation.ID, content); err != nil {
		t.Fatal(err)
	}

	return operations, run.Invocation.ID, provider
}
