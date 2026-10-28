package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowSliceValidationPurpose = "slice_validation"

// SliceCandidate retains immutable host evidence for a slice or an
// independent ticket; the ticket field preserves its actual integration unit.
type SliceCandidate struct {
	Schema       string                      `json:"schema"`
	ID           string                      `json:"id"`
	InvocationID string                      `json:"invocation_id"`
	Slice        workflowstate.SliceInput    `json:"slice,omitzero"`
	Ticket       *workflowstate.SliceInput   `json:"ticket,omitempty"`
	Target       workflowstate.PreviewTarget `json:"target"`
	Lane         workflowstate.Lane          `json:"lane"`
	Result       workflowstate.TaskResult    `json:"result"`
}

// ExecuteWorkflowSlice runs remaining tasks on the retained lane and validates
// one complete slice. It stops at that slice's integration barrier.
func (s *Store) ExecuteWorkflowSlice(ctx context.Context, id string, worker []string) (SliceCandidate, error) {
	for range workflowstate.RecordLimit {
		run, err := s.ReadWorkflow(ctx, id)
		if err != nil {
			return SliceCandidate{}, err
		}

		if run.Selection.Selector.Extent == "ticket" || run.Selection.Selector.Extent == "loose" {
			return SliceCandidate{}, fmt.Errorf("slice execution requires a slice-owned extent")
		}

		if workflowSliceValidationReady(run) {
			return s.ValidateWorkflowSlice(ctx, id)
		}

		if err := s.executeWorkflowSliceTask(ctx, id, worker); err != nil {
			return SliceCandidate{}, fmt.Errorf("workflow slice task: %w", err)
		}
	}

	return SliceCandidate{}, fmt.Errorf("workflow slice execution bound exceeded")
}

func (s *Store) executeWorkflowSliceTask(ctx context.Context, id string, worker []string) error {
	if err := workflowstate.ValidateCommand(worker); err != nil {
		return err
	}

	claim, err := s.ClaimWorkflow(ctx, id)
	if errors.Is(err, workflowstate.ErrOwned) {
		// Claiming may have reused the complete validated prefix without a worker.
		run, readErr := s.ReadWorkflow(ctx, id)
		if readErr == nil && workflowSliceValidationReady(run) {
			return nil
		}
	}

	if err != nil {
		return err
	}

	_, err = s.ExecuteWorkflowTask(ctx, claim, worker)

	return err
}

// ValidateWorkflowSlice runs only the host aggregate gate; retries retain tasks.
func (s *Store) ValidateWorkflowSlice(ctx context.Context, id string) (result SliceCandidate, err error) {
	slicePreparation, err := s.prepareWorkflowSliceValidation(ctx, id)

	if err != nil || slicePreparation.Cached != nil {
		if slicePreparation.Cached != nil {
			return *slicePreparation.Cached, err
		}

		return SliceCandidate{}, err
	}
	defer func() { err = errors.Join(err, closeWorkflowSliceCapacity(slicePreparation.Claim)) }()

	if err := beginWorkflowClaimExecution(slicePreparation.Claim); err != nil {
		return SliceCandidate{}, err
	}
	defer endWorkflowClaimExecution(slicePreparation.Claim)

	watched, stop := context.WithCancel(ctx)

	s.watchWorkflowStop(watched, workflowStopWatch{Claim: slicePreparation.Claim, Cancel: stop})
	defer stop()

	candidate, validationErr := s.runWorkflowSliceValidation(watched, slicePreparation.Claim, slicePreparation.Prepared)

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	err = s.finishWorkflowSliceValidation(cleanup, slicePreparation.Claim, slicePreparation.Prepared, &candidate, watched.Err() != nil)

	return candidate, errors.Join(validationErr, err)
}

func closeWorkflowSliceCapacity(claim *Claim) error {
	if claim.reservation == nil {
		return nil
	}

	// Failed journal/credential checks retain operation locks for reconciliation.
	// Closing this descriptor never releases a surviving supervisor's reservation.
	err := claim.reservation.Close()
	claim.reservation = nil

	return err
}

func workflowSliceValidationReady(run workflowstate.Invocation) bool {
	return run.Status == workflowstate.InvocationWaitingIntegration || run.Status == workflowstate.InvocationWaitingReview && run.CandidateRef != "" || run.Status == workflowstate.InvocationPaused && run.StopIntent == "" && run.Outcome == "slice_validation_failed"
}

func (s *Store) runWorkflowSliceValidation(ctx context.Context, claim *Claim, prepared workflowPreparedTask) (SliceCandidate, error) {
	deadline, cancel := context.WithDeadline(ctx, prepared.leaseUntil)
	defer cancel()

	result := newWorkflowTaskResult(prepared)
	result.Schema = "chr-slice-validation/v1"
	result.Outcome, result.IntegrationPending = "slice_validated", true

	err := s.validateWorkflowTask(deadline, claim, prepared, &result)

	result.Successful, result.FinishedAt = err == nil, time.Now().UTC()
	if err != nil {
		result.Outcome, result.Failure = "slice_validation_failed", workflowFailureText(err)
	}

	input := prepared.input
	candidate := SliceCandidate{Schema: "chr-slice-candidate/v1", ID: input.AttemptID, InvocationID: input.InvocationID, Slice: *input.Slice, Target: input.Target, Lane: input.Lane, Result: result}

	return candidate, err
}
