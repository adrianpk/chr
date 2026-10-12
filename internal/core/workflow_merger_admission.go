package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"
)

func (s *WorkflowStore) workflowMergerAdmission(ctx context.Context, run workflowInvocationFile, request WorkflowMergerRequest, harness WorkflowMergerHarness, provider WorkflowPullRequestProvider) (WorkflowIntegrationEntry, WorkflowIntegrationBudget, string, error) {
	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, request.WorkflowIntegrationBudgetGuard)
	if err != nil {
		return entry, budget, "", err
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.WorkflowIntegrationBudgetGuard); err != nil {
		return entry, budget, "", err
	}

	if err := s.checkWorkflowMergerAdmissionInputs(ctx, run, entry, budget, request.Input, harness, provider); err != nil {
		return entry, budget, "", err
	}

	seed, err := s.workflowMergerSeed(ctx, run, entry, request.Input)

	return entry, budget, seed, err
}

func (s *WorkflowStore) checkWorkflowMergerAdmissionInputs(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, budget WorkflowIntegrationBudget, input WorkflowMergerInput, harness WorkflowMergerHarness, provider WorkflowPullRequestProvider) error {
	if !workflowMergerRunnable(run, entry) {
		return ErrWorkflowOwned
	}

	if err := s.CheckWorkflowMergerEvidence(ctx, input, time.Now().UTC()); err != nil {
		return err
	}

	if _, err := s.checkWorkflowMergerHarness(ctx, input, harness); err != nil {
		return err
	}

	if err := checkWorkflowMergerEntry(run, entry, input); err != nil {
		return err
	}

	if err := checkWorkflowMergerRemote(ctx, provider, input); err != nil {
		return err
	}

	if err := s.checkWorkflowMergerCommands(ctx, run, input); err != nil {
		return err
	}

	return checkWorkflowMergerBudget(budget, input)
}

func checkWorkflowMergerEntry(run workflowInvocationFile, entry WorkflowIntegrationEntry, input WorkflowMergerInput) error {
	if run.Lane == nil || input.WorkerTimeoutSeconds != run.Policy.WorkerTimeoutSeconds {
		return ErrWorkflowConflict
	}

	hash := run.CandidateHash
	if entry.CandidateRef != nil {
		hash = entry.CandidateRef.Hash
	}

	current := WorkflowMergerFence{
		EntryID: entry.ID, CandidateHash: hash, ScopeHash: entry.FrozenScopeHash,
		OwnedWorkspaceID: workflowMergerWorkspaceID(input), Source: run.Lane.HeadCommit, Base: input.Base,
		Method: input.Method, Unit: entry.Unit, CandidateRevision: entry.StateRevision,
		ClaimGeneration: entry.ClaimGeneration, Repository: entry.Repository, SourceRef: run.Lane.Branch, TargetRef: entry.TargetRef,
	}
	if input.InvocationID != run.ID || current != workflowMergerFence(input) {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) workflowMergerSeed(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, input WorkflowMergerInput) (string, error) {
	policy, err := s.workflowReviewPolicy(run)
	if err != nil || policy.Method != input.Method {
		return "", ErrWorkflowConflict
	}

	if entry.CandidateRef == nil {
		return input.Source, nil
	}

	var candidate WorkflowIntegrationCandidate
	if err := s.readWorkflowNativeEvidence(run.ID, *entry.CandidateRef, &candidate); err != nil {
		return "", err
	}

	if candidate.Input.Git.SourceCommit != input.Source || candidate.Input.Git.BaseCommit != input.Base || candidate.Input.Git.Method != input.Method {
		return "", ErrWorkflowConflict
	}

	if err := s.checkWorkflowPreparationGit(ctx, candidate); err != nil {
		return "", err
	}

	return candidate.Input.Git.ResultCommit, nil
}

func freezeWorkflowMergerRequest(request WorkflowMergerRequest) (WorkflowMergerRequest, error) {
	data, err := json.Marshal(request)
	if err != nil || len(data) >= WorkflowFileLimit {
		return WorkflowMergerRequest{}, fmt.Errorf("%w: merger admission exceeds bounds", ErrWorkflowConflict)
	}

	var frozen WorkflowMergerRequest

	err = json.Unmarshal(data, &frozen)

	return frozen, err
}

func workflowMergerRunnable(run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return run.StopIntent == "" && entry.StopIntent == "" && entry.State == "merging" && slices.Contains([]string{"waiting_integration", "waiting_review"}, run.Status)
}
