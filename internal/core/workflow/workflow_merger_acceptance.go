package workflow

import (
	"context"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MergerAcceptanceOutcome is the host-observed repair acceptance state.
// Accepted repairs still require the exact candidate validation gate.
type MergerAcceptanceOutcome string

// Host acceptance outcomes preserve their existing retained strings.
const (
	MergerRejected                  MergerAcceptanceOutcome = "rejected"
	MergerAcceptedPendingValidation MergerAcceptanceOutcome = "accepted_pending_validation"
)

// MergerAcceptance records host observations, never a validation or
// review grant. Accepted checkpoints still require the complete candidate gate.
type MergerAcceptance struct {
	SchemaVersion int                                  `json:"schema_version"`
	AttemptID     string                               `json:"attempt_id"`
	InputHash     string                               `json:"input_hash"`
	CheckpointRef workflowstate.EvidenceRef            `json:"checkpoint_ref"`
	Outcome       MergerAcceptanceOutcome              `json:"outcome"`
	CheckedAt     time.Time                            `json:"checked_at"`
	Repair        *workflowstate.IntegrationTreeRepair `json:"repair,omitempty"`
	Next          *MergerRepairState                   `json:"next,omitempty"`
	Failure       string                               `json:"failure,omitempty"`
}

// MergerRepairPorts binds the provider fence and trusted repair Git capability.
type MergerRepairPorts struct {
	Provider PullRequestProvider
	Git      MergerRepairGit
}

// ExecuteMergerRepair runs a certified driver at a frozen checkpoint,
// independently verifies its edits and commits only through the trusted Git
// port. It retains accepted/rejected evidence without granting publication.
func ExecuteMergerRepair(ctx context.Context, store *Store, request MergerRequest, harness MergerHarness, ports MergerRepairPorts) (MergerAttempt, error) {
	if request.Input.RepairCheckpointRef == nil || ports.Git == nil {
		return MergerAttempt{}, workflowstate.ErrConflict
	}

	frozen, err := freezeWorkflowMergerRequest(request)
	if err != nil {
		return MergerAttempt{}, err
	}

	request = frozen
	if err := store.checkWorkflowMergerRepairRequest(ctx, request.Input, ports.Git); err != nil {
		return MergerAttempt{}, err
	}

	return (workflowMergerHost{store: store}).execute(ctx, request, harness, ports.Provider, ports.Git)
}

func (host workflowMergerHost) accept(ctx context.Context, claimed workflowClaimedMerger, provider PullRequestProvider, git MergerRepairGit, attempt *MergerAttempt) (err error) {
	claim := claimed.Claim
	acceptance := &MergerAcceptance{SchemaVersion: 1, AttemptID: claim.AttemptID, InputHash: attempt.Input.InputHash, CheckpointRef: *attempt.Input.RepairCheckpointRef, Outcome: MergerRejected, CheckedAt: time.Now().UTC()}

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

	if err := host.store.guardWorkflowMergerAcceptance(ctx, claimed, provider); err != nil {
		return err
	}

	state, err := claim.store.readWorkflowMergerCheckpoint(attempt.Input)
	if err != nil {
		return err
	}

	if err := git.CheckWorkflowMergerCheckpoint(ctx, state); err != nil {
		return err
	}

	repair, err := host.observeRepair(ctx, claimed, provider, state, attempt)
	acceptance.Repair = &repair

	if err != nil {
		return err
	}

	return host.commitRepair(ctx, claimed, provider, git, workflowMergerRepairCommit{State: state, Repair: repair, Acceptance: acceptance})
}

func (host workflowMergerHost) commitRepair(ctx context.Context, claimed workflowClaimedMerger, provider PullRequestProvider, git MergerRepairGit, commit workflowMergerRepairCommit) error {
	state, repair, acceptance := commit.State, commit.Repair, commit.Acceptance

	prepared := claimed.Prepared
	if err := importWorkflowMergerRepairTrees(ctx, git, state, prepared.input.Lane.Worktree, repair); err != nil {
		return err
	}

	if err := host.store.guardWorkflowMergerAcceptance(ctx, claimed, provider); err != nil {
		return err
	}

	next, err := git.ContinueWorkflowMergerRepair(ctx, state, repair)
	if err != nil {
		return fmt.Errorf("host repair commit: %w", err)
	}

	acceptance.Next = &next

	if err := host.store.guardWorkflowMergerAcceptance(ctx, claimed, provider); err != nil {
		return err
	}

	acceptance.Outcome = MergerAcceptedPendingValidation

	return nil
}

func (s *Store) guardWorkflowMergerAcceptance(ctx context.Context, claimed workflowClaimedMerger, provider PullRequestProvider) error {
	claim, prepared := claimed.Claim, claimed.Prepared
	if err := s.checkWorkflowMergerPrepared(ctx, claimed, provider); err != nil {
		return err
	}

	if err := claim.store.checkWorkflowPreparedInput(prepared); err != nil {
		return err
	}

	return claim.store.CheckWorkflowMergerEvidence(ctx, *prepared.input.Merger, time.Now().UTC())
}

func (s *Store) checkWorkflowMergerRepairRequest(ctx context.Context, input workflowstate.MergerInput, git MergerRepairGit) error {
	return s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, input.InvocationID)
		if err != nil {
			return err
		}

		state, err := invocationLocation.Owner.readWorkflowMergerCheckpoint(input)
		if err != nil {
			return err
		}

		return git.CheckWorkflowMergerCheckpoint(ctx, state)
	})
}
