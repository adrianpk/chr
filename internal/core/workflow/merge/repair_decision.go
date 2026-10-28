package merge

import (
	"fmt"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MergerRequiredDecisions identifies incompatible intent and unsupported
// procedures before a model call. These questions cannot be answered by time.
func MergerRequiredDecisions(input workflowstate.MergerInput) ([]workflowstate.MergerDecision, error) {
	if err := validateWorkflowMergerIntent(input); err != nil {
		return nil, err
	}

	decisions := workflowMergerIntentDecisions(input)
	for _, conflict := range input.Conflicts {
		if WorkflowMergerProcedureRequired(conflict) && !workflowMergerProcedureDeclared(input, conflict) {
			decisions = append(decisions, workflowstate.MergerDecision{ID: "procedure-" + contenthash.Sum([]byte(conflict.ID))[:16], Issue: fmt.Sprintf("%s conflict at %s has no declared deterministic regeneration and validation procedure.", conflict.Kind, conflict.Path), RequirementIDs: slices.Clone(conflict.RequirementIDs), Options: []string{"Declare a repository-owned resolution and validation procedure", "Leave this conflict unresolved"}, Authority: "maintainer"})
		}
	}

	if len(decisions) > 16 {
		return nil, workflowstate.ErrConflict
	}

	decisions = slices.DeleteFunc(decisions, func(d workflowstate.MergerDecision) bool {
		return input.OperatorDecisionRef != nil && strings.HasPrefix(d.ID, "intent-") && WorkflowDecisionResolved(input.OperatorAnswers, d.ID)
	})

	return decisions, ValidateWorkflowMergerDecisions(decisions, input)
}

func validateWorkflowMergerIntentConflicts(requirements []workflowstate.MergerRequirement) error {
	for _, requirement := range requirements {
		if len(requirement.ConflictsWith) == 0 {
			continue
		}

		if slices.Contains(requirement.ConflictsWith, requirement.ID) {
			return workflowstate.ErrConflict
		}

		if err := validateWorkflowMergerRequirementIDs(requirement.ConflictsWith, requirements); err != nil {
			return err
		}
	}

	return nil
}

func workflowMergerIntentDecisions(input workflowstate.MergerInput) []workflowstate.MergerDecision {
	var decisions []workflowstate.MergerDecision

	seen := map[string]bool{}

	for _, requirement := range input.Requirements {
		for _, other := range requirement.ConflictsWith {
			pair := []string{requirement.ID, other}
			slices.Sort(pair)

			key := pair[0] + ":" + pair[1]
			if seen[key] {
				continue
			}

			seen[key] = true
			decisions = append(decisions, workflowstate.MergerDecision{ID: "intent-" + contenthash.Sum([]byte(key))[:16], Issue: fmt.Sprintf("Requirements %s and %s declare incompatible intent; preserving both requires a product or ownership decision.", pair[0], pair[1]), RequirementIDs: pair, Options: []string{"Provide a compatible requirement interpretation", "Approve a scoped requirement change"}, Authority: "maintainer"})
		}
	}

	return decisions
}

func workflowMergerProcedureDeclared(input workflowstate.MergerInput, conflict workflowstate.MergerConflict) bool {
	return ValidCommand(conflict.RegenerationCommand) && ValidCommand(conflict.ValidationCommand) && slices.ContainsFunc(input.ValidationCommands, func(command []string) bool { return slices.Equal(command, conflict.ValidationCommand) })
}

// WorkflowMergerProcedureRequired identifies conflicts that require a declared deterministic procedure.
func WorkflowMergerProcedureRequired(conflict workflowstate.MergerConflict) bool {
	return conflict.Kind != "text" && conflict.Kind != "semantic"
}
