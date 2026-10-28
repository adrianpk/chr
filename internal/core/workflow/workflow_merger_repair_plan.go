package workflow

import (
	"context"
	"slices"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PlanWorkflowMergerRepair selects a source-owned semantic repair or requests
// an explicit integration fix or ownership decision. It does not accept edits.
func PlanWorkflowMergerRepair(ctx context.Context, directory string, input workflowstate.MergerInput, proof workflowstate.IntegrationProof, fix *workflowstate.ReviewFinding) (MergerRepairPlan, error) {
	plan := MergerRepairPlan{InputHash: input.InputHash}
	if workflowmerge.CheckInput(input, time.Now().UTC()) != nil || validateWorkflowIntegrationProof(proof.IntegrationQuery, proof) != nil || !workflowMergerRepairIdentity(input, proof) {
		return plan, workflowstate.ErrConflict
	}

	decisions, err := workflowmerge.RequiredDecisions(input)
	if err != nil {
		return plan, err
	}

	if len(decisions) > 0 {
		plan.Outcome, plan.Decisions = "needs_decision", decisions

		return plan, nil
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: directory}}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return plan, err
	}

	if proof.Method != "rebase" {
		plan.Outcome = "combined"

		return plan, nil
	}

	owners, err := reader.mergerSemanticOwners(ctx, input, proof)
	if err != nil {
		return plan, err
	}

	return workflowMergerSemanticPlan(plan, input, owners, fix)
}

func workflowMergerRepairIdentity(input workflowstate.MergerInput, proof workflowstate.IntegrationProof) bool {
	return input.Source == proof.SourceCommit && input.Base == proof.BaseCommit && input.Repository == proof.Repository && input.TargetRef == "refs/heads/"+proof.TargetBranch && input.Method == proof.Method
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

		paths, err := workflowMergerChangedPaths(data)
		if err != nil {
			return nil, err
		}

		if workflowMergerOwnsConflict(input.Conflicts, paths, owned) {
			owners = append(owners, entry.Source)
		}

		parent = entry.Source
	}

	if len(owned) != len(input.Conflicts) {
		owners = append(owners, "")
	}

	return owners, nil
}

func workflowMergerOwnsConflict(conflicts []workflowstate.MergerConflict, paths []string, owned map[string]bool) bool {
	matched := false

	for _, conflict := range conflicts {
		if slices.Contains(paths, conflict.Path) {
			owned[conflict.ID], matched = true, true
		}
	}

	return matched
}

func workflowMergerSemanticPlan(plan MergerRepairPlan, input workflowstate.MergerInput, owners []string, fix *workflowstate.ReviewFinding) (MergerRepairPlan, error) {
	if answered, found, err := workflowMergerAnsweredPlan(plan, input, owners); found || err != nil {
		return answered, err
	}

	if len(owners) == 1 && owners[0] != "" {
		plan.Outcome, plan.OriginalCommit = "replay_step", owners[0]

		return plan, nil
	}

	if fix != nil {
		if workflowstate.ValidateReviewFinding(*fix) != nil || fix.Category == workflowstate.AgenticReviewFindingCosmetic {
			return plan, workflowstate.ErrConflict
		}

		finding := *fix
		plan.Outcome, plan.IntegrationFix = "integration_fix", &finding

		return plan, nil
	}

	var requirements []string
	for _, requirement := range input.Requirements {
		requirements = append(requirements, requirement.ID)
	}

	plan.Outcome = "needs_decision"
	plan.Decisions = []workflowstate.MergerDecision{{ID: "replay-owner", Issue: "The semantic repair has no unique original replay owner. Select the owning source step or declare an actionable integration fix.", RequirementIDs: requirements, Options: []string{"Identify the original source step", "Declare an integration fix with an actionable finding"}, Authority: "maintainer"}}

	return plan, workflowmerge.CheckDecisions(plan.Decisions, input)
}
