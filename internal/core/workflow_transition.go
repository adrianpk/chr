package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const workflowJournalPath = "lock/workflow-transition.json"

// WorkflowChange binds a replacement record to its observed revision and image.
type WorkflowChange struct {
	Record           WorkflowRecord
	ExpectedRevision uint64
}

type workflowJournal struct {
	Schema  string               `json:"workflow_schema"`
	ID      string               `json:"id"`
	Changes []workflowFileChange `json:"changes"`
}

// Transition validates a complete namespace update before journaling its images.
func (s *WorkflowStore) Transition(ctx context.Context, changes []WorkflowChange) error {
	return s.withLock(ctx, func() error {
		if len(changes) == 0 || len(changes) > WorkflowRecordLimit {
			return fmt.Errorf("workflow transition record bound exceeded")
		}

		snapshot, err := s.readSnapshot(ctx)
		if err != nil {
			return err
		}

		files, err := s.prepareTransition(&snapshot, changes)
		if err != nil {
			return err
		}

		if err := snapshot.Validate(); err != nil {
			return err
		}

		return s.commitFiles(files)
	})
}

func (s *WorkflowStore) prepareTransition(snapshot *WorkflowSnapshot, changes []WorkflowChange) ([]workflowFileChange, error) {
	files := []workflowFileChange{}
	seen := map[string]bool{}

	for _, change := range changes {
		r := change.Record
		if change.ExpectedRevision == 0 || seen[r.Path] {
			return nil, fmt.Errorf("workflow transition requires unique adopted records")
		}

		seen[r.Path] = true

		if err := s.guardWrite(&r, change.ExpectedRevision); err != nil {
			return nil, err
		}

		for i, old := range snapshot.Records {
			if old.Path == r.Path {
				snapshot.Records[i] = r
			}
		}

		encoded, err := s.encodeChanges(r)
		if err != nil {
			return nil, err
		}

		files = append(files, encoded...)
	}

	return files, nil
}

func (s *WorkflowStore) commitFiles(changes []workflowFileChange) error {
	data, err := json.Marshal(changes)
	if err != nil {
		return err
	}

	journal := workflowJournal{Schema: WorkflowSchema, ID: workflowHash(data), Changes: changes}

	encoded, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	if len(encoded) > WorkflowFileLimit {
		return fmt.Errorf("workflow transition journal exceeds 1 MiB")
	}

	if err := s.checkJournalImages(journal); err != nil {
		return err
	}

	path, err := s.safePath(workflowJournalPath)
	if err != nil {
		return err
	}

	if err := atomicWorkflowFile(path, encoded); err != nil {
		return err
	}

	return s.replayJournal(journal)
}

func (s *WorkflowStore) recoverTransition() error {
	path, err := s.safePath(workflowJournalPath)
	if err != nil {
		return err
	}

	data, err := readWorkflowFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	journal := workflowJournal{}
	if err := json.Unmarshal(data, &journal); err != nil {
		return fmt.Errorf("decode workflow transition journal: %w", err)
	}

	return s.replayJournal(journal)
}

func (s *WorkflowStore) checkJournalImages(journal workflowJournal) error {
	data, err := json.Marshal(journal.Changes)
	if err != nil {
		return err
	}

	if journal.Schema != WorkflowSchema || workflowHash(data) != journal.ID || len(journal.Changes) == 0 {
		return fmt.Errorf("invalid workflow transition journal")
	}

	seen := map[string]bool{}
	for _, change := range journal.Changes {
		if seen[change.Path] || len(change.After) > WorkflowFileLimit {
			return fmt.Errorf("invalid workflow journal replacement")
		}

		seen[change.Path] = true
		if _, err := s.journalImage(change); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowStore) journalImage(change workflowFileChange) (bool, error) {
	path, err := s.safePath(change.Path)
	if err != nil {
		return false, err
	}

	data, err := readWorkflowFile(path)
	if errors.Is(err, os.ErrNotExist) && change.Before == "" {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	hash := workflowHash(data)
	if hash == workflowHash(change.After) {
		return true, nil
	}

	if hash != change.Before {
		return false, fmt.Errorf("%w: %s", ErrWorkflowConflict, change.Path)
	}

	return false, nil
}

func (s *WorkflowStore) replayJournal(journal workflowJournal) error {
	if err := s.checkJournalImages(journal); err != nil {
		return err
	}

	for i, change := range journal.Changes {
		alreadyApplied, err := s.journalImage(change)
		if err != nil {
			return err
		}

		if alreadyApplied {
			continue
		}

		if err := s.applyJournalImage(change, i); err != nil {
			return err
		}
	}

	if err := s.confirmJournalImages(journal); err != nil {
		return err
	}

	path, err := s.safePath(workflowJournalPath)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return syncWorkflowDirectory(s.layout.Root() + "/lock")
}

func (s *WorkflowStore) confirmJournalImages(journal workflowJournal) error {
	for _, change := range journal.Changes {
		applied, err := s.journalImage(change)
		if err != nil {
			return err
		}

		if !applied {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func (s *WorkflowStore) applyJournalImage(change workflowFileChange, index int) error {
	path, err := s.safePath(change.Path)
	if err != nil {
		return err
	}

	if err := atomicWorkflowFile(path, change.After); err != nil {
		return err
	}

	if s.afterReplace != nil {
		return s.afterReplace(index)
	}

	return nil
}
