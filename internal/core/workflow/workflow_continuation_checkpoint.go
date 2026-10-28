package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) retainWorkflowContinuation(ctx context.Context, observed workflowstate.Invocation, steps int, reason string) (result workflowstate.Invocation, err error) {
	result = observed

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, observed.ID)
		if readErr != nil {
			return readErr
		}

		if invocationLocation.Run.Revision != observed.Revision || invocationLocation.Run.RequestHash != observed.RequestHash {
			return workflowstate.ErrConflict
		}

		if readErr = invocationLocation.Owner.checkWorkflowInputs(ctx, invocationLocation.Run); readErr != nil {
			return readErr
		}

		if readErr = invocationLocation.Owner.commitWorkflowContinuation(&invocationLocation.Run, steps, reason); readErr != nil {
			return readErr
		}

		result = invocationLocation.Run.Invocation

		return nil
	})

	return result, err
}

func (s *Store) commitWorkflowContinuation(run *workflowInvocationFile, steps int, reason string, extra ...workflowstate.FileChange) error {
	previous := run.Continuation
	if len(extra) == 0 && steps == 0 && previous != nil && previous.SourceRevision+1 == run.Revision && previous.Reason == reason {
		return nil
	}

	checkpoint := workflowstate.ContinuationCheckpoint{Schema: "chr-continuation-checkpoint/v1", InvocationID: run.ID, RequestHash: run.RequestHash, Sequence: 1, SourceRevision: run.Revision, Cursor: run.Cursor, ConfirmedCursor: run.ConfirmedCursor, MemberCount: len(run.Members), AttemptCount: len(run.Attempts), Steps: steps, Reason: reason, RecordedAt: time.Now().UTC()}
	if previous != nil {
		checkpoint.Sequence = previous.Sequence + 1
		checkpoint.ParentRef, checkpoint.ParentHash = previous.EvidenceRef, previous.EvidenceHash
	}

	if checkpoint.Sequence > workflowstate.RecordLimit {
		return fmt.Errorf("continuation checkpoint history exhausted")
	}

	data, err := encodeWorkflowContinuationCheckpoint(checkpoint)
	if err != nil {
		return err
	}

	checkpoint.EvidenceRef = workflowContinuationCheckpointPath(run.ID, checkpoint.Sequence)
	checkpoint.EvidenceHash = contenthash.Sum(data)
	run.Continuation = &checkpoint

	return s.commitWorkflowPullRequestFiles(run, extra, workflowstate.FileChange{Path: checkpoint.EvidenceRef, After: data})
}

func workflowContinuationCheckpointPath(id string, sequence int) string {
	return filepath.ToSlash(filepath.Join("run", id, "evidence", fmt.Sprintf("continuation-%04d.json", sequence)))
}

func encodeWorkflowContinuationCheckpoint(checkpoint workflowstate.ContinuationCheckpoint) ([]byte, error) {
	checkpoint.EvidenceRef, checkpoint.EvidenceHash = "", ""

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return nil, errorsOrConflict(err)
	}

	return append(data, '\n'), nil
}
