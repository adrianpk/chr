package workflow

import (
	"context"
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) observeWorkflowFrontier(ctx context.Context, run workflowInvocationFile, receipt Record, candidate SliceCandidate, ports FrontierPorts) (workflowstate.PreviewTarget, workflowstate.IntegrationProof, error) {
	if ports.Provider == nil || ports.Observer == nil || ports.Target == nil {
		return workflowstate.PreviewTarget{}, workflowstate.IntegrationProof{}, ErrWorkflowObservationRequired
	}

	binding := run.PullRequest
	query := TargetQuery{Repository: binding.Request.Repository, Branch: binding.Request.BaseBranch, Directory: s.layout.ProjectDir()}

	target, err := ports.Target.ObserveWorkflowTarget(ctx, query)
	if err != nil {
		return target, workflowstate.IntegrationProof{}, err
	}

	if target.Repository != run.Selection.Target.Repository || !workflowCommitID(target.Commit) {
		return target, workflowstate.IntegrationProof{}, workflowstate.ErrConflict
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

func (s *Store) workflowFrontierReceiptProof(ctx context.Context, run workflowInvocationFile, receipt Record, candidate SliceCandidate, target workflowstate.PreviewTarget, ports FrontierPorts) (workflowstate.IntegrationProof, error) {
	if receipt.Receipt.Native == nil {
		return observeWorkflowFrontierProof(ctx, run, receipt, candidate, target, ports)
	}

	observation, err := s.observeWorkflowNativeReceipt(ctx, receipt, ports.Provider, ports.Observer)

	retained, retainErr := s.retainWorkflowNativeReceiptObservation(ctx, run, observation)
	if failure := errors.Join(err, retainErr); failure != nil {
		return workflowstate.IntegrationProof{}, failure
	}

	if retained.Blocker != nil || observation.State != "merged" || observation.Git == nil || !workflowFrontierReceiptProofMatches(*observation.Git, receipt, target) {
		return workflowstate.IntegrationProof{}, ErrWorkflowObservationRequired
	}

	return *observation.Git, nil
}

func observeWorkflowFrontierProof(ctx context.Context, run workflowInvocationFile, receipt Record, candidate SliceCandidate, target workflowstate.PreviewTarget, ports FrontierPorts) (workflowstate.IntegrationProof, error) {
	pr, err := ports.Provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return workflowstate.IntegrationProof{}, err
	}

	if workflowPullRequestMismatch(*run.PullRequest, []workflowstate.PullRequestSnapshot{pr}) != "" || !pr.Merged || pr.ResultCommit != receipt.Receipt.ResultCommit || pr.BaseCommit != target.Commit {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
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

func workflowFrontierReceiptProofMatches(proof workflowstate.IntegrationProof, receipt Record, target workflowstate.PreviewTarget) bool {
	return proof.ObservedTarget == target.Commit && proof.HistoryHash == receipt.Receipt.HistoryProof && proof.VerifiedTree == receipt.Receipt.VerifiedTree
}
