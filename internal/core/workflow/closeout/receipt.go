package closeout

import (
	"encoding/json"
	"errors"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Result distinguishes retained failures from confirmed closure.
type Result struct {
	Set        workflowstate.Record
	Receipt    workflowstate.ClosingReceipt
	ReceiptRef string
	Outcome    string
}

// InvocationResult is a successful set gate and its immutable reference.
type InvocationResult struct {
	Reference workflowstate.ClosingEvidence
	Receipt   workflowstate.ClosingReceipt
}

// ReadClosedSet attaches the exact confirmed receipt of an already closed set.
func ReadClosedSet(set workflowstate.Record, evidence Evidence) (Result, error) {
	if set.Set.Closing == nil || len(set.Set.Closing.Attempts) == 0 {
		return Result{}, workflowstate.ErrConflict
	}

	state, err := ReadAttempt(set.ID, set.Set.Closing, len(set.Set.Closing.Attempts)-1, evidence)
	if err != nil || state.Outcome != "completed" || state.Result.Ref != set.Set.ClosureRef {
		return Result{}, proofError(err)
	}

	receipt, err := ReadReceipt(state, evidence)
	if err != nil {
		return Result{}, err
	}

	return Result{Set: set, Receipt: receipt, ReceiptRef: state.Result.Ref, Outcome: state.Outcome}, nil
}

// ReadInvocationClosure verifies a retained gate result for result inspection.
func ReadInvocationClosure(closure workflowstate.InvocationClosure, evidence Evidence) (InvocationResult, error) {
	data, err := evidence.Read(closure.Receipt.Ref)
	if err != nil || contenthash.Sum(data) != closure.Receipt.Hash {
		return InvocationResult{}, errors.Join(workflowstate.ErrConflict, err)
	}

	var receipt workflowstate.ClosingReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return InvocationResult{}, err
	}

	if err := CheckResults(receipt.ClosingInput.Validation, receipt); err != nil {
		return InvocationResult{}, err
	}

	return InvocationResult{Reference: closure.Receipt, Receipt: receipt}, nil
}
