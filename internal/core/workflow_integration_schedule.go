package core

import (
	"context"
	"encoding/json"
	"slices"
)

// WorkflowIntegrationScheduleQuery bounds selection from admitted native units.
// Selection never admits later tasks, resumes a blocker or acquires an owner.
type WorkflowIntegrationScheduleQuery struct {
	Repository   string
	TargetRef    string
	Namespace    string
	InvocationID string
	Unit         WorkflowRef
	Offset       int
	Limit        int
	SnapshotHash string
}

// WorkflowIntegrationSelection binds one eligible task to its frozen unit.
// The observed revisions confer no claim; launch rechecks authoritative records.
type WorkflowIntegrationSelection struct {
	Entry              WorkflowIntegrationEntry
	InvocationRevision uint64
	Task               WorkflowPreviewMember
}

// WorkflowIntegrationSchedulePage reports queue order and bounded exclusions.
type WorkflowIntegrationSchedulePage struct {
	Items        []WorkflowIntegrationSelection
	SnapshotHash string
	NextOffset   int
	Excluded     map[string]int
}

type workflowIntegrationScheduler struct {
	store      *WorkflowStore
	snapshot   workflowIntegrationSnapshot
	selections map[string]workflowSelection
}

// QueryWorkflowIntegrationSchedule observes current admitted task frontiers.
// It verifies existing prerequisites without replaying a workflow transaction.
func (s *WorkflowStore) QueryWorkflowIntegrationSchedule(ctx context.Context, q WorkflowIntegrationScheduleQuery) (page WorkflowIntegrationSchedulePage, err error) {
	if err := validateWorkflowIntegrationQueueQuery(q.queueQuery()); err != nil {
		return page, err
	}

	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		page, err = s.workflowIntegrationSchedule(ctx, q)

		return err
	})
	if err != nil {
		page = WorkflowIntegrationSchedulePage{}
	}

	return page, integrationRecordError("integration schedule", err)
}

func (q WorkflowIntegrationScheduleQuery) queueQuery() WorkflowIntegrationQueueQuery {
	return WorkflowIntegrationQueueQuery{Repository: q.Repository, TargetRef: q.TargetRef, Namespace: q.Namespace, InvocationID: q.InvocationID, Unit: q.Unit, Offset: q.Offset, Limit: q.Limit}
}

func (s *WorkflowStore) workflowIntegrationSchedule(ctx context.Context, q WorkflowIntegrationScheduleQuery) (WorkflowIntegrationSchedulePage, error) {
	snapshot, err := s.readWorkflowIntegrationSnapshot(ctx)
	if err != nil {
		return WorkflowIntegrationSchedulePage{}, err
	}

	scheduler := workflowIntegrationScheduler{store: s, snapshot: snapshot, selections: map[string]workflowSelection{}}

	items, excluded, err := scheduler.selectTasks(ctx, q.queueQuery())
	if err != nil {
		return WorkflowIntegrationSchedulePage{}, err
	}

	hash, err := scheduler.hash()
	if err != nil || q.SnapshotHash != "" && q.SnapshotHash != hash {
		return WorkflowIntegrationSchedulePage{}, errorsOrConflict(err)
	}

	return workflowIntegrationSchedulePage(items, excluded, hash, q), nil
}

func workflowIntegrationSchedulePage(items []WorkflowIntegrationSelection, excluded map[string]int, hash string, q WorkflowIntegrationScheduleQuery) WorkflowIntegrationSchedulePage {
	start := min(q.Offset, len(items))
	end := min(start+q.Limit, len(items))

	page := WorkflowIntegrationSchedulePage{Items: items[start:end], Excluded: excluded, SnapshotHash: hash, NextOffset: -1}
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

	return workflowHash(data), err
}
