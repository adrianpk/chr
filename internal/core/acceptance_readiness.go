package core

import (
	"fmt"
	"slices"
	"strings"
)

const (
	// AcceptanceReadinessGenerate defines a package boundary constant.
	AcceptanceReadinessGenerate AcceptanceReadinessDecision = "generate"
	// AcceptanceReadinessNeedsMoreInfo defines a package boundary constant.
	AcceptanceReadinessNeedsMoreInfo AcceptanceReadinessDecision = "needs_more_info"
	// AcceptanceReadinessNotApplicable defines a package boundary constant.
	AcceptanceReadinessNotApplicable AcceptanceReadinessDecision = "not_applicable"
)

const (
	// AcceptanceReadinessSlotActor defines a package boundary constant.
	AcceptanceReadinessSlotActor AcceptanceReadinessSlot = "actor"
	// AcceptanceReadinessSlotTrigger defines a package boundary constant.
	AcceptanceReadinessSlotTrigger AcceptanceReadinessSlot = "trigger"
	// AcceptanceReadinessSlotInitialState defines a package boundary constant.
	AcceptanceReadinessSlotInitialState AcceptanceReadinessSlot = "initial_state"
	// AcceptanceReadinessSlotObservableOutcome defines a package boundary constant.
	AcceptanceReadinessSlotObservableOutcome AcceptanceReadinessSlot = "observable_outcome"
	// AcceptanceReadinessSlotSuccessPath defines a package boundary constant.
	AcceptanceReadinessSlotSuccessPath AcceptanceReadinessSlot = "success_path"
	// AcceptanceReadinessSlotFailurePath defines a package boundary constant.
	AcceptanceReadinessSlotFailurePath AcceptanceReadinessSlot = "failure_path"
	// AcceptanceReadinessSlotBoundaries defines a package boundary constant.
	AcceptanceReadinessSlotBoundaries AcceptanceReadinessSlot = "boundaries"
)

const (
	acceptanceReadinessPromptKind = "acceptance_readiness_question"
	acceptanceReadinessLocale     = "auto"
)

// AcceptanceReadinessDecision defines a package boundary type.
type AcceptanceReadinessDecision string

// AcceptanceReadinessSlot defines a package boundary type.
type AcceptanceReadinessSlot string

// AcceptanceReadiness describes whether context is ready for Gherkin generation.
type AcceptanceReadiness struct {
	Decision              AcceptanceReadinessDecision          `json:"decision"`
	Confidence            float64                              `json:"confidence"`
	Reasons               []string                             `json:"reasons"`
	Missing               []AcceptanceReadinessSlot            `json:"missing"`
	InferredElements      map[AcceptanceReadinessSlot][]string `json:"inferred_elements"`
	CandidateScenarios    []string                             `json:"candidate_scenarios"`
	BlockingOpenQuestions []string                             `json:"blocking_open_questions"`
}

// AcceptanceReadinessPromptIntent describes one semantic readiness question.
type AcceptanceReadinessPromptIntent struct {
	Kind        string                                  `json:"kind"`
	Slot        string                                  `json:"slot"`
	Locale      string                                  `json:"locale"`
	LLMRequired bool                                    `json:"llm_required"`
	Metadata    AcceptanceReadinessPromptIntentMetadata `json:"metadata"`
}

// AcceptanceReadinessPromptIntentMetadata describes prompt intent metadata.
type AcceptanceReadinessPromptIntentMetadata struct {
	SemanticContract AcceptanceReadinessSemanticContract `json:"semantic_contract"`
}

// AcceptanceReadinessSemanticContract describes a stable question target.
type AcceptanceReadinessSemanticContract struct {
	Slot               string   `json:"slot"`
	Intent             string   `json:"intent"`
	ExpectedAnswerKind string   `json:"expected_answer_kind"`
	QuestionGuidance   string   `json:"question_guidance"`
	ForbiddenSlots     []string `json:"forbidden_slots"`
}

// NewAcceptanceReadinessPromptIntent constructs a semantic prompt intent.
func NewAcceptanceReadinessPromptIntent(slot AcceptanceReadinessSlot) (AcceptanceReadinessPromptIntent, error) {
	if !slot.Valid() {
		return AcceptanceReadinessPromptIntent{}, fmt.Errorf("acceptance readiness slot is invalid: %s", slot)
	}

	return AcceptanceReadinessPromptIntent{
		Kind:        acceptanceReadinessPromptKind,
		Slot:        string(slot),
		Locale:      acceptanceReadinessLocale,
		LLMRequired: true,
		Metadata: AcceptanceReadinessPromptIntentMetadata{
			SemanticContract: AcceptanceReadinessSemanticContract{
				Slot:               string(slot),
				Intent:             acceptanceReadinessSlotIntent(slot),
				ExpectedAnswerKind: acceptanceReadinessSlotExpectedAnswerKind(slot),
				QuestionGuidance:   acceptanceReadinessSlotQuestionGuidance(slot),
				ForbiddenSlots:     acceptanceReadinessForbiddenSlots(slot),
			},
		},
	}, nil
}

// AcceptanceReadinessSlots returns the stable readiness slot order.
func AcceptanceReadinessSlots() []AcceptanceReadinessSlot {
	return []AcceptanceReadinessSlot{
		AcceptanceReadinessSlotActor,
		AcceptanceReadinessSlotTrigger,
		AcceptanceReadinessSlotInitialState,
		AcceptanceReadinessSlotObservableOutcome,
		AcceptanceReadinessSlotSuccessPath,
		AcceptanceReadinessSlotFailurePath,
		AcceptanceReadinessSlotBoundaries,
	}
}

// ValidateAcceptanceReadiness validates and normalizes the readiness payload.
func ValidateAcceptanceReadiness(readiness AcceptanceReadiness) (AcceptanceReadiness, error) {
	normalized := AcceptanceReadiness{
		Decision:              readiness.Decision,
		Confidence:            readiness.Confidence,
		Reasons:               normalizeStringList(readiness.Reasons),
		Missing:               normalizeAcceptanceReadinessSlots(readiness.Missing),
		InferredElements:      normalizeAcceptanceReadinessElements(readiness.InferredElements),
		CandidateScenarios:    normalizeStringList(readiness.CandidateScenarios),
		BlockingOpenQuestions: normalizeStringList(readiness.BlockingOpenQuestions),
	}

	if !normalized.Decision.Valid() {
		return AcceptanceReadiness{}, fmt.Errorf("acceptance readiness decision is invalid: %s", normalized.Decision)
	}

	if normalized.Confidence < 0 || normalized.Confidence > 1 {
		return AcceptanceReadiness{}, fmt.Errorf("acceptance readiness confidence must be between 0 and 1")
	}

	if len(normalized.Reasons) == 0 {
		return AcceptanceReadiness{}, fmt.Errorf("acceptance readiness reasons are required")
	}

	err := validateAcceptanceReadinessMissing(normalized.Missing)
	if err != nil {
		return AcceptanceReadiness{}, err
	}

	err = validateAcceptanceReadinessElements(normalized.InferredElements)
	if err != nil {
		return AcceptanceReadiness{}, err
	}

	err = validateAcceptanceReadinessDecision(normalized)
	if err != nil {
		return AcceptanceReadiness{}, err
	}

	return normalized, nil
}

// Valid reports whether the decision is supported.
func (d AcceptanceReadinessDecision) Valid() bool {
	switch d {
	case AcceptanceReadinessGenerate,
		AcceptanceReadinessNeedsMoreInfo,
		AcceptanceReadinessNotApplicable:
		return true
	default:
		return false
	}
}

// Valid reports whether the slot is supported.
func (s AcceptanceReadinessSlot) Valid() bool {
	return slices.Contains(AcceptanceReadinessSlots(), s)
}

func validateAcceptanceReadinessDecision(readiness AcceptanceReadiness) error {
	switch readiness.Decision {
	case AcceptanceReadinessGenerate:
		return validateAcceptanceReadinessGenerate(readiness)
	case AcceptanceReadinessNeedsMoreInfo:
		return validateAcceptanceReadinessNeedsMoreInfo(readiness)
	case AcceptanceReadinessNotApplicable:
		return validateAcceptanceReadinessNotApplicable(readiness)
	default:
		return fmt.Errorf("acceptance readiness decision is invalid: %s", readiness.Decision)
	}
}

func validateAcceptanceReadinessGenerate(readiness AcceptanceReadiness) error {
	if len(readiness.Missing) > 0 {
		return fmt.Errorf("acceptance readiness generate decision cannot have missing slots")
	}

	if len(readiness.BlockingOpenQuestions) > 0 {
		return fmt.Errorf("acceptance readiness generate decision cannot have blocking open questions")
	}

	for _, slot := range []AcceptanceReadinessSlot{
		AcceptanceReadinessSlotActor,
		AcceptanceReadinessSlotTrigger,
		AcceptanceReadinessSlotInitialState,
		AcceptanceReadinessSlotObservableOutcome,
		AcceptanceReadinessSlotSuccessPath,
	} {
		if len(readiness.InferredElements[slot]) == 0 {
			return fmt.Errorf("acceptance readiness generate decision requires inferred %s", slot)
		}
	}

	if len(readiness.CandidateScenarios) == 0 {
		return fmt.Errorf("acceptance readiness generate decision requires candidate scenarios")
	}

	return nil
}

func validateAcceptanceReadinessNeedsMoreInfo(readiness AcceptanceReadiness) error {
	if len(readiness.Missing) == 0 {
		return fmt.Errorf("acceptance readiness needs_more_info decision requires missing slots")
	}

	for _, slot := range readiness.Missing {
		if len(readiness.InferredElements[slot]) > 0 {
			return fmt.Errorf("acceptance readiness slot %s cannot be both missing and inferred", slot)
		}
	}

	return nil
}

func validateAcceptanceReadinessNotApplicable(readiness AcceptanceReadiness) error {
	if len(readiness.Missing) > 0 {
		return fmt.Errorf("acceptance readiness not_applicable decision cannot have missing slots")
	}

	if len(readiness.CandidateScenarios) > 0 {
		return fmt.Errorf("acceptance readiness not_applicable decision cannot have candidate scenarios")
	}

	return nil
}

func validateAcceptanceReadinessMissing(slots []AcceptanceReadinessSlot) error {
	seen := map[AcceptanceReadinessSlot]bool{}

	for _, slot := range slots {
		if !slot.Valid() {
			return fmt.Errorf("acceptance readiness missing slot is invalid: %s", slot)
		}

		if seen[slot] {
			return fmt.Errorf("acceptance readiness missing slot is duplicated: %s", slot)
		}

		seen[slot] = true
	}

	return nil
}

func validateAcceptanceReadinessElements(elements map[AcceptanceReadinessSlot][]string) error {
	for slot, values := range elements {
		if !slot.Valid() {
			return fmt.Errorf("acceptance readiness inferred slot is invalid: %s", slot)
		}

		if len(values) == 0 {
			return fmt.Errorf("acceptance readiness inferred slot %s values are required", slot)
		}
	}

	return nil
}

func normalizeAcceptanceReadinessSlots(slots []AcceptanceReadinessSlot) []AcceptanceReadinessSlot {
	normalized := make([]AcceptanceReadinessSlot, 0, len(slots))
	for _, slot := range slots {
		cleaned := AcceptanceReadinessSlot(strings.TrimSpace(string(slot)))
		if cleaned == "" {
			continue
		}

		normalized = append(normalized, cleaned)
	}

	return normalized
}

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}

	for _, value := range values {
		cleaned := strings.TrimSpace(value)
		if cleaned == "" || seen[cleaned] {
			continue
		}

		normalized = append(normalized, cleaned)
		seen[cleaned] = true
	}

	return normalized
}

func normalizeAcceptanceReadinessElements(elements map[AcceptanceReadinessSlot][]string) map[AcceptanceReadinessSlot][]string {
	if len(elements) == 0 {
		return map[AcceptanceReadinessSlot][]string{}
	}

	normalized := map[AcceptanceReadinessSlot][]string{}

	for slot, values := range elements {
		cleanedSlot := AcceptanceReadinessSlot(strings.TrimSpace(string(slot)))
		cleanedValues := normalizeStringList(values)

		if cleanedSlot == "" || len(cleanedValues) == 0 {
			continue
		}

		normalized[cleanedSlot] = cleanedValues
	}

	return normalized
}

func acceptanceReadinessSlotIntent(slot AcceptanceReadinessSlot) string {
	switch slot {
	case AcceptanceReadinessSlotActor:
		return "Ask who initiates, uses, receives, or observes the behavior."
	case AcceptanceReadinessSlotTrigger:
		return "Ask what action, command, request, event, or state transition starts the behavior."
	case AcceptanceReadinessSlotInitialState:
		return "Ask what state or context must exist before the behavior runs."
	case AcceptanceReadinessSlotObservableOutcome:
		return "Ask what result can be observed after the behavior runs."
	case AcceptanceReadinessSlotSuccessPath:
		return "Ask what should happen when the behavior succeeds."
	case AcceptanceReadinessSlotFailurePath:
		return "Ask what should happen when input is invalid or a dependency fails."
	case AcceptanceReadinessSlotBoundaries:
		return "Ask what is explicitly in scope or out of scope for this behavior."
	default:
		return "Ask for the next missing acceptance readiness detail."
	}
}

func acceptanceReadinessSlotExpectedAnswerKind(slot AcceptanceReadinessSlot) string {
	switch slot {
	case AcceptanceReadinessSlotActor:
		return "user, CLI caller, API client, worker, scheduler, reviewer, or other actor"
	case AcceptanceReadinessSlotTrigger:
		return "command, request, event, state transition, user action, or scheduled action"
	case AcceptanceReadinessSlotInitialState:
		return "relevant existing project, ticket, file, configuration, record, or runtime state"
	case AcceptanceReadinessSlotObservableOutcome:
		return "observable output, state, file, report, response, database row, or error"
	case AcceptanceReadinessSlotSuccessPath:
		return "expected behavior when inputs are valid"
	case AcceptanceReadinessSlotFailurePath:
		return "expected behavior when inputs are invalid or a dependency fails"
	case AcceptanceReadinessSlotBoundaries:
		return "scope, non-goals, exclusions, or implementation boundary"
	default:
		return "acceptance readiness detail"
	}
}

func acceptanceReadinessSlotQuestionGuidance(slot AcceptanceReadinessSlot) string {
	if slot == AcceptanceReadinessSlotFailurePath {
		return "Ask exactly one concise question about the most important failure or rejection case."
	}

	return "Ask exactly one concise question for this slot."
}

func acceptanceReadinessForbiddenSlots(slot AcceptanceReadinessSlot) []string {
	forbidden := make([]string, 0, len(AcceptanceReadinessSlots())-1)
	for _, candidate := range AcceptanceReadinessSlots() {
		if candidate == slot {
			continue
		}

		forbidden = append(forbidden, string(candidate))
	}

	return forbidden
}
