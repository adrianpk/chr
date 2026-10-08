package core

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *WorkflowStore) prepareWorkflowSliceValidation(ctx context.Context, id string) (claim *WorkflowClaim, prepared workflowPreparedTask, cached *WorkflowSliceCandidate, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if !workflowSliceValidationReady(run.WorkflowInvocation) || run.StopIntent != "" {
			return ErrWorkflowOwned
		}

		observation, readErr := owner.observeWorkflowSlice(ctx, run)
		if readErr != nil {
			return fmt.Errorf("observe workflow slice: %w", readErr)
		}

		if run.CandidateRef != "" {
			candidate, readErr := owner.readWorkflowSliceCandidate(ctx, run, observation)
			cached = &candidate

			return readErr
		}

		claim, prepared, readErr = owner.claimWorkflowSliceValidation(ctx, run, observation)
		if readErr != nil {
			return fmt.Errorf("prepare workflow slice: %w", readErr)
		}

		return nil
	})

	return claim, prepared, cached, err
}

func (s *WorkflowStore) claimWorkflowSliceValidation(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation) (claim *WorkflowClaim, prepared workflowPreparedTask, err error) {
	record := observation.tasks[len(observation.tasks)-1]
	if !workflowTaskClaimable(record.Task) {
		return nil, prepared, ErrWorkflowOwned
	}

	claim, err = s.reserveWorkflowClaim(ctx, run)
	if err != nil {
		return nil, prepared, err
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
		return nil, prepared, err
	}

	if err := s.beginWorkflowAttempt(&run, &record, claim); err != nil {
		return nil, prepared, err
	}

	run.Attempts[len(run.Attempts)-1].ContractHash = observation.descriptor.ContractHash
	prepared, err = s.commitWorkflowSliceInput(ctx, run, record, observation)

	return claim, prepared, err
}

func (s *WorkflowStore) commitWorkflowSliceInput(ctx context.Context, run workflowInvocationFile, record WorkflowRecord, observation workflowSliceObservation) (workflowPreparedTask, error) {
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

	prepared, file, err := encodeWorkflowTaskInput(&run, input)
	if err != nil {
		return prepared, err
	}

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(run)
	if err != nil {
		return prepared, err
	}

	files = append(files, file, state)

	return prepared, s.commitFiles(files)
}
