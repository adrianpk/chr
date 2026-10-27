package core

import (
	"fmt"
	"strings"

	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ValidateEvidenceConfig validates the corresponding package boundary value.
func ValidateEvidenceConfig(config workspaceconfig.EvidenceConfig) error {
	return ValidateEvidenceStrategyConfig(config.Strategy)
}

// ValidateEvidenceStrategyConfig validates the corresponding package boundary value.
func ValidateEvidenceStrategyConfig(config workspaceconfig.EvidenceStrategyConfig) error {
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
func IsValidEvidenceStrategyMode(mode workspaceconfig.EvidenceStrategyMode) bool {
	switch workspaceconfig.EvidenceStrategyMode(strings.TrimSpace(string(mode))) {
	case workspaceconfig.EvidenceStrategyModeInline,
		workspaceconfig.EvidenceStrategyModeSeparatePass:
		return true
	default:
		return false
	}
}

// IsValidEvidenceTicketChecklistPolicy executes the corresponding package operation.
func IsValidEvidenceTicketChecklistPolicy(policy workspaceconfig.EvidenceTicketChecklistPolicy) bool {
	switch workspaceconfig.EvidenceTicketChecklistPolicy(strings.TrimSpace(string(policy))) {
	case workspaceconfig.EvidenceTicketChecklistPolicyRequired,
		workspaceconfig.EvidenceTicketChecklistPolicyOptional:
		return true
	default:
		return false
	}
}

// IsValidEvidenceGherkinPolicy executes the corresponding package operation.
func IsValidEvidenceGherkinPolicy(policy workspaceconfig.EvidenceGherkinPolicy) bool {
	switch workspaceconfig.EvidenceGherkinPolicy(strings.TrimSpace(string(policy))) {
	case workspaceconfig.EvidenceGherkinPolicyActiveWhenApplicable,
		workspaceconfig.EvidenceGherkinPolicyDisabled,
		workspaceconfig.EvidenceGherkinPolicyManualOnly:
		return true
	default:
		return false
	}
}
