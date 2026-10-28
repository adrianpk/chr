package merge

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// RepairMatches compares the immutable source, target and method of a repair.
func RepairMatches(input workflowstate.MergerInput, proof workflowstate.IntegrationProof) bool {
	return input.Source == proof.SourceCommit && input.Base == proof.BaseCommit && input.Repository == proof.Repository && input.TargetRef == "refs/heads/"+proof.TargetBranch && input.Method == proof.Method
}

// OwnsConflict records which retained conflicts belong to the observed changed paths.
func OwnsConflict(conflicts []workflowstate.MergerConflict, paths []string, owned map[string]bool) bool {
	matched := false

	for _, conflict := range conflicts {
		if slices.Contains(paths, conflict.Path) {
			owned[conflict.ID], matched = true, true
		}
	}

	return matched
}

// PlanRepair selects an explicit replay owner or actionable fix and otherwise requests a decision.
func PlanRepair(plan workflowstate.MergerRepairPlan, input workflowstate.MergerInput, owners []string, fix *workflowstate.ReviewFinding) (workflowstate.MergerRepairPlan, error) {
	if answeredRepair, err := answeredPlan(plan, input, owners); answeredRepair.Found || err != nil {
		return answeredRepair.Plan, err
	}

	if len(owners) == 1 && owners[0] != "" {
		plan.Outcome, plan.OriginalCommit = workflowstate.MergerRepairReplayStep, owners[0]

		return plan, nil
	}

	if fix != nil {
		if workflowstate.ValidateReviewFinding(*fix) != nil || fix.Category == workflowstate.AgenticReviewFindingCosmetic {
			return plan, workflowstate.ErrConflict
		}

		finding := *fix
		plan.Outcome, plan.IntegrationFix = workflowstate.MergerRepairIntegrationFix, &finding

		return plan, nil
	}

	var requirements []string
	for _, requirement := range input.Requirements {
		requirements = append(requirements, requirement.ID)
	}

	plan.Outcome = workflowstate.MergerRepairNeedsDecision
	plan.Decisions = []workflowstate.MergerDecision{{ID: "replay-owner", Issue: "The semantic repair has no unique original replay owner. Select the owning source step or declare an actionable integration fix.", RequirementIDs: requirements, Options: []string{"Identify the original source step", "Declare an integration fix with an actionable finding"}, Authority: "maintainer"}}

	return plan, CheckDecisions(plan.Decisions, input)
}

func answeredPlan(plan workflowstate.MergerRepairPlan, input workflowstate.MergerInput, owners []string) (answeredRepair, error) {
	if input.OperatorDecisionRef == nil {
		return answeredRepair{Plan: plan, Found: false}, nil
	}

	for i := len(input.OperatorAnswers) - 1; i >= 0; i-- {
		answer := input.OperatorAnswers[i]
		if answer.DecisionID != "replay-owner" {
			continue
		}

		if answer.Action != "resolve" {
			return answeredRepair{Plan: plan, Found: false}, nil
		}

		if answer.OriginalCommit != "" {
			if !slices.Contains(owners, answer.OriginalCommit) {
				return answeredRepair{Plan: plan, Found: false}, workflowstate.ErrConflict
			}

			plan.Outcome, plan.OriginalCommit = workflowstate.MergerRepairReplayStep, answer.OriginalCommit

			return answeredRepair{Plan: plan, Found: true}, nil
		}

		if answer.IntegrationFix == nil || workflowstate.ValidateReviewFinding(*answer.IntegrationFix) != nil || answer.IntegrationFix.Category == workflowstate.AgenticReviewFindingCosmetic {
			return answeredRepair{Plan: plan, Found: false}, workflowstate.ErrConflict
		}

		finding := *answer.IntegrationFix
		plan.Outcome, plan.IntegrationFix = workflowstate.MergerRepairIntegrationFix, &finding

		return answeredRepair{Plan: plan, Found: true}, nil
	}

	return answeredRepair{Plan: plan, Found: false}, nil
}

// answeredRepair retains an explicit repair answer without granting acceptance.
type answeredRepair struct {
	Plan  workflowstate.MergerRepairPlan
	Found bool
}
