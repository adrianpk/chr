package workflow

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) prepareWorkflowIntegrationValidation(ctx context.Context, id string, provider PublicationGuardProvider, git PreparationGit) (claim *Claim, prepared workflowPreparedTask, input workflowstate.PreparationInput, round int, cached *workflowstate.IntegrationCandidate, err error) {
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

		restoreWorkflowCapacityWait(&run)

		observation, source, capability, readErr := owner.workflowPreparationPreflight(ctx, &run, provider)
		if readErr != nil {
			return readErr
		}

		cached, readErr = owner.cachedWorkflowPreparation(ctx, run, source, capability, provider)
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

func (s *Store) workflowPreparationPreflight(ctx context.Context, run *workflowInvocationFile, provider PublicationGuardProvider) (workflowSliceObservation, workflowstate.ReviewInput, workflowstate.PublicationCapability, error) {
	if !slices.Contains([]string{"waiting_integration", "waiting_review", "blocked"}, run.Status) {
		return workflowSliceObservation{}, workflowstate.ReviewInput{}, workflowstate.PublicationCapability{}, ErrWorkflowOwned
	}

	if err := s.guardWorkflowPreparationWork(ctx, *run); err != nil {
		return workflowSliceObservation{}, workflowstate.ReviewInput{}, workflowstate.PublicationCapability{}, err
	}

	if err := s.checkWorkflowNativeSourceRevision(ctx, run, provider); err != nil {
		return workflowSliceObservation{}, workflowstate.ReviewInput{}, workflowstate.PublicationCapability{}, err
	}

	observation, source, err := s.workflowReviewSourceInputs(ctx, *run, provider)
	if err != nil {
		return observation, source, workflowstate.PublicationCapability{}, err
	}

	capability, err := observeWorkflowPublicationCapability(ctx, provider, source)
	if errors.Is(err, ErrWorkflowAtomicGuard) {
		return observation, source, capability, s.blockWorkflowPublication(run, err)
	}

	if err == nil {
		err = s.reprepareWorkflowNativeInputs(ctx, run, source, capability, provider)
	}

	if err == nil {
		err = s.resumeWorkflowCandidateRepairGate(ctx, run, source)
	}

	if err == nil {
		err = s.checkWorkflowNativePreparationAdmission(*run, source, capability, provider)
	}

	return observation, source, capability, err
}

func (s *Store) cachedWorkflowPreparation(ctx context.Context, run workflowInvocationFile, source workflowstate.ReviewInput, capability workflowstate.PublicationCapability, provider PublicationGuardProvider) (*workflowstate.IntegrationCandidate, error) {
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
	run.Status, run.Outcome = "blocked", "publication_guard_unsupported"

	run.Blocker = &Blocker{Dependency: run.Lane.Unit, Reason: "atomic_source_base_guard_required", NextActionOwner: "workflow_operator"}
	if run.Review != nil {
		run.Review.State = "stale"
	}

	return errors.Join(reason, s.commitWorkflowState(run))
}

func (s *Store) claimWorkflowIntegrationValidation(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, source workflowstate.ReviewInput, capability workflowstate.PublicationCapability, provider PublicationGuardProvider, git PreparationGit) (claim *Claim, prepared workflowPreparedTask, input workflowstate.PreparationInput, err error) {
	claim, err = s.reserveWorkflowRole(ctx, *run, workflowIntegrationValidationPurpose)
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

func (s *Store) prepareWorkflowIntegrationGit(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, source workflowstate.ReviewInput, capability workflowstate.PublicationCapability, provider PublicationGuardProvider, git PreparationGit, lane workflowstate.Lane) (workflowstate.PreparationInput, workflowstate.Lane, error) {
	query, err := s.workflowPreparationQuery(ctx, *run, observation, source, lane)
	if err != nil {
		return workflowstate.PreparationInput{}, lane, err
	}

	query.RequiredSignatures, query.SigningProfile = capability.RequiredSignatures, capability.SigningProfile

	proof, refs, err := s.prepareWorkflowCandidateRepairGit(ctx, *run, git, query)
	if err != nil {
		return workflowstate.PreparationInput{}, lane, err
	}

	if err := checkWorkflowPreparedProof(query, proof); err != nil {
		return workflowstate.PreparationInput{}, lane, err
	}

	lane.HeadCommit = proof.ResultCommit

	input := workflowstate.PreparationInput{Source: source, Capability: capability, Git: proof, MergerAttemptRefs: refs}
	if err := s.bindWorkflowNativePreparation(ctx, *run, &input, provider); err != nil {
		return input, lane, err
	}

	if err := s.checkWorkflowPreparationGit(ctx, workflowstate.IntegrationCandidate{Input: input, Lane: lane}); err != nil {
		return input, lane, err
	}

	if err := s.checkWorkflowPreparationSource(ctx, *run, input, provider); err != nil {
		return input, lane, err
	}

	return input, lane, nil
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

	hash, err := workflowReviewInputHash(source)
	if err != nil {
		return workflowstate.Lane{}, err
	}

	run.PreparationCount++
	run.Preparation = &PreparationBinding{Round: run.PreparationCount, Unit: source.Unit, State: "preparing", SourceHash: hash}

	run.Status, run.Outcome, run.Blocker = "waiting_integration", "integration_preparing", nil
	if run.Review != nil {
		run.Review.State = "stale"
	}

	id := contenthash.Sum([]byte(fmt.Sprintf("%s:%d", run.ID, run.PreparationCount)))[:24]
	lane := workflowstate.Lane{Unit: source.Unit, Branch: "refs/heads/chr/workflow/integration-" + id, Worktree: filepath.Join(s.layout.WorktreeRoot(), "integration-"+id), BaseCommit: source.TargetCommit, HeadCommit: source.TargetCommit}

	return lane, nil
}

func (s *Store) workflowPreparationQuery(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, source workflowstate.ReviewInput, lane workflowstate.Lane) (PreparationQuery, error) {
	candidate, err := s.readWorkflowSliceCandidate(ctx, run, observation)
	if err != nil {
		return PreparationQuery{}, err
	}

	query := workflowIntegrationQuery(run, candidate, source.PullRequest, source.Policy.Method)
	query.ResultCommit = ""

	return PreparationQuery{IntegrationQuery: query, BaseCommit: source.TargetCommit, Lane: lane}, nil
}

func (s *Store) failClaimedWorkflowPreparation(run workflowInvocationFile, record Record, reason error) error {
	if err := s.checkpointWorkflowAttempt(run, record, "interrupted", nil); err != nil {
		return errors.Join(reason, err)
	}

	current, err := s.readWorkflowInvocation(run.ID)
	if err != nil {
		return errors.Join(reason, err)
	}

	current.Status, current.Outcome = "blocked", "integration_preparation_failed"
	current.Blocker = &Blocker{Dependency: run.Lane.Unit, Reason: "combined_result_preparation_failed", NextActionOwner: "workflow_operator"}

	return errors.Join(reason, s.commitWorkflowState(&current))
}
