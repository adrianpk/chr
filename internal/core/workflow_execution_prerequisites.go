package core

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) workflowPrerequisiteImages(ctx context.Context, run workflowInvocationFile, record WorkflowRecord) ([]WorkflowInputImage, map[string]string, error) {
	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return nil, nil, err
	}

	dependencies, err := newWorkflowDependencies(selection)
	if err != nil {
		return nil, nil, err
	}

	refs := dependencies.refs[WorkflowRef{Kind: "ticket", ID: record.Task.Key}]
	if len(refs) > WorkflowPageLimit {
		return nil, nil, fmt.Errorf("workflow prerequisite bounds exceeded")
	}

	request := run.ExecutionSelection()
	request.CumulativeBranch, request.CumulativeCommit = run.Lane.Branch, run.Lane.HeadCommit
	reader := workflowProofReader{store: s, selection: selection, request: request, observations: map[workflowProofKey]string{}}
	images := []WorkflowInputImage{}
	hashes := map[string]string{}

	for _, ref := range refs {
		if reason := reader.prerequisiteReason(ctx, record, ref); reason != "" {
			return nil, nil, fmt.Errorf("workflow prerequisite %s: %s", ref.ID, reason)
		}

		dependency := workflowPrerequisiteRecord(selection, ref)

		images = append(images, WorkflowInputImage{Path: dependency.Path, Revision: dependency.Revision, SourceHash: dependency.SourceHash})
		if dependency.Task != nil {
			for _, path := range dependency.Task.Evidence {
				hashes[path] = selection.evidenceHashes[path]
			}
		}
	}

	return images, hashes, nil
}

func workflowPrerequisiteRecord(selection workflowSelection, ref WorkflowRef) WorkflowRecord {
	if ref.Kind == "ticket" {
		return selection.records[ref.ID]
	}

	for _, record := range selection.sets {
		for _, descriptor := range record.Set.Slices {
			if descriptor.ID == ref.ID {
				return record
			}
		}
	}

	return WorkflowRecord{}
}

func (s *WorkflowStore) checkWorkflowPreparedInput(prepared workflowPreparedTask) error {
	path, err := s.safePath(prepared.inputRef)
	if err != nil {
		return err
	}

	data, err := readWorkflowFile(path)
	if err != nil || contenthash.Sum(data) != prepared.inputHash {
		return fmt.Errorf("%w: frozen worker input changed", workflowstate.ErrConflict)
	}

	for _, image := range prepared.input.Prerequisites {
		record, err := s.readRecord(image.Path)
		if err != nil || record.Revision != image.Revision || record.SourceHash != image.SourceHash {
			return fmt.Errorf("%w: frozen prerequisite changed", workflowstate.ErrConflict)
		}
	}

	return s.checkWorkflowPrerequisiteEvidence(prepared.input.EvidenceHashes)
}

func (s *WorkflowStore) checkWorkflowPrerequisiteEvidence(hashes map[string]string) error {
	for relative, hash := range hashes {
		path, err := s.safePath(relative)
		if err != nil {
			return err
		}

		data, err := readWorkflowFile(path)
		if err != nil || contenthash.Sum(data) != hash {
			return fmt.Errorf("%w: frozen prerequisite evidence changed", workflowstate.ErrConflict)
		}
	}

	return nil
}
