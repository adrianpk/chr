package integrate

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidationInput binds validation to one claimed round and prepared result.
// The workflow owner retains the claim, reservation and process resources.
type ValidationInput struct {
	InvocationID string
	Round        int
	Preparation  workflowstate.PreparationInput
	Lane         workflowstate.Lane
	Result       workflowstate.TaskResult
}

// ExecuteWorkflowValidation records the observed validation and requests atomic
// acceptance through the bound port. The execution deadline and stop context
// remain distinct: a deadline is a validation failure, not an operator stop.
func ExecuteWorkflowValidation(ctx, watched context.Context, input ValidationInput, validation Validation) (workflowstate.IntegrationCandidate, error) {
	result := input.Result
	result.Schema, result.Outcome, result.IntegrationPending = "chr-integration-validation/v1", "integration_validated", true

	validationErr := validation.Validate(ctx, &result)

	result.Successful, result.FinishedAt = validationErr == nil, time.Now().UTC()
	if validationErr != nil {
		result.Outcome, result.Failure = "integration_validation_failed", validationErr.Error()
		if len(result.Failure) > 4096 {
			result.Failure = result.Failure[:4096]
		}
	}

	candidate := workflowstate.IntegrationCandidate{Schema: "chr-integration-candidate/v1", InvocationID: input.InvocationID, Round: input.Round, Input: input.Preparation, Lane: input.Lane, Validation: result}

	cleanup, finish := context.WithTimeout(context.WithoutCancel(watched), 5*time.Second)
	defer finish()

	err := validation.Accept(cleanup, Completion{Candidate: &candidate, Failure: validationErr, Interrupted: watched.Err() != nil})

	return candidate, errors.Join(validationErr, err)
}
