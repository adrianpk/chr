package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (input WorkflowMergerInput) evidenceRefs() []WorkflowEvidenceRef {
	refs := []WorkflowEvidenceRef{input.SourceManifestRef, input.ValidationContractRef}
	if input.OperatorDecisionRef != nil {
		refs = append(refs, *input.OperatorDecisionRef)
	}

	if input.RepairCheckpointRef != nil {
		refs = append(refs, *input.RepairCheckpointRef)
	}

	refs = append(refs, input.IntentRefs...)
	refs = append(refs, input.TargetChangeRefs...)
	refs = append(refs, input.FindingRefs...)

	for _, requirement := range input.Requirements {
		refs = append(refs, requirement.Evidence)
	}

	for _, conflict := range input.Conflicts {
		refs = append(refs, conflict.Evidence)
	}

	if input.HistoryCheckpoint != nil {
		refs = append(refs, input.HistoryCheckpoint.Evidence)
	}

	return refs
}

func validateWorkflowMergerReferences(input WorkflowMergerInput) error {
	refs := input.evidenceRefs()
	if len(input.IntentRefs) == 0 || len(input.TargetChangeRefs) == 0 || len(refs) > WorkflowPageLimit {
		return workflowstate.ErrConflict
	}

	seen := map[string]string{}

	for _, ref := range refs {
		if err := validateWorkflowEvidenceRef(input.InvocationID, ref); err != nil {
			return err
		}

		if hash, present := seen[ref.Path]; present && hash != ref.Hash {
			return workflowstate.ErrConflict
		}

		seen[ref.Path] = ref.Hash
	}

	return nil
}

// CheckWorkflowMergerEvidence reads immutable evidence through the existing
// contained store port. It neither mutates an invocation nor acquires a slot.
func (s *WorkflowStore) CheckWorkflowMergerEvidence(ctx context.Context, input WorkflowMergerInput, now time.Time) error {
	if err := validateWorkflowMergerInput(input, now); err != nil {
		return err
	}

	for _, ref := range input.evidenceRefs() {
		if err := ctx.Err(); err != nil {
			return err
		}

		if _, err := s.readWorkflowReviewImage(ref.Path, ref.Hash); err != nil {
			return fmt.Errorf("%w: merger evidence changed or escaped containment", workflowstate.ErrConflict)
		}
	}

	if err := s.checkWorkflowMergerOperatorEvidence(input); err != nil {
		return err
	}

	return s.checkWorkflowMergerManifest(input)
}

func (s *WorkflowStore) checkWorkflowMergerManifest(input WorkflowMergerInput) error {
	var manifest WorkflowNativeSourceManifest
	if err := s.readWorkflowNativeEvidence(input.InvocationID, input.SourceManifestRef, &manifest); err != nil {
		return err
	}

	expected := [4]string{"chr-native-source/v1", input.Repository, input.SourceRef, input.Source}

	actual := [4]string{manifest.Schema, manifest.Repository, manifest.SourceRef, manifest.OriginalSource}
	if actual != expected || len(manifest.Commits) == 0 || len(manifest.Commits) > WorkflowRecordLimit {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}
	for _, commit := range manifest.Commits {
		if !workflowCommitID(commit) || len(commit) != len(input.Source) || seen[commit] {
			return workflowstate.ErrConflict
		}

		seen[commit] = true
	}

	if manifest.Commits[len(manifest.Commits)-1] != input.Source {
		return workflowstate.ErrConflict
	}

	return s.checkWorkflowMergerCheckpointEvidence(input, manifest)
}

func (s *WorkflowStore) checkWorkflowMergerCheckpointEvidence(input WorkflowMergerInput, manifest WorkflowNativeSourceManifest) error {
	checkpoint := input.HistoryCheckpoint
	if checkpoint == nil {
		return nil
	}

	if checkpoint.ReplayIndex >= len(manifest.Commits) || manifest.Commits[checkpoint.ReplayIndex] != checkpoint.OriginalCommit {
		return workflowstate.ErrConflict
	}

	for i, replay := range checkpoint.Predecessors {
		if manifest.Commits[i] != replay.Source {
			return workflowstate.ErrConflict
		}
	}

	data, err := s.readWorkflowReviewImage(checkpoint.Evidence.Path, checkpoint.Evidence.Hash)
	if err != nil {
		return err
	}

	var retained WorkflowMergerCheckpoint
	if err := json.Unmarshal(data, &retained); err != nil {
		return err
	}

	// The evidence cannot embed its own file hash. Compare the checkpoint payload
	// independently, then retain the external immutable reference in the input.
	expected := *checkpoint

	expected.Evidence, retained.Evidence = WorkflowEvidenceRef{}, WorkflowEvidenceRef{}
	if !workflowNativeJSONEqual(expected, retained) {
		return workflowstate.ErrConflict
	}

	return nil
}
