package app

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type serviceWorkflowPreparationFixture struct{ *serviceWorkflowReviewFixture }

func (p *serviceWorkflowPreparationFixture) WorkflowIntegrationCapability(_ context.Context, pr workflowstate.PullRequestSnapshot, method string) (workflowstate.PublicationCapability, error) {
	return workflowstate.PublicationCapability{Repository: pr.Repository, PullRequestURL: pr.URL, Method: method, Mechanism: "fixture_atomic_source_base", Revision: strings.Repeat("a", 64), AtomicSourceBase: true, ExactResult: true}, nil
}

// TestServiceWorkflowPreparation uses the central service to prepare a real
// linear combined result and authorize its exact manual approval. The service
// cannot substitute the source-only grant or complete delivery during preflight.
func TestServiceWorkflowPreparation(t *testing.T) {
	operations, id, provider := serviceWorkflowPreparationReady(t)
	if _, err := operations.CheckWorkflowPublicationGuard(t.Context(), id); err == nil {
		t.Fatal("unprepared content authorized publication")
	}

	candidate, err := operations.PrepareWorkflowSliceIntegration(t.Context(), id)
	if err != nil || len(candidate.Input.Git.History) != 3 {
		t.Fatalf("combined result: %+v, %v", candidate, err)
	}

	binding, err := operations.ReviewWorkflowSlice(t.Context(), id, nil)
	if err != nil || binding.State != workflowstate.ReviewWaitingApproval {
		t.Fatalf("manual wait: %+v, %v", binding, err)
	}

	if _, err := operations.ApproveWorkflowSlice(t.Context(), id, binding.InputHash); err != nil {
		t.Fatal(err)
	}

	guard, err := operations.CheckWorkflowPublicationGuard(t.Context(), id)
	if err != nil || guard.Git.ResultCommit != candidate.Input.Git.ResultCommit || provider.creates != 1 {
		t.Fatalf("publication guard: %+v, %v", guard, err)
	}
}

func serviceWorkflowPreparationReady(t *testing.T) (service, string, *serviceWorkflowPreparationFixture) {
	t.Helper()

	operations, id, original := serviceWorkflowReviewReady(t)
	provider := &serviceWorkflowPreparationFixture{original}
	operations.workflowPullRequests = provider

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	remote := filepath.Join(t.TempDir(), "canonical.git")
	serviceWorkflowDeliveryGit(t, run.Lane.Worktree, "init", "--bare", remote)
	serviceWorkflowDeliveryGit(t, run.Lane.Worktree, "push", remote, run.Lane.BaseCommit+":refs/heads/dev", run.Lane.HeadCommit+":"+run.Lane.Branch)

	operations.workflowPreparationGit = workflowcore.NewWorkflowPreparationGit(provider.pr.Repository, remote)

	return operations, id, provider
}
