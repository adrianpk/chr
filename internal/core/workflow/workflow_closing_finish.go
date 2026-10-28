package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) finishWorkflowClosing(ctx context.Context, claim *workflowClosingClaim, receipt workflowstate.ClosingReceipt, failure error, ports FrontierPorts) (workflowcloseout.Result, error) {
	if err := s.checkWorkflowClosingCredential(claim); err != nil {
		return workflowcloseout.Result{}, errors.Join(failure, err)
	}

	if err := workflowAttemptQuiescent(workflowstate.Attempt{Processes: claim.state.Processes}); err != nil {
		return workflowcloseout.Result{}, errors.Join(failure, err)
	}

	failure = errors.Join(failure, s.checkWorkflowClosingConfirmation(ctx, claim, ports))
	if failure != nil {
		receipt.Outcome = "closing_validation_failed"
	}

	receipt.ConfirmedAt = time.Now().UTC()

	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return workflowcloseout.Result{}, err
	}

	data = append(data, '\n')
	claim.state.Outcome, claim.state.FinishedAt = receipt.Outcome, receipt.ConfirmedAt
	claim.state.Result = workflowstate.ClosingEvidence{Ref: workflowcloseout.File(claim.set.ID, claim.state.Generation, "result.json"), Hash: contenthash.Sum(data)}

	claim.set.Set.Status = "blocked"
	if receipt.Outcome == "completed" {
		claim.set.Set.Status, claim.set.Set.ClosureRef = "completed", claim.state.Result.Ref
	}

	if err := s.commitWorkflowClosingState(claim, workflowstate.FileChange{Path: claim.state.Result.Ref, After: data}); err != nil {
		return workflowcloseout.Result{}, errors.Join(failure, err)
	}

	result := workflowcloseout.Result{Set: claim.set, Receipt: receipt, ReceiptRef: claim.state.Result.Ref, Outcome: receipt.Outcome}

	return result, failure
}

func (s *Store) checkWorkflowClosingConfirmation(ctx context.Context, claim *workflowClosingClaim, ports FrontierPorts) error {
	if err := s.checkWorkflowClosingAuthority(claim); err != nil {
		return err
	}

	if err := s.checkWorkflowClosingInputs(ctx, claim); err != nil {
		return err
	}

	request := workflowcloseout.Request{Repository: claim.input.Target.Repository, Branch: claim.input.Branch}

	target, err := s.observeWorkflowClosingTarget(ctx, request, ports)
	if err != nil || target != claim.input.Target {
		return errorsOrConflict(err)
	}

	if err := workflowClosingProofs(s, ports).CheckCanonical(ctx, claim.input); err != nil {
		return err
	}

	// Observation can cross the deadline or a host edit. Recheck authority and
	// the frozen input after that external operation before confirming closure.
	if err := s.checkWorkflowClosingAuthority(claim); err != nil {
		return err
	}

	return s.checkWorkflowClosingInputs(ctx, claim)
}

func (s *Store) checkWorkflowClosingAuthority(claim *workflowClosingClaim) error {
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

	if run.StopIntent != "" || workflowstate.InvocationTerminal(run.Status) {
		return workflowstate.ErrStop
	}

	return nil
}
