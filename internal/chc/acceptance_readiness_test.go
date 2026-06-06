package chc

import (
	"slices"
	"testing"
)

func TestValidateAcceptanceReadinessAcceptsGenerateDecision(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.Reasons = []string{"  CLI behavior has observable output.  ", "CLI behavior has observable output."}
	readiness.InferredElements[AcceptanceReadinessSlotActor] = []string{" CLI user ", "CLI user"}

	normalized, err := ValidateAcceptanceReadiness(readiness)
	if err != nil {
		t.Fatalf("ValidateAcceptanceReadiness returned error: %v", err)
	}

	if len(normalized.Reasons) != 1 {
		t.Fatalf("reasons = %#v, want deduplicated reasons", normalized.Reasons)
	}

	if got := normalized.InferredElements[AcceptanceReadinessSlotActor]; len(got) != 1 || got[0] != "CLI user" {
		t.Fatalf("actor inferred elements = %#v, want normalized CLI user", got)
	}
}

func TestValidateAcceptanceReadinessRejectsGenerateWithMissingSlot(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.Missing = []AcceptanceReadinessSlot{AcceptanceReadinessSlotFailurePath}

	_, err := ValidateAcceptanceReadiness(readiness)
	if err == nil {
		t.Fatal("ValidateAcceptanceReadiness returned nil error")
	}
}

func TestValidateAcceptanceReadinessRejectsGenerateWithoutRequiredElements(t *testing.T) {
	readiness := validAcceptanceReadiness()
	delete(readiness.InferredElements, AcceptanceReadinessSlotObservableOutcome)

	_, err := ValidateAcceptanceReadiness(readiness)
	if err == nil {
		t.Fatal("ValidateAcceptanceReadiness returned nil error")
	}
}

func TestValidateAcceptanceReadinessRejectsGenerateWithBlockingOpenQuestions(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.BlockingOpenQuestions = []string{"Which command is the real trigger?"}

	_, err := ValidateAcceptanceReadiness(readiness)
	if err == nil {
		t.Fatal("ValidateAcceptanceReadiness returned nil error")
	}
}

func TestValidateAcceptanceReadinessAcceptsNeedsMoreInfoDecision(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.Decision = AcceptanceReadinessNeedsMoreInfo
	readiness.Missing = []AcceptanceReadinessSlot{AcceptanceReadinessSlotFailurePath}
	delete(readiness.InferredElements, AcceptanceReadinessSlotFailurePath)

	readiness.CandidateScenarios = nil

	normalized, err := ValidateAcceptanceReadiness(readiness)
	if err != nil {
		t.Fatalf("ValidateAcceptanceReadiness returned error: %v", err)
	}

	if normalized.Decision != AcceptanceReadinessNeedsMoreInfo {
		t.Fatalf("decision = %q, want needs_more_info", normalized.Decision)
	}
}

func TestValidateAcceptanceReadinessRejectsContradictoryNeedsMoreInfoDecision(t *testing.T) {
	readiness := validAcceptanceReadiness()
	readiness.Decision = AcceptanceReadinessNeedsMoreInfo
	readiness.Missing = []AcceptanceReadinessSlot{AcceptanceReadinessSlotActor}

	_, err := ValidateAcceptanceReadiness(readiness)
	if err == nil {
		t.Fatal("ValidateAcceptanceReadiness returned nil error")
	}
}

func TestValidateAcceptanceReadinessAcceptsNotApplicableDecision(t *testing.T) {
	readiness := AcceptanceReadiness{
		Decision:   AcceptanceReadinessNotApplicable,
		Confidence: 0.92,
		Reasons:    []string{"The topic is documentation-only."},
	}

	normalized, err := ValidateAcceptanceReadiness(readiness)
	if err != nil {
		t.Fatalf("ValidateAcceptanceReadiness returned error: %v", err)
	}

	if normalized.Decision != AcceptanceReadinessNotApplicable {
		t.Fatalf("decision = %q, want not_applicable", normalized.Decision)
	}
}

func TestValidateAcceptanceReadinessRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name      string
		readiness AcceptanceReadiness
	}{
		{
			name: "invalid decision",
			readiness: AcceptanceReadiness{
				Decision:   "maybe",
				Confidence: 0.5,
				Reasons:    []string{"A reason."},
			},
		},
		{
			name: "invalid confidence",
			readiness: AcceptanceReadiness{
				Decision:   AcceptanceReadinessNotApplicable,
				Confidence: 1.5,
				Reasons:    []string{"A reason."},
			},
		},
		{
			name: "missing reasons",
			readiness: AcceptanceReadiness{
				Decision:   AcceptanceReadinessNotApplicable,
				Confidence: 0.5,
			},
		},
		{
			name: "invalid missing slot",
			readiness: AcceptanceReadiness{
				Decision:   AcceptanceReadinessNeedsMoreInfo,
				Confidence: 0.5,
				Reasons:    []string{"A reason."},
				Missing:    []AcceptanceReadinessSlot{"unknown"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ValidateAcceptanceReadiness(test.readiness)
			if err == nil {
				t.Fatal("ValidateAcceptanceReadiness returned nil error")
			}
		})
	}
}

func TestNewAcceptanceReadinessPromptIntent(t *testing.T) {
	intent, err := NewAcceptanceReadinessPromptIntent(AcceptanceReadinessSlotObservableOutcome)
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
	_, err := NewAcceptanceReadinessPromptIntent("unknown")
	if err == nil {
		t.Fatal("NewAcceptanceReadinessPromptIntent returned nil error")
	}
}

func validAcceptanceReadiness() AcceptanceReadiness {
	return AcceptanceReadiness{
		Decision:   AcceptanceReadinessGenerate,
		Confidence: 0.86,
		Reasons:    []string{"The topic has an actor, trigger, state, and observable outcome."},
		InferredElements: map[AcceptanceReadinessSlot][]string{
			AcceptanceReadinessSlotActor:             {"CLI user"},
			AcceptanceReadinessSlotTrigger:           {"runs the validation command"},
			AcceptanceReadinessSlotInitialState:      {"a configured project exists"},
			AcceptanceReadinessSlotObservableOutcome: {"validation report is written"},
			AcceptanceReadinessSlotSuccessPath:       {"validation passes and records the report"},
			AcceptanceReadinessSlotFailurePath:       {"invalid configuration fails before validation"},
			AcceptanceReadinessSlotBoundaries:        {"godog execution is out of scope"},
		},
		CandidateScenarios: []string{"Validation succeeds and writes a report"},
	}
}
