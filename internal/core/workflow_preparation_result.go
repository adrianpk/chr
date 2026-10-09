package core

import (
	"context"
	"errors"
)

func (s *WorkflowStore) commitWorkflowPreparationInput(ctx context.Context, run *workflowInvocationFile, record WorkflowRecord, observation workflowSliceObservation, input WorkflowPreparationInput, lane WorkflowLane) (workflowPreparedTask, error) {
	run.Attempts[len(run.Attempts)-1].ContractHash = workflowUnitContractHash(observation)

	taskInput, err := s.workflowSliceInput(ctx, *run, record, observation)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	files, err := s.workflowClaimRecordChanges(run, &record)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	taskInput.SourceHash, taskInput.Lane, taskInput.Preparation = record.SourceHash, lane, &input
	for i := range taskInput.Prerequisites {
		if taskInput.Prerequisites[i].Path == record.Path {
			taskInput.Prerequisites[i].Revision, taskInput.Prerequisites[i].SourceHash = record.Revision, record.SourceHash
		}
	}

	run.Preparation.State, run.Outcome = "validating", "integration_validating"

	prepared, file, err := encodeWorkflowTaskInput(run, taskInput)
	if err != nil {
		return prepared, err
	}

	return prepared, s.commitWorkflowPullRequestFiles(run, files, file)
}

func checkWorkflowPreparedProof(query WorkflowPreparationQuery, proof WorkflowIntegrationProof) error {
	expected := query.WorkflowIntegrationQuery

	expected.ResultCommit, expected.SourceDirectory = proof.ResultCommit, ""
	if proof.BaseCommit != query.BaseCommit || proof.ObservedTarget != query.BaseCommit {
		return ErrWorkflowConflict
	}

	return validateWorkflowIntegrationProof(expected, proof)
}

func (s *WorkflowStore) checkWorkflowPreparationSource(ctx context.Context, run workflowInvocationFile, input WorkflowPreparationInput, provider WorkflowPublicationGuardProvider) error {
	_, source, err := s.workflowReviewSourceInputs(ctx, run, provider)
	if err != nil {
		return err
	}

	if err := checkWorkflowReviewInputHash(source, run.Preparation.SourceHash); err != nil {
		return err
	}

	capability, err := observeWorkflowPublicationCapability(ctx, provider, source)
	if err != nil {
		return err
	}

	if capability != input.Capability {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) finishWorkflowIntegrationValidation(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, candidate WorkflowIntegrationCandidate, provider WorkflowPublicationGuardProvider, validationErr error, interrupted bool) error {
	return s.withLock(ctx, func() error {
		run, record, err := s.checkWorkflowCredential(ctx, claim, false)
		if err != nil {
			return err
		}

		owner := claim.store
		if err := owner.checkWorkflowPreparedInput(prepared); err != nil {
			return err
		}

		if err := checkWorkflowExecutionBinding(run, record, prepared); err != nil {
			return err
		}

		if workflowExecutionStopped(run, interrupted) {
			return owner.finishStoppedWorkflowTask(run, record, claim, &candidate.Validation)
		}

		if err := workflowAttemptQuiescent(run.Attempts[len(run.Attempts)-1]); err != nil {
			return err
		}

		failure := errors.Join(validationErr, owner.checkWorkflowPreparationSource(ctx, run, candidate.Input, provider), owner.checkWorkflowPreparationGit(ctx, candidate), validateWorkflowPreparationGate(prepared.input, candidate.Validation))
		if failure != nil {
			return errors.Join(failure, owner.checkpointWorkflowAttempt(run, record, "interrupted", &candidate.Validation), claim.releaseReservation())
		}

		if err := owner.persistWorkflowPreparation(&run, record, candidate); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}
