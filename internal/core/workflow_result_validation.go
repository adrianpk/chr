package core

import (
	"context"
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowResultValidation contains the latest host task validation, without
// worker output, launch commands, process identities or owner credentials.
type WorkflowResultValidation struct {
	Ticket, Title, Outcome string
	Reference              workflowstate.WorkflowEvidenceRef
	Commands               []WorkflowCommandResult
	Waiver                 string
}

func (s *WorkflowStore) workflowResultValidations(ctx context.Context, run workflowInvocationFile) ([]WorkflowResultValidation, error) {
	result := []WorkflowResultValidation{}
	seen := map[string]bool{}
	remaining := 8 << 20

	for i := len(run.Attempts) - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		attempt := run.Attempts[i]
		if attempt.Purpose != "" || attempt.ResultRef == "" || seen[attempt.TaskKey] {
			continue
		}

		if len(result) >= workflowstate.PageLimit {
			return nil, ErrWorkflowCapacity
		}

		validation, size, err := s.workflowResultValidation(attempt)
		if err != nil {
			return nil, err
		}

		remaining -= size
		if remaining < 0 {
			return nil, ErrWorkflowCapacity
		}

		seen[attempt.TaskKey] = true

		result = append(result, validation)
	}

	return result, nil
}

func (s *WorkflowStore) workflowResultValidation(attempt WorkflowAttempt) (WorkflowResultValidation, int, error) {
	path, err := s.safePath(attempt.ResultRef)
	if err != nil {
		return WorkflowResultValidation{}, 0, err
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != attempt.ResultHash {
		return WorkflowResultValidation{}, 0, errorsOrConflict(err)
	}

	var evidence WorkflowTaskResult
	if err := json.Unmarshal(data, &evidence); err != nil {
		return WorkflowResultValidation{}, 0, err
	}

	if !workflowResultMatchesAttempt(evidence, attempt) {
		return WorkflowResultValidation{}, 0, workflowstate.ErrConflict
	}

	for i := range evidence.Validation {
		evidence.Validation[i].Stdout, evidence.Validation[i].Stderr = "", ""
	}

	input, size, err := s.workflowResultTask(attempt)
	if err != nil {
		return WorkflowResultValidation{}, 0, err
	}

	return WorkflowResultValidation{Ticket: attempt.TaskKey, Title: input.Contract.Title, Outcome: evidence.Outcome, Reference: workflowstate.WorkflowEvidenceRef{Path: attempt.ResultRef, Hash: attempt.ResultHash}, Commands: evidence.Validation, Waiver: evidence.WaiverReason}, len(data) + size, nil
}

func workflowResultMatchesAttempt(evidence WorkflowTaskResult, attempt WorkflowAttempt) bool {
	return evidence.AttemptID == attempt.ID && evidence.Generation == attempt.Generation && evidence.InputRef == attempt.InputRef && evidence.InputHash == attempt.InputHash
}

func (s *WorkflowStore) workflowResultTask(attempt WorkflowAttempt) (WorkflowTaskInput, int, error) {
	path, err := s.safePath(attempt.InputRef)
	if err != nil {
		return WorkflowTaskInput{}, 0, err
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != attempt.InputHash {
		return WorkflowTaskInput{}, 0, errorsOrConflict(err)
	}

	var input WorkflowTaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		return input, 0, err
	}

	if input.AttemptID != attempt.ID || input.Generation != attempt.Generation || input.TaskKey != attempt.TaskKey {
		return input, 0, workflowstate.ErrConflict
	}

	return input, len(data), nil
}
