package app

import (
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func quizAcceptanceReadinessFromContext(state quizcore.RuntimeState, sections []quizmodel.ContextSection) generationcore.AcceptanceReadiness {
	content := strings.ToLower(quizContextText(sections))
	reasons := []string{"The quiz context was inspected for acceptance-test readiness."}

	if quizContextLooksNonAcceptanceTestable(content) {
		return generationcore.AcceptanceReadiness{
			Decision:         generationcore.AcceptanceReadinessNotApplicable,
			Confidence:       0.7,
			Reasons:          append(reasons, "The context appears internal, architectural, documentation-only, or refactor-oriented."),
			InferredElements: map[generationcore.AcceptanceReadinessSlot][]string{},
		}
	}

	inferred := quizAcceptanceInferredElements(content)
	missing := quizAcceptanceMissingSlots(inferred)

	if len(missing) == 0 {
		return generationcore.AcceptanceReadiness{
			Decision:           generationcore.AcceptanceReadinessGenerate,
			Confidence:         0.72,
			Reasons:            append(reasons, "The context includes actor, trigger, initial state, outcome, and success-path signals."),
			InferredElements:   inferred,
			CandidateScenarios: []string{"Primary behavior succeeds and produces the expected observable outcome."},
		}
	}

	return generationcore.AcceptanceReadiness{
		Decision:         generationcore.AcceptanceReadinessNeedsMoreInfo,
		Confidence:       0.64,
		Reasons:          append(reasons, "The context has some acceptance-test signals, but key scenario elements are missing."),
		Missing:          missing,
		InferredElements: inferred,
	}
}

func quizContextText(sections []quizmodel.ContextSection) string {
	var output strings.Builder
	for _, section := range sections {
		output.WriteString(section.Content)
		output.WriteString("\n")
	}

	return output.String()
}

func quizContextLooksNonAcceptanceTestable(content string) bool {
	nonAcceptanceSignals := []string{
		"documentation-only",
		"docs-only",
		"internal refactor",
		"refactor only",
		"architecture cleanup",
		"package boundary",
	}
	for _, signal := range nonAcceptanceSignals {
		if strings.Contains(content, signal) {
			return true
		}
	}

	return false
}

func quizAcceptanceInferredElements(content string) map[generationcore.AcceptanceReadinessSlot][]string {
	inferred := map[generationcore.AcceptanceReadinessSlot][]string{}

	if quizContainsAny(content, "user", "operator", "cli", "api client", "worker", "scheduler", "reviewer") {
		inferred[generationcore.AcceptanceReadinessSlotActor] = []string{"An actor is referenced by the source context."}
	}

	if quizContainsAny(content, "command", "request", "event", "runs", "starts", "trigger", "action") {
		inferred[generationcore.AcceptanceReadinessSlotTrigger] = []string{"A trigger is referenced by the source context."}
	}

	if quizContainsAny(content, "given", "existing", "configured", "before", "state", "project", "ticket", "file") {
		inferred[generationcore.AcceptanceReadinessSlotInitialState] = []string{"An initial state is referenced by the source context."}
	}

	if quizContainsAny(content, "output", "report", "status", "response", "error", "file", "written", "created") {
		inferred[generationcore.AcceptanceReadinessSlotObservableOutcome] = []string{"An observable outcome is referenced by the source context."}
	}

	if quizContainsAny(content, "success", "succeeds", "passes", "accepted", "valid", "expected") {
		inferred[generationcore.AcceptanceReadinessSlotSuccessPath] = []string{"A success path is referenced by the source context."}
	}

	if quizContainsAny(content, "failure", "fails", "invalid", "reject", "error", "missing") {
		inferred[generationcore.AcceptanceReadinessSlotFailurePath] = []string{"A failure path is referenced by the source context."}
	}

	if quizContainsAny(content, "out of scope", "non-goal", "boundary", "must not", "do not") {
		inferred[generationcore.AcceptanceReadinessSlotBoundaries] = []string{"A boundary is referenced by the source context."}
	}

	return inferred
}

func quizAcceptanceMissingSlots(inferred map[generationcore.AcceptanceReadinessSlot][]string) []generationcore.AcceptanceReadinessSlot {
	required := []generationcore.AcceptanceReadinessSlot{
		generationcore.AcceptanceReadinessSlotActor,
		generationcore.AcceptanceReadinessSlotTrigger,
		generationcore.AcceptanceReadinessSlotInitialState,
		generationcore.AcceptanceReadinessSlotObservableOutcome,
		generationcore.AcceptanceReadinessSlotSuccessPath,
	}
	missing := []generationcore.AcceptanceReadinessSlot{}

	for _, slot := range required {
		if len(inferred[slot]) == 0 {
			missing = append(missing, slot)
		}
	}

	return missing
}

func quizContainsAny(content string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(content, value) {
			return true
		}
	}

	return false
}
