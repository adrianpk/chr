package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (s *WorkflowStore) finishWorkflowSliceValidation(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, candidate *WorkflowSliceCandidate, interrupted bool) error {
	return s.withLock(ctx, func() error {
		run, record, err := s.checkWorkflowCredential(ctx, claim, false)
		if err != nil {
			return err
		}

		if err := claim.store.checkWorkflowPreparedInput(prepared); err != nil {
			return err
		}

		if err := checkWorkflowExecutionBinding(run, record, prepared); err != nil {
			return err
		}

		if workflowExecutionStopped(run, interrupted) {
			return claim.store.finishStoppedWorkflowTask(run, record, claim, &candidate.Result)
		}

		if err := workflowAttemptQuiescent(run.Attempts[len(run.Attempts)-1]); err != nil {
			return err
		}

		if err := claim.store.checkWorkflowTaskCandidate(ctx, &run, prepared, &candidate.Result); err != nil {
			return err
		}

		if err := claim.store.commitWorkflowSliceCandidate(run, record, *candidate); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}

func (s *WorkflowStore) commitWorkflowSliceCandidate(run workflowInvocationFile, record WorkflowRecord, candidate WorkflowSliceCandidate) error {
	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return fmt.Errorf("workflow slice candidate exceeds file bounds: %w", errorsOrConflict(err))
	}

	data = append(data, '\n')
	ref := workflowAttemptFile(run.ID, candidate.ID, "result.json")
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.ResultRef, attempt.ResultHash, attempt.FinishedAt = ref, workflowHash(data), time.Now().UTC()
	record.Task.AttemptID = ""

	run.Outcome = candidate.Result.Outcome
	if candidate.Result.Successful {
		attempt.Outcome, run.Status = "validated", "waiting_integration"
		run.CandidateRef, run.CandidateHash = ref, attempt.ResultHash
		run.ResumeStatus = ""
	} else {
		attempt.Outcome, run.Status, run.ResumeStatus = "failed", "paused", "waiting_integration"
	}

	return s.commitWorkflowClaim(run, record, workflowFileChange{Path: ref, After: data})
}

// ReadWorkflowSliceCandidate rechecks retained host proof and the actual lane.
func (s *WorkflowStore) ReadWorkflowSliceCandidate(ctx context.Context, id string) (candidate WorkflowSliceCandidate, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		observation, readErr := owner.observeWorkflowSlice(ctx, run)
		if readErr != nil {
			return readErr
		}

		candidate, readErr = owner.readWorkflowSliceCandidate(ctx, run, observation)

		return readErr
	})

	return candidate, err
}

func (s *WorkflowStore) readWorkflowSliceCandidate(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation) (WorkflowSliceCandidate, error) {
	path, err := s.safePath(run.CandidateRef)
	if err != nil || run.CandidateRef == "" {
		return WorkflowSliceCandidate{}, ErrWorkflowConflict
	}

	data, err := readWorkflowFile(path)
	if err != nil || workflowHash(data) != run.CandidateHash {
		return WorkflowSliceCandidate{}, ErrWorkflowConflict
	}

	var candidate WorkflowSliceCandidate
	if err := json.Unmarshal(data, &candidate); err != nil {
		return candidate, err
	}

	if err := checkWorkflowSliceCandidateBinding(run, observation, candidate); err != nil {
		return candidate, err
	}

	if err := s.checkWorkflowSliceCandidateProofs(ctx, run, observation, candidate); err != nil {
		return candidate, err
	}

	return candidate, nil
}

func checkWorkflowSliceCandidateBinding(run workflowInvocationFile, observation workflowSliceObservation, candidate WorkflowSliceCandidate) error {
	if !workflowSliceCandidateMatchesRun(run, candidate) {
		return ErrWorkflowConflict
	}

	if !workflowSliceCandidateMatchesContract(observation, candidate.Slice) {
		return ErrWorkflowConflict
	}

	for _, attempt := range run.Attempts {
		if workflowSliceCandidateMatchesAttempt(run, candidate, attempt) {
			return nil
		}
	}

	return ErrWorkflowConflict
}

func workflowSliceCandidateMatchesRun(run workflowInvocationFile, candidate WorkflowSliceCandidate) bool {
	return candidate.Schema == "chr-slice-candidate/v1" && candidate.InvocationID == run.ID && candidate.Target == run.Selection.Target && candidate.Lane == *run.Lane && candidate.Slice.Unit == run.Lane.Unit
}

func workflowSliceCandidateMatchesContract(observation workflowSliceObservation, input WorkflowSliceInput) bool {
	return input.SetID == observation.set.ID && input.ContractRef == observation.descriptor.ContractRef && input.ContractHash == observation.descriptor.ContractHash
}

func workflowSliceCandidateMatchesAttempt(run workflowInvocationFile, candidate WorkflowSliceCandidate, attempt WorkflowAttempt) bool {
	return attempt.ID == candidate.ID && attempt.Purpose == workflowSliceValidationPurpose && attempt.Outcome == "validated" && attempt.ResultRef == run.CandidateRef && attempt.ResultHash == run.CandidateHash && attempt.InputHash == candidate.Result.InputHash && attempt.InputRef == candidate.Result.InputRef
}

func (s *WorkflowStore) checkWorkflowSliceCandidateProofs(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, candidate WorkflowSliceCandidate) error {
	result := candidate.Result
	if !workflowSliceResultMatchesLane(result, *run.Lane) || len(candidate.Slice.Tasks) != len(observation.tasks) {
		return ErrWorkflowConflict
	}

	tree, err := s.observeWorkflowLane(ctx, candidate.Lane)
	if err != nil || tree != result.Tree {
		return ErrWorkflowConflict
	}

	if err := s.checkWorkflowPrerequisiteEvidence(map[string]string{result.InputRef: result.InputHash, candidate.Slice.ContractRef: candidate.Slice.ContractHash}); err != nil {
		return err
	}

	for i, record := range observation.tasks {
		proof, err := s.workflowSliceTaskProof(ctx, run, record)
		if err != nil || proof != candidate.Slice.Tasks[i] {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func workflowSliceResultMatchesLane(result WorkflowTaskResult, lane WorkflowLane) bool {
	return result.Successful && result.Schema == "chr-slice-validation/v1" && result.Commit == lane.HeadCommit
}
