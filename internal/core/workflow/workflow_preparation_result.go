package workflow

import (
	"context"
	"errors"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) commitWorkflowPreparationInput(ctx context.Context, run *workflowInvocationFile, record Record, observation workflowSliceObservation, input workflowstate.PreparationInput, lane workflowstate.Lane) (workflowPreparedTask, error) {
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

	encodedTaskInput, err := encodeWorkflowTaskInput(run, taskInput)
	if err != nil {
		return encodedTaskInput.Prepared, err
	}

	return encodedTaskInput.Prepared, s.commitWorkflowPullRequestFiles(run, files, encodedTaskInput.Change)
}

func checkWorkflowPreparedProof(query PreparationQuery, proof workflowstate.IntegrationProof) error {
	expected := query.IntegrationQuery

	expected.ResultCommit, expected.SourceDirectory = proof.ResultCommit, ""
	if proof.BaseCommit != query.BaseCommit || proof.ObservedTarget != query.BaseCommit {
		return workflowstate.ErrConflict
	}

	if err := checkWorkflowPreparationSigning(proof, query.RequiredSignatures, query.SigningProfile); err != nil {
		return err
	}

	return validateWorkflowIntegrationProof(expected, proof)
}

func (s *Store) checkWorkflowPreparationSource(ctx context.Context, run workflowInvocationFile, input workflowstate.PreparationInput, provider PublicationGuardProvider) error {
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
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) finishWorkflowIntegrationValidation(ctx context.Context, claim *Claim, prepared workflowPreparedTask, candidate *workflowstate.IntegrationCandidate, provider PublicationGuardProvider, validationErr error, interrupted bool) error {
	return s.withLock(ctx, func() error {
		claimObservation, err := s.checkWorkflowCredential(ctx, claim, workflowCredentialIdentity)
		if err != nil {
			return err
		}

		owner := claim.store
		if err := owner.checkWorkflowPreparedInput(prepared); err != nil {
			return err
		}

		if err := checkWorkflowExecutionBinding(claimObservation.Run, claimObservation.Record, prepared); err != nil {
			return err
		}

		if workflowExecutionStopped(claimObservation.Run, interrupted) {
			return owner.finishStoppedWorkflowTask(claimObservation.Run, claimObservation.Record, claim, &candidate.Validation)
		}

		if err := workflowAttemptQuiescent(claimObservation.Run.Attempts[len(claimObservation.Run.Attempts)-1]); err != nil {
			return err
		}

		failure := owner.checkWorkflowPreparationCompletion(ctx, claimObservation.Run, prepared, candidate, provider, validationErr)
		if failure != nil {
			return owner.checkpointWorkflowPreparationFailure(ctx, claimObservation.Run, claimObservation.Record, claim, *candidate, prepared, provider, validationErr, failure)
		}

		if err := owner.persistWorkflowPreparation(&claimObservation.Run, claimObservation.Record, *candidate); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}

func (s *Store) checkpointWorkflowPreparationFailure(ctx context.Context, run workflowInvocationFile, record Record, claim *Claim, candidate workflowstate.IntegrationCandidate, prepared workflowPreparedTask, provider PublicationGuardProvider, validationErr, failure error) error {
	// Only an observed command failure on the still-current clean result can
	// request semantic repair. Stops, mutation and stale fences remain failures.
	actionable := candidate.Input.Native != nil && workflowintegrate.WorkflowCandidateGateFailed(candidate.Validation) && s.checkWorkflowCandidateFailureCurrent(ctx, run, candidate, provider) == nil

	targetMoved := s.workflowNativePreparationTargetMoved(ctx, run, candidate, provider)
	if err := s.checkpointWorkflowAttempt(run, record, "interrupted", &candidate.Validation); err != nil {
		return errors.Join(failure, err)
	}

	repairErr := s.retainWorkflowPreparationFailureRepair(ctx, run, candidate, provider, actionable && validationErr != nil, targetMoved)

	return errors.Join(failure, repairErr, claim.releaseReservation())
}

func (s *Store) retainWorkflowPreparationFailureRepair(ctx context.Context, run workflowInvocationFile, candidate workflowstate.IntegrationCandidate, provider PublicationGuardProvider, actionable, targetMoved bool) error {
	if targetMoved {
		return s.retainWorkflowNativeTargetDrift(run.ID)
	}

	if actionable {
		return s.retainWorkflowFailedCandidate(ctx, run, candidate)
	}

	if candidate.Input.Native == nil {
		return nil
	}

	current, err := s.readWorkflowInvocation(run.ID)
	if err != nil || current.StopIntent != "" {
		return err
	}

	return s.checkWorkflowNativeSourceRevision(ctx, &current, provider)
}

func (s *Store) checkWorkflowCandidateFailureCurrent(ctx context.Context, run workflowInvocationFile, candidate workflowstate.IntegrationCandidate, provider PublicationGuardProvider) error {
	return errors.Join(s.checkWorkflowPreparationSource(ctx, run, candidate.Input, provider), s.checkWorkflowPreparationGit(ctx, candidate), s.checkWorkflowNativePreparationFence(run, candidate.Input.Native))
}

func (s *Store) checkWorkflowPreparationCompletion(ctx context.Context, run workflowInvocationFile, prepared workflowPreparedTask, candidate *workflowstate.IntegrationCandidate, provider PublicationGuardProvider, validationErr error) error {
	failure := errors.Join(validationErr, s.checkWorkflowPreparationSource(ctx, run, candidate.Input, provider), s.checkWorkflowPreparationGit(ctx, *candidate), workflowintegrate.ValidateWorkflowPreparationGate(prepared.input.Contract.Validation, candidate.Validation), s.checkWorkflowNativePreparationFence(run, candidate.Input.Native))
	if failure != nil {
		return failure
	}

	return s.sealWorkflowNativePreparation(run, candidate)
}
