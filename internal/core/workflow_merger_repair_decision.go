package core

import (
	"fmt"
	"slices"
	"strings"
)

// WorkflowMergerRequiredDecisions identifies incompatible intent and unsupported
// procedures before a model call. These questions cannot be answered by time.
func WorkflowMergerRequiredDecisions(input WorkflowMergerInput) ([]WorkflowMergerDecision, error) {
	if err := validateWorkflowMergerIntent(input); err != nil {
		return nil, err
	}

	decisions := workflowMergerIntentDecisions(input)
	for _, conflict := range input.Conflicts {
		if workflowMergerProcedureRequired(conflict) && !workflowMergerProcedureDeclared(input, conflict) {
			decisions = append(decisions, WorkflowMergerDecision{ID: "procedure-" + workflowHash([]byte(conflict.ID))[:16], Issue: fmt.Sprintf("%s conflict at %s has no declared deterministic regeneration and validation procedure.", conflict.Kind, conflict.Path), RequirementIDs: slices.Clone(conflict.RequirementIDs), Options: []string{"Declare a repository-owned resolution and validation procedure", "Leave this conflict unresolved"}, Authority: "maintainer"})
		}
	}

	if len(decisions) > 16 {
		return nil, ErrWorkflowConflict
	}

	decisions = slices.DeleteFunc(decisions, func(d WorkflowMergerDecision) bool {
		return input.OperatorDecisionRef != nil && strings.HasPrefix(d.ID, "intent-") && workflowDecisionResolved(input.OperatorAnswers, d.ID)
	})

	return decisions, validateWorkflowMergerDecisions(decisions, input)
}

func workflowMergerIntentDecisions(input WorkflowMergerInput) []WorkflowMergerDecision {
	var decisions []WorkflowMergerDecision

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
			decisions = append(decisions, WorkflowMergerDecision{ID: "intent-" + workflowHash([]byte(key))[:16], Issue: fmt.Sprintf("Requirements %s and %s declare incompatible intent; preserving both requires a product or ownership decision.", pair[0], pair[1]), RequirementIDs: pair, Options: []string{"Provide a compatible requirement interpretation", "Approve a scoped requirement change"}, Authority: "maintainer"})
		}
	}

	return decisions
}

func workflowMergerProcedureRequired(conflict WorkflowMergerConflict) bool {
	return conflict.Kind != "text" && conflict.Kind != "semantic"
}

func workflowMergerProcedureDeclared(input WorkflowMergerInput, conflict WorkflowMergerConflict) bool {
	return validWorkflowMergerCommand(conflict.RegenerationCommand) && validWorkflowMergerCommand(conflict.ValidationCommand) && slices.ContainsFunc(input.ValidationCommands, func(command []string) bool { return slices.Equal(command, conflict.ValidationCommand) })
}

func validateWorkflowMergerIntentConflicts(requirements []WorkflowMergerRequirement) error {
	for _, requirement := range requirements {
		if len(requirement.ConflictsWith) == 0 {
			continue
		}

		if slices.Contains(requirement.ConflictsWith, requirement.ID) {
			return ErrWorkflowConflict
		}

		if err := validateWorkflowMergerRequirementIDs(requirement.ConflictsWith, requirements); err != nil {
			return err
		}
	}

	return nil
}
