package core

import (
	"bytes"
	"slices"
)

func validateWorkflowContinuation(run workflowInvocationFile) error {
	checkpoint := run.Continuation
	if checkpoint == nil {
		return nil
	}

	if checkpoint.Schema != "chr-continuation-checkpoint/v1" || checkpoint.InvocationID != run.ID || checkpoint.RequestHash != run.RequestHash || checkpoint.RecordedAt.IsZero() {
		return ErrWorkflowConflict
	}

	if !validWorkflowContinuationPosition(*checkpoint, run) || !validWorkflowContinuationEvidence(*checkpoint) {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowContinuationPosition(checkpoint WorkflowContinuationCheckpoint, run workflowInvocationFile) bool {
	return checkpoint.SourceRevision > 0 && checkpoint.SourceRevision < run.Revision && checkpoint.MemberCount >= 0 && checkpoint.MemberCount <= len(run.Members) && checkpoint.Cursor >= 0 && checkpoint.Cursor <= checkpoint.MemberCount && checkpoint.ConfirmedCursor >= 0 && checkpoint.ConfirmedCursor <= checkpoint.Cursor && checkpoint.AttemptCount >= 0 && checkpoint.AttemptCount <= len(run.Attempts)
}

func validWorkflowContinuationEvidence(checkpoint WorkflowContinuationCheckpoint) bool {
	if checkpoint.Sequence < 1 || checkpoint.Sequence > WorkflowRecordLimit || checkpoint.EvidenceRef != workflowContinuationCheckpointPath(checkpoint.InvocationID, checkpoint.Sequence) || len(checkpoint.EvidenceHash) != 64 {
		return false
	}

	if checkpoint.Steps < 0 || checkpoint.Steps > WorkflowPageLimit || !slices.Contains([]string{"step_limit", "awaiting_authority", "stopped", "set_closure_required", "blocked", "waiting_capacity", "closing_validation_failed", "interrupted"}, checkpoint.Reason) {
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
	if err != nil || workflowHash(data) != checkpoint.EvidenceHash {
		return ErrWorkflowConflict
	}

	expected, err := encodeWorkflowContinuationCheckpoint(*checkpoint)
	if err != nil || !bytes.Equal(data, expected) {
		return ErrWorkflowConflict
	}

	if checkpoint.ParentRef != "" {
		return s.checkWorkflowPrerequisiteEvidence(map[string]string{checkpoint.ParentRef: checkpoint.ParentHash})
	}

	return nil
}
