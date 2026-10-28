package workflow

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowTrackerSnapshot(ctx context.Context, setID, schema string) (Snapshot, error) {
	if schema == "" {
		var err error

		schema, err = s.inferWorkflowTrackerSchema(setID)
		if err != nil {
			return Snapshot{}, err
		}
	}

	if schema != "chr-tracker/v2" {
		return s.readSnapshot(ctx)
	}

	return s.readWorkflowOrderSources(ctx, setID)
}

// Order-only inspection uses native set/task readers under the existing lock.
// Receipt contents and lifecycle evidence do not participate in map freshness.
func (s *Store) readWorkflowOrderSources(ctx context.Context, setID string) (Snapshot, error) {
	set, err := s.readWorkflowOrderSet(setID)
	snapshot := Snapshot{Records: []Record{set}}

	if err != nil {
		return Snapshot{}, err
	}

	err = s.walkWorkflowOrderTasks(ctx, &snapshot, setID)

	return snapshot, err
}

func (s *Store) readWorkflowOrderSet(setID string) (Record, error) {
	if _, err := workflowstate.TrackerMetadataPath(setID); err != nil {
		return Record{}, err
	}

	set, err := s.readRecord("plan/" + setID + "/metadata.json")
	if err != nil {
		return set, err
	}

	if set.ID != setID || set.Set == nil || set.ProjectID != s.projectID || set.Namespace != s.layout.Namespace() {
		return set, workflowstate.ErrConflict
	}

	return set, nil
}

func (s *Store) walkWorkflowOrderTasks(ctx context.Context, snapshot *Snapshot, setID string) error {
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

func (s *Store) appendWorkflowOrderTask(snapshot *Snapshot, path, setID string) error {
	reference, err := filepath.Rel(s.layout.Root(), path)
	if err != nil {
		return err
	}

	record, err := s.readRecord(filepath.ToSlash(reference))
	if errors.Is(err, workflowstate.ErrNotWorkflow) {
		return nil
	}

	if err != nil {
		return err
	}

	if record.Task == nil || record.Task.SetID != setID {
		return nil
	}

	if record.ProjectID != s.projectID || record.Namespace != s.layout.Namespace() {
		return workflowstate.ErrConflict
	}

	if len(snapshot.Records) >= workflowstate.RecordLimit {
		return fmt.Errorf("workflow order source bound exceeded")
	}

	snapshot.Records = append(snapshot.Records, record)

	return nil
}

func (s *Store) inferWorkflowTrackerSchema(setID string) (string, error) {
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
