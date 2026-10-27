package core

import (
	"context"
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// All callers hold the common repository lock. A purpose selects scheduling
// policy, never a separate worker pool or credentials.
func (s *WorkflowStore) guardWorkflowIntegrationRole(ctx context.Context, run workflowInvocationFile, purpose string) error {
	if !slices.Contains([]string{"", workflowSliceValidationPurpose, workflowIntegrationValidationPurpose, workflowSliceReviewPurpose, workflowSliceRepairPurpose, workflowMergerPurpose, "closing", "set_amendment"}, purpose) {
		return workflowstate.ErrConflict
	}

	if err := s.guardWorkflowIntegrationRoleOwner(run); err != nil {
		return err
	}

	snapshot, err := s.readWorkflowIntegrationSnapshot(ctx)
	if err != nil {
		return err
	}

	roles, err := s.workflowIntegrationRoles(ctx, snapshot)
	if err != nil {
		return err
	}

	if err := guardWorkflowIntegrationPriority(roles, run, purpose); err != nil {
		return err
	}

	if purpose == workflowIntegrationValidationPurpose || purpose == workflowMergerPurpose {
		return s.guardWorkflowPreparedCapacity(ctx, snapshot, run)
	}

	return nil
}

func workflowIntegrationLowerPriority(purpose string) bool {
	return purpose == "" || purpose == workflowIntegrationValidationPurpose || purpose == workflowMergerPurpose || purpose == workflowSliceRepairPurpose
}

func (s *WorkflowStore) guardWorkflowPreparedCapacity(ctx context.Context, snapshot workflowIntegrationSnapshot, run workflowInvocationFile) error {
	limits, err := s.workflowRepositoryCapacity(ctx, snapshot.Runs)
	if err != nil {
		return err
	}

	repository, target, err := s.workflowIntegrationPreparationTarget(run)
	if err != nil {
		return err
	}

	count, retained, err := s.workflowPreparedTargetCount(snapshot, run, repository, target)
	if err != nil || retained {
		return err
	}

	if count >= limits.Prepared {
		return fmt.Errorf("%w: prepared candidate limit reached for target", ErrWorkflowCapacity)
	}

	return nil
}

func (s *WorkflowStore) workflowIntegrationPreparationTarget(run workflowInvocationFile) (string, string, error) {
	if run.Integration != nil {
		owner, err := s.workflowOwnerStore(run)
		if err != nil {
			return "", "", err
		}

		entry, err := owner.readWorkflowIntegrationEntry(run, *run.Integration)

		return entry.Repository, entry.TargetRef, err
	}

	if run.PullRequest == nil {
		return "", "", workflowstate.ErrConflict
	}

	repository, err := workflowCanonicalRepository(run.PullRequest.Request.Repository)

	return repository, "refs/heads/" + run.PullRequest.Request.BaseBranch, err
}

func workflowIntegrationPrepared(run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	if entry.PreparationReserved || entry.CandidateRef != nil || slices.Contains([]string{"preparing", "merging", "validating", "reviewing", "ready_to_publish", "publishing", "observing", "finalizing"}, entry.State) {
		return true
	}

	return run.Preparation != nil && run.Preparation.Unit == entry.Unit
}

func guardWorkflowIntegrationPriority(roles []WorkflowIntegrationRoleSelection, run workflowInvocationFile, purpose string) error {
	if !workflowIntegrationLowerPriority(purpose) {
		return nil
	}

	for _, role := range roles {
		if role.Role == "preparation" && purpose != "" || workflowIntegrationOwnMerger(role, run, purpose) {
			continue
		}

		return fmt.Errorf("%w: eligible integration %s requires %s", ErrWorkflowCapacity, role.Entry.ID, role.Role)
	}

	return nil
}

func workflowIntegrationOwnMerger(role WorkflowIntegrationRoleSelection, run workflowInvocationFile, purpose string) bool {
	return purpose == workflowMergerPurpose && role.Role == workflowMergerPurpose && role.Entry.InvocationID == run.ID && role.Entry.Namespace == run.Namespace
}

func (s *WorkflowStore) workflowPreparedTargetCount(snapshot workflowIntegrationSnapshot, run workflowInvocationFile, repository, target string) (int, bool, error) {
	count := 0

	for _, view := range snapshot.Views {
		entry := view.Entry
		if entry.Repository != repository || entry.TargetRef != target || workflowIntegrationTerminal(entry.State) {
			continue
		}

		owner, err := workflowIntegrationScheduleRun(snapshot.Runs, entry)
		if err != nil {
			return 0, false, err
		}

		if !workflowIntegrationPrepared(owner, entry) {
			continue
		}

		if workflowIntegrationOwnPreparation(owner, run, entry) {
			return count, true, nil
		}

		count++
	}

	legacy, retained, err := s.workflowLegacyPreparedCount(snapshot.Runs, run, repository, target)

	return count + legacy, retained, err
}

func workflowIntegrationOwnPreparation(owner, run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return owner.ID == run.ID && owner.OwnerRoot == run.OwnerRoot && run.Integration != nil && run.Integration.EntryID == entry.ID
}

func (s *WorkflowStore) workflowLegacyPreparedCount(runs []workflowInvocationFile, run workflowInvocationFile, repository, target string) (int, bool, error) {
	count := 0

	for _, peer := range runs {
		if peer.Integration != nil || workflowInvocationTerminal(peer.Status) || peer.Preparation == nil {
			continue
		}

		peerRepository, peerTarget, err := s.workflowIntegrationPreparationTarget(peer)
		if err != nil {
			return 0, false, err
		}

		if peerRepository != repository || peerTarget != target {
			continue
		}

		if peer.ID == run.ID && peer.OwnerRoot == run.OwnerRoot {
			return count, true, nil
		}

		count++
	}

	return count, false, nil
}

func (s *WorkflowStore) guardWorkflowIntegrationRoleOwner(run workflowInvocationFile) error {
	if run.Integration == nil {
		return nil
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if run.StopIntent != "" || entry.StopIntent != "" || !slices.Contains([]string{"queued", "preparing", "merging", "validating", "reviewing"}, entry.State) {
		return ErrWorkflowOwned
	}

	return nil
}
