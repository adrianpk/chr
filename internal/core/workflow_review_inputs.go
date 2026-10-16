package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
)

func (s *WorkflowStore) workflowReviewInputs(ctx context.Context, run workflowInvocationFile, provider WorkflowPullRequestProvider) (workflowSliceObservation, WorkflowReviewInput, error) {
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

func (s *WorkflowStore) workflowReviewSourceInputs(ctx context.Context, run workflowInvocationFile, provider WorkflowPullRequestProvider) (workflowSliceObservation, WorkflowReviewInput, error) {
	if err := checkWorkflowReviewReady(run, provider); err != nil {
		return workflowSliceObservation{}, WorkflowReviewInput{}, err
	}

	if _, err := s.readWorkflowPullRequestObservation(*run.PullRequest); err != nil {
		return workflowSliceObservation{}, WorkflowReviewInput{}, err
	}

	observation, err := s.observeWorkflowSlice(ctx, run)
	if err != nil {
		return observation, WorkflowReviewInput{}, err
	}

	candidate, err := s.readWorkflowSliceCandidate(ctx, run, observation)
	if err != nil {
		return observation, WorkflowReviewInput{}, err
	}

	policy, err := s.workflowReviewPolicy(run)
	if err != nil {
		return observation, WorkflowReviewInput{}, err
	}

	pr, target, err := observeWorkflowReviewPR(ctx, run, provider)
	if err != nil {
		return observation, WorkflowReviewInput{}, err
	}

	if err := s.checkWorkflowReviewAfterIO(ctx, run, observation, policy); err != nil {
		return observation, WorkflowReviewInput{}, err
	}

	input := WorkflowReviewInput{Schema: "chr-workflow-review-input/v1", InvocationID: run.ID, Unit: candidate.UnitInput().Unit, CandidateID: candidate.ID, CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, Slice: candidate.UnitInput(), PullRequest: pr, TargetCommit: target, Policy: policy}

	return observation, input, nil
}

func observeWorkflowReviewPR(ctx context.Context, run workflowInvocationFile, provider WorkflowPullRequestProvider) (WorkflowPullRequestSnapshot, string, error) {
	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return pr, "", err
	}

	if reason := workflowPullRequestMismatch(*run.PullRequest, []WorkflowPullRequestSnapshot{pr}); reason != "" || pr.Merged {
		return pr, "", fmt.Errorf("%w: canonical PR is changed or requires delivery observation", ErrWorkflowConflict)
	}

	target, err := provider.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)
	if err != nil {
		return pr, target, err
	}

	if !workflowCommitID(target) || target != pr.BaseCommit {
		return pr, target, fmt.Errorf("%w: canonical target differs from review observation", ErrWorkflowConflict)
	}

	return pr, target, nil
}

func (s *WorkflowStore) checkWorkflowReviewAfterIO(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, policy WorkflowReviewPolicy) error {
	if _, err := s.readWorkflowSliceCandidate(ctx, run, observation); err != nil {
		return err
	}

	current, err := s.workflowReviewPolicy(run)
	if err != nil || current != policy {
		return ErrWorkflowConflict
	}

	return nil
}

func workflowReviewInputBytes(input WorkflowReviewInput) ([]byte, error) {
	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return nil, fmt.Errorf("review input exceeds file bounds")
	}

	return append(data, '\n'), nil
}

func workflowReviewInputHash(input WorkflowReviewInput) (string, error) {
	data, err := workflowReviewInputBytes(input)

	return workflowHash(data), err
}

func (s *WorkflowStore) checkWorkflowReviewGrant(ctx context.Context, run workflowInvocationFile, provider WorkflowPullRequestProvider) (WorkflowReviewInput, error) {
	if run.Review == nil || run.Review.State != "authorized" {
		return WorkflowReviewInput{}, fmt.Errorf("exact candidate review is not authorized")
	}

	if _, _, err := s.readWorkflowReview(run); err != nil {
		return WorkflowReviewInput{}, err
	}

	_, input, err := s.workflowReviewInputs(ctx, run, provider)
	if err != nil {
		return input, err
	}

	hash, err := workflowReviewInputHash(input)
	if err != nil || hash != run.Review.InputHash {
		return input, fmt.Errorf("%w: review grant inputs changed", ErrWorkflowConflict)
	}

	return input, nil
}

func (s *WorkflowStore) approveWorkflowSlice(ctx context.Context, run workflowInvocationFile, bindingHash string, provider WorkflowPullRequestProvider, control *workflowControlCommit) (WorkflowReviewBinding, error) {
	if run.Review == nil || bindingHash != run.Review.InputHash || !slices.Contains([]string{"waiting_approval", "human_required", "authorized"}, run.Review.State) {
		return WorkflowReviewBinding{}, fmt.Errorf("%w: approval does not name the current human decision", ErrWorkflowConflict)
	}

	if _, _, err := s.readWorkflowReview(run); err != nil {
		return *run.Review, err
	}

	observation, input, err := s.workflowReviewInputs(ctx, run, provider)
	if err != nil {
		return *run.Review, err
	}

	hash, err := workflowReviewInputHash(input)
	if err != nil || hash != bindingHash {
		return *run.Review, fmt.Errorf("%w: approval inputs changed", ErrWorkflowConflict)
	}

	if run.Review.State == "authorized" {
		return *run.Review, s.retainWorkflowControlNoop(ctx, run, control)
	}

	files, err := control.files(ctx, run.Revision+1)
	if err != nil {
		return *run.Review, err
	}

	decision := WorkflowReviewDecision{BindingHash: hash, Result: AgenticReviewResult{Decision: IntegrationReviewDecisionPass, Summary: "Explicit human approval of this exact candidate and current policy."}}
	err = s.persistWorkflowReviewDecision(&run, observation, decision, "authorized", files...)

	return *run.Review, err
}

func checkWorkflowReviewReady(run workflowInvocationFile, provider WorkflowPullRequestProvider) error {
	if provider == nil || run.StopIntent != "" || workflowInvocationTerminal(run.Status) || run.Status == "paused" {
		return ErrWorkflowOwned
	}

	if run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.URL == "" {
		return fmt.Errorf("review requires a bound canonical integration-unit PR")
	}

	return nil
}
