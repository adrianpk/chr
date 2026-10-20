package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (r *fileAuthoringSessionRepository) projectDumpImages(source ProjectDumpArtifact) ([]byte, []byte, error) {
	body, err := r.readAuthoringDocument(source.Reference)
	if err != nil {
		return nil, nil, err
	}

	metadata, err := r.readAuthoringDocument(filepath.Join(filepath.Dir(source.Reference), "metadata.json"))
	if err != nil {
		return nil, nil, err
	}

	if workflowHash(body) != source.SourceHash || workflowHash(metadata) != source.MetadataHash {
		return nil, nil, ErrWorkflowConflict
	}

	return body, metadata, nil
}

func (r *fileAuthoringSessionRepository) composeProjectDumps(sources []ProjectDumpArtifact) ([]byte, error) {
	var composed bytes.Buffer

	for index, source := range sources {
		body, _, err := r.projectDumpImages(source)
		if err != nil {
			return nil, err
		}

		if bytes.HasPrefix(body, []byte("---\n")) {
			front, content, err := splitWorkflowMarkdown(body)
			if err != nil {
				return nil, err
			}

			if index == 0 {
				composed.WriteString("---\n" + front + "\n---\n")
			}

			body = content
		}

		fmt.Fprintf(&composed, "\n## Source: %s\n\nSource hash: %s\n\n", source.Reference, source.SourceHash)
		composed.Write(body)
		composed.WriteByte('\n')

		if composed.Len() > WorkflowFileLimit {
			return nil, fmt.Errorf("reconciled dump exceeds 1 MiB")
		}
	}

	return composed.Bytes(), nil
}

func projectDumpHistoryRef(source ProjectDumpArtifact) string {
	image := workflowHash([]byte(source.SourceHash + "\x00" + source.MetadataHash))

	return filepath.ToSlash(filepath.Join(filepath.Dir(source.Reference), "files", "history", image))
}

func (r *fileAuthoringSessionRepository) projectDumpHistoryChange(reference string, content []byte) (workflowFileChange, error) {
	change := workflowFileChange{Path: filepath.FromSlash(reference), After: content}

	before, err := r.readAuthoringDocument(reference)
	if errors.Is(err, os.ErrNotExist) {
		return change, nil
	}

	if err != nil {
		return change, err
	}

	if !bytes.Equal(before, content) {
		return change, ErrWorkflowConflict
	}

	change.Before = workflowHash(before)

	return change, nil
}

func projectDumpReconciledMetadata(source ProjectDumpArtifact, original []byte, preview ProjectDumpReconciliationPreview, operationID string) ([]byte, error) {
	var metadata map[string]any
	if err := json.Unmarshal(original, &metadata); err != nil {
		return nil, err
	}

	if err := advanceAuthoringMetadata(metadata); err != nil {
		return nil, err
	}

	metadata["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	metadata["dump_history_ref"] = projectDumpHistoryRef(source)

	metadata["reconciliation_operation"] = operationID
	if source.Reference == preview.Destination {
		metadata["dump_reconciliation"] = projectDumpReconciliationProvenance(preview, operationID)
	} else {
		metadata["canonical_ref"] = preview.Destination
	}

	encoded, err := json.MarshalIndent(metadata, "", "  ")

	return append(encoded, '\n'), err
}

func projectDumpReconciliationProvenance(preview ProjectDumpReconciliationPreview, operationID string) map[string]any {
	sources := make([]map[string]string, 0, len(preview.Sources))
	for _, source := range preview.Sources {
		sources = append(sources, map[string]string{
			"reference": source.Reference, "source_hash": source.SourceHash, "metadata_hash": source.MetadataHash,
			"history_ref": projectDumpHistoryRef(source), "timestamp_issue": source.TimestampIssue,
		})
	}

	return map[string]any{
		"operation_id": operationID, "observed_revision": preview.Revision, "target": preview.Destination,
		"target_source_hash": preview.Sources[0].SourceHash, "sources": sources,
	}
}

func (r *fileAuthoringSessionRepository) projectDumpReconciliationChanges(observed ProjectDumpObservation, preview ProjectDumpReconciliationPreview, operationID string) ([]workflowFileChange, ProjectDumpMutationResult, error) {
	result := ProjectDumpMutationResult{OperationID: operationID, Destination: preview.Destination, SourceHash: workflowHash([]byte(preview.Content))}

	changes := []workflowFileChange{{Path: filepath.FromSlash(preview.Destination), Before: preview.Sources[0].SourceHash, After: []byte(preview.Content)}}
	for _, source := range preview.Sources {
		sourceChanges, err := r.projectDumpReconciliationSourceChanges(source, preview, operationID)
		if err != nil {
			return nil, result, err
		}

		if source.Reference == preview.Destination {
			result.MetadataHash = workflowHash(sourceChanges[2].After)
		}

		changes = append(changes, sourceChanges...)
	}

	project := observed.Project
	project.ProjectDumpRef = preview.Destination
	changes = append(changes, workflowFileChange{
		Path: filepath.Join(DirectoryName, UserDirectoryName, "project.md"), Before: project.SourceHash, After: []byte(RenderProjectMetadata(project)),
	})

	return changes, result, nil
}

func (r *fileAuthoringSessionRepository) projectDumpReconciliationSourceChanges(source ProjectDumpArtifact, preview ProjectDumpReconciliationPreview, operationID string) ([]workflowFileChange, error) {
	body, metadata, err := r.projectDumpImages(source)
	if err != nil {
		return nil, err
	}

	changes := make([]workflowFileChange, 0, 3)

	for _, image := range []struct {
		name    string
		content []byte
	}{{"artifact.md", body}, {"metadata.json", metadata}} {
		change, err := r.projectDumpHistoryChange(projectDumpHistoryRef(source)+"/"+image.name, image.content)
		if err != nil {
			return nil, err
		}

		changes = append(changes, change)
	}

	updated, err := projectDumpReconciledMetadata(source, metadata, preview, operationID)
	if err != nil {
		return nil, err
	}

	changes = append(changes, workflowFileChange{Path: filepath.Join(filepath.Dir(source.Reference), "metadata.json"), Before: source.MetadataHash, After: updated})

	return changes, nil
}
