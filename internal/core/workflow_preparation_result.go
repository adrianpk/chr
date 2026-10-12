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

	if err := checkWorkflowPreparationSigning(proof, query.RequiredSignatures, query.SigningProfile); err != nil {
		return err
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

	if err := s.checkWorkflowNativePreparationAdmission(run, source, input.Capability, provider); err != nil {
		return err
	}

	current := input
	current.Source = source

	capability, err := observeWorkflowPreparedCapability(ctx, provider, current)
	if err != nil {
		return err
	}

	if capability != input.Capability {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) finishWorkflowIntegrationValidation(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, candidate *WorkflowIntegrationCandidate, provider WorkflowPublicationGuardProvider, validationErr error, interrupted bool) error {
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

		failure := owner.checkWorkflowPreparationCompletion(ctx, run, prepared, candidate, provider, validationErr)
		if failure != nil {
			return owner.checkpointWorkflowPreparationFailure(ctx, run, record, claim, *candidate, prepared, provider, validationErr, failure)
		}

		if err := owner.persistWorkflowPreparation(&run, record, *candidate); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}

func (s *WorkflowStore) checkpointWorkflowPreparationFailure(ctx context.Context, run workflowInvocationFile, record WorkflowRecord, claim *WorkflowClaim, candidate WorkflowIntegrationCandidate, prepared workflowPreparedTask, provider WorkflowPublicationGuardProvider, validationErr, failure error) error {
	// Only an observed command failure on the still-current clean result can
	// request semantic repair. Stops, mutation and stale fences remain failures.
	actionable := candidate.Input.Native != nil && workflowCandidateGateFailed(candidate.Validation) && s.checkWorkflowCandidateFailureCurrent(ctx, run, candidate, provider) == nil
	if err := s.checkpointWorkflowAttempt(run, record, "interrupted", &candidate.Validation); err != nil {
		return errors.Join(failure, err)
	}

	var repairErr error
	if validationErr != nil && actionable {
		repairErr = s.retainWorkflowFailedCandidate(ctx, run, candidate)
	}

	return errors.Join(failure, repairErr, claim.releaseReservation())
}

func (s *WorkflowStore) checkWorkflowCandidateFailureCurrent(ctx context.Context, run workflowInvocationFile, candidate WorkflowIntegrationCandidate, provider WorkflowPublicationGuardProvider) error {
	return errors.Join(s.checkWorkflowPreparationSource(ctx, run, candidate.Input, provider), s.checkWorkflowPreparationGit(ctx, candidate), s.checkWorkflowNativePreparationFence(run, candidate.Input.Native))
}

func workflowCandidateGateFailed(result WorkflowTaskResult) bool {
	for _, command := range result.Validation {
		if command.ExitCode != 0 && !command.Truncated {
			return true
		}
	}

	return false
}

func (s *WorkflowStore) checkWorkflowPreparationCompletion(ctx context.Context, run workflowInvocationFile, prepared workflowPreparedTask, candidate *WorkflowIntegrationCandidate, provider WorkflowPublicationGuardProvider, validationErr error) error {
	failure := errors.Join(validationErr, s.checkWorkflowPreparationSource(ctx, run, candidate.Input, provider), s.checkWorkflowPreparationGit(ctx, *candidate), validateWorkflowPreparationGate(prepared.input, candidate.Validation), s.checkWorkflowNativePreparationFence(run, candidate.Input.Native))
	if failure != nil {
		return failure
	}

	return s.sealWorkflowNativePreparation(run, candidate)
}
