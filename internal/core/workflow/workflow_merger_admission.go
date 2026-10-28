package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowMergerAdmission(ctx context.Context, run workflowInvocationFile, request MergerRequest, harness MergerHarness, provider PullRequestProvider) (workflowMergerAdmissionResult, error) {
	budgetObservation, err := s.currentWorkflowIntegrationBudget(ctx, run, request.IntegrationBudgetGuard)

	entry, budget := budgetObservation.Entry, budgetObservation.Budget

	admission := workflowMergerAdmissionResult{Run: run, Entry: entry, Budget: budget}

	if err != nil {
		return admission, err
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.IntegrationBudgetGuard); err != nil {
		return admission, err
	}

	if err := s.checkWorkflowMergerAdmissionInputs(ctx, admission, request.Input, harness, provider); err != nil {
		return admission, err
	}

	seed, err := s.workflowMergerSeed(ctx, run, entry, request.Input)
	if err == nil && request.Input.RepairCheckpointRef != nil {
		_, err = s.readWorkflowMergerCheckpoint(request.Input)
	}

	admission.Seed = seed

	return admission, err
}

func (s *Store) checkWorkflowMergerAdmissionInputs(ctx context.Context, admission workflowMergerAdmissionResult, input workflowstate.MergerInput, harness MergerHarness, provider PullRequestProvider) error {
	run, entry, budget := admission.Run, admission.Entry, admission.Budget

	if !workflowMergerRunnable(run, entry) {
		return ErrWorkflowOwned
	}

	if err := s.CheckWorkflowMergerEvidence(ctx, input, time.Now().UTC()); err != nil {
		return err
	}

	if _, err := s.checkWorkflowMergerHarness(ctx, input, harness); err != nil {
		return err
	}

	if err := checkWorkflowMergerDecisionBinding(entry, input); err != nil {
		return err
	}

	if err := checkWorkflowMergerEntry(run, entry, input); err != nil {
		return err
	}

	if err := checkWorkflowMergerRemote(ctx, provider, input); err != nil {
		return err
	}

	if err := s.checkWorkflowCandidateRepairAdmission(ctx, run, entry, input, provider); err != nil {
		return err
	}

	if err := s.checkWorkflowMergerCommands(ctx, run, input); err != nil {
		return err
	}

	return checkWorkflowMergerBudget(budget, input)
}

func checkWorkflowMergerEntry(run workflowInvocationFile, entry IntegrationEntry, input workflowstate.MergerInput) error {
	if run.Lane == nil || input.WorkerTimeoutSeconds != run.Policy.WorkerTimeoutSeconds {
		return workflowstate.ErrConflict
	}

	hash := run.CandidateHash
	if entry.CandidateRef != nil {
		hash = entry.CandidateRef.Hash
	}

	current := workflowmerge.Fence{
		EntryID: entry.ID, CandidateHash: hash, ScopeHash: entry.FrozenScopeHash,
		OwnedWorkspaceID: workflowMergerWorkspaceID(input), Source: run.Lane.HeadCommit, Base: input.Base,
		Method: input.Method, Unit: entry.Unit, CandidateRevision: entry.StateRevision,
		ClaimGeneration: entry.ClaimGeneration, Repository: entry.Repository, SourceRef: run.Lane.Branch, TargetRef: entry.TargetRef,
	}
	if input.InvocationID != run.ID || current != workflowmerge.InputFence(input) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) workflowMergerSeed(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, input workflowstate.MergerInput) (string, error) {
	policy, err := s.workflowReviewPolicy(run)
	if err != nil || policy.Method != input.Method {
		return "", workflowstate.ErrConflict
	}

	if run.CandidateRepair != nil {
		return s.workflowCandidateRepairSeed(ctx, run)
	}

	if entry.CandidateRef == nil {
		return input.Source, nil
	}

	var candidate workflowstate.IntegrationCandidate
	if err := s.readWorkflowNativeEvidence(run.ID, *entry.CandidateRef, &candidate); err != nil {
		return "", err
	}

	if candidate.Input.Git.SourceCommit != input.Source || candidate.Input.Git.BaseCommit != input.Base || candidate.Input.Git.Method != input.Method {
		return "", workflowstate.ErrConflict
	}

	if err := s.checkWorkflowPreparationGit(ctx, candidate); err != nil {
		return "", err
	}

	return candidate.Input.Git.ResultCommit, nil
}

func freezeWorkflowMergerRequest(request MergerRequest) (MergerRequest, error) {
	data, err := json.Marshal(request)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return MergerRequest{}, fmt.Errorf("%w: merger admission exceeds bounds", workflowstate.ErrConflict)
	}

	var frozen MergerRequest

	err = json.Unmarshal(data, &frozen)

	return frozen, err
}

func workflowMergerRunnable(run workflowInvocationFile, entry IntegrationEntry) bool {
	return run.StopIntent == "" && entry.StopIntent == "" && entry.State == "merging" && slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationWaitingIntegration, workflowstate.InvocationWaitingReview}, run.Status)
}

func (s *Store) workflowCandidateRepairSeed(ctx context.Context, run workflowInvocationFile) (string, error) {
	repair, err := s.readWorkflowCandidateRepair(run)
	if err != nil {
		return "", err
	}

	return repair.Candidate.Input.Git.ResultCommit, s.checkWorkflowPreparationGit(ctx, repair.Candidate)
}
