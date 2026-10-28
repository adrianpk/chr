package merge

import (
	"slices"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validOptions(options []string) bool {
	if len(options) < 2 || len(options) > 5 {
		return false
	}

	seen := map[string]bool{}
	for _, option := range options {
		if !ValidText(option) || seen[option] {
			return false
		}

		seen[option] = true
	}

	return true
}

// ValidText checks bounded nonempty protocol text.
func ValidText(text string) bool {
	return strings.TrimSpace(text) != "" && len(text) <= 4096 && !strings.ContainsRune(text, '\x00')
}

// CheckDecisions checks bounded questions, requirements, options and authority identities.
func CheckDecisions(decisions []workflowstate.MergerDecision, input workflowstate.MergerInput) error {
	if len(decisions) > 16 {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}
	for _, decision := range decisions {
		if !workflowstate.ValidRequestKey(decision.ID) || seen[decision.ID] || !ValidText(decision.Issue) || !workflowstate.ValidRequestKey(decision.Authority) {
			return workflowstate.ErrConflict
		}

		if err := checkRequirementIDs(decision.RequirementIDs, input.Requirements); err != nil {
			return err
		}

		if !validOptions(decision.Options) {
			return workflowstate.ErrConflict
		}

		seen[decision.ID] = true
	}

	return nil
}

func checkOutcome(result workflowstate.MergerResult, input workflowstate.MergerInput) error {
	switch result.Outcome {
	case "resolved":
		return checkResolved(result, input)
	case "needs_decision":
		if len(result.DecisionRequests) == 0 || result.Failure != "" {
			return workflowstate.ErrConflict
		}
	case "failed":
		if !ValidText(result.Failure) || len(result.DecisionRequests) > 0 {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func checkResolution(note workflowstate.MergerResolution, conflict workflowstate.MergerConflict, input workflowstate.MergerInput) error {
	if note.Path != conflict.Path || !slices.Equal(note.RegenerationCommand, conflict.RegenerationCommand) {
		return workflowstate.ErrConflict
	}

	if err := checkRequirementIDs(note.RequirementIDs, input.Requirements); err != nil {
		return err
	}

	for _, requirement := range conflict.RequirementIDs {
		if !slices.Contains(note.RequirementIDs, requirement) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func checkResolutions(result workflowstate.MergerResult, input workflowstate.MergerInput) error {
	if len(result.ResolutionNotes) > len(input.Conflicts) {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}

	for _, note := range result.ResolutionNotes {
		index := slices.IndexFunc(input.Conflicts, func(c workflowstate.MergerConflict) bool { return c.ID == note.ConflictID })
		if index < 0 || seen[note.ConflictID] || !ValidText(note.Rationale) {
			return workflowstate.ErrConflict
		}

		if err := checkResolution(note, input.Conflicts[index], input); err != nil {
			return err
		}

		seen[note.ConflictID] = true
	}

	return CheckDecisions(result.DecisionRequests, input)
}

func checkResolved(result workflowstate.MergerResult, input workflowstate.MergerInput) error {
	if len(result.DecisionRequests) > 0 || result.Failure != "" || len(result.ChangedPaths) == 0 || len(result.ResolutionNotes) != len(input.Conflicts) || len(result.PreservedRequirements) != len(input.Requirements) {
		return workflowstate.ErrConflict
	}

	decisions, err := RequiredDecisions(input)
	if err != nil || len(decisions) != 0 {
		return workflowstate.ErrConflict
	}

	return nil
}
