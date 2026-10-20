package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
)

func (s *WorkflowStore) workflowTrackerSnapshot(ctx context.Context, setID, schema string) (WorkflowSnapshot, error) {
	if schema == "" {
		var err error

		schema, err = s.inferWorkflowTrackerSchema(setID)
		if err != nil {
			return WorkflowSnapshot{}, err
		}
	}

	if schema != "chr-tracker/v2" {
		return s.readSnapshot(ctx)
	}

	return s.readWorkflowOrderSources(ctx, setID)
}

// Order-only inspection uses native set/task readers under the existing lock.
// Receipt contents and lifecycle evidence do not participate in map freshness.
func (s *WorkflowStore) readWorkflowOrderSources(ctx context.Context, setID string) (WorkflowSnapshot, error) {
	set, err := s.readWorkflowOrderSet(setID)
	snapshot := WorkflowSnapshot{Records: []WorkflowRecord{set}}

	if err != nil {
		return WorkflowSnapshot{}, err
	}

	err = s.walkWorkflowOrderTasks(ctx, &snapshot, setID)

	return snapshot, err
}

func (s *WorkflowStore) readWorkflowOrderSet(setID string) (WorkflowRecord, error) {
	if _, err := workflowTrackerMetadataPath(setID); err != nil {
		return WorkflowRecord{}, err
	}

	set, err := s.readRecord("plan/" + setID + "/metadata.json")
	if err != nil {
		return set, err
	}

	if set.ID != setID || set.Set == nil || set.ProjectID != s.projectID || set.Namespace != s.layout.Namespace() {
		return set, ErrWorkflowConflict
	}

	return set, nil
}

func (s *WorkflowStore) walkWorkflowOrderTasks(ctx context.Context, snapshot *WorkflowSnapshot, setID string) error {
	root := filepath.Join(s.layout.Root(), "ticket")

	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if path == root && errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}

			return walkErr
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() || !workflowCandidatePath(path) {
			return nil
		}

		return s.appendWorkflowOrderTask(snapshot, path, setID)
	})
}

func (s *WorkflowStore) appendWorkflowOrderTask(snapshot *WorkflowSnapshot, path, setID string) error {
	reference, err := filepath.Rel(s.layout.Root(), path)
	if err != nil {
		return err
	}

	record, err := s.readRecord(filepath.ToSlash(reference))
	if errors.Is(err, ErrNotWorkflow) {
		return nil
	}

	if err != nil {
		return err
	}

	if record.Task == nil || record.Task.SetID != setID {
		return nil
	}

	if record.ProjectID != s.projectID || record.Namespace != s.layout.Namespace() {
		return ErrWorkflowConflict
	}

	if len(snapshot.Records) >= WorkflowRecordLimit {
		return fmt.Errorf("workflow order source bound exceeded")
	}

	snapshot.Records = append(snapshot.Records, record)

	return nil
}

func (s *WorkflowStore) inferWorkflowTrackerSchema(setID string) (string, error) {
	previous, err := s.readWorkflowTracker(setID)
	if err == nil {
		return previous.Schema, nil
	}

	if !errors.Is(err, errWorkflowTrackerAbsent) {
		return "", err
	}

	set, err := s.readWorkflowOrderSet(setID)
	if err == nil && set.Set.PlanRevision != nil {
		return "chr-tracker/v2", nil
	}
	// A legacy set can use a different retained path. Preserve its original
	// namespace reader and validation rather than inventing a migration.
	return "chr-tracker/v1", nil
}
