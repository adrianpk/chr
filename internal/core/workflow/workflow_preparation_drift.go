package workflow

import (
	"context"
	"encoding/json"
	"slices"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s *Store) guardWorkflowPreparationWork(ctx context.Context, run workflowInvocationFile) error {
	if err := s.guardWorkflowPublicationJournal(ctx, run); err != nil {
		return err
	}

	if run.Integration == nil {
		return nil
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if run.StopIntent != "" || entry.StopIntent != "" {
		return ErrWorkflowStop
	}

	if entry.State == "blocked" && entry.Blocker.Reason == "native_not_applied" {
		return nil
	}

	if !slices.Contains([]IntegrationState{"queued", "preparing", "merging", "validating", "reviewing", "waiting_review", "ready_to_publish"}, entry.State) {
		return ErrWorkflowOwned
	}

	return nil
}

func (s *Store) reprepareWorkflowNativeInputs(ctx context.Context, run *workflowInvocationFile, source workflowstate.ReviewInput, capability workflowstate.PublicationCapability, provider PublicationGuardProvider) error {
	if capability.Mechanism != workspaceconfig.WorkflowPublicationGitAtomicDualRef || run.Preparation == nil {
		return nil
	}

	previous, err := s.retainedWorkflowNativePreparation(*run)
	if err != nil || previous.Native == nil {
		return err
	}

	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil || previous.Native.ClaimGeneration < entry.ClaimGeneration {
		return err
	}

	if previous.Native.ClaimGeneration != entry.ClaimGeneration {
		return workflowstate.ErrConflict
	}

	changed, err := workflowNativePreparationChanged(ctx, *run, previous, source, provider)
	if err != nil || !changed {
		return err
	}

	return s.replaceWorkflowNativePreparation(ctx, run, entry, previous, source)
}

func workflowNativePreparationChanged(ctx context.Context, run workflowInvocationFile, previous workflowstate.PreparationInput, source workflowstate.ReviewInput, provider PublicationGuardProvider) (bool, error) {
	hash, err := workflowreview.WorkflowReviewInputHash(source)
	if err != nil || hash != run.Preparation.SourceHash {
		return err == nil, err
	}

	capability, err := observeWorkflowPreparedCapability(ctx, provider, previous)

	return capability != previous.Capability, err
}

func (s *Store) retainedWorkflowNativePreparation(run workflowInvocationFile) (workflowstate.PreparationInput, error) {
	if run.Preparation.ResultRef != "" {
		candidate, err := s.readWorkflowPreparation(run)

		return candidate.Input, err
	}

	for i := len(run.Attempts) - 1; i >= 0; i-- {
		attempt := run.Attempts[i]
		if attempt.Purpose != workflowIntegrationValidationPurpose || attempt.InputRef == "" {
			continue
		}

		data, err := s.readWorkflowReviewImage(attempt.InputRef, attempt.InputHash)
		if err != nil {
			return workflowstate.PreparationInput{}, err
		}

		var input TaskInput
		if err := json.Unmarshal(data, &input); err != nil || input.Preparation == nil {
			return workflowstate.PreparationInput{}, errorsOrConflict(err)
		}

		if err := workflowreview.CheckWorkflowReviewInputHash(input.Preparation.Source, run.Preparation.SourceHash); err != nil {
			return workflowstate.PreparationInput{}, err
		}

		return *input.Preparation, nil
	}

	return workflowstate.PreparationInput{}, nil
}
