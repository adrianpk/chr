package core

import (
	"context"
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) observeWorkflowFrontier(ctx context.Context, run workflowInvocationFile, receipt WorkflowRecord, candidate WorkflowSliceCandidate, ports WorkflowFrontierPorts) (WorkflowPreviewTarget, WorkflowIntegrationProof, error) {
	if ports.Provider == nil || ports.Observer == nil || ports.Target == nil {
		return WorkflowPreviewTarget{}, WorkflowIntegrationProof{}, ErrWorkflowObservationRequired
	}

	binding := run.PullRequest
	query := WorkflowTargetQuery{Repository: binding.Request.Repository, Branch: binding.Request.BaseBranch, Directory: s.layout.ProjectDir()}

	target, err := ports.Target.ObserveWorkflowTarget(ctx, query)
	if err != nil {
		return target, WorkflowIntegrationProof{}, err
	}

	if target.Repository != run.Selection.Target.Repository || !workflowCommitID(target.Commit) {
		return target, WorkflowIntegrationProof{}, workflowstate.ErrConflict
	}

	proof, err := s.workflowFrontierReceiptProof(ctx, run, receipt, candidate, target, ports)
	if err != nil {
		return target, proof, err
	}

	observed, err := ports.Provider.ReadWorkflowBranch(ctx, query.Repository, query.Branch)
	if err != nil || observed != target.Commit {
		return target, proof, errorsOrConflict(err)
	}

	return target, proof, s.checkWorkflowTarget(ctx, target)
}

func (s *WorkflowStore) workflowFrontierReceiptProof(ctx context.Context, run workflowInvocationFile, receipt WorkflowRecord, candidate WorkflowSliceCandidate, target WorkflowPreviewTarget, ports WorkflowFrontierPorts) (WorkflowIntegrationProof, error) {
	if receipt.Receipt.Native == nil {
		return observeWorkflowFrontierProof(ctx, run, receipt, candidate, target, ports)
	}

	observation, err := s.observeWorkflowNativeReceipt(ctx, receipt, ports.Provider, ports.Observer)

	retained, retainErr := s.retainWorkflowNativeReceiptObservation(ctx, run, observation)
	if failure := errors.Join(err, retainErr); failure != nil {
		return WorkflowIntegrationProof{}, failure
	}

	if retained.Blocker != nil || observation.State != "merged" || observation.Git == nil || !workflowFrontierReceiptProofMatches(*observation.Git, receipt, target) {
		return WorkflowIntegrationProof{}, ErrWorkflowObservationRequired
	}

	return *observation.Git, nil
}

func observeWorkflowFrontierProof(ctx context.Context, run workflowInvocationFile, receipt WorkflowRecord, candidate WorkflowSliceCandidate, target WorkflowPreviewTarget, ports WorkflowFrontierPorts) (WorkflowIntegrationProof, error) {
	pr, err := ports.Provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return WorkflowIntegrationProof{}, err
	}

	if workflowPullRequestMismatch(*run.PullRequest, []WorkflowPullRequestSnapshot{pr}) != "" || !pr.Merged || pr.ResultCommit != receipt.Receipt.ResultCommit || pr.BaseCommit != target.Commit {
		return WorkflowIntegrationProof{}, workflowstate.ErrConflict
	}

	integration := workflowIntegrationQuery(run, candidate, pr, receipt.Receipt.Method)

	proof, err := ports.Observer.VerifyWorkflowIntegration(ctx, integration)
	if err != nil {
		return proof, err
	}

	if err := validateWorkflowIntegrationProof(integration, proof); err != nil {
		return proof, err
	}

	if !workflowFrontierReceiptProofMatches(proof, receipt, target) {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}

func workflowFrontierReceiptProofMatches(proof WorkflowIntegrationProof, receipt WorkflowRecord, target WorkflowPreviewTarget) bool {
	return proof.ObservedTarget == target.Commit && proof.HistoryHash == receipt.Receipt.HistoryProof && proof.VerifiedTree == receipt.Receipt.VerifiedTree
}
