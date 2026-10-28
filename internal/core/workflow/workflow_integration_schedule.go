package workflow

import (
	"context"
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationScheduleQuery bounds selection from admitted native units.
// Selection never admits later tasks, resumes a blocker or acquires an owner.
type IntegrationScheduleQuery struct {
	Repository   string
	TargetRef    string
	Namespace    string
	InvocationID string
	Unit         workflowstate.Ref
	Offset       int
	Limit        int
	SnapshotHash string
}

// IntegrationSelection binds one eligible task to its frozen unit.
// The observed revisions confer no claim; launch rechecks authoritative records.
type IntegrationSelection struct {
	Entry              IntegrationEntry
	InvocationRevision uint64
	Task               PreviewMember
}

// IntegrationSchedulePage reports queue order and bounded exclusions.
type IntegrationSchedulePage struct {
	Items        []IntegrationSelection
	SnapshotHash string
	NextOffset   int
	Excluded     map[string]int
}

type workflowIntegrationScheduler struct {
	store      *Store
	snapshot   workflowIntegrationSnapshot
	selections map[string]workflowSelection
}

// QueryWorkflowIntegrationSchedule observes current admitted task frontiers.
// It verifies existing prerequisites without replaying a workflow transaction.
func (s *Store) QueryWorkflowIntegrationSchedule(ctx context.Context, q IntegrationScheduleQuery) (page IntegrationSchedulePage, err error) {
	if err := validateWorkflowIntegrationQueueQuery(q.queueQuery()); err != nil {
		return page, err
	}

	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		page, err = s.workflowIntegrationSchedule(ctx, q)

		return err
	})
	if err != nil {
		page = IntegrationSchedulePage{}
	}

	return page, integrationRecordError("integration schedule", err)
}

func (q IntegrationScheduleQuery) queueQuery() IntegrationQueueQuery {
	return IntegrationQueueQuery{Repository: q.Repository, TargetRef: q.TargetRef, Namespace: q.Namespace, InvocationID: q.InvocationID, Unit: q.Unit, Offset: q.Offset, Limit: q.Limit}
}

func (s *Store) workflowIntegrationSchedule(ctx context.Context, q IntegrationScheduleQuery) (IntegrationSchedulePage, error) {
	snapshot, err := s.readWorkflowIntegrationSnapshot(ctx)
	if err != nil {
		return IntegrationSchedulePage{}, err
	}

	scheduler := workflowIntegrationScheduler{store: s, snapshot: snapshot, selections: map[string]workflowSelection{}}

	taskSelections, err := scheduler.selectTasks(ctx, q.queueQuery())

	items, excluded := taskSelections.Items, taskSelections.Excluded

	if err != nil {
		return IntegrationSchedulePage{}, err
	}

	hash, err := scheduler.hash()
	if err != nil || q.SnapshotHash != "" && q.SnapshotHash != hash {
		return IntegrationSchedulePage{}, errorsOrConflict(err)
	}

	return workflowIntegrationSchedulePage(items, excluded, hash, q), nil
}

func workflowIntegrationSchedulePage(items []IntegrationSelection, excluded map[string]int, hash string, q IntegrationScheduleQuery) IntegrationSchedulePage {
	start := min(q.Offset, len(items))
	end := min(start+q.Limit, len(items))

	page := IntegrationSchedulePage{Items: items[start:end], Excluded: excluded, SnapshotHash: hash, NextOffset: -1}
	if end < len(items) {
		page.NextOffset = end
	}

	return page
}

func (scheduler *workflowIntegrationScheduler) hash() (string, error) {
	images := []string{scheduler.snapshot.Hash}
	for owner, selection := range scheduler.selections {
		images = append(images, owner+"\x00"+selection.hash())
	}

	slices.Sort(images)

	data, err := json.Marshal(images)

	return contenthash.Sum(data), err
}
