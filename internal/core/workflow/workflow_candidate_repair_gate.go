package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowCandidateRepairGit struct {
	Proof    workflowstate.IntegrationProof
	Evidence []workflowstate.EvidenceRef
}

func (s *Store) resumeWorkflowCandidateRepairGate(ctx context.Context, run *workflowInvocationFile, source workflowstate.ReviewInput) error {
	if run.Integration == nil {
		return nil
	}

	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil || entry.State != "merging" {
		return err
	}

	acceptedRepair, err := s.acceptedWorkflowCandidateRepair(*run)

	state, _ := acceptedRepair.State, acceptedRepair.Refs

	if err != nil || state == nil || !workflowMergerRepairSourceMatches(source, *state.Result) {
		return errorsOrConflict(err)
	}

	request := IntegrationTransition{EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, State: "validating"}
	if entry.CandidateRef != nil {
		request.CandidateHash = entry.CandidateRef.Hash
	}

	if err := s.checkWorkflowCandidateRepairTransition(ctx, *run, entry, request); err != nil {
		return err
	}

	return s.commitWorkflowCandidateRepairGate(run, entry, request)
}

func (s *Store) commitWorkflowCandidateRepairGate(run *workflowInvocationFile, entry IntegrationEntry, request IntegrationTransition) error {
	if run.Preparation != nil {
		run.Preparation.State = "stale"
	}

	if run.Review != nil {
		run.Review.State = workflowstate.ReviewStale
	}

	updateWorkflowIntegrationEntry(&entry, request)
	projectWorkflowIntegrationPhase(run, entry)

	if err := s.commitWorkflowIntegrationEntry(run, entry, run.Integration.EntryRef.Hash); err != nil {
		return err
	}

	var err error

	*run, err = s.readWorkflowInvocation(run.ID)

	return err
}

func workflowMergerRepairSourceMatches(source workflowstate.ReviewInput, proof workflowstate.IntegrationProof) bool {
	return source.PullRequest.HeadCommit == proof.SourceCommit && source.TargetCommit == proof.BaseCommit && source.Policy.Method == proof.Method && source.PullRequest.Repository == proof.Repository && source.PullRequest.BaseBranch == proof.TargetBranch
}
