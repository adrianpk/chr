package workflow

import (
	"context"
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowManualReplay struct {
	Result workflowstate.TaskResult
	Found  bool
}

func (s *Store) reuseWorkflowManualResult(ctx context.Context, claim *Claim, candidate ManualCandidate) (observed workflowManualReplay, err error) {
	var (
		result workflowstate.TaskResult
		found  bool
	)

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, claim.InvocationID)
		if readErr != nil {
			return readErr
		}

		if len(invocationLocation.Run.Attempts) == 0 {
			return workflowstate.ErrOwned
		}

		attempt := invocationLocation.Run.Attempts[len(invocationLocation.Run.Attempts)-1]
		if attempt.Outcome != "validated" {
			return nil
		}

		found = true

		if !workflowManualResultOwnerMatches(invocationLocation.Run, attempt, claim) {
			return workflowstate.ErrOwned
		}

		if readErr = invocationLocation.Owner.checkWorkflowInputs(ctx, invocationLocation.Run); readErr != nil {
			return readErr
		}

		if readErr = workflowAttemptQuiescent(attempt); readErr != nil {
			return readErr
		}

		result, readErr = invocationLocation.Owner.checkWorkflowManualProof(ctx, invocationLocation.Run, attempt, candidate)
		if readErr != nil {
			return readErr
		}

		return claim.releaseReservation()
	})

	return workflowManualReplay{Result: result, Found: found}, err
}

func (s *Store) checkWorkflowManualProof(ctx context.Context, run workflowInvocationFile, attempt workflowstate.Attempt, candidate ManualCandidate) (workflowstate.TaskResult, error) {
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

func (s *Store) checkWorkflowManualRetainedImage(ctx context.Context, attempt workflowstate.Attempt, candidate ManualCandidate) error {
	prepared, err := s.readWorkflowManualInput(attempt)
	if err != nil {
		return err
	}

	if err := s.checkWorkflowPreparedInput(prepared); err != nil {
		return err
	}

	laneImage, err := s.workflowLaneCandidate(ctx, prepared.input.Lane)
	if err != nil {
		return err
	}

	if laneImage.Commit != candidate.Commit || laneImage.Tree != candidate.Tree {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowManualResultOwnerMatches(run workflowInvocationFile, attempt workflowstate.Attempt, claim *Claim) bool {
	return run.OwnerRoot == claim.store.layout.Root() && attempt.ID == claim.AttemptID && attempt.Generation == claim.Generation && run.Owners[attempt.ID] == contenthash.Sum([]byte(claim.Token))
}

func workflowAttemptTaskPath(run workflowInvocationFile, attempt workflowstate.Attempt) string {
	for _, member := range run.Members {
		if member.Admission.Key == attempt.TaskKey {
			return member.Admission.Path
		}
	}

	return ""
}

func (s *Store) readWorkflowManualResult(attempt workflowstate.Attempt, candidate ManualCandidate) (workflowstate.TaskResult, error) {
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

func workflowManualResultMatches(result workflowstate.TaskResult, attempt workflowstate.Attempt, candidate ManualCandidate) bool {
	return result.ExecutionMode == "manual" && result.AttemptID == attempt.ID && result.Generation == attempt.Generation && result.InputHash == candidate.InputHash && result.InputHash == attempt.InputHash && result.Branch == candidate.Branch && result.Commit == candidate.Commit && result.Tree == candidate.Tree
}
