package core

import (
	"fmt"
	"strings"
)

// ValidateEvidenceConfig validates the corresponding package boundary value.
func ValidateEvidenceConfig(config EvidenceConfig) error {
	return ValidateEvidenceStrategyConfig(config.Strategy)
}

// ValidateEvidenceStrategyConfig validates the corresponding package boundary value.
func ValidateEvidenceStrategyConfig(config EvidenceStrategyConfig) error {
	if !IsValidEvidenceStrategyMode(config.Mode) {
		return fmt.Errorf("invalid evidence strategy mode: %s", config.Mode)
	}

	if !IsValidEvidenceTicketChecklistPolicy(config.TicketChecklist) {
		return fmt.Errorf("invalid evidence ticket_checklist policy: %s", config.TicketChecklist)
	}

	if !IsValidEvidenceGherkinPolicy(config.GherkinPolicy) {
		return fmt.Errorf("invalid evidence gherkin_policy: %s", config.GherkinPolicy)
	}

	return nil
}

// IsValidEvidenceStrategyMode executes the corresponding package operation.
func IsValidEvidenceStrategyMode(mode EvidenceStrategyMode) bool {
	switch EvidenceStrategyMode(strings.TrimSpace(string(mode))) {
	case EvidenceStrategyModeInline,
		EvidenceStrategyModeSeparatePass:
		return true
	default:
		return false
	}
}

// IsValidEvidenceTicketChecklistPolicy executes the corresponding package operation.
func IsValidEvidenceTicketChecklistPolicy(policy EvidenceTicketChecklistPolicy) bool {
	switch EvidenceTicketChecklistPolicy(strings.TrimSpace(string(policy))) {
	case EvidenceTicketChecklistPolicyRequired,
		EvidenceTicketChecklistPolicyOptional:
		return true
	default:
		return false
	}
}

// IsValidEvidenceGherkinPolicy executes the corresponding package operation.
func IsValidEvidenceGherkinPolicy(policy EvidenceGherkinPolicy) bool {
	switch EvidenceGherkinPolicy(strings.TrimSpace(string(policy))) {
	case EvidenceGherkinPolicyActiveWhenApplicable,
		EvidenceGherkinPolicyDisabled,
		EvidenceGherkinPolicyManualOnly:
		return true
	default:
		return false
	}
}
