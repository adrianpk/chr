package core

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

func (s *WorkflowStore) finishWorkflowClosing(ctx context.Context, claim *workflowClosingClaim, receipt WorkflowClosingReceipt, failure error, ports WorkflowFrontierPorts) (WorkflowSetClosureResult, error) {
	if err := s.checkWorkflowClosingCredential(claim); err != nil {
		return WorkflowSetClosureResult{}, errors.Join(failure, err)
	}

	if err := workflowAttemptQuiescent(WorkflowAttempt{Processes: claim.state.Processes}); err != nil {
		return WorkflowSetClosureResult{}, errors.Join(failure, err)
	}

	failure = errors.Join(failure, s.checkWorkflowClosingConfirmation(ctx, claim, ports))
	if failure != nil {
		receipt.Outcome = "closing_validation_failed"
	}

	receipt.ConfirmedAt = time.Now().UTC()

	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return WorkflowSetClosureResult{}, err
	}

	data = append(data, '\n')
	claim.state.Outcome, claim.state.FinishedAt = receipt.Outcome, receipt.ConfirmedAt
	claim.state.Result = WorkflowClosingEvidence{Ref: workflowClosingFile(claim.set.ID, claim.state.Generation, "result.json"), Hash: workflowHash(data)}

	claim.set.Set.Status = "blocked"
	if receipt.Outcome == "completed" {
		claim.set.Set.Status, claim.set.Set.ClosureRef = "completed", claim.state.Result.Ref
	}

	if err := s.commitWorkflowClosingState(ctx, claim, workflowFileChange{Path: claim.state.Result.Ref, After: data}); err != nil {
		return WorkflowSetClosureResult{}, errors.Join(failure, err)
	}

	result := WorkflowSetClosureResult{Set: claim.set, Receipt: receipt, ReceiptRef: claim.state.Result.Ref, Outcome: receipt.Outcome}

	return result, failure
}

func (s *WorkflowStore) checkWorkflowClosingConfirmation(ctx context.Context, claim *workflowClosingClaim, ports WorkflowFrontierPorts) error {
	if err := s.checkWorkflowClosingAuthority(claim); err != nil {
		return err
	}

	if err := s.checkWorkflowClosingInputs(ctx, claim); err != nil {
		return err
	}

	request := WorkflowSetClosureRequest{Repository: claim.input.Target.Repository, Branch: claim.input.Branch}

	target, err := s.observeWorkflowClosingTarget(ctx, request, ports)
	if err != nil || target != claim.input.Target {
		return errorsOrConflict(err)
	}

	// Observation can cross the deadline or a host edit. Recheck authority and
	// the frozen input after that external operation before confirming closure.
	if err := s.checkWorkflowClosingAuthority(claim); err != nil {
		return err
	}

	return s.checkWorkflowClosingInputs(ctx, claim)
}

func (s *WorkflowStore) checkWorkflowClosingAuthority(claim *workflowClosingClaim) error {
	if time.Now().After(claim.state.LeaseUntil) {
		return context.DeadlineExceeded
	}

	if claim.state.InvocationID == "" {
		return nil
	}

	run, err := s.readWorkflowInvocation(claim.state.InvocationID)
	if err != nil {
		return err
	}

	if run.StopIntent != "" || workflowInvocationTerminal(run.Status) {
		return ErrWorkflowStop
	}

	return nil
}
