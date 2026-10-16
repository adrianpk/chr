package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type authoringMutationJournal struct {
	Schema   string
	Changes  []workflowFileChange
	Checksum string
}

func (r *fileAuthoringSessionRepository) writeAuthoringJournal(changes []workflowFileChange) error {
	journal := authoringMutationJournal{Schema: "chr-authoring-journal/v1", Changes: changes}

	encoded, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	journal.Checksum = workflowHash(encoded)

	encoded, err = json.Marshal(journal)
	if err != nil {
		return err
	}

	if err := r.checkAuthoringJournal(journal); err != nil {
		return err
	}

	path, err := r.authoringPath(authoringRuntimeRef("pending.json"))
	if err != nil {
		return err
	}

	return atomicWorkflowFile(path, encoded)
}

func (r *fileAuthoringSessionRepository) recoverAuthoringMutation() error {
	path, err := r.authoringPath(authoringRuntimeRef("pending.json"))
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

	var journal authoringMutationJournal
	if err := decodeAuthoringState(data, &journal); err != nil {
		return err
	}

	if err := r.checkAuthoringJournal(journal); err != nil {
		return err
	}

	if err := r.replayAuthoringJournal(journal); err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return syncWorkflowDirectory(filepath.Dir(path))
}

func (r *fileAuthoringSessionRepository) checkAuthoringJournal(journal authoringMutationJournal) error {
	checksum := journal.Checksum
	journal.Checksum = ""

	encoded, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	if journal.Schema != "chr-authoring-journal/v1" || len(journal.Changes) < 2 || len(journal.Changes) > 12 || checksum != workflowHash(encoded) {
		return errors.New("invalid authoring journal")
	}

	return r.checkAuthoringJournalImages(journal.Changes)
}

func validAuthoringJournalPath(path string) bool {
	normalized := filepath.ToSlash(path)
	if strings.HasPrefix(normalized, filepath.ToSlash(authoringRuntimeRef("sessions/"))+"/") {
		return true
	}

	if strings.HasPrefix(normalized, filepath.ToSlash(authoringRuntimeRef("receipts/"))+"/") {
		return true
	}

	return validAuthoringOpsJournalPath(normalized)
}

func (r *fileAuthoringSessionRepository) authoringJournalImage(change workflowFileChange) (bool, error) {
	data, err := r.authoringRead(change.Path)
	if errors.Is(err, os.ErrNotExist) && change.Before == "" {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if workflowHash(data) == workflowHash(change.After) {
		return true, nil
	}

	if workflowHash(data) != change.Before {
		return false, ErrWorkflowConflict
	}

	return false, nil
}

func (r *fileAuthoringSessionRepository) applyAuthoringImage(change workflowFileChange, index int) error {
	applied, err := r.authoringJournalImage(change)
	if err != nil || applied {
		return err
	}

	path, err := r.authoringPath(change.Path)
	if err != nil {
		return err
	}

	if err := r.ensureAuthoringEnvelope(change.Path); err != nil {
		return err
	}

	if err := atomicWorkflowFile(path, change.After); err != nil {
		return err
	}

	if r.afterMutationReplace != nil {
		return r.afterMutationReplace(index)
	}

	return nil
}

func (r *fileAuthoringSessionRepository) ensureAuthoringEnvelope(relative string) error {
	parts := strings.Split(filepath.ToSlash(relative), "/")
	if len(parts) != 5 || parts[0] != OpsDirectoryName {
		return nil
	}

	dir := filepath.Dir(relative)
	for _, suffix := range []string{"files", "evidence", filepath.Join("extension", "chr")} {
		if _, err := r.authoringPath(filepath.Join(dir, suffix)); err != nil {
			return err
		}
	}

	path, err := r.authoringPath(dir)
	if err != nil {
		return err
	}

	return createOpsArtifactEnvelope(path, "chr")
}

func (r *fileAuthoringSessionRepository) replayAuthoringJournal(journal authoringMutationJournal) error {
	for index, change := range journal.Changes {
		if err := r.applyAuthoringImage(change, index); err != nil {
			return err
		}
	}

	for _, change := range journal.Changes {
		applied, err := r.authoringJournalImage(change)
		if err != nil || !applied {
			return errors.Join(ErrWorkflowConflict, err)
		}
	}

	return nil
}

func (r *fileAuthoringSessionRepository) checkAuthoringJournalImages(changes []workflowFileChange) error {
	seen := map[string]bool{}
	for _, change := range changes {
		if seen[change.Path] || len(change.After) > WorkflowFileLimit || !validAuthoringJournalPath(change.Path) {
			return errors.New("invalid authoring journal target")
		}

		seen[change.Path] = true
		if _, err := r.authoringJournalImage(change); err != nil {
			return err
		}
	}

	return nil
}

func validAuthoringOpsJournalPath(path string) bool {
	parts := strings.Split(path, "/")
	if len(parts) != 5 || parts[0] != OpsDirectoryName || !IsValidOpsNamespace(parts[1]) {
		return false
	}

	if parts[4] != "artifact.md" && parts[4] != "metadata.json" {
		return false
	}

	return validAuthoringContinuationKind(AuthoringArtifactKind(parts[2]))
}
