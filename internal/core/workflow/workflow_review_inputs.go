package workflow

import (
	"context"
	"fmt"
	"slices"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowReviewInputs(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider) (workflowSliceObservation, workflowstate.ReviewInput, error) {
	observation, input, err := s.workflowReviewSourceInputs(ctx, run, provider)
	if err != nil {
		return observation, input, err
	}

	if run.Preparation == nil {
		return observation, input, s.checkWorkflowSourceReviewProfile()
	}

	candidate, err := s.currentWorkflowPreparation(ctx, run, input)
	if err != nil {
		return observation, input, err
	}

	input.PreparationRef, input.PreparationHash = run.Preparation.ResultRef, run.Preparation.ResultHash
	input.PreparedGit = &candidate.Input.Git

	err = s.bindWorkflowNativeReview(ctx, run, candidate, provider, &input)
	if err == nil && input.Native != nil {
		err = s.checkWorkflowReviewAfterIO(ctx, run, observation, input.Policy)
	}

	return observation, input, err
}

func (s *Store) workflowReviewSourceInputs(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider) (workflowSliceObservation, workflowstate.ReviewInput, error) {
	if err := checkWorkflowReviewReady(run, provider); err != nil {
		return workflowSliceObservation{}, workflowstate.ReviewInput{}, err
	}

	if _, err := s.readWorkflowPullRequestObservation(*run.PullRequest); err != nil {
		return workflowSliceObservation{}, workflowstate.ReviewInput{}, err
	}

	observation, err := s.observeWorkflowSlice(ctx, run)
	if err != nil {
		return observation, workflowstate.ReviewInput{}, err
	}

	candidate, err := s.readWorkflowSliceCandidate(ctx, run, observation)
	if err != nil {
		return observation, workflowstate.ReviewInput{}, err
	}

	policy, err := s.workflowReviewPolicy(run)
	if err != nil {
		return observation, workflowstate.ReviewInput{}, err
	}

	pr, target, err := observeWorkflowReviewPR(ctx, run, provider)
	if err != nil {
		return observation, workflowstate.ReviewInput{}, err
	}

	if err := s.checkWorkflowReviewAfterIO(ctx, run, observation, policy); err != nil {
		return observation, workflowstate.ReviewInput{}, err
	}

	input := workflowstate.ReviewInput{Schema: "chr-workflow-review-input/v1", InvocationID: run.ID, Unit: candidate.UnitInput().Unit, CandidateID: candidate.ID, CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, Slice: candidate.UnitInput(), PullRequest: pr, TargetCommit: target, Policy: policy}

	return observation, input, nil
}

func observeWorkflowReviewPR(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider) (workflowstate.PullRequestSnapshot, string, error) {
	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return pr, "", err
	}

	if reason := workflowPullRequestMismatch(*run.PullRequest, []workflowstate.PullRequestSnapshot{pr}); reason != "" || pr.Merged {
		return pr, "", fmt.Errorf("%w: canonical PR is changed or requires delivery observation", workflowstate.ErrConflict)
	}

	target, err := provider.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)
	if err != nil {
		return pr, target, err
	}

	if !workflowstate.ValidCommitID(target) || target != pr.BaseCommit {
		return pr, target, fmt.Errorf("%w: canonical target differs from review observation", workflowstate.ErrConflict)
	}

	return pr, target, nil
}

func (s *Store) checkWorkflowReviewAfterIO(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, policy workflowstate.ReviewPolicy) error {
	if _, err := s.readWorkflowSliceCandidate(ctx, run, observation); err != nil {
		return err
	}

	current, err := s.workflowReviewPolicy(run)
	if err != nil || current != policy {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowReviewGrant(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider) (workflowstate.ReviewInput, error) {
	if run.Review == nil || run.Review.State != "authorized" {
		return workflowstate.ReviewInput{}, fmt.Errorf("exact candidate review is not authorized")
	}

	if _, _, err := s.readWorkflowReview(run); err != nil {
		return workflowstate.ReviewInput{}, err
	}

	_, input, err := s.workflowReviewInputs(ctx, run, provider)
	if err != nil {
		return input, err
	}

	hash, err := workflowreview.WorkflowReviewInputHash(input)
	if err != nil || hash != run.Review.InputHash {
		return input, fmt.Errorf("%w: review grant inputs changed", workflowstate.ErrConflict)
	}

	return input, nil
}

func (s *Store) approveWorkflowSlice(ctx context.Context, run workflowInvocationFile, bindingHash string, provider PullRequestProvider, control *workflowControlCommit) (workflowstate.ReviewBinding, error) {
	if run.Review == nil || bindingHash != run.Review.InputHash || !slices.Contains([]string{"waiting_approval", "human_required", "authorized"}, run.Review.State) {
		return workflowstate.ReviewBinding{}, fmt.Errorf("%w: approval does not name the current human decision", workflowstate.ErrConflict)
	}

	if _, _, err := s.readWorkflowReview(run); err != nil {
		return *run.Review, err
	}

	observation, input, err := s.workflowReviewInputs(ctx, run, provider)
	if err != nil {
		return *run.Review, err
	}

	hash, err := workflowreview.WorkflowReviewInputHash(input)
	if err != nil || hash != bindingHash {
		return *run.Review, fmt.Errorf("%w: approval inputs changed", workflowstate.ErrConflict)
	}

	if run.Review.State == "authorized" {
		return *run.Review, s.retainWorkflowControlNoop(ctx, run, control)
	}

	files, err := control.files(ctx, run.Revision+1)
	if err != nil {
		return *run.Review, err
	}

	decision := workflowstate.ReviewDecision{BindingHash: hash, Result: workflowstate.AgenticReviewResult{Decision: workflowstate.IntegrationReviewDecisionPass, Summary: "Explicit human approval of this exact candidate and current policy."}}
	err = s.persistWorkflowReviewDecision(&run, observation, decision, "authorized", files...)

	return *run.Review, err
}

func checkWorkflowReviewReady(run workflowInvocationFile, provider PullRequestProvider) error {
	if provider == nil || run.StopIntent != "" || workflowInvocationTerminal(run.Status) || run.Status == workflowstate.InvocationPaused {
		return ErrWorkflowOwned
	}

	if run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.URL == "" {
		return fmt.Errorf("review requires a bound canonical integration-unit PR")
	}

	return nil
}
