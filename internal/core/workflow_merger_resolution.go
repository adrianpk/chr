package core

import (
	"slices"
	"strings"
)

func validWorkflowMergerText(text string) bool {
	return strings.TrimSpace(text) != "" && len(text) <= 4096 && !strings.ContainsRune(text, '\x00')
}

func validateWorkflowMergerResolutions(result WorkflowMergerResult, input WorkflowMergerInput) error {
	if len(result.ResolutionNotes) > len(input.Conflicts) {
		return ErrWorkflowConflict
	}

	seen := map[string]bool{}

	for _, note := range result.ResolutionNotes {
		index := slices.IndexFunc(input.Conflicts, func(c WorkflowMergerConflict) bool { return c.ID == note.ConflictID })
		if index < 0 || seen[note.ConflictID] || !validWorkflowMergerText(note.Rationale) {
			return ErrWorkflowConflict
		}

		if err := validateWorkflowMergerResolution(note, input.Conflicts[index], input); err != nil {
			return err
		}

		seen[note.ConflictID] = true
	}

	return validateWorkflowMergerDecisions(result.DecisionRequests, input)
}

func validateWorkflowMergerResolution(note WorkflowMergerResolution, conflict WorkflowMergerConflict, input WorkflowMergerInput) error {
	if note.Path != conflict.Path || !slices.Equal(note.RegenerationCommand, conflict.RegenerationCommand) {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowMergerRequirementIDs(note.RequirementIDs, input.Requirements); err != nil {
		return err
	}

	for _, requirement := range conflict.RequirementIDs {
		if !slices.Contains(note.RequirementIDs, requirement) {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func validateWorkflowMergerDecisions(decisions []WorkflowMergerDecision, input WorkflowMergerInput) error {
	if len(decisions) > 16 {
		return ErrWorkflowConflict
	}

	seen := map[string]bool{}
	for _, decision := range decisions {
		if !validWorkflowRequestKey(decision.ID) || seen[decision.ID] || !validWorkflowMergerText(decision.Issue) || !validWorkflowRequestKey(decision.Authority) {
			return ErrWorkflowConflict
		}

		if err := validateWorkflowMergerRequirementIDs(decision.RequirementIDs, input.Requirements); err != nil {
			return err
		}

		if !validWorkflowMergerOptions(decision.Options) {
			return ErrWorkflowConflict
		}

		seen[decision.ID] = true
	}

	return nil
}

func validWorkflowMergerOptions(options []string) bool {
	if len(options) < 2 || len(options) > 5 {
		return false
	}

	seen := map[string]bool{}
	for _, option := range options {
		if !validWorkflowMergerText(option) || seen[option] {
			return false
		}

		seen[option] = true
	}

	return true
}

func validateWorkflowMergerOutcome(result WorkflowMergerResult, input WorkflowMergerInput) error {
	switch result.Outcome {
	case "resolved":
		return validateWorkflowMergerResolved(result, input)
	case "needs_decision":
		if len(result.DecisionRequests) == 0 || result.Failure != "" {
			return ErrWorkflowConflict
		}
	case "failed":
		if !validWorkflowMergerText(result.Failure) || len(result.DecisionRequests) > 0 {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func validateWorkflowMergerResolved(result WorkflowMergerResult, input WorkflowMergerInput) error {
	if len(result.DecisionRequests) > 0 || result.Failure != "" || len(result.ChangedPaths) == 0 || len(result.ResolutionNotes) != len(input.Conflicts) || len(result.PreservedRequirements) != len(input.Requirements) {
		return ErrWorkflowConflict
	}

	decisions, err := WorkflowMergerRequiredDecisions(input)
	if err != nil || len(decisions) != 0 {
		return ErrWorkflowConflict
	}

	return nil
}
