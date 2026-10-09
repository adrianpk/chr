package core

import (
	"fmt"
	"path/filepath"
	"slices"
)

func validateWorkflowClosingBinding(set WorkflowSet) error {
	if set.Closing == nil {
		return nil
	}

	if len(set.Closing.Attempts) < 1 || len(set.Closing.Attempts) > WorkflowRecordLimit {
		return fmt.Errorf("set closing history bound exceeded")
	}

	for _, binding := range set.Closing.Attempts {
		if !filepath.IsLocal(binding.Ref) || len(binding.Hash) != 64 || filepath.Base(binding.Ref) != "state.json" {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func validateWorkflowClosingAttempt(set WorkflowRecord, index int, state workflowClosingAttempt) error {
	if !workflowClosingAttemptIdentity(set, index, state) {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowClosingOutcome(state, index < len(set.Set.Closing.Attempts)-1); err != nil {
		return err
	}

	if state.Input.Ref != workflowClosingFile(set.ID, index+1, "input.json") || len(state.Input.Hash) != 64 {
		return ErrWorkflowConflict
	}

	if state.Result.Ref != "" && (state.Result.Ref != workflowClosingFile(set.ID, index+1, "result.json") || len(state.Result.Hash) != 64) {
		return ErrWorkflowConflict
	}

	for i, process := range state.Processes {
		if err := validateWorkflowClosingProcess(set.ID, state, i, process); err != nil {
			return err
		}
	}

	return nil
}

func workflowClosingAttemptIdentity(set WorkflowRecord, index int, state workflowClosingAttempt) bool {
	return state.Schema == "chr-set-closing-attempt/v1" && state.SetID == set.ID && state.Generation == index+1 && filepath.IsAbs(state.OwnerRoot) && validWorkflowProcessIdentity(state.Owner) && len(state.TokenHash) == 64 && !state.LeaseUntil.IsZero() && len(state.Processes) <= 16 && len(state.BudgetID) >= 1 && len(state.BudgetID) <= 128
}

func validateWorkflowClosingOutcome(state workflowClosingAttempt, historical bool) error {
	if state.Outcome == "running" {
		if !workflowClosingRunningOutcome(state, historical) {
			return ErrWorkflowConflict
		}

		return nil
	}

	if state.FinishedAt.IsZero() {
		return ErrWorkflowConflict
	}

	switch state.Outcome {
	case "completed", "closing_validation_failed":
		if state.Result.Ref == "" {
			return ErrWorkflowConflict
		}
	case "interrupted":
		if state.Result.Ref != "" {
			return ErrWorkflowConflict
		}
	default:
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowClosingProcess(id string, state workflowClosingAttempt, i int, process WorkflowProcess) error {
	if !validWorkflowProcessIdentity(process.Identity) || process.GroupID != process.Identity.PID || !slices.Contains([]string{"running", "stopped"}, process.State) {
		return ErrWorkflowConflict
	}

	base := workflowClosingFile(id, state.Generation, fmt.Sprintf("command-%d", i+1))
	if process.StdoutRef != base+".stdout" || process.StderrRef != base+".stderr" {
		return ErrWorkflowConflict
	}

	if process.State == "stopped" {
		if len(process.StdoutHash) != 64 || len(process.StderrHash) != 64 {
			return ErrWorkflowConflict
		}
	} else if state.Outcome != "running" {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowClosingEvidence(set WorkflowRecord) error {
	if set.Set.Closing == nil {
		return nil
	}

	budgets := map[string]int{}

	for index := range set.Set.Closing.Attempts {
		state, err := s.readWorkflowClosingAttempt(set, index)
		if err != nil {
			return err
		}

		budgets[state.BudgetID]++
		if budgets[state.BudgetID] > WorkflowPageLimit {
			return fmt.Errorf("set invocation closing attempt budget exceeded")
		}

		if err := s.checkWorkflowClosingAttemptEvidence(state); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowClosingAttemptEvidence(state workflowClosingAttempt) error {
	if err := s.checkWorkflowClosingResult(state); err != nil {
		return err
	}

	refs := map[string]string{state.Input.Ref: state.Input.Hash}
	if state.Result.Ref != "" {
		refs[state.Result.Ref] = state.Result.Hash
	}

	for _, process := range state.Processes {
		if process.State == "stopped" {
			refs[process.StdoutRef], refs[process.StderrRef] = process.StdoutHash, process.StderrHash
		}
	}

	return s.checkWorkflowPrerequisiteEvidence(refs)
}

func workflowClosingRunningOutcome(state workflowClosingAttempt, historical bool) bool {
	return !historical && state.FinishedAt.IsZero() && state.Result.Ref == ""
}
