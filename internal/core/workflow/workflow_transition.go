package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Change binds a replacement record to its observed revision and image.
type Change struct {
	Record           workflowstate.Record
	ExpectedRevision uint64
}

type workflowJournal struct {
	AllocationGuard string                     `json:"allocation_guard,omitempty"`
	Schema          string                     `json:"workflow_schema"`
	ID              string                     `json:"id"`
	Changes         []workflowstate.FileChange `json:"changes"`
}

// Transition validates a complete namespace update before journaling its images.
func (s *Store) Transition(ctx context.Context, changes []Change) error {
	return s.withLock(ctx, func() error {
		if len(changes) == 0 || len(changes) > workflowstate.RecordLimit {
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

		if err := validateWorkflowSnapshot(snapshot); err != nil {
			return err
		}

		return s.commitFiles(files)
	})
}

func (s *Store) prepareTransition(snapshot *workflowstate.Snapshot, changes []Change) ([]workflowstate.FileChange, error) {
	files := []workflowstate.FileChange{}
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

func (s *Store) commitFiles(changes []workflowstate.FileChange) error {
	return s.commitAllocatedFiles(changes, "")
}

func (s *Store) commitAllocatedFiles(changes []workflowstate.FileChange, guard string) error {
	data, err := json.Marshal(changes)
	if err != nil {
		return err
	}

	journal := workflowJournal{Schema: workflowstate.WorkflowSchema, ID: contenthash.Sum(append([]byte(guard), data...)), Changes: changes, AllocationGuard: guard}

	encoded, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	if len(encoded) > workflowstate.FileLimit {
		return fmt.Errorf("workflow transition journal exceeds 1 MiB")
	}

	if err := s.checkJournalImages(journal); err != nil {
		return err
	}

	path, err := s.safePath(workflowstate.JournalPath)
	if err != nil {
		return err
	}

	if err := opsfile.WriteAtomic(path, encoded); err != nil {
		return err
	}

	if err := s.replayJournal(journal); err != nil {
		return err
	}

	s.markWorkflowTrackersStale(changes)

	return nil
}

func (s *Store) recoverTransition() error {
	path, err := s.safePath(workflowstate.JournalPath)
	if err != nil {
		return err
	}

	data, err := opsfile.Read(path)
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

func (s *Store) checkJournalImages(journal workflowJournal) error {
	data, err := json.Marshal(journal.Changes)
	if err != nil {
		return err
	}

	if journal.Schema != workflowstate.WorkflowSchema || contenthash.Sum(append([]byte(journal.AllocationGuard), data...)) != journal.ID || len(journal.Changes) == 0 {
		return fmt.Errorf("invalid workflow transition journal")
	}

	seen := map[string]bool{}
	for _, change := range journal.Changes {
		if seen[change.Path] || len(change.After) > workflowstate.FileLimit {
			return fmt.Errorf("invalid workflow journal replacement")
		}

		seen[change.Path] = true
		if _, err := s.journalImage(change); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) journalImage(change workflowstate.FileChange) (bool, error) {
	path, err := s.safePath(change.Path)
	if err != nil {
		return false, err
	}

	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) && change.Before == "" {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	hash := contenthash.Sum(data)
	if hash == contenthash.Sum(change.After) {
		return true, nil
	}

	if hash != change.Before {
		return false, fmt.Errorf("%w: %s", workflowstate.ErrConflict, change.Path)
	}

	return false, nil
}

func (s *Store) replayJournal(journal workflowJournal) error {
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

	if err := s.releaseWorkflowAllocation(journal.AllocationGuard); err != nil {
		return err
	}

	path, err := s.safePath(workflowstate.JournalPath)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return opsfile.SyncDirectory(s.layout.Root() + "/lock")
}

func (s *Store) confirmJournalImages(journal workflowJournal) error {
	for _, change := range journal.Changes {
		applied, err := s.journalImage(change)
		if err != nil {
			return err
		}

		if !applied {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func (s *Store) applyJournalImage(change workflowstate.FileChange, index int) error {
	path, err := s.safePath(change.Path)
	if err != nil {
		return err
	}

	if err := opsfile.WriteAtomic(path, change.After); err != nil {
		return err
	}

	if s.afterReplace != nil {
		return s.afterReplace(index)
	}

	return nil
}
