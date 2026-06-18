package core

import (
	"fmt"
	"strings"
)

// ValidateWorkConfig validates the corresponding package boundary value.
func ValidateWorkConfig(config WorkConfig) error {
	return ValidateWorkBranchPolicy(config.BranchPolicy)
}

// ValidateWorkBranchPolicy validates the corresponding package boundary value.
func ValidateWorkBranchPolicy(policy WorkBranchPolicy) error {
	if !IsValidWorkBranchPolicyMode(policy.Mode) {
		return fmt.Errorf("invalid work branch policy mode: %s", policy.Mode)
	}

	if policy.Mode == WorkBranchPolicyModeIntegrationBranch {
		if strings.TrimSpace(policy.IntegrationBranch) == "" {
			return fmt.Errorf("work integration_branch is required")
		}

		if policy.RequireRemoteTracking && strings.TrimSpace(policy.CanonicalRemote) == "" {
			return fmt.Errorf("work canonical_remote is required when remote tracking is required")
		}
	}

	return nil
}

// IsValidWorkBranchPolicyMode executes the corresponding package operation.
func IsValidWorkBranchPolicyMode(mode WorkBranchPolicyMode) bool {
	switch WorkBranchPolicyMode(strings.TrimSpace(string(mode))) {
	case WorkBranchPolicyModeIntegrationBranch,
		WorkBranchPolicyModeCurrentBranch,
		WorkBranchPolicyModeNoGitEnforcement:
		return true
	default:
		return false
	}
}
