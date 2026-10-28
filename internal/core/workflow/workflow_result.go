package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ResultReceipt is validated delivery evidence from the native owner.
type ResultReceipt struct {
	Reference workflowstate.EvidenceRef
	Receipt   Receipt
	Proof     SliceDeliveryProof
}

// ResultClosure is a successful set gate and its immutable reference.
type ResultClosure struct {
	Reference workflowstate.ClosingEvidence
	Receipt   workflowstate.ClosingReceipt
}

// ResultInspection distinguishes evidence-backed completion from counters
// and preserves partial delivery when an invocation is stopped or still waiting.
type ResultInspection struct {
	Invocation  Invocation
	Receipts    []ResultReceipt
	Closures    []ResultClosure
	Validations []ResultValidation
	Complete    bool
}

// InspectWorkflowResult reads and verifies retained evidence without recovery,
// provider calls, publication, native admission or changes to current inputs.
func (s *Store) InspectWorkflowResult(ctx context.Context, id string) (result ResultInspection, err error) {
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

func (s *Store) workflowResult(ctx context.Context, run workflowInvocationFile) (ResultInspection, error) {
	result := ResultInspection{Invocation: run.Invocation}
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

func (s *Store) workflowResultReceipt(run workflowInvocationFile) (ResultReceipt, error) {
	proof, err := s.readWorkflowDeliveryProof(run)
	if err != nil {
		return ResultReceipt{}, err
	}

	record, err := s.readRecord(run.DeliveryRef)
	if err != nil || !workflowSliceReceiptHeaderMatches(record, run) || !workflowReceiptMatchesProof(record, proof, run.DeliveryProofRef) {
		return ResultReceipt{}, errors.Join(workflowstate.ErrConflict, err)
	}

	return ResultReceipt{Reference: workflowstate.EvidenceRef{Path: record.Path, Hash: record.SourceHash}, Receipt: *record.Receipt, Proof: proof}, nil
}

func (s *Store) workflowResultClosure(closure InvocationClosure) (ResultClosure, error) {
	data, err := s.workflowClosingRead(closure.Receipt.Ref)
	if err != nil || contenthash.Sum(data) != closure.Receipt.Hash {
		return ResultClosure{}, errors.Join(workflowstate.ErrConflict, err)
	}

	var receipt workflowstate.ClosingReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return ResultClosure{}, err
	}

	if err := validateWorkflowClosingResults(receipt.ClosingInput.Validation, receipt); err != nil {
		return ResultClosure{}, err
	}

	return ResultClosure{Reference: closure.Receipt, Receipt: receipt}, nil
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
