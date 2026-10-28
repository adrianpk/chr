package workflow

import (
	"context"
	"encoding/json"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) commitWorkflowClosingStart(claim *workflowClosingClaim) error {
	if claim.set.Set.Closing == nil {
		claim.set.Set.Closing = &workflowstate.WorkflowClosingBinding{}
	}

	claim.set.Set.Closing.Attempts = append(claim.set.Set.Closing.Attempts, workflowstate.ClosingEvidence{})
	claim.set.Set.Status = "closing"

	return s.commitWorkflowClosingState(claim, workflowstate.FileChange{Path: claim.state.Input.Ref, After: claim.data})
}

func (s *Store) commitWorkflowClosingState(claim *workflowClosingClaim, extra ...workflowstate.FileChange) error {
	previous := InputImage{Path: claim.set.Path, Revision: claim.set.Revision, SourceHash: claim.set.SourceHash}
	index := claim.state.Generation - 1
	binding := &claim.set.Set.Closing.Attempts[index]

	data, err := json.MarshalIndent(claim.state, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	stateFile := workflowstate.FileChange{Path: workflowClosingFile(claim.set.ID, claim.state.Generation, "state.json"), Before: binding.Hash, After: data}
	*binding = workflowstate.ClosingEvidence{Ref: stateFile.Path, Hash: contenthash.Sum(data)}
	claim.set.Revision++
	claim.set.UpdatedAt = time.Now().UTC()

	files, err := s.encodeChanges(claim.set)
	if err != nil {
		return err
	}

	claim.set.SourceHash = workflowClaimRecordHash(claim.set, files)
	files = append(files, stateFile)
	files = append(files, extra...)

	files, err = s.workflowClosingInvocationFiles(claim, previous, files)
	if err != nil {
		return err
	}

	return s.commitFiles(files)
}

func (s *Store) checkWorkflowClosingClaim(ctx context.Context, claim *workflowClosingClaim, lease bool) error {
	if claim.state.OwnerRoot != s.layout.Root() || claim.state.TokenHash != contenthash.Sum([]byte(claim.worker.Token)) || claim.state.Outcome != "running" {
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

func (s *Store) checkWorkflowClosingCredential(claim *workflowClosingClaim) error {
	set, err := s.readRecord(claim.set.Path)
	if err != nil || set.SourceHash != claim.set.SourceHash || set.Revision != claim.set.Revision {
		return errorsOrConflict(err)
	}

	state, err := s.readWorkflowClosingAttempt(claim.set, claim.state.Generation-1)
	if err != nil || state.TokenHash != contenthash.Sum([]byte(claim.worker.Token)) || state.Outcome != "running" {
		return errorsOrConflict(err)
	}

	return nil
}
