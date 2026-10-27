package core

import (
	"fmt"
	"strings"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// ValidateWorkConfig validates the corresponding package boundary value.
func ValidateWorkConfig(config WorkConfig) error {
	if !workflowConfigChoice(config.Integration.PublicationProfile, []string{"", WorkflowPublicationDisabled, WorkflowPublicationGitAtomicDualRef}) {
		return fmt.Errorf("unsupported work.integration.publication_profile")
	}

	if err := ValidateWorkStateAuthority(config.StateAuthority); err != nil {
		return err
	}

	if err := ValidateWorkBranchPolicy(config.BranchPolicy); err != nil {
		return err
	}

	return workspacecore.ValidateWorkWorkspaceConfig(config.Workspace)
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
