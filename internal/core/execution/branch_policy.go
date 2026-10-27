package execution

import (
	"fmt"
	"strings"
)

// WorkBranchPolicy describes a package boundary value.
type WorkBranchPolicy struct {
	Mode                  WorkBranchPolicyMode
	IntegrationBranch     string
	CanonicalRemote       string
	RequireCleanWorktree  bool
	RequireRemoteTracking bool
	AllowSafeRepair       bool
	AllowBranchCreation   bool
}

// WorkBranchPolicyMode defines a package boundary type.
type WorkBranchPolicyMode string

// DefaultWorkBranchPolicy executes the corresponding package operation.
func DefaultWorkBranchPolicy() WorkBranchPolicy {
	return WorkBranchPolicy{
		Mode:                  WorkBranchPolicyModeIntegrationBranch,
		IntegrationBranch:     DefaultWorkIntegrationBranch,
		CanonicalRemote:       DefaultWorkCanonicalRemote,
		RequireCleanWorktree:  true,
		RequireRemoteTracking: true,
		AllowSafeRepair:       true,
		AllowBranchCreation:   false,
	}
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

const (
	// WorkBranchPolicyModeIntegrationBranch defines a package boundary constant.
	WorkBranchPolicyModeIntegrationBranch WorkBranchPolicyMode = "integration_branch"
	// WorkBranchPolicyModeCurrentBranch defines a package boundary constant.
	WorkBranchPolicyModeCurrentBranch WorkBranchPolicyMode = "current_branch"
	// WorkBranchPolicyModeNoGitEnforcement defines a package boundary constant.
	WorkBranchPolicyModeNoGitEnforcement WorkBranchPolicyMode = "no_git_enforcement"
	// DefaultWorkIntegrationBranch defines a package boundary constant.
	DefaultWorkIntegrationBranch = "dev"
	// DefaultWorkCanonicalRemote defines a package boundary constant.
	DefaultWorkCanonicalRemote = "origin"
)
