package workflow

import (
	"context"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) retainWorkflowCandidateRepair(ctx context.Context, run *workflowInvocationFile, candidate workflowstate.IntegrationCandidate, reason string, refs []workflowstate.EvidenceRef) error {
	if run.Integration == nil || candidate.Input.Native == nil || len(refs) == 0 {
		return workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return err
	}

	if !workflowCandidateRepairOwnerMatches(*run, entry, candidate) {
		return workflowstate.ErrConflict
	}

	request := IntegrationTransition{EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, CandidateHash: workflowMergerCandidateHash(*run, entry), State: "merging"}
	if entry.CandidateRef == nil {
		request.CandidateHash = ""
	}

	if err := s.checkWorkflowCandidateRepairTransition(ctx, *run, entry, request); err != nil {
		return err
	}

	record := CandidateRepair{Schema: "chr-candidate-repair/v1", Reason: reason, Candidate: candidate, FindingRefs: refs}

	file, ref, err := workflowNativePreparationEvidenceFile(run.ID, record)
	if err != nil {
		return err
	}

	run.CandidateRepair = &CandidateRepairBinding{Unit: entry.Unit, EntryID: entry.ID, ClaimGeneration: entry.ClaimGeneration, CandidateHash: workflowMergerCandidateHash(*run, entry), FindingRef: refs[0], RecordRef: ref}
	before := run.Integration.EntryRef.Hash

	updateWorkflowIntegrationEntry(&entry, request)

	run.Status, run.Outcome = workflowstate.InvocationWaitingIntegration, "candidate_repair_required"
	run.Blocker = &Blocker{Dependency: entry.Unit, Reason: reason, NextActionOwner: "workflow_merger"}

	queue, err := workflowIntegrationEntryChange(entry, before)
	if err != nil {
		return err
	}

	updateWorkflowIntegrationBinding(run, workflowstate.IntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: workflowstate.EvidenceRef{Path: queue.Path, Hash: contenthash.Sum(queue.After)}})

	return s.commitWorkflowPullRequestFiles(run, []workflowstate.FileChange{queue}, file)
}

func (s *Store) checkWorkflowCandidateRepairTransition(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, request IntegrationTransition) error {
	if err := guardWorkflowIntegrationTransition(entry, request); err != nil {
		return err
	}

	if err := s.checkWorkflowIntegrationTransition(ctx, run, entry, request); err != nil {
		return err
	}

	return s.guardWorkflowPreparationTransition(ctx, run, request)
}

func (s *Store) retainWorkflowFailedCandidate(ctx context.Context, run workflowInvocationFile, candidate workflowstate.IntegrationCandidate) error {
	current, err := s.readWorkflowInvocation(run.ID)
	if err != nil {
		return err
	}

	last := current.Attempts[len(current.Attempts)-1]
	ref := workflowstate.EvidenceRef{Path: last.CheckpointRef, Hash: last.CheckpointHash}

	return s.retainWorkflowCandidateRepair(ctx, &current, candidate, "combined_gate_regression", []workflowstate.EvidenceRef{ref})
}

func workflowCandidateRepairOwnerMatches(run workflowInvocationFile, entry IntegrationEntry, candidate workflowstate.IntegrationCandidate) bool {
	return entry.ClaimGeneration == candidate.Input.Native.ClaimGeneration && entry.StopIntent == "" && run.StopIntent == ""
}
