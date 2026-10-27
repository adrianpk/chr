package core

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func (r *fileAuthoringSessionRepository) commitDumpMutation(ctx context.Context, binding dumpMutationBinding) (ProjectDumpMutationResult, error) {
	target, err := r.dumpMutationTarget(ctx, binding.Request)
	if err != nil {
		return ProjectDumpMutationResult{}, err
	}

	if target.Destination != binding.Destination || target.RawReference != binding.RawReference {
		return ProjectDumpMutationResult{}, workflowstate.ErrConflict
	}

	changes, result, err := r.dumpMutationChanges(binding, target)
	if err != nil {
		return result, err
	}

	receipt, err := projectDumpReceiptChange(binding.Digest, result)
	if err != nil {
		return result, err
	}

	changes = append(changes, receipt)

	if err := ctx.Err(); err != nil {
		return result, err
	}

	if err := r.writeAuthoringJournalVersion(projectDumpJournalSchema, changes); err != nil {
		return result, err
	}

	return result, r.recoverAuthoringMutation()
}

func (r *fileAuthoringSessionRepository) dumpMutationChanges(binding dumpMutationBinding, target dumpMutationTarget) ([]workflowstate.FileChange, ProjectDumpMutationResult, error) {
	result := ProjectDumpMutationResult{OperationID: binding.Request.OperationID, Destination: target.Destination}

	body, metadata, history, err := r.dumpMutationBefore(target)
	if err != nil {
		return nil, result, err
	}

	content, err := dumpMutationBody(binding.Request, target, binding.Prepared, body)
	if err != nil {
		return nil, result, err
	}

	updated, err := r.dumpMutationMetadata(binding, target, metadata)
	if err != nil {
		return nil, result, err
	}

	changes := append(history, workflowstate.FileChange{Path: filepath.FromSlash(target.Destination), Before: target.Artifact.SourceHash, After: content},
		workflowstate.FileChange{Path: filepath.Join(filepath.Dir(target.Destination), "metadata.json"), Before: target.Artifact.MetadataHash, After: updated})
	result.SourceHash, result.MetadataHash = contenthash.Sum(content), contenthash.Sum(updated)

	components, err := r.dumpCaptureChanges(binding, target)
	if err != nil {
		return nil, result, err
	}

	changes = append(changes, components...)

	if target.Project.Name != "" && target.Project.ProjectDumpRef != target.Destination {
		project := target.Project
		project.ProjectDumpRef = target.Destination
		changes = append(changes, workflowstate.FileChange{Path: filepath.Join(workspacecore.DirectoryName, UserDirectoryName, "project.md"), Before: project.SourceHash, After: []byte(RenderProjectMetadata(project))})
	}

	return changes, result, nil
}

func (r *fileAuthoringSessionRepository) dumpMutationBefore(target dumpMutationTarget) ([]byte, []byte, []workflowstate.FileChange, error) {
	if target.Artifact.Reference == "" {
		return nil, nil, nil, nil
	}

	fileImages, err := r.projectDumpImages(target.Artifact)
	body, metadata := fileImages.Body, fileImages.Metadata

	if err != nil {
		return nil, nil, nil, err
	}

	var changes []workflowstate.FileChange

	for _, image := range []struct {
		name    string
		content []byte
	}{{"artifact.md", body}, {"metadata.json", metadata}} {
		change, err := r.projectDumpHistoryChange(projectDumpHistoryRef(target.Artifact)+"/"+image.name, image.content)
		if err != nil {
			return nil, nil, nil, err
		}

		changes = append(changes, change)
	}

	return body, metadata, changes, nil
}

func (r *fileAuthoringSessionRepository) dumpCaptureChanges(binding dumpMutationBinding, target dumpMutationTarget) ([]workflowstate.FileChange, error) {
	if binding.Prepared == nil {
		return nil, nil
	}

	metadata, err := r.newDumpMutationMetadata(target.RawReference, binding.Request)
	if err != nil {
		return nil, err
	}

	metadata["status"] = "raw"
	metadata["curated_ref"] = target.Destination
	metadata["source_request_id"] = binding.Request.OperationID

	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, err
	}

	return []workflowstate.FileChange{
		{Path: filepath.FromSlash(target.RawReference), After: []byte(binding.Prepared.RawContent)},
		{Path: filepath.Join(filepath.Dir(target.RawReference), "metadata.json"), After: append(encoded, '\n')},
	}, nil
}

func (r *fileAuthoringSessionRepository) newDumpMutationMetadata(reference string, request DumpMutationRequest) (map[string]any, error) {
	parts := strings.Split(reference, "/")

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: r.projectDir, Namespace: parts[1]})
	if err != nil {
		return nil, err
	}

	metadata := opscore.NewArtifactMetadata(opscore.ArtifactOptions{Layout: layout, Type: "dump", Title: request.Title, Producer: "chr"}, parts[3], time.Now().UTC())
	if request.CaseSlug != "" {
		metadata["case"] = request.CaseSlug
	}

	return metadata, nil
}

func (r *fileAuthoringSessionRepository) dumpMutationMetadata(binding dumpMutationBinding, target dumpMutationTarget, before []byte) ([]byte, error) {
	var (
		metadata map[string]any
		err      error
	)
	if before == nil {
		metadata, err = r.newDumpMutationMetadata(target.Destination, binding.Request)
	} else {
		err = json.Unmarshal(before, &metadata)
		if err == nil {
			err = advanceAuthoringMetadata(metadata)
		}
	}

	if err != nil {
		return nil, err
	}

	metadata["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)

	metadata["source_request_id"] = binding.Request.OperationID
	if before != nil {
		metadata["dump_history_ref"] = projectDumpHistoryRef(target.Artifact)
	}

	if binding.Prepared != nil {
		if err := appendDumpCaptureMetadata(metadata, binding); err != nil {
			return nil, err
		}
	}

	encoded, err := json.MarshalIndent(metadata, "", "  ")

	return append(encoded, '\n'), err
}

func appendDumpCaptureMetadata(metadata map[string]any, binding dumpMutationBinding) error {
	var captures []any

	if retained, exists := metadata["dump_captures"]; exists {
		var valid bool

		captures, valid = retained.([]any)
		if !valid {
			return errors.New("invalid dump capture provenance")
		}
	}

	if len(captures) >= workflowstate.RecordLimit {
		return errors.New("dump capture provenance capacity exhausted")
	}

	metadata["dump_captures"] = append(captures, map[string]string{
		"operation_id": binding.Request.OperationID, "raw_ref": binding.RawReference,
		"input_hash": contenthash.Sum([]byte(binding.Request.Content)), "raw_hash": contenthash.Sum([]byte(binding.Prepared.RawContent)),
	})

	return nil
}
