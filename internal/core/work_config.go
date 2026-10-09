package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateWorkConfig validates the corresponding package boundary value.
func ValidateWorkConfig(config WorkConfig) error {
	if err := ValidateWorkStateAuthority(config.StateAuthority); err != nil {
		return err
	}

	if err := ValidateWorkBranchPolicy(config.BranchPolicy); err != nil {
		return err
	}

	return ValidateWorkWorkspaceConfig(config.Workspace)
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

// ValidateWorkWorkspaceConfig validates the corresponding package boundary value.
func ValidateWorkWorkspaceConfig(config WorkWorkspaceConfig) error {
	if config == (WorkWorkspaceConfig{}) {
		return nil
	}

	if !IsValidWorkWorkspaceMode(config.Mode) {
		return fmt.Errorf("invalid work workspace_mode: %s", config.Mode)
	}

	if config.Mode != WorkWorkspaceModeWorktree {
		return nil
	}

	if strings.TrimSpace(config.WorktreeRoot) == "" {
		return fmt.Errorf("work worktree_root is required")
	}

	if filepath.ToSlash(filepath.Clean(config.WorktreeRoot)) != DefaultWorktreeRoot {
		return fmt.Errorf("work worktree_root must be %s", DefaultWorktreeRoot)
	}

	if !IsValidWorktreeBasePolicy(config.WorktreeBase) {
		return fmt.Errorf("invalid work worktree_base: %s", config.WorktreeBase)
	}

	if !IsValidWorktreeCleanupPolicy(config.WorktreeCleanup) {
		return fmt.Errorf("invalid work worktree_cleanup: %s", config.WorktreeCleanup)
	}

	return nil
}

// IsValidWorkWorkspaceMode executes the corresponding package operation.
func IsValidWorkWorkspaceMode(mode WorkWorkspaceMode) bool {
	switch WorkWorkspaceMode(strings.TrimSpace(string(mode))) {
	case WorkWorkspaceModeProject,
		WorkWorkspaceModeWorktree:
		return true
	default:
		return false
	}
}

// IsValidWorktreeBasePolicy executes the corresponding package operation.
func IsValidWorktreeBasePolicy(policy WorktreeBasePolicy) bool {
	switch WorktreeBasePolicy(strings.TrimSpace(string(policy))) {
	case WorktreeBasePolicyRemoteIntegrationBranch,
		WorktreeBasePolicyLocalIntegrationBranch,
		WorktreeBasePolicyCurrentBranch:
		return true
	default:
		return false
	}
}

// IsValidWorktreeCleanupPolicy executes the corresponding package operation.
func IsValidWorktreeCleanupPolicy(policy WorktreeCleanupPolicy) bool {
	switch WorktreeCleanupPolicy(strings.TrimSpace(string(policy))) {
	case WorktreeCleanupPolicyAfterMerge,
		WorktreeCleanupPolicyManual,
		WorktreeCleanupPolicyNever:
		return true
	default:
		return false
	}
}
