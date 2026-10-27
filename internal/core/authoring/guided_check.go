package authoring

import (
	"errors"
	"strings"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

// ValidateGuided checks the existing complete guided session contract.
func ValidateGuided(session Session) error {
	state := session.Guided
	if state == nil {
		return nil
	}

	if !ValidGuidedIntent(session) {
		return errors.New("invalid guided use-case schema, scope or progression intent")
	}

	if !boundedGuidedState(state) {
		return errors.New("guided use-case state exceeds its bounds")
	}

	if !guidedStageValid(state.Stage) || !guidedStatusValid(state.Status) {
		return errors.New("invalid guided use-case stage or status")
	}

	if state.Stage != "name" && (state.CaseSlug == "" || state.CaseName == "") {
		return errors.New("guided use case requires a canonical name")
	}

	return nil
}

func guidedStageValid(value string) bool {
	switch value {
	case "name", "dump", "quiz", "specification", "preparation", "execution", "results":
		return true
	default:
		return false
	}
}

func guidedStatusValid(value string) bool {
	switch value {
	case "waiting_input", "running", "blocked", "paused", "completed", "cancelled", "failed":
		return true
	default:
		return false
	}
}

// ValidGuidedIntent checks the original managed intent fields of a guided session.
func ValidGuidedIntent(session Session) bool {
	state := session.Guided

	return state.SchemaVersion == 1 && session.Revision > 0 && opsnamespace.Valid(session.Namespace) && strings.TrimSpace(state.Actor) != "" && len(state.Actor) <= 128 && state.ProgressionIntent
}

func boundedGuidedState(state *GuidedUseCaseState) bool {
	return len(state.Capture) <= 65536 && len(state.CompletedOperations) <= 256 && len(state.Checkpoints) <= 8 && len(state.LaterIdeas) <= 64 && boundedGuidedControls(state.Control)
}

func boundedGuidedControls(control *GuidedControlState) bool {
	return control == nil || len(control.CompletedOperations) <= 128
}
