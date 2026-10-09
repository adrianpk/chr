package core

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"time"
)

func (s *WorkflowStore) prepareWorkflowIntegrationValidation(ctx context.Context, id string, provider WorkflowPublicationGuardProvider, git WorkflowPreparationGit) (claim *WorkflowClaim, prepared workflowPreparedTask, input WorkflowPreparationInput, round int, cached *WorkflowIntegrationCandidate, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if provider == nil || git == nil {
		return nil, prepared, input, 0, nil, fmt.Errorf("guarded provider and host Git preparer required")
	}

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		observation, source, capability, readErr := owner.workflowPreparationPreflight(ctx, &run, provider)
		if readErr != nil {
			return readErr
		}

		cached, readErr = owner.cachedWorkflowPreparation(ctx, run, source, capability)
		if readErr != nil || cached != nil {
			return readErr
		}

		claim, prepared, input, readErr = owner.claimWorkflowIntegrationValidation(ctx, &run, observation, source, capability, provider, git)
		if run.Preparation != nil {
			round = run.Preparation.Round
		}

		return readErr
	})

	return claim, prepared, input, round, cached, err
}

func (s *WorkflowStore) workflowPreparationPreflight(ctx context.Context, run *workflowInvocationFile, provider WorkflowPublicationGuardProvider) (workflowSliceObservation, WorkflowReviewInput, WorkflowPublicationCapability, error) {
	if !slices.Contains([]string{"waiting_integration", "waiting_review", "blocked"}, run.Status) {
		return workflowSliceObservation{}, WorkflowReviewInput{}, WorkflowPublicationCapability{}, ErrWorkflowOwned
	}

	if err := s.guardWorkflowPublicationWork(ctx, *run); err != nil {
		return workflowSliceObservation{}, WorkflowReviewInput{}, WorkflowPublicationCapability{}, err
	}

	observation, source, err := s.workflowReviewSourceInputs(ctx, *run, provider)
	if err != nil {
		return observation, source, WorkflowPublicationCapability{}, err
	}

	capability, err := observeWorkflowPublicationCapability(ctx, provider, source)
	if errors.Is(err, ErrWorkflowAtomicGuard) {
		return observation, source, capability, s.blockWorkflowPublication(run, err)
	}

	return observation, source, capability, err
}

func (s *WorkflowStore) cachedWorkflowPreparation(ctx context.Context, run workflowInvocationFile, source WorkflowReviewInput, capability WorkflowPublicationCapability) (*WorkflowIntegrationCandidate, error) {
	if run.Preparation == nil || run.Preparation.State != "validated" {
		return nil, nil
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil {
		return nil, err
	}

	hash, err := workflowReviewInputHash(source)
	if err != nil {
		return nil, err
	}

	if hash != run.Preparation.SourceHash || capability != candidate.Input.Capability {
		return nil, nil
	}

	if err := s.checkWorkflowPreparationGit(ctx, candidate); err != nil {
		return nil, err
	}

	return &candidate, nil
}

func (s *WorkflowStore) blockWorkflowPublication(run *workflowInvocationFile, reason error) error {
	run.Status, run.Outcome = "blocked", "publication_guard_unsupported"

	run.Blocker = &WorkflowBlocker{Dependency: run.Lane.Unit, Reason: "atomic_source_base_guard_required", NextActionOwner: "workflow_operator"}
	if run.Review != nil {
		run.Review.State = "stale"
	}

	return errors.Join(reason, s.commitWorkflowState(run))
}

func (s *WorkflowStore) claimWorkflowIntegrationValidation(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, source WorkflowReviewInput, capability WorkflowPublicationCapability, provider WorkflowPublicationGuardProvider, git WorkflowPreparationGit) (claim *WorkflowClaim, prepared workflowPreparedTask, input WorkflowPreparationInput, err error) {
	claim, err = s.reserveWorkflowClaim(ctx, *run)
	if err != nil {
		return nil, prepared, input, err
	}

	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.releaseReservation())
		}
	}()

	lane, err := s.beginWorkflowPreparationClaim(run, observation, source, claim)
	if err != nil {
		return nil, prepared, input, err
	}

	*run, err = s.readWorkflowInvocation(run.ID)
	if err != nil {
		return nil, prepared, input, err
	}

	observation, err = s.observeWorkflowSlice(ctx, *run)
	if err != nil {
		return nil, prepared, input, err
	}

	record := observation.tasks[len(observation.tasks)-1]

	preparationContext, cancel := context.WithDeadline(ctx, run.Attempts[len(run.Attempts)-1].LeaseUntil)
	defer cancel()

	input, lane, err = s.prepareWorkflowIntegrationGit(preparationContext, run, observation, source, capability, provider, git, lane)
	if err != nil {
		return nil, prepared, input, s.failClaimedWorkflowPreparation(*run, record, err)
	}

	prepared, err = s.commitWorkflowPreparationInput(ctx, run, record, observation, input, lane)

	return claim, prepared, input, err
}

func (s *WorkflowStore) prepareWorkflowIntegrationGit(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, source WorkflowReviewInput, capability WorkflowPublicationCapability, provider WorkflowPublicationGuardProvider, git WorkflowPreparationGit, lane WorkflowLane) (WorkflowPreparationInput, WorkflowLane, error) {
	query, err := s.workflowPreparationQuery(ctx, *run, observation, source, lane)
	if err != nil {
		return WorkflowPreparationInput{}, lane, err
	}

	proof, err := git.PrepareWorkflowIntegration(ctx, query)
	if err != nil {
		return WorkflowPreparationInput{}, lane, err
	}

	if err := checkWorkflowPreparedProof(query, proof); err != nil {
		return WorkflowPreparationInput{}, lane, err
	}

	lane.HeadCommit = proof.ResultCommit

	input := WorkflowPreparationInput{Source: source, Capability: capability, Git: proof}
	if err := s.checkWorkflowPreparationGit(ctx, WorkflowIntegrationCandidate{Input: input, Lane: lane}); err != nil {
		return input, lane, err
	}

	if err := s.checkWorkflowPreparationSource(ctx, *run, input, provider); err != nil {
		return input, lane, err
	}

	return input, lane, nil
}

func (s *WorkflowStore) beginWorkflowPreparationClaim(run *workflowInvocationFile, observation workflowSliceObservation, source WorkflowReviewInput, claim *WorkflowClaim) (WorkflowLane, error) {
	record := observation.tasks[len(observation.tasks)-1]
	if !workflowTaskClaimable(record.Task) {
		return WorkflowLane{}, ErrWorkflowOwned
	}

	lane, err := s.beginWorkflowPreparation(run, source)
	if err != nil {
		return lane, err
	}

	claim.purpose = workflowIntegrationValidationPurpose

	claim.Token, err = newWorkflowClaimToken()
	if err != nil {
		return lane, err
	}

	if err := s.beginWorkflowAttempt(run, &record, claim); err != nil {
		return lane, err
	}

	return lane, s.commitWorkflowClaim(*run, record)
}

func (s *WorkflowStore) beginWorkflowPreparation(run *workflowInvocationFile, source WorkflowReviewInput) (WorkflowLane, error) {
	if run.PreparationCount >= WorkflowPageLimit {
		return WorkflowLane{}, fmt.Errorf("workflow preparation history budget exhausted")
	}

	hash, err := workflowReviewInputHash(source)
	if err != nil {
		return WorkflowLane{}, err
	}

	run.PreparationCount++
	run.Preparation = &WorkflowPreparationBinding{Round: run.PreparationCount, Unit: source.Unit, State: "preparing", SourceHash: hash}

	run.Status, run.Outcome, run.Blocker = "waiting_integration", "integration_preparing", nil
	if run.Review != nil {
		run.Review.State = "stale"
	}

	id := workflowHash([]byte(fmt.Sprintf("%s:%d", run.ID, run.PreparationCount)))[:24]
	lane := WorkflowLane{Unit: source.Unit, Branch: "refs/heads/chr/workflow/integration-" + id, Worktree: filepath.Join(s.layout.WorktreeRoot(), "integration-"+id), BaseCommit: source.TargetCommit, HeadCommit: source.TargetCommit}

	return lane, nil
}

func (s *WorkflowStore) workflowPreparationQuery(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, source WorkflowReviewInput, lane WorkflowLane) (WorkflowPreparationQuery, error) {
	candidate, err := s.readWorkflowSliceCandidate(ctx, run, observation)
	if err != nil {
		return WorkflowPreparationQuery{}, err
	}

	query := workflowIntegrationQuery(run, candidate, source.PullRequest, source.Policy.Method)
	query.ResultCommit = ""

	return WorkflowPreparationQuery{WorkflowIntegrationQuery: query, BaseCommit: source.TargetCommit, Lane: lane}, nil
}

func (s *WorkflowStore) failClaimedWorkflowPreparation(run workflowInvocationFile, record WorkflowRecord, reason error) error {
	if err := s.checkpointWorkflowAttempt(run, record, "interrupted", nil); err != nil {
		return errors.Join(reason, err)
	}

	current, err := s.readWorkflowInvocation(run.ID)
	if err != nil {
		return errors.Join(reason, err)
	}

	current.Status, current.Outcome = "blocked", "integration_preparation_failed"
	current.Blocker = &WorkflowBlocker{Dependency: run.Lane.Unit, Reason: "combined_result_preparation_failed", NextActionOwner: "workflow_operator"}

	return errors.Join(reason, s.commitWorkflowState(&current))
}
