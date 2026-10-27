package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowContinuationCheckpoint binds a bounded call to frozen request identity.
// Prior images remain immutable; ParentRef/Hash retain the previous checkpoint.
type WorkflowContinuationCheckpoint struct {
	Schema          string    `json:"schema"`
	InvocationID    string    `json:"invocation_id"`
	RequestHash     string    `json:"request_hash"`
	Sequence        int       `json:"sequence"`
	SourceRevision  uint64    `json:"source_revision"`
	Cursor          int       `json:"cursor"`
	ConfirmedCursor int       `json:"confirmed_cursor"`
	MemberCount     int       `json:"member_count"`
	AttemptCount    int       `json:"attempt_count"`
	Steps           int       `json:"steps"`
	Reason          string    `json:"reason"`
	ParentRef       string    `json:"parent_ref,omitempty"`
	ParentHash      string    `json:"parent_hash,omitempty"`
	RecordedAt      time.Time `json:"recorded_at"`
	EvidenceRef     string    `json:"evidence_ref,omitempty"`
	EvidenceHash    string    `json:"evidence_hash,omitempty"`
}

func (s *WorkflowStore) retainWorkflowContinuation(ctx context.Context, observed WorkflowInvocation, steps int, reason string) (result WorkflowInvocation, err error) {
	result = observed

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, observed.ID)
		if readErr != nil {
			return readErr
		}

		if run.Revision != observed.Revision || run.RequestHash != observed.RequestHash {
			return workflowstate.ErrConflict
		}

		if readErr = owner.checkWorkflowInputs(ctx, run); readErr != nil {
			return readErr
		}

		if readErr = owner.commitWorkflowContinuation(&run, steps, reason); readErr != nil {
			return readErr
		}

		result = run.WorkflowInvocation

		return nil
	})

	return result, err
}

func (s *WorkflowStore) commitWorkflowContinuation(run *workflowInvocationFile, steps int, reason string, extra ...workflowFileChange) error {
	previous := run.Continuation
	if len(extra) == 0 && steps == 0 && previous != nil && previous.SourceRevision+1 == run.Revision && previous.Reason == reason {
		return nil
	}

	checkpoint := WorkflowContinuationCheckpoint{Schema: "chr-continuation-checkpoint/v1", InvocationID: run.ID, RequestHash: run.RequestHash, Sequence: 1, SourceRevision: run.Revision, Cursor: run.Cursor, ConfirmedCursor: run.ConfirmedCursor, MemberCount: len(run.Members), AttemptCount: len(run.Attempts), Steps: steps, Reason: reason, RecordedAt: time.Now().UTC()}
	if previous != nil {
		checkpoint.Sequence = previous.Sequence + 1
		checkpoint.ParentRef, checkpoint.ParentHash = previous.EvidenceRef, previous.EvidenceHash
	}

	if checkpoint.Sequence > workflowstate.WorkflowRecordLimit {
		return fmt.Errorf("continuation checkpoint history exhausted")
	}

	data, err := encodeWorkflowContinuationCheckpoint(checkpoint)
	if err != nil {
		return err
	}

	checkpoint.EvidenceRef = workflowContinuationCheckpointPath(run.ID, checkpoint.Sequence)
	checkpoint.EvidenceHash = contenthash.Sum(data)
	run.Continuation = &checkpoint

	return s.commitWorkflowPullRequestFiles(run, extra, workflowFileChange{Path: checkpoint.EvidenceRef, After: data})
}

func workflowContinuationCheckpointPath(id string, sequence int) string {
	return filepath.ToSlash(filepath.Join("run", id, "evidence", fmt.Sprintf("continuation-%04d.json", sequence)))
}

func encodeWorkflowContinuationCheckpoint(checkpoint WorkflowContinuationCheckpoint) ([]byte, error) {
	checkpoint.EvidenceRef, checkpoint.EvidenceHash = "", ""

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil || len(data) >= workflowstate.WorkflowFileLimit {
		return nil, errorsOrConflict(err)
	}

	return append(data, '\n'), nil
}
