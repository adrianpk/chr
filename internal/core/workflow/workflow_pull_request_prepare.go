package workflow

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) prepareWorkflowPullRequest(ctx context.Context, run *workflowInvocationFile, request SlicePullRequestRequest) (workflowSliceObservation, error) {
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

func (s *Store) checkWorkflowPullRequestReady(run workflowInvocationFile) error {
	if run.StopIntent != "" || !slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationWaitingIntegration, workflowstate.InvocationWaitingReview, workflowstate.InvocationBlocked}, run.Status) {
		return ErrWorkflowOwned
	}

	if run.PullRequest != nil && run.PullRequest.State == "blocked" {
		return fmt.Errorf("%w: %s requires explicit reconciliation", workflowstate.ErrConflict, run.PullRequest.Reason)
	}

	if run.PullRequest != nil && run.PullRequest.ObservationRef != "" {
		if _, err := s.readWorkflowPullRequestObservation(*run.PullRequest); err != nil {
			return err
		}
	}

	return nil
}

func prepareWorkflowPullRequestBinding(run *workflowInvocationFile, observation workflowSliceObservation, candidate SliceCandidate, request SlicePullRequestRequest) error {
	repository, err := workflowCanonicalRepository(run.ExecutionSelection().Target.Repository)
	if err != nil {
		return err
	}

	prRequest := PullRequestRequest{PullRequestQuery: PullRequestQuery{Repository: repository, HeadBranch: strings.TrimPrefix(run.Lane.Branch, "refs/heads/"), BaseBranch: request.BaseBranch}, Title: request.Title, Body: request.Body}
	if !validWorkflowPRBranch(prRequest.HeadBranch) || prRequest.HeadBranch == prRequest.BaseBranch {
		return workflowstate.ErrConflict
	}

	if err := initializeWorkflowPullRequestBinding(run, observation.descriptor, prRequest); err != nil {
		return err
	}

	binding := run.PullRequest
	if observation.descriptor.PullRequest != "" && observation.descriptor.PullRequest != binding.URL {
		return workflowstate.ErrConflict
	}

	if !workflowPullRequestCandidateMayAdvance(*binding, candidate.ID) {
		return fmt.Errorf("%w: uncertain creation still owns its candidate", workflowstate.ErrConflict)
	}

	binding.CandidateID, binding.CandidateRef, binding.CandidateHash = candidate.ID, run.CandidateRef, run.CandidateHash
	binding.SourceCommit, binding.SourceTree = candidate.Result.Commit, candidate.Result.Tree

	return nil
}

func workflowPullRequestCandidateMayAdvance(binding PullRequestBinding, id string) bool {
	return binding.CandidateID == "" || binding.CandidateID == id || binding.State == "prepared" || binding.State == "bound" && binding.URL != ""
}

func initializeWorkflowPullRequestBinding(run *workflowInvocationFile, descriptor workflowstate.WorkflowSlice, request PullRequestRequest) error {
	if run.PullRequest != nil {
		if run.PullRequest.Request != request || run.PullRequest.Unit != run.Lane.Unit || run.PullRequest.TargetCommit != run.ExecutionSelection().Target.Commit {
			return fmt.Errorf("%w: retained PR request changed", workflowstate.ErrConflict)
		}

		return nil
	}

	if descriptor.Branch != "" && descriptor.Branch != run.Lane.Branch {
		return workflowstate.ErrConflict
	}

	run.PullRequest = &PullRequestBinding{Schema: "chr-slice-pr/v1", ID: workflowPullRequestBindingID(*run, request.Repository), Unit: run.Lane.Unit, Request: request, TargetCommit: run.ExecutionSelection().Target.Commit, State: "prepared", URL: descriptor.PullRequest}
	if run.PullRequest.URL != "" {
		run.PullRequest.State = "bound"
	}

	return nil
}

func workflowPullRequestBindingID(run workflowInvocationFile, repository string) string {
	return contenthash.Sum([]byte(run.ProjectID + "\x00" + run.Namespace + "\x00" + run.Lane.Unit.ID + "\x00" + repository))
}
