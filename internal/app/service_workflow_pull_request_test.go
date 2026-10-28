package app

import (
	"context"
	"strings"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type serviceWorkflowPRFixture struct {
	pr      workflowstate.PullRequestSnapshot
	creates int
}

func (p *serviceWorkflowPRFixture) FindWorkflowPullRequests(context.Context, workflowcore.PullRequestQuery) ([]workflowstate.PullRequestSnapshot, error) {
	return nil, nil
}

func (p *serviceWorkflowPRFixture) ReadWorkflowBranch(context.Context, string, string) (string, error) {
	return p.pr.HeadCommit, nil
}

func (p *serviceWorkflowPRFixture) ReadWorkflowPullRequest(context.Context, string) (workflowstate.PullRequestSnapshot, error) {
	return p.pr, nil
}

func (p *serviceWorkflowPRFixture) CreateWorkflowPullRequest(context.Context, workflowcore.PullRequestRequest) (workflowstate.PullRequestSnapshot, error) {
	p.creates++

	return p.pr, nil
}

// TestServiceWorkflowSlicePR uses the shared application service to bind one PR
// and retain review status without task completion or duplicate provider creation.
func TestServiceWorkflowSlicePR(t *testing.T) {
	operations, hierarchy, target := serviceWorkflowSliceFixture(t)
	request := workflowcore.StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "slice-pr",
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

	provider := &serviceWorkflowPRFixture{pr: workflowstate.PullRequestSnapshot{URL: request.Selection.Target.Repository + "/pulls/7", Repository: request.Selection.Target.Repository, HeadRepository: request.Selection.Target.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: target, State: "open"}}
	operations.workflowPullRequests = provider

	content := workflowcore.SlicePullRequestRequest{BaseBranch: "dev", Title: "Deliver slice", Body: "Deliver the aggregate candidate."}
	for range 2 {
		pr, err := operations.EnsureWorkflowSlicePullRequest(t.Context(), run.Invocation.ID, content)
		if err != nil || pr != provider.pr {
			t.Fatalf("service PR: %+v, %v", pr, err)
		}
	}

	retained, err := operations.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || retained.Status != workflowstate.InvocationWaitingReview || provider.creates != 1 || retained.PullRequest.URL != provider.pr.URL {
		t.Fatal("service did not share durable canonical PR state")
	}
}
