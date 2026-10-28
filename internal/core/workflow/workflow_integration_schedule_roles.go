package workflow

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowMergerPurpose = "merger"

// IntegrationRoleSelection names an eligible local continuation. It is
// observational; the existing role owner must recheck inputs and acquire a slot.
type IntegrationRoleSelection struct {
	Entry              IntegrationEntry
	InvocationRevision uint64
	Role               string
}

// IntegrationRolePage lists finishable roles ahead of new preparation.
// It never contacts a forge, runs a gate, resumes a human wait or grants a claim.
type IntegrationRolePage struct {
	Items        []IntegrationRoleSelection
	SnapshotHash string
	NextOffset   int
}

// QueryWorkflowIntegrationRoles observes admitted integration continuations.
// Provider and current-policy checks still belong to each existing launch owner.
func (s *Store) QueryWorkflowIntegrationRoles(ctx context.Context, q IntegrationScheduleQuery) (page IntegrationRolePage, err error) {
	if err := validateWorkflowIntegrationQueueQuery(q.queueQuery()); err != nil {
		return page, err
	}

	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		snapshot, readErr := s.readWorkflowIntegrationSnapshot(ctx)
		if readErr != nil {
			return readErr
		}

		items, readErr := s.workflowIntegrationRoles(ctx, snapshot)
		if readErr != nil {
			return readErr
		}

		limits, readErr := s.workflowRepositoryCapacity(ctx, snapshot.Runs)
		if readErr != nil {
			return readErr
		}

		data, readErr := json.Marshal(struct {
			Snapshot string
			Items    []IntegrationRoleSelection
			Limits   workflowCapacityLimits
		}{snapshot.Hash, items, limits})
		if readErr != nil {
			return readErr
		}

		hash := contenthash.Sum(data)
		if q.SnapshotHash != "" && q.SnapshotHash != hash {
			return workflowstate.ErrConflict
		}

		items = slices.DeleteFunc(items, func(item IntegrationRoleSelection) bool {
			return !workflowIntegrationQueryMatches(item.Entry, q.queueQuery())
		})
		start := min(q.Offset, len(items))
		end := min(start+q.Limit, len(items))

		page = IntegrationRolePage{Items: items[start:end], SnapshotHash: hash, NextOffset: -1}
		if end < len(items) {
			page.NextOffset = end
		}

		return nil
	})
	if err != nil {
		page = IntegrationRolePage{}
	}

	return page, integrationRecordError("integration roles", err)
}

func (s *Store) workflowIntegrationRoles(ctx context.Context, snapshot workflowIntegrationSnapshot) ([]IntegrationRoleSelection, error) {
	items := []IntegrationRoleSelection{}

	for _, view := range snapshot.Views {
		run, err := workflowIntegrationScheduleRun(snapshot.Runs, view.Entry)
		if err != nil {
			return nil, err
		}

		if !workflowIntegrationRoleReady(run, view.Entry) || workflowIntegrationTargetBlocked(snapshot, view.Entry) {
			continue
		}

		owner, err := s.workflowOwnerStore(run)
		if err != nil {
			return nil, err
		}

		role, err := owner.workflowIntegrationAvailableRole(ctx, snapshot, run, view.Entry)
		if err != nil {
			return nil, integrationRecordError(workflowInvocationPath(run.ID), err)
		}

		if role != "" {
			items = append(items, IntegrationRoleSelection{Entry: view.Entry, InvocationRevision: run.Revision, Role: role})
		}
	}
	// Stable sorting retains admission sequence and ID within each priority class.
	slices.SortStableFunc(items, func(a, b IntegrationRoleSelection) int {
		return cmp.Compare(workflowIntegrationRoleRank(a.Role), workflowIntegrationRoleRank(b.Role))
	})

	return items, nil
}

func workflowIntegrationRoleReady(run workflowInvocationFile, entry IntegrationEntry) bool {
	restoreWorkflowCapacityWait(&run)

	if run.Integration == nil || run.Integration.EntryID != entry.ID || run.StopIntent != "" || entry.StopIntent != "" {
		return false
	}

	if !slices.Contains([]IntegrationState{"queued", "preparing", "merging", "validating", "reviewing"}, entry.State) || !slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationWaitingIntegration, workflowstate.InvocationWaitingReview}, run.Status) {
		return false
	}

	return workflowIntegrationFinishedFrontier(run, entry)
}

func (s *Store) workflowIntegrationNextRole(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry) (string, error) {
	observation, err := s.observeWorkflowSlicePrepared(ctx, run, (*Store).requireNoWorkflowTransition)
	if err != nil {
		return "", err
	}

	if !workflowTaskClaimable(observation.tasks[len(observation.tasks)-1].Task) {
		return "", nil
	}

	if run.CandidateRef == "" {
		if entry.Unit.Kind == "slice" {
			return "validation", nil
		}

		return "", nil
	}

	if _, err := s.readWorkflowSliceCandidate(ctx, run, observation); err != nil {
		return "", err
	}

	if run.PullRequest == nil || run.PullRequest.State != "bound" {
		return "", nil
	}

	if entry.State == "merging" {
		return workflowMergerPurpose, nil
	}

	return s.workflowIntegrationPreparedRole(ctx, run)
}

func (s *Store) workflowIntegrationPreparedRole(ctx context.Context, run workflowInvocationFile) (string, error) {
	if run.Preparation == nil || run.Preparation.State == "stale" {
		return "preparation", nil
	}

	if run.Preparation.State != "validated" {
		return "", nil
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil {
		return "", err
	}

	if err := s.checkWorkflowPreparationGit(ctx, candidate); err != nil {
		return "", err
	}

	policy, err := s.readWorkflowReviewPolicy(run)
	if err != nil {
		return "", err
	}

	if policy.Mode != workflowstate.ReviewAgentic {
		return "", nil
	}

	if run.Review != nil && !slices.Contains([]workflowstate.ReviewState{workflowstate.ReviewPending, workflowstate.ReviewStale}, run.Review.State) {
		return "", nil
	}

	return "reviewer", nil
}

func workflowIntegrationRoleRank(role string) int {
	if role == "preparation" {
		return 1
	}

	return 0
}

func workflowIntegrationTargetBlocked(snapshot workflowIntegrationSnapshot, entry IntegrationEntry) bool {
	target := contenthash.Sum([]byte(entry.Repository + "\n" + entry.TargetRef))

	return slices.ContainsFunc(snapshot.Journals, func(j workflowSubmissionJournal) bool {
		return j.TargetKey == target && j.Attempts[len(j.Attempts)-1].State == workflowstate.SubmissionAwaitingObservation
	})
}

func workflowIntegrationFinishedFrontier(run workflowInvocationFile, entry IntegrationEntry) bool {
	if run.Lane == nil || run.Lane.Unit != entry.Unit || run.Cursor <= run.ConfirmedCursor {
		return false
	}

	return run.Cursor == len(run.Members) || !workflowMemberOwnsUnit(run.Members[run.Cursor], entry.Unit)
}

func (s *Store) workflowIntegrationAvailableRole(ctx context.Context, snapshot workflowIntegrationSnapshot, run workflowInvocationFile, entry IntegrationEntry) (string, error) {
	role, err := s.workflowIntegrationNextRole(ctx, run, entry)
	if err != nil || role != "preparation" {
		return role, err
	}

	err = s.guardWorkflowPreparedCapacity(ctx, snapshot, run)
	if errors.Is(err, ErrWorkflowCapacity) {
		return "", nil
	}

	return role, err
}
