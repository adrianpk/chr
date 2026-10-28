package workflow

import (
	"context"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// workflowPreparationHost owns trusted preparation Git IO under the admission guard.
type workflowPreparationHost struct{ store *Store }

func (host workflowPreparationHost) prepareGit(ctx context.Context, run *workflowInvocationFile, request workflowPreparationGitRequest, services workflowPreparationServices) (workflowPreparedIntegration, error) {
	result := workflowPreparedIntegration{Lane: request.Lane}
	preflight := request.Preflight

	query, err := host.store.workflowPreparationQuery(ctx, *run, preflight.Observation, preflight.Source, result.Lane)
	if err != nil {
		return result, err
	}

	query.RequiredSignatures, query.SigningProfile = preflight.Capability.RequiredSignatures, preflight.Capability.SigningProfile

	repairGit, err := host.prepareCandidateGit(ctx, *run, services.Git, query)
	if err != nil {
		return result, err
	}

	if err := workflowintegrate.CheckPrepared(query, repairGit.Proof, workflowIntegrationRepairs{}); err != nil {
		return result, err
	}

	result.Lane.HeadCommit = repairGit.Proof.ResultCommit

	result.Input = workflowstate.PreparationInput{Source: preflight.Source, Capability: preflight.Capability, Git: repairGit.Proof, MergerAttemptRefs: repairGit.Evidence}
	if err := host.store.bindWorkflowNativePreparation(ctx, *run, &result.Input, services.Provider); err != nil {
		return result, err
	}

	if err := host.store.checkWorkflowPreparationGit(ctx, workflowstate.IntegrationCandidate{Input: result.Input, Lane: result.Lane}); err != nil {
		return result, err
	}

	if err := host.store.checkWorkflowPreparationSource(ctx, *run, result.Input, services.Provider); err != nil {
		return result, err
	}

	return result, nil
}

func (host workflowPreparationHost) prepareCandidateGit(ctx context.Context, run workflowInvocationFile, git PreparationGit, query workflowstate.PreparationQuery) (workflowCandidateRepairGit, error) {
	acceptedRepair, err := host.store.acceptedWorkflowCandidateRepair(run)

	state, refs := acceptedRepair.State, acceptedRepair.Refs

	if err != nil {
		return workflowCandidateRepairGit{Proof: workflowstate.IntegrationProof{}, Evidence: nil}, err
	}

	if state == nil {
		proof, err := git.PrepareWorkflowIntegration(ctx, query)

		return workflowCandidateRepairGit{Proof: proof, Evidence: nil}, err
	}

	repaired, ok := git.(RepairedPreparationGit)
	if !ok {
		return workflowCandidateRepairGit{Proof: workflowstate.IntegrationProof{}, Evidence: nil}, workflowstate.ErrConflict
	}

	proof, err := repaired.PrepareWorkflowRepairedIntegration(ctx, query, *state)
	if err == nil && !workflowNativeJSONEqual(proof, *state.Result) {
		err = workflowstate.ErrConflict
	}

	return workflowCandidateRepairGit{Proof: proof, Evidence: refs}, err
}
