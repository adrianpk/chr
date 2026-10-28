package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckWorkflowMergerEvidence reads immutable evidence through the existing
// contained store port. It neither mutates an invocation nor acquires a slot.
func (s *Store) CheckWorkflowMergerEvidence(ctx context.Context, input workflowstate.MergerInput, now time.Time) error {
	if err := workflowmerge.CheckInput(input, now); err != nil {
		return err
	}

	for _, ref := range input.EvidenceRefs() {
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

func (s *Store) checkWorkflowMergerManifest(input workflowstate.MergerInput) error {
	var manifest NativeSourceManifest
	if err := s.readWorkflowNativeEvidence(input.InvocationID, input.SourceManifestRef, &manifest); err != nil {
		return err
	}

	expected := [4]string{"chr-native-source/v1", input.Repository, input.SourceRef, input.Source}

	actual := [4]string{manifest.Schema, manifest.Repository, manifest.SourceRef, manifest.OriginalSource}
	if actual != expected || len(manifest.Commits) == 0 || len(manifest.Commits) > workflowstate.RecordLimit {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}
	for _, commit := range manifest.Commits {
		if !workflowstate.ValidCommitID(commit) || len(commit) != len(input.Source) || seen[commit] {
			return workflowstate.ErrConflict
		}

		seen[commit] = true
	}

	if manifest.Commits[len(manifest.Commits)-1] != input.Source {
		return workflowstate.ErrConflict
	}

	return s.checkWorkflowMergerCheckpointEvidence(input, manifest)
}

func (s *Store) checkWorkflowMergerCheckpointEvidence(input workflowstate.MergerInput, manifest NativeSourceManifest) error {
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

	var retained workflowstate.MergerCheckpoint
	if err := json.Unmarshal(data, &retained); err != nil {
		return err
	}

	// The evidence cannot embed its own file hash. Compare the checkpoint payload
	// independently, then retain the external immutable reference in the input.
	expected := *checkpoint

	expected.Evidence, retained.Evidence = workflowstate.EvidenceRef{}, workflowstate.EvidenceRef{}
	if !workflowNativeJSONEqual(expected, retained) {
		return workflowstate.ErrConflict
	}

	return nil
}
