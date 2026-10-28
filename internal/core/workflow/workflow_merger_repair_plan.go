package workflow

import (
	"context"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PlanWorkflowMergerRepair selects a source-owned semantic repair or requests
// an explicit integration fix or ownership decision. It does not accept edits.
func PlanWorkflowMergerRepair(ctx context.Context, directory string, input workflowstate.MergerInput, proof workflowstate.IntegrationProof, fix *workflowstate.ReviewFinding) (workflowstate.MergerRepairPlan, error) {
	plan := workflowstate.MergerRepairPlan{InputHash: input.InputHash}
	if workflowmerge.CheckInput(input, time.Now().UTC()) != nil || workflowintegrate.CheckProof(proof.IntegrationQuery, proof, workflowIntegrationRepairs{}) != nil || !workflowmerge.RepairMatches(input, proof) {
		return plan, workflowstate.ErrConflict
	}

	decisions, err := workflowmerge.RequiredDecisions(input)
	if err != nil {
		return plan, err
	}

	if len(decisions) > 0 {
		plan.Outcome, plan.Decisions = workflowstate.MergerRepairNeedsDecision, decisions

		return plan, nil
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: directory}}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return plan, err
	}

	if proof.Method != "rebase" {
		plan.Outcome = workflowstate.MergerRepairCombined

		return plan, nil
	}

	owners, err := reader.mergerSemanticOwners(ctx, input, proof)
	if err != nil {
		return plan, err
	}

	return workflowmerge.PlanRepair(plan, input, owners, fix)
}

func (reader workflowIntegrationReader) mergerSemanticOwners(ctx context.Context, input workflowstate.MergerInput, proof workflowstate.IntegrationProof) ([]string, error) {
	var owners []string

	owned := map[string]bool{}

	parent := proof.OriginCommit
	for _, entry := range proof.History {
		data, err := reader.RawOutput(ctx, "diff", "--name-only", "-z", "--no-renames", "--no-ext-diff", "--no-textconv", parent, entry.Source, "--")
		if err != nil {
			return nil, err
		}

		paths, err := workflowmerge.ChangedPaths(data)
		if err != nil {
			return nil, err
		}

		if workflowmerge.OwnsConflict(input.Conflicts, paths, owned) {
			owners = append(owners, entry.Source)
		}

		parent = entry.Source
	}

	if len(owned) != len(input.Conflicts) {
		owners = append(owners, "")
	}

	return owners, nil
}
