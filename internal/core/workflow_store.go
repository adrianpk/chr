package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

// ErrWorkflowConflict requires fresh observation or explicit reconciliation.
var ErrWorkflowConflict = errors.New("workflow revision or image conflict")

// ErrNotWorkflow excludes unadopted documents from executable state.
var ErrNotWorkflow = errors.New("document is not an adopted workflow record")

// WorkflowStore owns guarded filesystem records; SQLite is never its authority.
type WorkflowStore struct {
	afterReplace func(int) error
	layout       OpsLayout
	projectID    string
}

// NewWorkflowStore stores dependencies without performing IO.
func NewWorkflowStore(layout OpsLayout, projectID string) *WorkflowStore {
	return &WorkflowStore{layout: layout, projectID: projectID}
}

// Read observes one record under repository-wide mutation exclusion.
func (s *WorkflowStore) Read(ctx context.Context, path string) (record WorkflowRecord, err error) {
	err = s.withLock(ctx, func() error {
		var readErr error

		record, readErr = s.readRecord(path)

		return readErr
	})

	return record, err
}

// Snapshot validates all adopted records without trusting a tracker or cache.
func (s *WorkflowStore) Snapshot(ctx context.Context) (snapshot WorkflowSnapshot, err error) {
	err = s.withLock(ctx, func() error {
		var readErr error

		snapshot, readErr = s.readSnapshot(ctx)

		return readErr
	})

	return snapshot, err
}

// Create allocates the existing ops envelope and attaches a workflow domain.
func (s *WorkflowStore) Create(ctx context.Context, title string, record WorkflowRecord) (created WorkflowRecord, err error) {
	err = s.withLock(ctx, func() error {
		if err := s.preflightCreation(ctx, record); err != nil {
			return err
		}

		artifact, createErr := CreateOpsArtifact(CreateOpsArtifactOptions{Layout: s.layout, Type: record.Kind, Title: title, Producer: "chr"})
		if createErr != nil {
			return createErr
		}

		relative, relErr := filepath.Rel(s.layout.Root(), filepath.Join(artifact.Dir, "metadata.json"))
		if relErr != nil {
			return relErr
		}

		base, readErr := readWorkflowEnvelope(filepath.Join(artifact.Dir, "metadata.json"))
		if readErr != nil {
			return readErr
		}

		record.WorkflowIdentity, record.Metadata = base.WorkflowIdentity, base.Metadata
		record.ProjectID, record.Path = s.projectID, filepath.ToSlash(relative)
		created, readErr = s.writeRecord(ctx, record, 0)

		return readErr
	})

	return created, err
}

func (s *WorkflowStore) preflightCreation(ctx context.Context, r WorkflowRecord) error {
	r.WorkflowIdentity = WorkflowIdentity{ID: "new-workflow-artifact", ProjectID: s.projectID, Namespace: s.layout.Namespace(), Revision: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := r.Validate(); err != nil {
		return err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	snapshot.Records = append(snapshot.Records, r)

	return snapshot.Validate()
}

// CreateReceipt stores new immutable evidence beneath its owning invocation.
func (s *WorkflowStore) CreateReceipt(ctx context.Context, record WorkflowRecord) (created WorkflowRecord, err error) {
	if record.Kind != "receipt" || !strings.HasPrefix(record.Path, "run/") || !strings.Contains(record.Path, "/evidence/") || !strings.HasSuffix(record.Path, ".json") {
		return WorkflowRecord{}, fmt.Errorf("receipt requires owning run evidence path")
	}

	err = s.withLock(ctx, func() error {
		var createErr error

		created, createErr = s.writeRecord(ctx, record, 0)

		return createErr
	})

	return created, err
}

// Write requires the observed revision and source image; it returns the next revision.
func (s *WorkflowStore) Write(ctx context.Context, record WorkflowRecord, expectedRevision uint64) (written WorkflowRecord, err error) {
	if expectedRevision == 0 {
		return WorkflowRecord{}, fmt.Errorf("use explicit workflow creation or adoption")
	}

	err = s.withLock(ctx, func() error {
		var writeErr error

		written, writeErr = s.writeRecord(ctx, record, expectedRevision)

		return writeErr
	})

	return written, err
}

func (s *WorkflowStore) writeRecord(ctx context.Context, r WorkflowRecord, expected uint64) (WorkflowRecord, error) {
	if err := s.guardWrite(&r, expected); err != nil {
		return WorkflowRecord{}, err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if err := replaceWorkflowSnapshot(&snapshot, r); err != nil {
		return WorkflowRecord{}, err
	}

	changes, err := s.encodeChanges(r)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if err := s.commitFiles(changes); err != nil {
		return WorkflowRecord{}, err
	}

	return s.readRecord(r.Path)
}

func (s *WorkflowStore) guardWrite(r *WorkflowRecord, expected uint64) error {
	if r.ProjectID != s.projectID || r.Namespace != s.layout.Namespace() {
		return fmt.Errorf("foreign workflow write scope")
	}

	old, err := s.readRecord(r.Path)
	if expected == 0 && (errors.Is(err, fs.ErrNotExist) || errors.Is(err, ErrNotWorkflow)) {
		r.Revision = 1

		return r.Validate()
	}

	if err != nil {
		return err
	}

	if conflictingWorkflowImage(old, *r, expected) {
		return ErrWorkflowConflict
	}

	if err := guardWorkflowDomain(old, *r); err != nil {
		return err
	}

	r.Revision = expected + 1
	r.UpdatedAt = time.Now().UTC()

	return r.Validate()
}

func conflictingWorkflowImage(old, next WorkflowRecord, expected uint64) bool {
	return old.Revision != expected || old.SourceHash != next.SourceHash || old.ID != next.ID || !old.CreatedAt.Equal(next.CreatedAt)
}

func guardWorkflowDomain(old, next WorkflowRecord) error {
	if old.Kind != next.Kind || old.Receipt != nil {
		return fmt.Errorf("workflow domain kind and receipts are immutable")
	}

	if old.Set != nil && next.Set != nil {
		return guardWorkflowSetChange(*old.Set, *next.Set)
	}

	if old.Task != nil && next.Task != nil {
		if old.Task.AttemptID != "" {
			return fmt.Errorf("%w: owned task requires a fenced operation", ErrWorkflowOwned)
		}

		return guardWorkflowTaskChange(old.Task, next.Task)
	}

	return nil
}

func guardWorkflowTaskChange(oldTask, nextTask *WorkflowTask) error {
	if oldTask.Key != nextTask.Key || oldTask.ClaimGeneration != nextTask.ClaimGeneration || oldTask.AttemptID != nextTask.AttemptID {
		return ErrWorkflowConflict
	}

	changed := oldTask.SetID != nextTask.SetID || oldTask.SliceID != nextTask.SliceID || oldTask.Order != nextTask.Order
	if changed && oldTask.AttemptID != "" {
		return fmt.Errorf("workflow membership amendment requires quiescence")
	}

	return nil
}

func replaceWorkflowSnapshot(snapshot *WorkflowSnapshot, r WorkflowRecord) error {
	found := false

	for i, old := range snapshot.Records {
		if old.Path == r.Path {
			snapshot.Records[i], found = r, true
		}
	}

	if !found {
		snapshot.Records = append(snapshot.Records, r)
	}

	return snapshot.Validate()
}

func (s *WorkflowStore) readSnapshot(ctx context.Context) (WorkflowSnapshot, error) {
	snapshot := WorkflowSnapshot{}

	for _, kind := range []string{"ticket", "plan", "run"} {
		err := filepath.WalkDir(filepath.Join(s.layout.Root(), kind), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if err := ctx.Err(); err != nil {
				return err
			}

			if entry.IsDir() || !workflowCandidatePath(path) {
				return nil
			}

			return s.appendSnapshotRecord(&snapshot, path)
		})
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return WorkflowSnapshot{}, err
		}
	}

	return snapshot, snapshot.Validate()
}

func workflowCandidatePath(path string) bool {
	return filepath.Base(path) == "metadata.json" || strings.HasSuffix(path, ".md") || strings.Contains(filepath.ToSlash(path), "/evidence/") && strings.HasSuffix(path, ".json")
}

func (s *WorkflowStore) appendSnapshotRecord(snapshot *WorkflowSnapshot, path string) error {
	relative, err := filepath.Rel(s.layout.Root(), path)
	if err != nil {
		return err
	}

	r, err := s.readRecord(filepath.ToSlash(relative))
	if errors.Is(err, ErrNotWorkflow) {
		return nil
	}

	if err != nil {
		return err
	}

	if r.ProjectID != s.projectID || r.Namespace != s.layout.Namespace() {
		return fmt.Errorf("foreign workflow record scope")
	}

	snapshot.Records = append(snapshot.Records, r)
	if len(snapshot.Records) > WorkflowRecordLimit {
		return fmt.Errorf("workflow namespace record limit exceeded")
	}

	return nil
}
