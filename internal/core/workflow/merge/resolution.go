package merge

import (
	"slices"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validWorkflowMergerOptions(options []string) bool {
	if len(options) < 2 || len(options) > 5 {
		return false
	}

	seen := map[string]bool{}
	for _, option := range options {
		if !ValidWorkflowMergerText(option) || seen[option] {
			return false
		}

		seen[option] = true
	}

	return true
}

// ValidWorkflowMergerText checks bounded nonempty protocol text.
func ValidWorkflowMergerText(text string) bool {
	return strings.TrimSpace(text) != "" && len(text) <= 4096 && !strings.ContainsRune(text, '\x00')
}

// ValidateWorkflowMergerDecisions checks bounded questions, requirements, options and authority identities.
func ValidateWorkflowMergerDecisions(decisions []workflowstate.MergerDecision, input workflowstate.MergerInput) error {
	if len(decisions) > 16 {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}
	for _, decision := range decisions {
		if !workflowstate.ValidRequestKey(decision.ID) || seen[decision.ID] || !ValidWorkflowMergerText(decision.Issue) || !workflowstate.ValidRequestKey(decision.Authority) {
			return workflowstate.ErrConflict
		}

		if err := validateWorkflowMergerRequirementIDs(decision.RequirementIDs, input.Requirements); err != nil {
			return err
		}

		if !validWorkflowMergerOptions(decision.Options) {
			return workflowstate.ErrConflict
		}

		seen[decision.ID] = true
	}

	return nil
}

func validateWorkflowMergerOutcome(result workflowstate.MergerResult, input workflowstate.MergerInput) error {
	switch result.Outcome {
	case "resolved":
		return validateWorkflowMergerResolved(result, input)
	case "needs_decision":
		if len(result.DecisionRequests) == 0 || result.Failure != "" {
			return workflowstate.ErrConflict
		}
	case "failed":
		if !ValidWorkflowMergerText(result.Failure) || len(result.DecisionRequests) > 0 {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func validateWorkflowMergerResolution(note workflowstate.MergerResolution, conflict workflowstate.MergerConflict, input workflowstate.MergerInput) error {
	if note.Path != conflict.Path || !slices.Equal(note.RegenerationCommand, conflict.RegenerationCommand) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowMergerRequirementIDs(note.RequirementIDs, input.Requirements); err != nil {
		return err
	}

	for _, requirement := range conflict.RequirementIDs {
		if !slices.Contains(note.RequirementIDs, requirement) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func validateWorkflowMergerResolutions(result workflowstate.MergerResult, input workflowstate.MergerInput) error {
	if len(result.ResolutionNotes) > len(input.Conflicts) {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}

	for _, note := range result.ResolutionNotes {
		index := slices.IndexFunc(input.Conflicts, func(c workflowstate.MergerConflict) bool { return c.ID == note.ConflictID })
		if index < 0 || seen[note.ConflictID] || !ValidWorkflowMergerText(note.Rationale) {
			return workflowstate.ErrConflict
		}

		if err := validateWorkflowMergerResolution(note, input.Conflicts[index], input); err != nil {
			return err
		}

		seen[note.ConflictID] = true
	}

	return ValidateWorkflowMergerDecisions(result.DecisionRequests, input)
}

func validateWorkflowMergerResolved(result workflowstate.MergerResult, input workflowstate.MergerInput) error {
	if len(result.DecisionRequests) > 0 || result.Failure != "" || len(result.ChangedPaths) == 0 || len(result.ResolutionNotes) != len(input.Conflicts) || len(result.PreservedRequirements) != len(input.Requirements) {
		return workflowstate.ErrConflict
	}

	decisions, err := MergerRequiredDecisions(input)
	if err != nil || len(decisions) != 0 {
		return workflowstate.ErrConflict
	}

	return nil
}
