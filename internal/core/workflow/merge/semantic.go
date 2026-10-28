package merge

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowMergerRepairIdentity checks the retained repair observations without executing Git or accepting a candidate.
func WorkflowMergerRepairIdentity(input workflowstate.MergerInput, proof workflowstate.IntegrationProof) bool {
	return input.Source == proof.SourceCommit && input.Base == proof.BaseCommit && input.Repository == proof.Repository && input.TargetRef == "refs/heads/"+proof.TargetBranch && input.Method == proof.Method
}

// WorkflowMergerOwnsConflict checks the retained repair observations without executing Git or accepting a candidate.
func WorkflowMergerOwnsConflict(conflicts []workflowstate.MergerConflict, paths []string, owned map[string]bool) bool {
	matched := false

	for _, conflict := range conflicts {
		if slices.Contains(paths, conflict.Path) {
			owned[conflict.ID], matched = true, true
		}
	}

	return matched
}

// WorkflowMergerSemanticPlan checks the retained repair observations without executing Git or accepting a candidate.
func WorkflowMergerSemanticPlan(plan workflowstate.MergerRepairPlan, input workflowstate.MergerInput, owners []string, fix *workflowstate.ReviewFinding) (workflowstate.MergerRepairPlan, error) {
	if answeredRepair, err := workflowMergerAnsweredPlan(plan, input, owners); answeredRepair.Found || err != nil {
		return answeredRepair.Plan, err
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

	return plan, CheckDecisions(plan.Decisions, input)
}

func workflowMergerAnsweredPlan(plan workflowstate.MergerRepairPlan, input workflowstate.MergerInput, owners []string) (workflowAnsweredRepair, error) {
	if input.OperatorDecisionRef == nil {
		return workflowAnsweredRepair{Plan: plan, Found: false}, nil
	}

	for i := len(input.OperatorAnswers) - 1; i >= 0; i-- {
		answer := input.OperatorAnswers[i]
		if answer.DecisionID != "replay-owner" {
			continue
		}

		if answer.Action != "resolve" {
			return workflowAnsweredRepair{Plan: plan, Found: false}, nil
		}

		if answer.OriginalCommit != "" {
			if !slices.Contains(owners, answer.OriginalCommit) {
				return workflowAnsweredRepair{Plan: plan, Found: false}, workflowstate.ErrConflict
			}

			plan.Outcome, plan.OriginalCommit = "replay_step", answer.OriginalCommit

			return workflowAnsweredRepair{Plan: plan, Found: true}, nil
		}

		if answer.IntegrationFix == nil || workflowstate.ValidateReviewFinding(*answer.IntegrationFix) != nil || answer.IntegrationFix.Category == workflowstate.AgenticReviewFindingCosmetic {
			return workflowAnsweredRepair{Plan: plan, Found: false}, workflowstate.ErrConflict
		}

		finding := *answer.IntegrationFix
		plan.Outcome, plan.IntegrationFix = "integration_fix", &finding

		return workflowAnsweredRepair{Plan: plan, Found: true}, nil
	}

	return workflowAnsweredRepair{Plan: plan, Found: false}, nil
}

// workflowAnsweredRepair retains an explicit repair answer without granting acceptance.
type workflowAnsweredRepair struct {
	Plan  workflowstate.MergerRepairPlan
	Found bool
}
