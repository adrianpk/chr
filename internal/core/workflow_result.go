package core

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowResultReceipt is validated delivery evidence from the native owner.
type WorkflowResultReceipt struct {
	Reference workflowstate.WorkflowEvidenceRef
	Receipt   WorkflowReceipt
	Proof     WorkflowSliceDeliveryProof
}

// WorkflowResultClosure is a successful set gate and its immutable reference.
type WorkflowResultClosure struct {
	Reference workflowstate.WorkflowClosingEvidence
	Receipt   WorkflowClosingReceipt
}

// WorkflowResultInspection distinguishes evidence-backed completion from counters
// and preserves partial delivery when an invocation is stopped or still waiting.
type WorkflowResultInspection struct {
	Invocation  WorkflowInvocation
	Receipts    []WorkflowResultReceipt
	Closures    []WorkflowResultClosure
	Validations []WorkflowResultValidation
	Complete    bool
}

// InspectWorkflowResult reads and verifies retained evidence without recovery,
// provider calls, publication, native admission or changes to current inputs.
func (s *WorkflowStore) InspectWorkflowResult(ctx context.Context, id string) (result WorkflowResultInspection, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withWorkflowInspection(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		result, err = owner.workflowResult(ctx, run)

		return err
	})

	return result, err
}

func (s *WorkflowStore) workflowResult(ctx context.Context, run workflowInvocationFile) (WorkflowResultInspection, error) {
	result := WorkflowResultInspection{Invocation: run.WorkflowInvocation}
	if err := s.checkWorkflowContinuationEvidence(ctx, run); err != nil {
		return result, err
	}

	cursor := 0

	for _, delivery := range run.Deliveries {
		receipt, err := s.workflowResultReceipt(workflowDeliveredUnitView(run, delivery))
		if err != nil {
			return result, err
		}

		result.Receipts = append(result.Receipts, receipt)
		cursor = delivery.Cursor
	}

	if run.ConfirmedCursor > cursor {
		receipt, err := s.workflowResultReceipt(run)
		if err != nil {
			return result, err
		}

		result.Receipts = append(result.Receipts, receipt)
		cursor = run.ConfirmedCursor
	}

	for _, closure := range run.Closures {
		receipt, err := s.workflowResultClosure(closure)
		if err != nil {
			return result, err
		}

		result.Closures = append(result.Closures, receipt)
	}

	var err error

	result.Validations, err = s.workflowResultValidations(ctx, run)
	if err != nil {
		return result, err
	}

	result.Complete = workflowResultComplete(run, cursor)

	return result, nil
}

func (s *WorkflowStore) workflowResultReceipt(run workflowInvocationFile) (WorkflowResultReceipt, error) {
	proof, err := s.readWorkflowDeliveryProof(run)
	if err != nil {
		return WorkflowResultReceipt{}, err
	}

	record, err := s.readRecord(run.DeliveryRef)
	if err != nil || !workflowSliceReceiptHeaderMatches(record, run) || !workflowReceiptMatchesProof(record, proof, run.DeliveryProofRef) {
		return WorkflowResultReceipt{}, errors.Join(workflowstate.ErrConflict, err)
	}

	return WorkflowResultReceipt{Reference: workflowstate.WorkflowEvidenceRef{Path: record.Path, Hash: record.SourceHash}, Receipt: *record.Receipt, Proof: proof}, nil
}

func (s *WorkflowStore) workflowResultClosure(closure WorkflowInvocationClosure) (WorkflowResultClosure, error) {
	data, err := s.workflowClosingRead(closure.Receipt.Ref)
	if err != nil || contenthash.Sum(data) != closure.Receipt.Hash {
		return WorkflowResultClosure{}, errors.Join(workflowstate.ErrConflict, err)
	}

	var receipt WorkflowClosingReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return WorkflowResultClosure{}, err
	}

	if err := validateWorkflowClosingResults(receipt.WorkflowClosingInput.Validation, receipt); err != nil {
		return WorkflowResultClosure{}, err
	}

	return WorkflowResultClosure{Reference: closure.Receipt, Receipt: receipt}, nil
}

func workflowResultComplete(run workflowInvocationFile, cursor int) bool {
	if len(run.Members) == 0 && len(run.SetImages) == 0 {
		return false
	}

	if run.Status != "completed" || run.StopIntent != "" || cursor != len(run.Members) || run.ConfirmedCursor != cursor || run.Cursor != cursor {
		return false
	}

	return len(run.SetImages) == len(run.Closures)
}
