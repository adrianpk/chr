package app

import (
	"fmt"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func tuiGuidedStagePrompt(observed GuidedUseCaseInspection) string {
	state := observed.Session.Guided
	if !authoringcore.GuidedDispatchAllowed(*observed.Session) {
		return fmt.Sprintf("%s is suspended or stopped. Saved documents and inputs are retained.", guidedStageName(state.Stage))
	}

	if state.PendingOperation != nil {
		return fmt.Sprintf("%s has a pending operation. Saved input is retained; Continue reconciles that operation.", guidedStageName(state.Stage))
	}

	prompts := map[string]string{
		"name":          "What is the use-case name?",
		"dump":          "Add the known context. Continue to quiz submits the current editor content.",
		"quiz":          "Clarification needs a current question. Resume clarification after correcting its setup.",
		"specification": "Specification generation and automatic approval use the committed Dump and Quiz.",
		"preparation":   "Plan and ticket preparation uses the approved specification. Continue the retained preparation when ready.",
	}

	if state.Stage == "quiz" && observed.Question != "" {
		return observed.Question
	}

	if prompt, ok := prompts[state.Stage]; ok {
		return prompt
	}

	return "The shared workflow is progressing or waiting for its recorded prerequisite."
}

func guidedStageName(stage string) string {
	names := map[string]string{"name": "Naming", "dump": "Capture", "quiz": "Clarification", "specification": "Specification", "preparation": "Plan and ticket preparation", "execution": "Execution", "results": "Results"}

	return names[stage]
}

func tuiGuidedOperationLines(observed GuidedUseCaseInspection) []string {
	lines := []string{}

	state := observed.Session.Guided

	lines = append(lines, tuiGuidedPendingLines(state.PendingOperation)...)
	if state.PendingCheckpoint != nil {
		lines = append(lines, "Checkpoint acknowledgement is pending; retry reconciles the original committed documents.")
	}

	if observed.Quiz != nil && observed.Quiz.State.ContinuationError != "" {
		lines = append(lines, observed.Quiz.State.ContinuationError)
	}

	return lines
}

func tuiGuidedPendingLines(pending *authoringcore.GuidedUseCaseOperation) []string {
	if pending == nil {
		return nil
	}

	lines := []string{}

	lines = append(lines, "The original operation and consumed limits remain unchanged.")
	if pending.Content != "" {
		lines = append(lines, "Saved input:", pending.Content)
	}

	if pending.PreparedSpec != nil {
		lines = append(lines, "Generated specification retained; retry reuses its original content.")
	}

	if pending.PreparedPlan != nil {
		lines = append(lines, "Generated plan retained; retry reuses its original content.")
	}

	if pending.PreparedWorkflow != nil {
		lines = append(lines, "Ticket contracts retained; retry reconciles their original materialization.")
	}

	return lines
}
