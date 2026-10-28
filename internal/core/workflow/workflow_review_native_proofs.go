package workflow

import (
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Native review composes the existing record and preparation checks without
// granting the review stage access to mutable workflow or publication state.
type workflowNativeReviewProofs struct{}

func (workflowNativeReviewProofs) CheckNativeRecordBinding(binding workflowstate.NativeRecordBinding) error {
	return validateWorkflowNativeRecordBinding(binding)
}

func (workflowNativeReviewProofs) CheckNativePreparation(input workflowstate.ReviewInput) error {
	preparation := workflowstate.PreparationInput{Source: workflowintegrate.SourceInput(input), Git: *input.PreparedGit}

	return validateWorkflowPreparedNativeCapability(preparation, input.Native.Capability)
}
