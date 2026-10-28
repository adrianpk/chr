package workflow

import (
	"context"
	"errors"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowFrontierUnit struct {
	Run       workflowInvocationFile
	Receipt   workflowstate.Record
	Candidate SliceCandidate
}

type workflowFrontierObservation struct {
	Target workflowstate.PreviewTarget
	Proof  workflowstate.IntegrationProof
}

// workflowFrontierObserver sequences canonical IO inside the Store-owned frontier transaction.
type workflowFrontierObserver struct{ store *Store }

func (observer workflowFrontierObserver) observe(ctx context.Context, unit workflowFrontierUnit, ports FrontierPorts) (workflowFrontierObservation, error) {
	if ports.Provider == nil || ports.Observer == nil || ports.Target == nil {
		return workflowFrontierObservation{}, workflowstate.ErrObservationRequired
	}

	run := unit.Run
	binding := run.PullRequest
	query := TargetQuery{Repository: binding.Request.Repository, Branch: binding.Request.BaseBranch, Directory: observer.store.layout.ProjectDir()}

	target, err := ports.Target.ObserveWorkflowTarget(ctx, query)
	if err != nil {
		return workflowFrontierObservation{Target: target}, err
	}

	if target.Repository != run.Selection.Target.Repository || !workflowstate.ValidCommitID(target.Commit) {
		return workflowFrontierObservation{Target: target}, workflowstate.ErrConflict
	}

	proof, err := observer.receiptProof(ctx, unit, target, ports)
	if err != nil {
		return workflowFrontierObservation{Target: target, Proof: proof}, err
	}

	observed, err := ports.Provider.ReadWorkflowBranch(ctx, query.Repository, query.Branch)
	if err != nil || observed != target.Commit {
		return workflowFrontierObservation{Target: target, Proof: proof}, errorsOrConflict(err)
	}

	return workflowFrontierObservation{Target: target, Proof: proof}, observer.store.checkWorkflowTarget(ctx, target)
}

func (observer workflowFrontierObserver) receiptProof(ctx context.Context, unit workflowFrontierUnit, target workflowstate.PreviewTarget, ports FrontierPorts) (workflowstate.IntegrationProof, error) {
	run, receipt := unit.Run, unit.Receipt

	if receipt.Receipt.Native == nil {
		return observeWorkflowFrontierProof(ctx, unit, target, ports)
	}

	observation, err := workflowNativeObservationHost(observer).receipt(ctx, receipt, ports.Provider, ports.Observer)

	retained, retainErr := observer.store.retainWorkflowNativeReceiptObservation(ctx, run, observation)
	if failure := errors.Join(err, retainErr); failure != nil {
		return workflowstate.IntegrationProof{}, failure
	}

	if retained.Blocker != nil || observation.State != workflowstate.NativeMerged || observation.Git == nil || !workflowstate.ReceiptTargetProofMatches(*observation.Git, receipt, target) {
		return workflowstate.IntegrationProof{}, workflowstate.ErrObservationRequired
	}

	return *observation.Git, nil
}

func observeWorkflowFrontierProof(ctx context.Context, unit workflowFrontierUnit, target workflowstate.PreviewTarget, ports FrontierPorts) (workflowstate.IntegrationProof, error) {
	run, receipt, candidate := unit.Run, unit.Receipt, unit.Candidate

	pr, err := ports.Provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return workflowstate.IntegrationProof{}, err
	}

	if workflowforge.PullRequestMismatch(*run.PullRequest, []workflowstate.PullRequestSnapshot{pr}) != "" || !pr.Merged || pr.ResultCommit != receipt.Receipt.ResultCommit || pr.BaseCommit != target.Commit {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
	}

	integration := workflowIntegrationQuery(run, candidate, pr, receipt.Receipt.Method)

	proof, err := ports.Observer.VerifyWorkflowIntegration(ctx, integration)
	if err != nil {
		return proof, err
	}

	if err := workflowintegrate.CheckProof(integration, proof, workflowIntegrationRepairs{}); err != nil {
		return proof, err
	}

	if !workflowstate.ReceiptTargetProofMatches(proof, receipt, target) {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}
