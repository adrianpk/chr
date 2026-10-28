package closeout

import (
	"encoding/json"
	"reflect"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadAttempt verifies the retained image hash before decoding and checking its binding.
func ReadAttempt(setID string, closing *workflowstate.WorkflowClosingBinding, index int, evidence Evidence) (Attempt, error) {
	binding := closing.Attempts[index]

	data, err := evidence.Read(binding.Ref)
	if err != nil {
		return Attempt{}, err
	}

	if contenthash.Sum(data) != binding.Hash {
		return Attempt{}, workflowstate.ErrConflict
	}

	var state Attempt
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}

	return state, checkAttempt(setID, closing, index, state)
}

// ReadReceipt checks the exact attempt, input and confirmation binding.
func ReadReceipt(state Attempt, evidence Evidence) (workflowstate.ClosingReceipt, error) {
	data, err := evidence.Read(state.Result.Ref)
	if err != nil {
		return workflowstate.ClosingReceipt{}, err
	}

	if contenthash.Sum(data) != state.Result.Hash {
		return workflowstate.ClosingReceipt{}, workflowstate.ErrConflict
	}

	var receipt workflowstate.ClosingReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return receipt, err
	}

	if receipt.Schema != "chr-set-closure/v1" || receipt.SetID != state.SetID || receipt.Outcome != state.Outcome || receipt.Input != state.Input || receipt.ConfirmedAt != state.FinishedAt {
		return receipt, workflowstate.ErrConflict
	}

	return receipt, nil
}

func checkResult(state Attempt, evidence Evidence) error {
	if state.Result.Ref == "" {
		return nil
	}

	data, err := evidence.Read(state.Input.Ref)
	if err != nil {
		return err
	}

	var input workflowstate.ClosingInput
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	receipt, err := ReadReceipt(state, evidence)
	if err != nil {
		return err
	}

	retained := receipt.ClosingInput

	retained.Schema = input.Schema
	if input.Schema != "chr-set-closing-input/v1" || !reflect.DeepEqual(retained, input) || receipt.Waiver != input.Validation.WaiverReason {
		return workflowstate.ErrConflict
	}

	return CheckResults(input.Validation, receipt)
}
