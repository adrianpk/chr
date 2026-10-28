package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowSliceValidationPurpose = "slice_validation"

// SliceTaskProof binds ordered membership to host-accepted task evidence.
type SliceTaskProof struct {
	Key          string `json:"ticket_key"`
	ContractHash string `json:"contract_hash"`
	EvidenceRef  string `json:"evidence_ref"`
	EvidenceHash string `json:"evidence_hash"`
	InputRef     string `json:"input_ref"`
	InputHash    string `json:"input_hash"`
}

// SliceInput binds an integration contract and its task proofs.
// Independent ticket inputs omit set and aggregate-contract references.
type SliceInput struct {
	Unit         workflowstate.Ref `json:"unit"`
	SetID        string            `json:"set_id,omitempty"`
	ContractRef  string            `json:"contract_ref,omitempty"`
	ContractHash string            `json:"contract_hash"`
	Tasks        []SliceTaskProof  `json:"tasks"`
}

// SliceCandidate retains immutable host evidence for a slice or an
// independent ticket; the ticket field preserves its actual integration unit.
type SliceCandidate struct {
	Schema       string                      `json:"schema"`
	ID           string                      `json:"id"`
	InvocationID string                      `json:"invocation_id"`
	Slice        SliceInput                  `json:"slice,omitzero"`
	Ticket       *SliceInput                 `json:"ticket,omitempty"`
	Target       workflowstate.PreviewTarget `json:"target"`
	Lane         Lane                        `json:"lane"`
	Result       TaskResult                  `json:"result"`
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
	if errors.Is(err, ErrWorkflowOwned) {
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
	claim, prepared, cached, err := s.prepareWorkflowSliceValidation(ctx, id)
	if err != nil || cached != nil {
		if cached != nil {
			return *cached, err
		}

		return SliceCandidate{}, err
	}
	defer func() { err = errors.Join(err, closeWorkflowSliceCapacity(claim)) }()

	if err := beginWorkflowClaimExecution(claim); err != nil {
		return SliceCandidate{}, err
	}
	defer endWorkflowClaimExecution(claim)

	watched, stop := s.watchWorkflowStop(ctx, claim)
	defer stop()

	candidate, validationErr := s.runWorkflowSliceValidation(watched, claim, prepared)

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	err = s.finishWorkflowSliceValidation(cleanup, claim, prepared, &candidate, watched.Err() != nil)

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

func workflowSliceValidationReady(run Invocation) bool {
	return run.Status == "waiting_integration" || run.Status == "waiting_review" && run.CandidateRef != "" || run.Status == "paused" && run.StopIntent == "" && run.Outcome == "slice_validation_failed"
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
