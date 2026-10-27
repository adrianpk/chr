package core

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowBoundTaskWorker(run workflowInvocationFile, worker []string, mode string) ([]string, error) {
	if run.HarnessBinding == nil {
		return worker, validateWorkflowInputWorker(mode, worker)
	}

	if mode == "manual" && len(worker) == 0 {
		return nil, nil
	}

	if len(worker) != 0 && !slices.Equal(worker, run.HarnessCommand) {
		return nil, workflowstate.ErrConflict
	}

	if err := checkHarnessExecutable(run.HarnessBinding, run.HarnessCommand); err != nil {
		return nil, err
	}

	return slices.Clone(run.HarnessCommand), validateWorkflowInputWorker(mode, run.HarnessCommand)
}

func validateWorkflowTaskHarnessSchema(input WorkflowTaskInput) error {
	switch input.Schema {
	case "chr-task-input/v1":
		if input.HarnessBinding != nil {
			return workflowstate.ErrConflict
		}
	case "chr-task-input/v2":
		if input.HarnessBinding == nil {
			return workflowstate.ErrConflict
		}
	default:
		return workflowstate.ErrConflict
	}

	if input.HarnessBinding == nil {
		return nil
	}

	if input.ExecutionMode == "manual" && len(input.Worker) == 0 {
		return ValidateHarnessBinding(*input.HarnessBinding)
	}

	return validateWorkflowHarnessPair(input.HarnessBinding, input.Worker)
}

func checkWorkflowTaskHarness(run workflowInvocationFile, input WorkflowTaskInput) error {
	if err := validateWorkflowTaskHarnessSchema(input); err != nil {
		return err
	}

	attempt := run.Attempts[len(run.Attempts)-1]
	if attempt.Purpose != "" {
		if input.HarnessBinding != nil {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if !sameHarnessBinding(run.HarnessBinding, input.HarnessBinding) {
		return workflowstate.ErrConflict
	}

	return nil
}
