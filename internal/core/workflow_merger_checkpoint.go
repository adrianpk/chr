package core

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowMergerRepairGit is a trusted host port. Model tools cannot invoke it.
type WorkflowMergerRepairGit interface {
	CheckWorkflowMergerCheckpoint(context.Context, WorkflowMergerRepairState) error
	ImportWorkflowMergerTree(context.Context, WorkflowMergerRepairState, string, string) error
	ContinueWorkflowMergerRepair(context.Context, WorkflowMergerRepairState, WorkflowIntegrationTreeRepair) (WorkflowMergerRepairState, error)
}

// WorkflowMergerCheckpointRecord binds a private replay checkpoint to the queue
// generation that admitted it. Its immutable image stays outside model tools.
type WorkflowMergerCheckpointRecord struct {
	SchemaVersion     int                       `json:"schema_version"`
	InvocationID      string                    `json:"invocation_id"`
	EntryID           string                    `json:"entry_id"`
	CandidateRevision uint64                    `json:"candidate_revision"`
	ClaimGeneration   uint64                    `json:"claim_generation"`
	CandidateHash     string                    `json:"candidate_hash"`
	ScopeHash         string                    `json:"scope_hash"`
	State             WorkflowMergerRepairState `json:"state"`
}

// RetainWorkflowMergerCheckpoint accepts only a host-prepared, independently
// verified private checkpoint. It creates neither an agent call nor approval.
func (s *WorkflowStore) RetainWorkflowMergerCheckpoint(ctx context.Context, guard WorkflowIntegrationBudgetGuard, state WorkflowMergerRepairState, git WorkflowMergerRepairGit) (ref WorkflowEvidenceRef, err error) {
	if git == nil {
		return ref, workflowstate.ErrConflict
	}

	state, err = freezeWorkflowMergerCheckpoint(state)
	if err != nil {
		return ref, err
	}

	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if err != nil {
			return err
		}

		entry, budget, err := owner.currentWorkflowIntegrationBudget(ctx, run, guard)
		if err != nil {
			return err
		}

		if err := guardWorkflowIntegrationBudget(run, entry, budget, guard); err != nil {
			return err
		}

		if !workflowMergerRunnable(run, entry) {
			return ErrWorkflowOwned
		}

		if err := owner.checkWorkflowMergerCheckpointOwner(run, entry, state); err != nil {
			return err
		}

		if err := git.CheckWorkflowMergerCheckpoint(ctx, state); err != nil {
			return err
		}

		record := WorkflowMergerCheckpointRecord{1, run.ID, entry.ID, entry.StateRevision, entry.ClaimGeneration, workflowMergerCandidateHash(run, entry), entry.FrozenScopeHash, state}
		ref, err = owner.persistWorkflowMergerCheckpoint(record)

		return err
	})

	return ref, err
}

func workflowMergerCandidateHash(run workflowInvocationFile, entry WorkflowIntegrationEntry) string {
	if entry.CandidateRef != nil {
		return entry.CandidateRef.Hash
	}

	return run.CandidateHash
}

func (s *WorkflowStore) readWorkflowMergerCheckpoint(input WorkflowMergerInput) (WorkflowMergerRepairState, error) {
	var record WorkflowMergerCheckpointRecord
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

func (s *WorkflowStore) checkWorkflowMergerCheckpointOwner(run workflowInvocationFile, entry WorkflowIntegrationEntry, state WorkflowMergerRepairState) error {
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

func workflowMergerCheckpointMatches(record WorkflowMergerCheckpointRecord, input WorkflowMergerInput) bool {
	expected := [5]string{input.InvocationID, input.EntryID, input.CandidateHash, input.ScopeHash, input.Repository}

	actual := [5]string{record.InvocationID, record.EntryID, record.CandidateHash, record.ScopeHash, record.State.Query.Repository}
	if record.SchemaVersion != 1 || expected != actual || record.CandidateRevision != input.CandidateRevision || record.ClaimGeneration != input.ClaimGeneration {
		return false
	}

	q := record.State.Query

	return [4]string{q.SourceCommit, q.BaseCommit, q.Method, q.TargetBranch} == [4]string{input.Source, input.Base, input.Method, strings.TrimPrefix(input.TargetRef, "refs/heads/")} && q.Lane.Unit == input.Unit
}

func (s *WorkflowStore) persistWorkflowMergerCheckpoint(record WorkflowMergerCheckpointRecord) (WorkflowEvidenceRef, error) {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return WorkflowEvidenceRef{}, workflowstate.ErrConflict
	}

	data = append(data, '\n')
	ref := WorkflowEvidenceRef{Path: filepath.ToSlash(filepath.Join("run", record.InvocationID, "extension", "chr", "merger", "checkpoints", contenthash.Sum(data)+".json")), Hash: contenthash.Sum(data)}

	return ref, s.commitFiles([]workflowFileChange{{Path: ref.Path, After: data}})
}

func freezeWorkflowMergerCheckpoint(state WorkflowMergerRepairState) (WorkflowMergerRepairState, error) {
	data, err := json.Marshal(state)
	if err != nil || len(data) >= WorkflowFileLimit {
		return WorkflowMergerRepairState{}, workflowstate.ErrConflict
	}

	var frozen WorkflowMergerRepairState

	err = json.Unmarshal(data, &frozen)

	return frozen, err
}
