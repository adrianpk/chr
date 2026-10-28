package workflow

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Store owns guarded filesystem records; SQLite is never its authority.
type Store struct {
	afterReplace        func(int) error
	afterTrackerReplace func(int) error
	layout              opscore.Layout
	projectID           string
}

// NewWorkflowStore stores dependencies without performing IO.
func NewWorkflowStore(layout opscore.Layout, projectID string) *Store {
	return &Store{layout: layout, projectID: projectID}
}

// Read observes one record under repository-wide mutation exclusion.
func (s *Store) Read(ctx context.Context, path string) (record workflowstate.Record, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		var readErr error

		record, readErr = s.readRecord(path)

		return readErr
	})

	return record, err
}

// Snapshot validates all adopted records without trusting a tracker or cache.
func (s *Store) Snapshot(ctx context.Context) (snapshot workflowstate.Snapshot, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		var readErr error

		snapshot, readErr = s.readSnapshot(ctx)

		return readErr
	})

	return snapshot, err
}

// Create allocates the existing ops envelope and attaches a workflow domain.
func (s *Store) Create(ctx context.Context, title string, record workflowstate.Record) (created workflowstate.Record, err error) {
	err = s.withLock(ctx, func() error {
		if err := s.preflightCreation(ctx, record); err != nil {
			return err
		}

		if _, err := s.workflowCoordinatorFiles(ctx); err != nil {
			return err
		}

		artifact, createErr := opscore.CreateArtifact(opscore.ArtifactOptions{Layout: s.layout, Type: record.Kind, Title: title, Producer: "chr"})
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

		record.Identity, record.Metadata = base.Identity, base.Metadata
		record.ProjectID, record.Path = s.projectID, filepath.ToSlash(relative)
		created, readErr = s.writeRecord(ctx, record, 0)

		return readErr
	})

	return created, err
}

func (s *Store) preflightCreation(ctx context.Context, r workflowstate.Record) error {
	r.Identity = workflowstate.Identity{ID: "new-workflow-artifact", ProjectID: s.projectID, Namespace: s.layout.Namespace(), Revision: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := validateWorkflowRecord(r); err != nil {
		return err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	snapshot.Records = append(snapshot.Records, r)

	return validateWorkflowSnapshot(snapshot)
}

// CreateReceipt stores new immutable evidence beneath its owning invocation.
func (s *Store) CreateReceipt(ctx context.Context, record workflowstate.Record) (created workflowstate.Record, err error) {
	if record.Kind != "receipt" || !strings.HasPrefix(record.Path, "run/") || !strings.Contains(record.Path, "/evidence/") || !strings.HasSuffix(record.Path, ".json") {
		return workflowstate.Record{}, fmt.Errorf("receipt requires owning run evidence path")
	}

	err = s.withLock(ctx, func() error {
		var createErr error

		created, createErr = s.writeRecord(ctx, record, 0)

		return createErr
	})

	return created, err
}

// Write requires the observed revision and source image; it returns the next revision.
func (s *Store) Write(ctx context.Context, record workflowstate.Record, expectedRevision uint64) (written workflowstate.Record, err error) {
	if expectedRevision == 0 {
		return workflowstate.Record{}, fmt.Errorf("use explicit workflow creation or adoption")
	}

	err = s.withLock(ctx, func() error {
		var writeErr error

		written, writeErr = s.writeRecord(ctx, record, expectedRevision)

		return writeErr
	})

	return written, err
}

func (s *Store) writeRecord(ctx context.Context, r workflowstate.Record, expected uint64) (workflowstate.Record, error) {
	if err := s.guardWrite(&r, expected); err != nil {
		return workflowstate.Record{}, err
	}

	if r.Receipt != nil {
		if err := s.checkWorkflowNativeReceipt(r); err != nil {
			return workflowstate.Record{}, err
		}
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return workflowstate.Record{}, err
	}

	if err := replaceWorkflowSnapshot(&snapshot, r); err != nil {
		return workflowstate.Record{}, err
	}

	changes, err := s.encodeChanges(r)
	if err != nil {
		return workflowstate.Record{}, err
	}

	changes, err = s.workflowLooseWriteChanges(ctx, r, expected, changes)
	if err != nil {
		return workflowstate.Record{}, err
	}

	coordinator, err := s.workflowCoordinatorFiles(ctx)
	if err != nil {
		return workflowstate.Record{}, err
	}

	changes = append(changes, coordinator...)

	if err := s.commitFiles(changes); err != nil {
		return workflowstate.Record{}, err
	}

	return s.readRecord(r.Path)
}

func (s *Store) guardWrite(r *workflowstate.Record, expected uint64) error {
	if r.ProjectID != s.projectID || r.Namespace != s.layout.Namespace() {
		return fmt.Errorf("foreign workflow write scope")
	}

	old, err := s.readRecord(r.Path)
	if expected == 0 && (errors.Is(err, fs.ErrNotExist) || errors.Is(err, workflowstate.ErrNotWorkflow)) {
		r.Revision = 1

		return validateWorkflowRecord(*r)
	}

	if err != nil {
		return err
	}

	if conflictingWorkflowImage(old, *r, expected) {
		return workflowstate.ErrConflict
	}

	if err := s.guardBoundWorkflowTaskOrder(old, *r); err != nil {
		return err
	}

	if err := guardWorkflowDomain(old, *r); err != nil {
		return err
	}

	r.Revision = expected + 1
	r.UpdatedAt = time.Now().UTC()

	return validateWorkflowRecord(*r)
}

func conflictingWorkflowImage(old, next workflowstate.Record, expected uint64) bool {
	return old.Revision != expected || old.SourceHash != next.SourceHash || old.ID != next.ID || !old.CreatedAt.Equal(next.CreatedAt)
}

func guardWorkflowDomain(old, next workflowstate.Record) error {
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

func guardWorkflowTaskChange(oldTask, nextTask *workflowstate.Task) error {
	if oldTask.Key != nextTask.Key || oldTask.ClaimGeneration != nextTask.ClaimGeneration || oldTask.AttemptID != nextTask.AttemptID {
		return workflowstate.ErrConflict
	}

	changed := oldTask.SetID != nextTask.SetID || oldTask.SliceID != nextTask.SliceID || oldTask.Order != nextTask.Order
	if changed && oldTask.AttemptID != "" {
		return fmt.Errorf("workflow membership amendment requires quiescence")
	}

	return nil
}

func replaceWorkflowSnapshot(snapshot *workflowstate.Snapshot, r workflowstate.Record) error {
	found := false

	for i, old := range snapshot.Records {
		if old.Path == r.Path {
			snapshot.Records[i], found = r, true
		}
	}

	if !found {
		snapshot.Records = append(snapshot.Records, r)
	}

	return validateWorkflowSnapshot(*snapshot)
}

func (s *Store) readSnapshot(ctx context.Context) (workflowstate.Snapshot, error) {
	snapshot := workflowstate.Snapshot{}

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
			return workflowstate.Snapshot{}, err
		}
	}

	return snapshot, validateWorkflowSnapshot(snapshot)
}

func workflowCandidatePath(path string) bool {
	return filepath.Base(path) == "metadata.json" || strings.HasSuffix(path, ".md") || strings.Contains(filepath.ToSlash(path), "/evidence/") && strings.HasSuffix(path, ".json")
}

func (s *Store) appendSnapshotRecord(snapshot *workflowstate.Snapshot, path string) error {
	relative, err := filepath.Rel(s.layout.Root(), path)
	if err != nil {
		return err
	}

	r, err := s.readRecord(filepath.ToSlash(relative))
	if errors.Is(err, workflowstate.ErrNotWorkflow) {
		return nil
	}

	if err != nil {
		return err
	}

	if r.ProjectID != s.projectID || r.Namespace != s.layout.Namespace() {
		return fmt.Errorf("foreign workflow record scope")
	}

	snapshot.Records = append(snapshot.Records, r)
	if len(snapshot.Records) > workflowstate.RecordLimit {
		return fmt.Errorf("workflow namespace record limit exceeded")
	}

	return nil
}
