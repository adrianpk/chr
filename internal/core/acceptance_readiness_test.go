package core

import (
	"slices"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
)

func TestValidateAcceptanceReadinessAcceptsGenerateDecision(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.Reasons = []string{"  CLI behavior has observable output.  ", "CLI behavior has observable output."}
	readiness.InferredElements[generationcore.AcceptanceReadinessSlotActor] = []string{" CLI user ", "CLI user"}

	normalized, err := generationcore.ValidateAcceptanceReadiness(readiness)
	if err != nil {
		t.Fatalf("ValidateAcceptanceReadiness returned error: %v", err)
	}

	if len(normalized.Reasons) != 1 {
		t.Fatalf("reasons = %#v, want deduplicated reasons", normalized.Reasons)
	}

	if got := normalized.InferredElements[generationcore.AcceptanceReadinessSlotActor]; len(got) != 1 || got[0] != "CLI user" {
		t.Fatalf("actor inferred elements = %#v, want normalized CLI user", got)
	}
}

func TestValidateAcceptanceReadinessRejectsGenerateWithMissingSlot(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.Missing = []generationcore.AcceptanceReadinessSlot{generationcore.AcceptanceReadinessSlotFailurePath}

	_, err := generationcore.ValidateAcceptanceReadiness(readiness)
	if err == nil {
		t.Fatal("ValidateAcceptanceReadiness returned nil error")
	}
}

func TestValidateAcceptanceReadinessRejectsGenerateWithoutRequiredElements(t *testing.T) {
	readiness := validAcceptanceReadiness()
	delete(readiness.InferredElements, generationcore.AcceptanceReadinessSlotObservableOutcome)

	_, err := generationcore.ValidateAcceptanceReadiness(readiness)
	if err == nil {
		t.Fatal("ValidateAcceptanceReadiness returned nil error")
	}
}

func TestValidateAcceptanceReadinessRejectsGenerateWithBlockingOpenQuestions(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.BlockingOpenQuestions = []string{"Which command is the real trigger?"}

	_, err := generationcore.ValidateAcceptanceReadiness(readiness)
	if err == nil {
		t.Fatal("ValidateAcceptanceReadiness returned nil error")
	}
}

func TestValidateAcceptanceReadinessAcceptsNeedsMoreInfoDecision(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.Decision = generationcore.AcceptanceReadinessNeedsMoreInfo
	readiness.Missing = []generationcore.AcceptanceReadinessSlot{generationcore.AcceptanceReadinessSlotFailurePath}
	delete(readiness.InferredElements, generationcore.AcceptanceReadinessSlotFailurePath)

	readiness.CandidateScenarios = nil

	normalized, err := generationcore.ValidateAcceptanceReadiness(readiness)
	if err != nil {
		t.Fatalf("ValidateAcceptanceReadiness returned error: %v", err)
	}

	if normalized.Decision != generationcore.AcceptanceReadinessNeedsMoreInfo {
		t.Fatalf("decision = %q, want needs_more_info", normalized.Decision)
	}
}

func TestValidateAcceptanceReadinessRejectsContradictoryNeedsMoreInfoDecision(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.Decision = generationcore.AcceptanceReadinessNeedsMoreInfo
	readiness.Missing = []generationcore.AcceptanceReadinessSlot{generationcore.AcceptanceReadinessSlotActor}

	_, err := generationcore.ValidateAcceptanceReadiness(readiness)
	if err == nil {
		t.Fatal("ValidateAcceptanceReadiness returned nil error")
	}
}

func TestValidateAcceptanceReadinessAcceptsNotApplicableDecision(t *testing.T) {
	readiness := generationcore.AcceptanceReadiness{
		Decision:   generationcore.AcceptanceReadinessNotApplicable,
		Confidence: 0.92,
		Reasons:    []string{"The topic is documentation-only."},
	}

	normalized, err := generationcore.ValidateAcceptanceReadiness(readiness)
	if err != nil {
		t.Fatalf("ValidateAcceptanceReadiness returned error: %v", err)
	}

	if normalized.Decision != generationcore.AcceptanceReadinessNotApplicable {
		t.Fatalf("decision = %q, want not_applicable", normalized.Decision)
	}
}

func TestValidateAcceptanceReadinessRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name      string
		readiness generationcore.AcceptanceReadiness
	}{
		{
			name: "invalid decision",
			readiness: generationcore.AcceptanceReadiness{
				Decision:   "maybe",
				Confidence: 0.5,
				Reasons:    []string{"A reason."},
			},
		},
		{
			name: "invalid confidence",
			readiness: generationcore.AcceptanceReadiness{
				Decision:   generationcore.AcceptanceReadinessNotApplicable,
				Confidence: 1.5,
				Reasons:    []string{"A reason."},
			},
		},
		{
			name: "missing reasons",
			readiness: generationcore.AcceptanceReadiness{
				Decision:   generationcore.AcceptanceReadinessNotApplicable,
				Confidence: 0.5,
			},
		},
		{
			name: "invalid missing slot",
			readiness: generationcore.AcceptanceReadiness{
				Decision:   generationcore.AcceptanceReadinessNeedsMoreInfo,
				Confidence: 0.5,
				Reasons:    []string{"A reason."},
				Missing:    []generationcore.AcceptanceReadinessSlot{"unknown"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := generationcore.ValidateAcceptanceReadiness(test.readiness)
			if err == nil {
				t.Fatal("ValidateAcceptanceReadiness returned nil error")
			}
		})
	}
}

func TestNewAcceptanceReadinessPromptIntent(t *testing.T) {
	intent, err := generationcore.NewAcceptanceReadinessPromptIntent(generationcore.AcceptanceReadinessSlotObservableOutcome)
	if err != nil {
		t.Fatalf("NewAcceptanceReadinessPromptIntent returned error: %v", err)
	}

	if intent.Kind != "acceptance_readiness_question" {
		t.Fatalf("kind = %q", intent.Kind)
	}

	if intent.Locale != "auto" {
		t.Fatalf("locale = %q, want auto", intent.Locale)
	}

	contract := intent.Metadata.SemanticContract
	if contract.Slot != "observable_outcome" {
		t.Fatalf("contract slot = %q", contract.Slot)
	}

	if !slices.Contains(contract.ForbiddenSlots, "actor") {
		t.Fatalf("forbidden slots = %#v, want actor", contract.ForbiddenSlots)
	}

	if slices.Contains(contract.ForbiddenSlots, "observable_outcome") {
		t.Fatalf("forbidden slots = %#v, should not include selected slot", contract.ForbiddenSlots)
	}
}

func TestNewAcceptanceReadinessPromptIntentRejectsInvalidSlot(t *testing.T) {
	_, err := generationcore.NewAcceptanceReadinessPromptIntent("unknown")
	if err == nil {
		t.Fatal("NewAcceptanceReadinessPromptIntent returned nil error")
	}
}

func validAcceptanceReadiness() generationcore.AcceptanceReadiness {
	return generationcore.AcceptanceReadiness{
		Decision:   generationcore.AcceptanceReadinessGenerate,
		Confidence: 0.86,
		Reasons:    []string{"The topic has an actor, trigger, state, and observable outcome."},
		InferredElements: map[generationcore.AcceptanceReadinessSlot][]string{
			generationcore.AcceptanceReadinessSlotActor:             {"CLI user"},
			generationcore.AcceptanceReadinessSlotTrigger:           {"runs the validation command"},
			generationcore.AcceptanceReadinessSlotInitialState:      {"a configured project exists"},
			generationcore.AcceptanceReadinessSlotObservableOutcome: {"validation report is written"},
			generationcore.AcceptanceReadinessSlotSuccessPath:       {"validation passes and records the report"},
			generationcore.AcceptanceReadinessSlotFailurePath:       {"invalid configuration fails before validation"},
			generationcore.AcceptanceReadinessSlotBoundaries:        {"godog execution is out of scope"},
		},
		CandidateScenarios: []string{"Validation succeeds and writes a report"},
	}
}
