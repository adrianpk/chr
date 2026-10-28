package workflow

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MergerRepairGit is a trusted host port. Model tools cannot invoke it.
type MergerRepairGit interface {
	CheckWorkflowMergerCheckpoint(context.Context, MergerRepairState) error
	ImportWorkflowMergerTree(context.Context, MergerRepairState, string, string) error
	ContinueWorkflowMergerRepair(context.Context, MergerRepairState, workflowstate.IntegrationTreeRepair) (MergerRepairState, error)
}

// MergerCheckpointRecord binds a private replay checkpoint to the queue
// generation that admitted it. Its immutable image stays outside model tools.
type MergerCheckpointRecord struct {
	SchemaVersion     int               `json:"schema_version"`
	InvocationID      string            `json:"invocation_id"`
	EntryID           string            `json:"entry_id"`
	CandidateRevision uint64            `json:"candidate_revision"`
	ClaimGeneration   uint64            `json:"claim_generation"`
	CandidateHash     string            `json:"candidate_hash"`
	ScopeHash         string            `json:"scope_hash"`
	State             MergerRepairState `json:"state"`
}

// RetainWorkflowMergerCheckpoint accepts only a host-prepared, independently
// verified private checkpoint. It creates neither an agent call nor approval.
func (s *Store) RetainWorkflowMergerCheckpoint(ctx context.Context, guard IntegrationBudgetGuard, state MergerRepairState, git MergerRepairGit) (ref workflowstate.EvidenceRef, err error) {
	if git == nil {
		return ref, workflowstate.ErrConflict
	}

	state, err = freezeWorkflowMergerCheckpoint(state)
	if err != nil {
		return ref, err
	}

	err = s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if err != nil {
			return err
		}

		budgetObservation, err := invocationLocation.Owner.currentWorkflowIntegrationBudget(ctx, invocationLocation.Run, guard)

		entry, budget := budgetObservation.Entry, budgetObservation.Budget

		if err != nil {
			return err
		}

		if err := guardWorkflowIntegrationBudget(invocationLocation.Run, entry, budget, guard); err != nil {
			return err
		}

		if !workflowMergerRunnable(invocationLocation.Run, entry) {
			return ErrWorkflowOwned
		}

		if err := invocationLocation.Owner.checkWorkflowMergerCheckpointOwner(invocationLocation.Run, entry, state); err != nil {
			return err
		}

		if err := git.CheckWorkflowMergerCheckpoint(ctx, state); err != nil {
			return err
		}

		record := MergerCheckpointRecord{1, invocationLocation.Run.ID, entry.ID, entry.StateRevision, entry.ClaimGeneration, workflowMergerCandidateHash(invocationLocation.Run, entry), entry.FrozenScopeHash, state}
		ref, err = invocationLocation.Owner.persistWorkflowMergerCheckpoint(record)

		return err
	})

	return ref, err
}

func workflowMergerCandidateHash(run workflowInvocationFile, entry IntegrationEntry) string {
	if entry.CandidateRef != nil {
		return entry.CandidateRef.Hash
	}

	return run.CandidateHash
}

func (s *Store) readWorkflowMergerCheckpoint(input workflowstate.MergerInput) (MergerRepairState, error) {
	var record MergerCheckpointRecord
	if input.RepairCheckpointRef == nil {
		return record.State, workflowstate.ErrConflict
	}

	if err := s.readWorkflowNativeEvidence(input.InvocationID, *input.RepairCheckpointRef, &record); err != nil {
		return record.State, err
	}

	if !workflowMergerCheckpointMatches(record, input) {
		return record.State, workflowstate.ErrConflict
	}

	return record.State, checkWorkflowRepairPredecessors(record.State, input)
}

func (s *Store) checkWorkflowMergerCheckpointOwner(run workflowInvocationFile, entry IntegrationEntry, state MergerRepairState) error {
	policy, err := s.workflowReviewPolicy(run)
	if err != nil || run.Lane == nil {
		return workflowstate.ErrConflict
	}

	q := state.Query
	expected := [5]string{entry.Repository, strings.TrimPrefix(entry.TargetRef, "refs/heads/"), run.Lane.HeadCommit, run.Lane.BaseCommit, policy.Method}

	actual := [5]string{q.Repository, q.TargetBranch, q.SourceCommit, q.OriginCommit, q.Method}
	if actual != expected || q.Lane.Unit != entry.Unit || q.Lane.Worktree == run.Lane.Worktree {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowMergerCheckpointMatches(record MergerCheckpointRecord, input workflowstate.MergerInput) bool {
	expected := [5]string{input.InvocationID, input.EntryID, input.CandidateHash, input.ScopeHash, input.Repository}

	actual := [5]string{record.InvocationID, record.EntryID, record.CandidateHash, record.ScopeHash, record.State.Query.Repository}
	if record.SchemaVersion != 1 || expected != actual || record.CandidateRevision != input.CandidateRevision || record.ClaimGeneration != input.ClaimGeneration {
		return false
	}

	q := record.State.Query

	return [4]string{q.SourceCommit, q.BaseCommit, q.Method, q.TargetBranch} == [4]string{input.Source, input.Base, input.Method, strings.TrimPrefix(input.TargetRef, "refs/heads/")} && q.Lane.Unit == input.Unit
}

func (s *Store) persistWorkflowMergerCheckpoint(record MergerCheckpointRecord) (workflowstate.EvidenceRef, error) {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return workflowstate.EvidenceRef{}, workflowstate.ErrConflict
	}

	data = append(data, '\n')
	ref := workflowstate.EvidenceRef{Path: filepath.ToSlash(filepath.Join("run", record.InvocationID, "extension", "chr", "merger", "checkpoints", contenthash.Sum(data)+".json")), Hash: contenthash.Sum(data)}

	return ref, s.commitFiles([]workflowstate.FileChange{{Path: ref.Path, After: data}})
}

func freezeWorkflowMergerCheckpoint(state MergerRepairState) (MergerRepairState, error) {
	data, err := json.Marshal(state)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return MergerRepairState{}, workflowstate.ErrConflict
	}

	var frozen MergerRepairState

	err = json.Unmarshal(data, &frozen)

	return frozen, err
}
