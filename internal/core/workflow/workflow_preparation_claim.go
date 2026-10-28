package workflow

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) prepareWorkflowIntegrationValidation(ctx context.Context, id string, provider PublicationGuardProvider, git PreparationGit) (selection workflowValidationPreparation, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if provider == nil || git == nil {
		return selection, fmt.Errorf("guarded provider and host Git preparer required")
	}

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		restoreWorkflowCapacityWait(&invocationLocation.Run)

		preflight, readErr := invocationLocation.Owner.workflowPreparationPreflight(ctx, &invocationLocation.Run, provider)
		if readErr != nil {
			return readErr
		}

		selection.Cached, readErr = invocationLocation.Owner.cachedWorkflowPreparation(ctx, invocationLocation.Run, preflight.Source, preflight.Capability, provider)
		if readErr != nil || selection.Cached != nil {
			return readErr
		}

		selection.Claimed, readErr = invocationLocation.Owner.claimWorkflowIntegrationValidation(ctx, &invocationLocation.Run, preflight, workflowPreparationServices{Provider: provider, Git: git})
		if invocationLocation.Run.Preparation != nil {
			selection.Round = invocationLocation.Run.Preparation.Round
		}

		return readErr
	})

	return selection, err
}

func (s *Store) workflowPreparationPreflight(ctx context.Context, run *workflowInvocationFile, provider PublicationGuardProvider) (result workflowPreparationObservation, err error) {
	if !slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationWaitingIntegration, workflowstate.InvocationWaitingReview, workflowstate.InvocationBlocked}, run.Status) {
		return result, ErrWorkflowOwned
	}

	if err := s.guardWorkflowPreparationWork(ctx, *run); err != nil {
		return result, err
	}

	if err := s.checkWorkflowNativeSourceRevision(ctx, run, provider); err != nil {
		return result, err
	}

	result.Observation, result.Source, err = s.workflowReviewSourceInputs(ctx, *run, provider)
	if err != nil {
		return result, err
	}

	result.Capability, err = observeWorkflowPublicationCapability(ctx, provider, result.Source)
	if errors.Is(err, workflowstate.ErrAtomicGuard) {
		return result, s.blockWorkflowPublication(run, err)
	}

	if err == nil {
		err = s.reprepareWorkflowNativeInputs(ctx, run, result.Source, result.Capability, provider)
	}

	if err == nil {
		err = s.resumeWorkflowCandidateRepairGate(ctx, run, result.Source)
	}

	if err == nil {
		err = s.checkWorkflowNativePreparationAdmission(*run, result.Source, result.Capability, provider)
	}

	return result, err
}

func (s *Store) cachedWorkflowPreparation(ctx context.Context, run workflowInvocationFile, source workflowstate.ReviewInput, capability workflowstate.PublicationCapability, provider PublicationGuardProvider) (*workflowstate.IntegrationCandidate, error) {
	if run.Preparation == nil || run.Preparation.State != "validated" {
		return nil, nil
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil {
		return nil, err
	}

	hash, err := workflowreview.InputHash(source)
	if err != nil {
		return nil, err
	}

	capability, err = s.cachedWorkflowPreparationCapability(ctx, run, candidate, capability, provider)
	if err != nil {
		return nil, err
	}

	if !workflowPreparationInputsCurrent(run, candidate, hash, capability) {
		if candidate.Input.Native != nil {
			return nil, workflowstate.ErrConflict
		}

		return nil, nil
	}

	if err := s.checkWorkflowPreparationGit(ctx, candidate); err != nil {
		return nil, err
	}

	return &candidate, nil
}

func workflowPreparationInputsCurrent(run workflowInvocationFile, candidate workflowstate.IntegrationCandidate, hash string, capability workflowstate.PublicationCapability) bool {
	return hash == run.Preparation.SourceHash && capability == candidate.Input.Capability
}

func (s *Store) blockWorkflowPublication(run *workflowInvocationFile, reason error) error {
	run.Status, run.Outcome = workflowstate.InvocationBlocked, "publication_guard_unsupported"

	run.Blocker = &Blocker{Dependency: run.Lane.Unit, Reason: "atomic_source_base_guard_required", NextActionOwner: "workflow_operator"}
	if run.Review != nil {
		run.Review.State = "stale"
	}

	return errors.Join(reason, s.commitWorkflowState(run))
}

func (s *Store) claimWorkflowIntegrationValidation(ctx context.Context, run *workflowInvocationFile, preflight workflowPreparationObservation, services workflowPreparationServices) (result workflowClaimedPreparation, err error) {
	claim, err := s.reserveWorkflowRole(ctx, *run, workflowIntegrationValidationPurpose)
	if err != nil {
		return result, err
	}

	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.releaseReservation())
		}
	}()

	lane, err := s.beginWorkflowPreparationClaim(run, preflight.Observation, preflight.Source, claim)
	if err != nil {
		return result, err
	}

	*run, err = s.readWorkflowInvocation(run.ID)
	if err != nil {
		return result, err
	}

	preflight.Observation, err = s.observeWorkflowSlice(ctx, *run)
	if err != nil {
		return result, err
	}

	record := preflight.Observation.tasks[len(preflight.Observation.tasks)-1]

	preparationContext, cancel := context.WithDeadline(ctx, run.Attempts[len(run.Attempts)-1].LeaseUntil)
	defer cancel()

	preparedGit, err := s.prepareWorkflowIntegrationGit(preparationContext, run, workflowPreparationGitRequest{Preflight: preflight, Lane: lane}, services)

	result.Input = preparedGit.Input
	if err != nil {
		return result, s.failClaimedWorkflowPreparation(*run, record, err)
	}

	result.Prepared, err = s.commitWorkflowPreparationInput(ctx, run, workflowPreparationInputChange{Record: record, Observation: preflight.Observation, Preparation: preparedGit})
	result.Claim = claim

	return result, err
}

func (s *Store) prepareWorkflowIntegrationGit(ctx context.Context, run *workflowInvocationFile, request workflowPreparationGitRequest, services workflowPreparationServices) (workflowPreparedIntegration, error) {
	result := workflowPreparedIntegration{Lane: request.Lane}
	preflight := request.Preflight

	query, err := s.workflowPreparationQuery(ctx, *run, preflight.Observation, preflight.Source, result.Lane)
	if err != nil {
		return result, err
	}

	query.RequiredSignatures, query.SigningProfile = preflight.Capability.RequiredSignatures, preflight.Capability.SigningProfile

	repairGit, err := s.prepareWorkflowCandidateRepairGit(ctx, *run, services.Git, query)
	if err != nil {
		return result, err
	}

	if err := workflowintegrate.CheckPrepared(query, repairGit.Proof, workflowIntegrationRepairs{}); err != nil {
		return result, err
	}

	result.Lane.HeadCommit = repairGit.Proof.ResultCommit

	result.Input = workflowstate.PreparationInput{Source: preflight.Source, Capability: preflight.Capability, Git: repairGit.Proof, MergerAttemptRefs: repairGit.Evidence}
	if err := s.bindWorkflowNativePreparation(ctx, *run, &result.Input, services.Provider); err != nil {
		return result, err
	}

	if err := s.checkWorkflowPreparationGit(ctx, workflowstate.IntegrationCandidate{Input: result.Input, Lane: result.Lane}); err != nil {
		return result, err
	}

	if err := s.checkWorkflowPreparationSource(ctx, *run, result.Input, services.Provider); err != nil {
		return result, err
	}

	return result, nil
}

func (s *Store) beginWorkflowPreparationClaim(run *workflowInvocationFile, observation workflowSliceObservation, source workflowstate.ReviewInput, claim *Claim) (workflowstate.Lane, error) {
	record := observation.tasks[len(observation.tasks)-1]
	if !workflowTaskClaimable(record.Task) {
		return workflowstate.Lane{}, ErrWorkflowOwned
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

func (s *Store) beginWorkflowPreparation(run *workflowInvocationFile, source workflowstate.ReviewInput) (workflowstate.Lane, error) {
	if run.PreparationCount >= workflowstate.PageLimit {
		return workflowstate.Lane{}, fmt.Errorf("workflow preparation history budget exhausted")
	}

	hash, err := workflowreview.InputHash(source)
	if err != nil {
		return workflowstate.Lane{}, err
	}

	run.PreparationCount++
	run.Preparation = &PreparationBinding{Round: run.PreparationCount, Unit: source.Unit, State: "preparing", SourceHash: hash}

	run.Status, run.Outcome, run.Blocker = workflowstate.InvocationWaitingIntegration, "integration_preparing", nil
	if run.Review != nil {
		run.Review.State = "stale"
	}

	id := contenthash.Sum([]byte(fmt.Sprintf("%s:%d", run.ID, run.PreparationCount)))[:24]
	lane := workflowstate.Lane{Unit: source.Unit, Branch: "refs/heads/chr/workflow/integration-" + id, Worktree: filepath.Join(s.layout.WorktreeRoot(), "integration-"+id), BaseCommit: source.TargetCommit, HeadCommit: source.TargetCommit}

	return lane, nil
}

func (s *Store) workflowPreparationQuery(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, source workflowstate.ReviewInput, lane workflowstate.Lane) (workflowstate.PreparationQuery, error) {
	candidate, err := s.readWorkflowSliceCandidate(ctx, run, observation)
	if err != nil {
		return workflowstate.PreparationQuery{}, err
	}

	query := workflowIntegrationQuery(run, candidate, source.PullRequest, source.Policy.Method)
	query.ResultCommit = ""

	return workflowstate.PreparationQuery{IntegrationQuery: query, BaseCommit: source.TargetCommit, Lane: lane}, nil
}

func (s *Store) failClaimedWorkflowPreparation(run workflowInvocationFile, record Record, reason error) error {
	if err := s.checkpointWorkflowAttempt(run, record, "interrupted", nil); err != nil {
		return errors.Join(reason, err)
	}

	current, err := s.readWorkflowInvocation(run.ID)
	if err != nil {
		return errors.Join(reason, err)
	}

	current.Status, current.Outcome = workflowstate.InvocationBlocked, "integration_preparation_failed"
	current.Blocker = &Blocker{Dependency: run.Lane.Unit, Reason: "combined_result_preparation_failed", NextActionOwner: "workflow_operator"}

	return errors.Join(reason, s.commitWorkflowState(&current))
}
