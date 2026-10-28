package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) finishWorkflowSliceValidation(ctx context.Context, claim *Claim, prepared workflowPreparedTask, candidate *SliceCandidate, interrupted bool) error {
	return s.withLock(ctx, func() error {
		claimObservation, err := s.checkWorkflowCredential(ctx, claim, workflowCredentialIdentity)
		if err != nil {
			return err
		}

		if err := claim.store.checkWorkflowPreparedInput(prepared); err != nil {
			return err
		}

		if err := checkWorkflowExecutionBinding(claimObservation.Run, claimObservation.Record, prepared); err != nil {
			return err
		}

		if workflowExecutionStopped(claimObservation.Run, interrupted) {
			return claim.store.finishStoppedWorkflowTask(claimObservation.Run, claimObservation.Record, claim, &candidate.Result)
		}

		if err := workflowAttemptQuiescent(claimObservation.Run.Attempts[len(claimObservation.Run.Attempts)-1]); err != nil {
			return err
		}

		if err := claim.store.checkWorkflowTaskCandidate(ctx, &claimObservation.Run, prepared, &candidate.Result); err != nil {
			return err
		}

		if err := claim.store.commitWorkflowSliceCandidate(claimObservation.Run, claimObservation.Record, *candidate); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}

func (s *Store) commitWorkflowSliceCandidate(run workflowInvocationFile, record Record, candidate SliceCandidate) error {
	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return fmt.Errorf("workflow slice candidate exceeds file bounds: %w", errorsOrConflict(err))
	}

	data = append(data, '\n')
	ref := workflowAttemptFile(run.ID, candidate.ID, "result.json")
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.ResultRef, attempt.ResultHash, attempt.FinishedAt = ref, contenthash.Sum(data), time.Now().UTC()
	record.Task.AttemptID = ""

	run.Outcome = candidate.Result.Outcome
	if candidate.Result.Successful {
		attempt.Outcome, run.Status = "validated", workflowstate.InvocationWaitingIntegration
		run.CandidateRef, run.CandidateHash = ref, attempt.ResultHash
		run.ResumeStatus = workflowstate.InvocationUnspecified
	} else {
		attempt.Outcome, run.Status, run.ResumeStatus = "failed", workflowstate.InvocationPaused, workflowstate.InvocationWaitingIntegration
	}

	return s.commitWorkflowClaim(run, record, workflowstate.FileChange{Path: ref, After: data})
}

// ReadWorkflowSliceCandidate rechecks retained host proof and the actual lane.
func (s *Store) ReadWorkflowSliceCandidate(ctx context.Context, id string) (candidate SliceCandidate, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		observation, readErr := invocationLocation.Owner.observeWorkflowSlice(ctx, invocationLocation.Run)
		if readErr != nil {
			return readErr
		}

		candidate, readErr = invocationLocation.Owner.readWorkflowSliceCandidate(ctx, invocationLocation.Run, observation)

		return readErr
	})

	return candidate, err
}

func (s *Store) readWorkflowSliceCandidate(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation) (SliceCandidate, error) {
	path, err := s.safePath(run.CandidateRef)
	if err != nil || run.CandidateRef == "" {
		return SliceCandidate{}, workflowstate.ErrConflict
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != run.CandidateHash {
		return SliceCandidate{}, workflowstate.ErrConflict
	}

	var candidate SliceCandidate
	if err := json.Unmarshal(data, &candidate); err != nil {
		return candidate, err
	}

	if run.Lane.Unit.Kind == "ticket" {
		return candidate, s.checkWorkflowTicketCandidate(ctx, run, observation, candidate)
	}

	if err := checkWorkflowSliceCandidateBinding(run, observation, candidate); err != nil {
		return candidate, err
	}

	if err := s.checkWorkflowSliceCandidateProofs(ctx, run, observation, candidate); err != nil {
		return candidate, err
	}

	return candidate, nil
}

func checkWorkflowSliceCandidateBinding(run workflowInvocationFile, observation workflowSliceObservation, candidate SliceCandidate) error {
	if !workflowSliceCandidateMatchesRun(run, candidate) {
		return workflowstate.ErrConflict
	}

	if !workflowSliceCandidateMatchesContract(observation, candidate.UnitInput()) {
		return workflowstate.ErrConflict
	}

	for _, attempt := range run.Attempts {
		if workflowSliceCandidateMatchesAttempt(run, candidate, attempt) {
			return nil
		}
	}

	return workflowstate.ErrConflict
}

func workflowSliceCandidateMatchesRun(run workflowInvocationFile, candidate SliceCandidate) bool {
	return candidate.Ticket == nil && candidate.Schema == "chr-slice-candidate/v1" && candidate.InvocationID == run.ID && candidate.Target == run.ExecutionSelection().Target && candidate.Lane == *run.Lane && candidate.UnitInput().Unit == run.Lane.Unit
}

func workflowSliceCandidateMatchesContract(observation workflowSliceObservation, input workflowstate.SliceInput) bool {
	return input.SetID == observation.set.ID && input.ContractRef == observation.descriptor.ContractRef && input.ContractHash == observation.descriptor.ContractHash
}

func workflowSliceCandidateMatchesAttempt(run workflowInvocationFile, candidate SliceCandidate, attempt Attempt) bool {
	return attempt.ID == candidate.ID && attempt.Purpose == workflowSliceValidationPurpose && attempt.Outcome == "validated" && attempt.ResultRef == run.CandidateRef && attempt.ResultHash == run.CandidateHash && attempt.InputHash == candidate.Result.InputHash && attempt.InputRef == candidate.Result.InputRef
}

func (s *Store) checkWorkflowSliceCandidateProofs(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, candidate SliceCandidate) error {
	result := candidate.Result
	if !workflowSliceResultMatchesLane(result, *run.Lane) || len(candidate.UnitInput().Tasks) != len(observation.tasks) {
		return workflowstate.ErrConflict
	}

	tree, err := s.observeWorkflowLane(ctx, candidate.Lane)
	if err != nil || tree != result.Tree {
		return workflowstate.ErrConflict
	}

	if err := s.checkWorkflowPrerequisiteEvidence(map[string]string{result.InputRef: result.InputHash, candidate.UnitInput().ContractRef: candidate.UnitInput().ContractHash}); err != nil {
		return err
	}

	for i, record := range observation.tasks {
		proof, err := s.workflowSliceTaskProof(ctx, run, record)
		if err != nil || proof != candidate.UnitInput().Tasks[i] {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func workflowSliceResultMatchesLane(result workflowstate.TaskResult, lane workflowstate.Lane) bool {
	return result.Successful && result.Schema == "chr-slice-validation/v1" && result.Commit == lane.HeadCommit
}
