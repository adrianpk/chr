package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowSliceValidationPreparation struct {
	Claim    *Claim
	Prepared workflowPreparedTask
	Cached   *SliceCandidate
}

type workflowSliceValidationClaim struct {
	Claim    *Claim
	Prepared workflowPreparedTask
}

func (s *Store) prepareWorkflowSliceValidation(ctx context.Context, id string) (observed workflowSliceValidationPreparation, err error) {
	var (
		claim    *Claim
		prepared workflowPreparedTask
		cached   *SliceCandidate
	)

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		restoreWorkflowCapacityWait(&invocationLocation.Run)

		if !workflowSliceValidationReady(invocationLocation.Run.Invocation) || invocationLocation.Run.StopIntent != "" || invocationLocation.Run.Lane == nil || invocationLocation.Run.Lane.Unit.Kind != "slice" {
			return workflowstate.ErrOwned
		}

		observation, readErr := invocationLocation.Owner.observeWorkflowSlice(ctx, invocationLocation.Run)
		if readErr != nil {
			return fmt.Errorf("observe workflow slice: %w", readErr)
		}

		if invocationLocation.Run.CandidateRef != "" {
			candidate, readErr := invocationLocation.Owner.readWorkflowSliceCandidate(ctx, invocationLocation.Run, observation)
			cached = &candidate

			return readErr
		}

		sliceClaim, sliceClaimErr := invocationLocation.Owner.claimWorkflowSliceValidation(ctx, invocationLocation.Run, observation)

		claim, prepared, readErr = sliceClaim.Claim, sliceClaim.Prepared, sliceClaimErr
		if readErr != nil {
			return fmt.Errorf("prepare workflow slice: %w", readErr)
		}

		return nil
	})

	return workflowSliceValidationPreparation{Claim: claim, Prepared: prepared, Cached: cached}, err
}

func (s *Store) claimWorkflowSliceValidation(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation) (observed workflowSliceValidationClaim, err error) {
	var (
		claim    *Claim
		prepared workflowPreparedTask
	)

	record := observation.tasks[len(observation.tasks)-1]
	if !workflowTaskClaimable(record.Task) {
		return workflowSliceValidationClaim{Claim: nil, Prepared: prepared}, workflowstate.ErrOwned
	}

	claim, err = s.reserveWorkflowRole(ctx, run, workflowSliceValidationPurpose)
	if err != nil {
		return workflowSliceValidationClaim{Claim: nil, Prepared: prepared}, err
	}

	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.releaseReservation())
		}
	}()

	claim.purpose = workflowSliceValidationPurpose

	claim.Token, err = newWorkflowClaimToken()
	if err != nil {
		return workflowSliceValidationClaim{Claim: nil, Prepared: prepared}, err
	}

	if err := s.beginWorkflowAttempt(&run, &record, claim); err != nil {
		return workflowSliceValidationClaim{Claim: nil, Prepared: prepared}, err
	}

	run.Attempts[len(run.Attempts)-1].ContractHash = observation.descriptor.ContractHash
	prepared, err = s.commitWorkflowSliceInput(ctx, run, record, observation)

	return workflowSliceValidationClaim{Claim: claim, Prepared: prepared}, err
}

func (s *Store) commitWorkflowSliceInput(ctx context.Context, run workflowInvocationFile, record workflowstate.Record, observation workflowSliceObservation) (workflowPreparedTask, error) {
	// Proofs are observed before the journal changes the anchor's claim image.
	input, err := s.workflowSliceInput(ctx, run, record, observation)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	files, err := s.workflowClaimRecordChanges(&run, &record)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	input.SourceHash = record.SourceHash
	for i := range input.Prerequisites {
		if input.Prerequisites[i].Path == record.Path {
			input.Prerequisites[i].Revision, input.Prerequisites[i].SourceHash = record.Revision, record.SourceHash
		}
	}

	encodedTaskInput, err := encodeWorkflowTaskInput(&run, input)
	if err != nil {
		return encodedTaskInput.Prepared, err
	}

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(run)
	if err != nil {
		return encodedTaskInput.Prepared, err
	}

	files = append(files, encodedTaskInput.Change, state)

	return encodedTaskInput.Prepared, s.commitFiles(files)
}
