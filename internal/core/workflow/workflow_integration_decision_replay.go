package workflow

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) checkWorkflowDecisionReplayAnswer(record IntegrationDecisions, answer workflowstate.IntegrationAnswer) error {
	if answer.OriginalCommit == "" {
		return nil
	}

	var manifest NativeSourceManifest
	if err := s.readWorkflowNativeEvidence(record.Input.InvocationID, record.Input.SourceManifestRef, &manifest); err != nil {
		return err
	}

	if !slices.Contains(manifest.Commits, answer.OriginalCommit) {
		return workflowstate.ErrConflict
	}

	return nil
}
