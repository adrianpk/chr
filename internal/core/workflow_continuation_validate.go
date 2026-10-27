package core

import (
	"bytes"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowContinuation(run workflowInvocationFile) error {
	checkpoint := run.Continuation
	if checkpoint == nil {
		return nil
	}

	if checkpoint.Schema != "chr-continuation-checkpoint/v1" || checkpoint.InvocationID != run.ID || checkpoint.RequestHash != run.RequestHash || checkpoint.RecordedAt.IsZero() {
		return workflowstate.ErrConflict
	}

	if !validWorkflowContinuationPosition(*checkpoint, run) || !validWorkflowContinuationEvidence(*checkpoint) {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowContinuationPosition(checkpoint WorkflowContinuationCheckpoint, run workflowInvocationFile) bool {
	return checkpoint.SourceRevision > 0 && checkpoint.SourceRevision < run.Revision && checkpoint.MemberCount >= 0 && checkpoint.MemberCount <= len(run.Members) && checkpoint.Cursor >= 0 && checkpoint.Cursor <= checkpoint.MemberCount && checkpoint.ConfirmedCursor >= 0 && checkpoint.ConfirmedCursor <= checkpoint.Cursor && checkpoint.AttemptCount >= 0 && checkpoint.AttemptCount <= len(run.Attempts)
}

func validWorkflowContinuationEvidence(checkpoint WorkflowContinuationCheckpoint) bool {
	if checkpoint.Sequence < 1 || checkpoint.Sequence > workflowstate.RecordLimit || checkpoint.EvidenceRef != workflowContinuationCheckpointPath(checkpoint.InvocationID, checkpoint.Sequence) || len(checkpoint.EvidenceHash) != 64 {
		return false
	}

	if checkpoint.Steps < 0 || checkpoint.Steps > workflowstate.PageLimit || !slices.Contains([]string{"accepted", "step_limit", "awaiting_authority", "stopped", "set_closure_required", "blocked", "waiting_capacity", "closing_validation_failed", "interrupted"}, checkpoint.Reason) {
		return false
	}

	return validWorkflowContinuationParent(checkpoint)
}

func validWorkflowContinuationParent(checkpoint WorkflowContinuationCheckpoint) bool {
	if checkpoint.Sequence == 1 {
		return checkpoint.ParentRef == "" && checkpoint.ParentHash == ""
	}

	return checkpoint.ParentRef == workflowContinuationCheckpointPath(checkpoint.InvocationID, checkpoint.Sequence-1) && len(checkpoint.ParentHash) == 64
}

func (s *WorkflowStore) checkWorkflowContinuationCheckpoint(run workflowInvocationFile) error {
	checkpoint := run.Continuation
	if checkpoint == nil {
		return nil
	}

	path, err := s.safePath(checkpoint.EvidenceRef)
	if err != nil {
		return err
	}

	data, err := readWorkflowFile(path)
	if err != nil || contenthash.Sum(data) != checkpoint.EvidenceHash {
		return workflowstate.ErrConflict
	}

	expected, err := encodeWorkflowContinuationCheckpoint(*checkpoint)
	if err != nil || !bytes.Equal(data, expected) {
		return workflowstate.ErrConflict
	}

	if checkpoint.ParentRef != "" {
		return s.checkWorkflowPrerequisiteEvidence(map[string]string{checkpoint.ParentRef: checkpoint.ParentHash})
	}

	return nil
}
