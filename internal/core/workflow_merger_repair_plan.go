package core

import (
	"context"
	"slices"
	"time"
)

// PlanWorkflowMergerRepair selects a source-owned semantic repair or requests
// an explicit integration fix or ownership decision. It does not accept edits.
func PlanWorkflowMergerRepair(ctx context.Context, directory string, input WorkflowMergerInput, proof WorkflowIntegrationProof, fix *AgenticReviewFinding) (WorkflowMergerRepairPlan, error) {
	plan := WorkflowMergerRepairPlan{InputHash: input.InputHash}
	if validateWorkflowMergerInput(input, time.Now().UTC()) != nil || validateWorkflowIntegrationProof(proof.WorkflowIntegrationQuery, proof) != nil || !workflowMergerRepairIdentity(input, proof) {
		return plan, ErrWorkflowConflict
	}

	decisions, err := WorkflowMergerRequiredDecisions(input)
	if err != nil {
		return plan, err
	}

	if len(decisions) > 0 {
		plan.Outcome, plan.Decisions = "needs_decision", decisions

		return plan, nil
	}

	reader := workflowIntegrationReader{dir: directory}

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

func workflowMergerRepairIdentity(input WorkflowMergerInput, proof WorkflowIntegrationProof) bool {
	return input.Source == proof.SourceCommit && input.Base == proof.BaseCommit && input.Repository == proof.Repository && input.TargetRef == "refs/heads/"+proof.TargetBranch && input.Method == proof.Method
}

func (reader workflowIntegrationReader) mergerSemanticOwners(ctx context.Context, input WorkflowMergerInput, proof WorkflowIntegrationProof) ([]string, error) {
	var owners []string

	owned := map[string]bool{}

	parent := proof.OriginCommit
	for _, entry := range proof.History {
		data, err := reader.rawOutput(ctx, "diff", "--name-only", "-z", "--no-renames", "--no-ext-diff", "--no-textconv", parent, entry.Source, "--")
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

func workflowMergerOwnsConflict(conflicts []WorkflowMergerConflict, paths []string, owned map[string]bool) bool {
	matched := false

	for _, conflict := range conflicts {
		if slices.Contains(paths, conflict.Path) {
			owned[conflict.ID], matched = true, true
		}
	}

	return matched
}

func workflowMergerSemanticPlan(plan WorkflowMergerRepairPlan, input WorkflowMergerInput, owners []string, fix *AgenticReviewFinding) (WorkflowMergerRepairPlan, error) {
	if len(owners) == 1 && owners[0] != "" {
		plan.Outcome, plan.OriginalCommit = "replay_step", owners[0]

		return plan, nil
	}

	if fix != nil {
		if validateAgenticReviewFinding(*fix) != nil || fix.Category == AgenticReviewFindingCosmetic {
			return plan, ErrWorkflowConflict
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
	plan.Decisions = []WorkflowMergerDecision{{ID: "replay-owner", Issue: "The semantic repair has no unique original replay owner. Select the owning source step or declare an actionable integration fix.", RequirementIDs: requirements, Options: []string{"Identify the original source step", "Declare an integration fix with an actionable finding"}, Authority: "maintainer"}}

	return plan, validateWorkflowMergerDecisions(plan.Decisions, input)
}
