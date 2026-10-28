package workflow

import (
	"context"
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) reuseWorkflowManualResult(ctx context.Context, claim *Claim, candidate ManualCandidate) (result workflowstate.TaskResult, found bool, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, claim.InvocationID)
		if readErr != nil {
			return readErr
		}

		if len(run.Attempts) == 0 {
			return ErrWorkflowOwned
		}

		attempt := run.Attempts[len(run.Attempts)-1]
		if attempt.Outcome != "validated" {
			return nil
		}

		found = true

		if !workflowManualResultOwnerMatches(run, attempt, claim) {
			return ErrWorkflowOwned
		}

		if readErr = owner.checkWorkflowInputs(ctx, run); readErr != nil {
			return readErr
		}

		if readErr = workflowAttemptQuiescent(attempt); readErr != nil {
			return readErr
		}

		result, readErr = owner.checkWorkflowManualProof(ctx, run, attempt, candidate)
		if readErr != nil {
			return readErr
		}

		return claim.releaseReservation()
	})

	return result, found, err
}

func (s *Store) checkWorkflowManualProof(ctx context.Context, run workflowInvocationFile, attempt Attempt, candidate ManualCandidate) (workflowstate.TaskResult, error) {
	result, err := s.readWorkflowManualResult(attempt, candidate)
	if err != nil {
		return result, err
	}

	if err := s.checkWorkflowManualRetainedImage(ctx, attempt, candidate); err != nil {
		return result, err
	}

	record, err := s.readRecord(workflowAttemptTaskPath(run, attempt))
	if err != nil {
		return result, err
	}

	request := run.ExecutionSelection()
	request.CumulativeBranch, request.CumulativeCommit = candidate.Branch, candidate.Commit

	reader := workflowProofReader{store: s, request: request}
	if !workflowAttemptValidatesTask(attempt, record) || !s.workflowRetainedInputMatches(attempt, record) || !reader.validTaskProof(ctx, record, result.TaskValidation) {
		return result, workflowstate.ErrConflict
	}

	return result, nil
}

func (s *Store) checkWorkflowManualRetainedImage(ctx context.Context, attempt Attempt, candidate ManualCandidate) error {
	prepared, err := s.readWorkflowManualInput(attempt)
	if err != nil {
		return err
	}

	if err := s.checkWorkflowPreparedInput(prepared); err != nil {
		return err
	}

	commit, tree, err := s.workflowLaneCandidate(ctx, prepared.input.Lane)
	if err != nil {
		return err
	}

	if commit != candidate.Commit || tree != candidate.Tree {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowManualResultOwnerMatches(run workflowInvocationFile, attempt Attempt, claim *Claim) bool {
	return run.OwnerRoot == claim.store.layout.Root() && attempt.ID == claim.AttemptID && attempt.Generation == claim.Generation && run.Owners[attempt.ID] == contenthash.Sum([]byte(claim.Token))
}

func workflowAttemptTaskPath(run workflowInvocationFile, attempt Attempt) string {
	for _, member := range run.Members {
		if member.Admission.Key == attempt.TaskKey {
			return member.Admission.Path
		}
	}

	return ""
}

func (s *Store) readWorkflowManualResult(attempt Attempt, candidate ManualCandidate) (workflowstate.TaskResult, error) {
	path, err := s.safePath(attempt.ResultRef)
	if err != nil {
		return workflowstate.TaskResult{}, err
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return workflowstate.TaskResult{}, err
	}

	if contenthash.Sum(data) != attempt.ResultHash {
		return workflowstate.TaskResult{}, workflowstate.ErrConflict
	}

	var result workflowstate.TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}

	if !workflowManualResultMatches(result, attempt, candidate) {
		return result, workflowstate.ErrConflict
	}

	return result, nil
}

func workflowManualResultMatches(result workflowstate.TaskResult, attempt Attempt, candidate ManualCandidate) bool {
	return result.ExecutionMode == "manual" && result.AttemptID == attempt.ID && result.Generation == attempt.Generation && result.InputHash == candidate.InputHash && result.InputHash == attempt.InputHash && result.Branch == candidate.Branch && result.Commit == candidate.Commit && result.Tree == candidate.Tree
}
