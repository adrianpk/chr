package core

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

func (s *WorkflowStore) prepareWorkflowPullRequest(ctx context.Context, run *workflowInvocationFile, request WorkflowSlicePullRequestRequest) (workflowSliceObservation, error) {
	if err := s.checkWorkflowPullRequestReady(*run); err != nil {
		return workflowSliceObservation{}, err
	}

	observation, err := s.observeWorkflowSlice(ctx, *run)
	if err != nil {
		return observation, err
	}

	candidate, err := s.readWorkflowSliceCandidate(ctx, *run, observation)
	if err != nil {
		return observation, err
	}

	if err := prepareWorkflowPullRequestBinding(run, observation, candidate, request); err != nil {
		return observation, err
	}

	return observation, s.commitWorkflowState(run)
}

func (s *WorkflowStore) checkWorkflowPullRequestReady(run workflowInvocationFile) error {
	if run.StopIntent != "" || !slices.Contains([]string{"waiting_integration", "waiting_review", "blocked"}, run.Status) {
		return ErrWorkflowOwned
	}

	if run.PullRequest != nil && run.PullRequest.State == "blocked" {
		return fmt.Errorf("%w: %s requires explicit reconciliation", ErrWorkflowConflict, run.PullRequest.Reason)
	}

	if run.PullRequest != nil && run.PullRequest.ObservationRef != "" {
		if _, err := s.readWorkflowPullRequestObservation(*run.PullRequest); err != nil {
			return err
		}
	}

	return nil
}

func prepareWorkflowPullRequestBinding(run *workflowInvocationFile, observation workflowSliceObservation, candidate WorkflowSliceCandidate, request WorkflowSlicePullRequestRequest) error {
	repository, err := workflowCanonicalRepository(run.ExecutionSelection().Target.Repository)
	if err != nil {
		return err
	}

	prRequest := WorkflowPullRequestRequest{WorkflowPullRequestQuery: WorkflowPullRequestQuery{Repository: repository, HeadBranch: strings.TrimPrefix(run.Lane.Branch, "refs/heads/"), BaseBranch: request.BaseBranch}, Title: request.Title, Body: request.Body}
	if !validWorkflowPRBranch(prRequest.HeadBranch) || prRequest.HeadBranch == prRequest.BaseBranch {
		return ErrWorkflowConflict
	}

	if err := initializeWorkflowPullRequestBinding(run, observation.descriptor, prRequest); err != nil {
		return err
	}

	binding := run.PullRequest
	if observation.descriptor.PullRequest != "" && observation.descriptor.PullRequest != binding.URL {
		return ErrWorkflowConflict
	}

	if !workflowPullRequestCandidateMayAdvance(*binding, candidate.ID) {
		return fmt.Errorf("%w: uncertain creation still owns its candidate", ErrWorkflowConflict)
	}

	binding.CandidateID, binding.CandidateRef, binding.CandidateHash = candidate.ID, run.CandidateRef, run.CandidateHash
	binding.SourceCommit, binding.SourceTree = candidate.Result.Commit, candidate.Result.Tree

	return nil
}

func workflowPullRequestCandidateMayAdvance(binding WorkflowPullRequestBinding, id string) bool {
	return binding.CandidateID == "" || binding.CandidateID == id || binding.State == "prepared" || binding.State == "bound" && binding.URL != ""
}

func initializeWorkflowPullRequestBinding(run *workflowInvocationFile, descriptor WorkflowSlice, request WorkflowPullRequestRequest) error {
	if run.PullRequest != nil {
		if run.PullRequest.Request != request || run.PullRequest.Unit != run.Lane.Unit || run.PullRequest.TargetCommit != run.ExecutionSelection().Target.Commit {
			return fmt.Errorf("%w: retained PR request changed", ErrWorkflowConflict)
		}

		return nil
	}

	if descriptor.Branch != "" && descriptor.Branch != run.Lane.Branch {
		return ErrWorkflowConflict
	}

	run.PullRequest = &WorkflowPullRequestBinding{Schema: "chr-slice-pr/v1", ID: workflowPullRequestBindingID(*run, request.Repository), Unit: run.Lane.Unit, Request: request, TargetCommit: run.ExecutionSelection().Target.Commit, State: "prepared", URL: descriptor.PullRequest}
	if run.PullRequest.URL != "" {
		run.PullRequest.State = "bound"
	}

	return nil
}

func workflowPullRequestBindingID(run workflowInvocationFile, repository string) string {
	return workflowHash([]byte(run.ProjectID + "\x00" + run.Namespace + "\x00" + run.Lane.Unit.ID + "\x00" + repository))
}
