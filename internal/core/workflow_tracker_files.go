package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var errWorkflowTrackerAbsent = errors.New("tracker is not registered")

func workflowTrackerMetadataPath(setID string) (string, error) {
	if setID == "" || setID == "." || setID == ".." || strings.ContainsAny(setID, "/\\") {
		return "", fmt.Errorf("invalid tracker set identity")
	}

	return "tracker/projections/" + setID + ".json", nil
}

func (s *WorkflowStore) readWorkflowTrackerFile(relative string) ([]byte, error) {
	path, err := s.safePath(relative)
	if err != nil {
		return nil, err
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("tracker source/output must be a regular file")
	}

	return readWorkflowFile(path)
}

func (s *WorkflowStore) workflowTrackerOutput(setID string) (string, error) {
	if _, err := workflowTrackerMetadataPath(setID); err != nil {
		return "", err
	}

	data, err := s.readWorkflowTrackerFile("tracker/" + setID + ".md")
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return workflowHash(data), nil
}

func (s *WorkflowStore) readWorkflowTracker(setID string) (WorkflowTrackerStatus, error) {
	relative, err := workflowTrackerMetadataPath(setID)
	if err != nil {
		return WorkflowTrackerStatus{}, err
	}

	data, err := s.readWorkflowTrackerFile(relative)
	if errors.Is(err, os.ErrNotExist) {
		return WorkflowTrackerStatus{}, errWorkflowTrackerAbsent
	}

	var result WorkflowTrackerStatus
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}

	if err := s.validateWorkflowTracker(result, setID); err != nil {
		return result, err
	}

	return result, nil
}

func (s *WorkflowStore) validateWorkflowTracker(result WorkflowTrackerStatus, setID string) error {
	checksum := result.Checksum
	result.Checksum = ""

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if !validWorkflowTrackerSchema(result.Schema) || result.SetID != setID || result.ProjectID != s.projectID || result.Namespace != s.layout.Namespace() || result.Path != "tracker/"+setID+".md" || workflowHash(data) != checksum {
		return fmt.Errorf("%w: invalid or edited tracker metadata", ErrWorkflowProjectionStale)
	}

	if result.Status != "current" && result.Status != "stale" {
		return fmt.Errorf("invalid tracker projection state")
	}

	return nil
}

func (s *WorkflowStore) persistWorkflowTracker(result WorkflowTrackerStatus) error {
	relative, err := workflowTrackerMetadataPath(result.SetID)
	if err != nil {
		return err
	}

	path, err := s.safePath(relative)
	if err != nil {
		return err
	}

	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("tracker metadata must be a regular file")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	result.Checksum = ""

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	result.Checksum = workflowHash(data)

	data, err = json.Marshal(result)
	if err != nil {
		return err
	}

	return atomicWorkflowFile(path, data)
}

func (s *WorkflowStore) writeWorkflowTracker(result WorkflowTrackerStatus, output []byte) (WorkflowTrackerStatus, error) {
	if err := s.persistWorkflowTracker(result); err != nil {
		return result, errors.Join(ErrWorkflowProjectionStale, err)
	}

	if err := s.workflowTrackerBoundary(0); err != nil {
		return result, errors.Join(ErrWorkflowProjectionStale, err)
	}

	if err := s.checkWorkflowTrackerOutput(result.SetID, result.ObservedHash); err != nil {
		return s.workflowTrackerFailure(result, err)
	}

	if err := s.workflowTrackerReplace(result.Path, output, 1); err != nil {
		return s.workflowTrackerFailure(result, err)
	}

	if err := s.checkWorkflowTrackerOutput(result.SetID, result.PendingHash); err != nil {
		return s.workflowTrackerFailure(result, err)
	}

	return s.confirmWorkflowTracker(result, output)
}

func (s *WorkflowStore) confirmWorkflowTracker(result WorkflowTrackerStatus, output []byte) (WorkflowTrackerStatus, error) {
	result.OutputHash, result.PendingHash, result.ObservedHash = workflowHash(output), "", workflowHash(output)
	result.Status, result.Reason, result.Detail = "current", "", ""

	if result.Schema != "chr-tracker/v2" || result.GeneratedAt.IsZero() {
		result.GeneratedAt = time.Now().UTC()
	}

	if err := s.persistWorkflowTracker(result); err != nil {
		result.Status, result.Reason = "stale", "projection_stale"

		return result, errors.Join(ErrWorkflowProjectionStale, err)
	}

	if err := s.workflowTrackerBoundary(2); err != nil {
		return result, err
	}

	return s.readWorkflowTracker(result.SetID)
}

func (s *WorkflowStore) checkWorkflowTrackerOutput(setID, expected string) error {
	observed, err := s.workflowTrackerOutput(setID)
	if err != nil || observed != expected {
		return errors.Join(ErrWorkflowConflict, err)
	}

	return nil
}

func (s *WorkflowStore) workflowTrackerBoundary(stage int) error {
	if s.afterTrackerReplace != nil {
		return s.afterTrackerReplace(stage)
	}

	return nil
}

func (s *WorkflowStore) workflowTrackerReplace(relative string, output []byte, stage int) error {
	path, err := s.safePath(relative)
	if err != nil {
		return err
	}

	if err := atomicWorkflowFile(path, output); err != nil {
		return err
	}

	return s.workflowTrackerBoundary(stage)
}

func (s *WorkflowStore) workflowTrackerRegistrations() ([]string, error) {
	path, err := s.safePath("tracker/projections")
	if err != nil {
		return nil, err
	}

	directory, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	entries, err := directory.ReadDir(WorkflowPageLimit + 1)
	if errors.Is(err, io.EOF) {
		err = nil
	}

	if err := errors.Join(err, directory.Close()); err != nil {
		return nil, err
	}

	if len(entries) > WorkflowPageLimit {
		return nil, fmt.Errorf("tracker projection limit exceeded")
	}

	var ids []string

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			ids = append(ids, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}

	return ids, nil
}

// Projection failures cannot change the result of a committed domain transition.
// Read-only status independently verifies bytes if the stale marker cannot persist.
func (s *WorkflowStore) markWorkflowTrackersStale(changes []workflowFileChange) {
	ids, err := s.workflowTrackerRegistrations()
	if err != nil {
		return
	}

	changed := map[string]bool{}
	for _, change := range changes {
		changed[change.Path] = true
	}

	for _, id := range ids {
		result, err := s.readWorkflowTracker(id)
		if err != nil || result.Schema == "chr-tracker/v2" || !workflowTrackerSourcesChanged(result.Sources, changed) {
			continue
		}

		result.Status, result.Reason, result.Detail = "stale", "projection_stale", "authoritative records changed; export or repair the tracker"
		//nolint:errcheck // Derived state cannot roll back successful domain writes.
		_ = s.persistWorkflowTracker(result)
	}
}

func workflowTrackerSourcesChanged(sources []WorkflowTrackerSource, changed map[string]bool) bool {
	for _, source := range sources {
		if changed[source.Path] {
			return true
		}
	}

	return false
}
