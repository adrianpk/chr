package core

import (
	"context"
	"fmt"
	"time"
)

// WorkflowMergerAcceptance records host observations, never a validation or
// review grant. Accepted checkpoints still require the complete candidate gate.
type WorkflowMergerAcceptance struct {
	SchemaVersion int                            `json:"schema_version"`
	AttemptID     string                         `json:"attempt_id"`
	InputHash     string                         `json:"input_hash"`
	CheckpointRef WorkflowEvidenceRef            `json:"checkpoint_ref"`
	Outcome       string                         `json:"outcome"`
	CheckedAt     time.Time                      `json:"checked_at"`
	Repair        *WorkflowIntegrationTreeRepair `json:"repair,omitempty"`
	Next          *WorkflowMergerRepairState     `json:"next,omitempty"`
	Failure       string                         `json:"failure,omitempty"`
}

// ExecuteWorkflowMergerRepair runs a certified driver at a frozen checkpoint,
// independently verifies its edits and commits only through the trusted Git
// port. It retains accepted/rejected evidence without granting publication.
func (s *WorkflowStore) ExecuteWorkflowMergerRepair(ctx context.Context, request WorkflowMergerRequest, harness WorkflowMergerHarness, provider WorkflowPullRequestProvider, git WorkflowMergerRepairGit) (WorkflowMergerAttempt, error) {
	if request.Input.RepairCheckpointRef == nil || git == nil {
		return WorkflowMergerAttempt{}, ErrWorkflowConflict
	}

	frozen, err := freezeWorkflowMergerRequest(request)
	if err != nil {
		return WorkflowMergerAttempt{}, err
	}

	request = frozen
	if err := s.checkWorkflowMergerRepairRequest(ctx, request.Input, git); err != nil {
		return WorkflowMergerAttempt{}, err
	}

	return s.executeWorkflowMerger(ctx, request, harness, provider, git)
}

func (s *WorkflowStore) acceptWorkflowMerger(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, provider WorkflowPullRequestProvider, git WorkflowMergerRepairGit, attempt *WorkflowMergerAttempt) (err error) {
	acceptance := &WorkflowMergerAcceptance{SchemaVersion: 1, AttemptID: claim.AttemptID, InputHash: attempt.Input.InputHash, CheckpointRef: *attempt.Input.RepairCheckpointRef, Outcome: "rejected", CheckedAt: time.Now().UTC()}

	attempt.Acceptance = acceptance
	defer func() {
		acceptance.CheckedAt = time.Now().UTC()
		if err != nil {
			acceptance.Failure = workflowFailureText(err)
		}
	}()

	if attempt.Proposal == nil || attempt.Proposal.Outcome != "resolved" {
		return nil
	}

	if err := s.guardWorkflowMergerAcceptance(ctx, claim, prepared, provider); err != nil {
		return err
	}

	state, err := claim.store.readWorkflowMergerCheckpoint(attempt.Input)
	if err != nil {
		return err
	}

	if err := git.CheckWorkflowMergerCheckpoint(ctx, state); err != nil {
		return err
	}

	repair, err := s.observeWorkflowMergerRepair(ctx, claim, prepared, provider, state, attempt)
	acceptance.Repair = &repair

	if err != nil {
		return err
	}

	return s.commitWorkflowMergerRepair(ctx, claim, prepared, provider, git, state, repair, acceptance)
}

func (s *WorkflowStore) commitWorkflowMergerRepair(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, provider WorkflowPullRequestProvider, git WorkflowMergerRepairGit, state WorkflowMergerRepairState, repair WorkflowIntegrationTreeRepair, acceptance *WorkflowMergerAcceptance) error {
	if err := importWorkflowMergerRepairTrees(ctx, git, state, prepared.input.Lane.Worktree, repair); err != nil {
		return err
	}

	if err := s.guardWorkflowMergerAcceptance(ctx, claim, prepared, provider); err != nil {
		return err
	}

	next, err := git.ContinueWorkflowMergerRepair(ctx, state, repair)
	if err != nil {
		return fmt.Errorf("host repair commit: %w", err)
	}

	acceptance.Next = &next

	if err := s.guardWorkflowMergerAcceptance(ctx, claim, prepared, provider); err != nil {
		return err
	}

	acceptance.Outcome = "accepted_pending_validation"

	return nil
}

func (s *WorkflowStore) guardWorkflowMergerAcceptance(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, provider WorkflowPullRequestProvider) error {
	if err := s.checkWorkflowMergerPrepared(ctx, claim, prepared, provider); err != nil {
		return err
	}

	if err := claim.store.checkWorkflowPreparedInput(prepared); err != nil {
		return err
	}

	return claim.store.CheckWorkflowMergerEvidence(ctx, *prepared.input.Merger, time.Now().UTC())
}

func (s *WorkflowStore) checkWorkflowMergerRepairRequest(ctx context.Context, input WorkflowMergerInput, git WorkflowMergerRepairGit) error {
	return s.withLock(ctx, func() error {
		_, owner, err := s.locateWorkflowInvocation(ctx, input.InvocationID)
		if err != nil {
			return err
		}

		state, err := owner.readWorkflowMergerCheckpoint(input)
		if err != nil {
			return err
		}

		return git.CheckWorkflowMergerCheckpoint(ctx, state)
	})
}
