package workflow

import (
	"context"
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowValidationRead struct {
	Validation ResultValidation
	Bytes      int
}

type workflowTaskInputRead struct {
	Input TaskInput
	Bytes int
}

// ResultValidation contains the latest host task validation, without
// worker output, launch commands, process identities or owner credentials.
type ResultValidation struct {
	Ticket, Title, Outcome string
	Reference              workflowstate.EvidenceRef
	Commands               []workflowstate.CommandResult
	Waiver                 string
}

func (s *Store) workflowResultValidations(ctx context.Context, run workflowInvocationFile) ([]ResultValidation, error) {
	result := []ResultValidation{}
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

		validationRead, err := s.workflowResultValidation(attempt)
		if err != nil {
			return nil, err
		}

		remaining -= validationRead.Bytes
		if remaining < 0 {
			return nil, ErrWorkflowCapacity
		}

		seen[attempt.TaskKey] = true

		result = append(result, validationRead.Validation)
	}

	return result, nil
}

func (s *Store) workflowResultValidation(attempt workflowstate.Attempt) (workflowValidationRead, error) {
	path, err := s.safePath(attempt.ResultRef)
	if err != nil {
		return workflowValidationRead{Validation: ResultValidation{}, Bytes: 0}, err
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != attempt.ResultHash {
		return workflowValidationRead{Validation: ResultValidation{}, Bytes: 0}, errorsOrConflict(err)
	}

	var evidence workflowstate.TaskResult
	if err := json.Unmarshal(data, &evidence); err != nil {
		return workflowValidationRead{Validation: ResultValidation{}, Bytes: 0}, err
	}

	if !workflowResultMatchesAttempt(evidence, attempt) {
		return workflowValidationRead{Validation: ResultValidation{}, Bytes: 0}, workflowstate.ErrConflict
	}

	for i := range evidence.Validation {
		evidence.Validation[i].Stdout, evidence.Validation[i].Stderr = "", ""
	}

	taskInputRead, err := s.workflowResultTask(attempt)
	if err != nil {
		return workflowValidationRead{Validation: ResultValidation{}, Bytes: 0}, err
	}

	return workflowValidationRead{Validation: ResultValidation{Ticket: attempt.TaskKey, Title: taskInputRead.Input.Contract.Title, Outcome: evidence.Outcome, Reference: workflowstate.EvidenceRef{Path: attempt.ResultRef, Hash: attempt.ResultHash}, Commands: evidence.Validation, Waiver: evidence.WaiverReason}, Bytes: len(data) + taskInputRead.Bytes}, nil
}

func workflowResultMatchesAttempt(evidence workflowstate.TaskResult, attempt workflowstate.Attempt) bool {
	return evidence.AttemptID == attempt.ID && evidence.Generation == attempt.Generation && evidence.InputRef == attempt.InputRef && evidence.InputHash == attempt.InputHash
}

func (s *Store) workflowResultTask(attempt workflowstate.Attempt) (workflowTaskInputRead, error) {
	path, err := s.safePath(attempt.InputRef)
	if err != nil {
		return workflowTaskInputRead{Input: TaskInput{}, Bytes: 0}, err
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != attempt.InputHash {
		return workflowTaskInputRead{Input: TaskInput{}, Bytes: 0}, errorsOrConflict(err)
	}

	var input TaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		return workflowTaskInputRead{Input: input, Bytes: 0}, err
	}

	if input.AttemptID != attempt.ID || input.Generation != attempt.Generation || input.TaskKey != attempt.TaskKey {
		return workflowTaskInputRead{Input: input, Bytes: 0}, workflowstate.ErrConflict
	}

	return workflowTaskInputRead{Input: input, Bytes: len(data)}, nil
}
