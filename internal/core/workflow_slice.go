package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowSliceValidationPurpose = "slice_validation"

// WorkflowSliceTaskProof binds ordered membership to host-accepted task evidence.
type WorkflowSliceTaskProof struct {
	Key          string `json:"ticket_key"`
	ContractHash string `json:"contract_hash"`
	EvidenceRef  string `json:"evidence_ref"`
	EvidenceHash string `json:"evidence_hash"`
	InputRef     string `json:"input_ref"`
	InputHash    string `json:"input_hash"`
}

// WorkflowSliceInput binds an integration contract and its task proofs.
// Independent ticket inputs omit set and aggregate-contract references.
type WorkflowSliceInput struct {
	Unit         workflowstate.Ref        `json:"unit"`
	SetID        string                   `json:"set_id,omitempty"`
	ContractRef  string                   `json:"contract_ref,omitempty"`
	ContractHash string                   `json:"contract_hash"`
	Tasks        []WorkflowSliceTaskProof `json:"tasks"`
}

// WorkflowSliceCandidate retains immutable host evidence for a slice or an
// independent ticket; the ticket field preserves its actual integration unit.
type WorkflowSliceCandidate struct {
	Schema       string                `json:"schema"`
	ID           string                `json:"id"`
	InvocationID string                `json:"invocation_id"`
	Slice        WorkflowSliceInput    `json:"slice,omitzero"`
	Ticket       *WorkflowSliceInput   `json:"ticket,omitempty"`
	Target       WorkflowPreviewTarget `json:"target"`
	Lane         WorkflowLane          `json:"lane"`
	Result       WorkflowTaskResult    `json:"result"`
}

// ExecuteWorkflowSlice runs remaining tasks on the retained lane and validates
// one complete slice. It stops at that slice's integration barrier.
func (s *WorkflowStore) ExecuteWorkflowSlice(ctx context.Context, id string, worker []string) (WorkflowSliceCandidate, error) {
	for range workflowstate.RecordLimit {
		run, err := s.ReadWorkflow(ctx, id)
		if err != nil {
			return WorkflowSliceCandidate{}, err
		}

		if run.Selection.Selector.Extent == "ticket" || run.Selection.Selector.Extent == "loose" {
			return WorkflowSliceCandidate{}, fmt.Errorf("slice execution requires a slice-owned extent")
		}

		if workflowSliceValidationReady(run) {
			return s.ValidateWorkflowSlice(ctx, id)
		}

		if err := s.executeWorkflowSliceTask(ctx, id, worker); err != nil {
			return WorkflowSliceCandidate{}, fmt.Errorf("workflow slice task: %w", err)
		}
	}

	return WorkflowSliceCandidate{}, fmt.Errorf("workflow slice execution bound exceeded")
}

func (s *WorkflowStore) executeWorkflowSliceTask(ctx context.Context, id string, worker []string) error {
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
func (s *WorkflowStore) ValidateWorkflowSlice(ctx context.Context, id string) (result WorkflowSliceCandidate, err error) {
	claim, prepared, cached, err := s.prepareWorkflowSliceValidation(ctx, id)
	if err != nil || cached != nil {
		if cached != nil {
			return *cached, err
		}

		return WorkflowSliceCandidate{}, err
	}
	defer func() { err = errors.Join(err, closeWorkflowSliceCapacity(claim)) }()

	if err := beginWorkflowClaimExecution(claim); err != nil {
		return WorkflowSliceCandidate{}, err
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

func closeWorkflowSliceCapacity(claim *WorkflowClaim) error {
	if claim.reservation == nil {
		return nil
	}

	// Failed journal/credential checks retain operation locks for reconciliation.
	// Closing this descriptor never releases a surviving supervisor's reservation.
	err := claim.reservation.Close()
	claim.reservation = nil

	return err
}

func workflowSliceValidationReady(run WorkflowInvocation) bool {
	return run.Status == "waiting_integration" || run.Status == "waiting_review" && run.CandidateRef != "" || run.Status == "paused" && run.StopIntent == "" && run.Outcome == "slice_validation_failed"
}

func (s *WorkflowStore) runWorkflowSliceValidation(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask) (WorkflowSliceCandidate, error) {
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
	candidate := WorkflowSliceCandidate{Schema: "chr-slice-candidate/v1", ID: input.AttemptID, InvocationID: input.InvocationID, Slice: *input.Slice, Target: input.Target, Lane: input.Lane, Result: result}

	return candidate, err
}
