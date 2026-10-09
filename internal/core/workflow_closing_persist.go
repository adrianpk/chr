package core

import (
	"context"
	"encoding/json"
	"time"
)

func (s *WorkflowStore) commitWorkflowClosingStart(ctx context.Context, claim *workflowClosingClaim) error {
	if claim.set.Set.Closing == nil {
		claim.set.Set.Closing = &WorkflowClosingBinding{}
	}

	claim.set.Set.Closing.Attempts = append(claim.set.Set.Closing.Attempts, WorkflowClosingEvidence{})
	claim.set.Set.Status = "closing"

	return s.commitWorkflowClosingState(ctx, claim, workflowFileChange{Path: claim.state.Input.Ref, After: claim.data})
}

func (s *WorkflowStore) commitWorkflowClosingState(ctx context.Context, claim *workflowClosingClaim, extra ...workflowFileChange) error {
	previous := WorkflowInputImage{Path: claim.set.Path, Revision: claim.set.Revision, SourceHash: claim.set.SourceHash}
	index := claim.state.Generation - 1
	binding := &claim.set.Set.Closing.Attempts[index]

	data, err := json.MarshalIndent(claim.state, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	stateFile := workflowFileChange{Path: workflowClosingFile(claim.set.ID, claim.state.Generation, "state.json"), Before: binding.Hash, After: data}
	*binding = WorkflowClosingEvidence{Ref: stateFile.Path, Hash: workflowHash(data)}
	claim.set.Revision++
	claim.set.UpdatedAt = time.Now().UTC()

	files, err := s.encodeChanges(claim.set)
	if err != nil {
		return err
	}

	claim.set.SourceHash = workflowClaimRecordHash(claim.set, files)
	files = append(files, stateFile)
	files = append(files, extra...)

	files, err = s.workflowClosingInvocationFiles(ctx, claim, previous, files)
	if err != nil {
		return err
	}

	return s.commitFiles(files)
}

func (s *WorkflowStore) checkWorkflowClosingClaim(ctx context.Context, claim *workflowClosingClaim, lease bool) error {
	if claim.state.OwnerRoot != s.layout.Root() || claim.state.TokenHash != workflowHash([]byte(claim.worker.Token)) || claim.state.Outcome != "running" {
		return ErrWorkflowOwned
	}

	if lease && time.Now().After(claim.state.LeaseUntil) {
		return ErrWorkflowOwned
	}

	if err := s.checkWorkflowClosingCredential(claim); err != nil {
		return err
	}

	return s.checkWorkflowClosingInputs(ctx, claim)
}

func (s *WorkflowStore) checkWorkflowClosingCredential(claim *workflowClosingClaim) error {
	set, err := s.readRecord(claim.set.Path)
	if err != nil || set.SourceHash != claim.set.SourceHash || set.Revision != claim.set.Revision {
		return errorsOrConflict(err)
	}

	state, err := s.readWorkflowClosingAttempt(claim.set, claim.state.Generation-1)
	if err != nil || state.TokenHash != workflowHash([]byte(claim.worker.Token)) || state.Outcome != "running" {
		return errorsOrConflict(err)
	}

	return nil
}
