package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
)

// WorkWorkspaceConfig describes workspace ownership for agent execution.
type WorkWorkspaceConfig struct {
	Mode            WorkWorkspaceMode
	WorktreeRoot    string
	WorktreeBase    WorktreeBasePolicy
	WorktreeCleanup WorktreeCleanupPolicy
}

// WorkWorkspaceMode defines a package boundary type.
type WorkWorkspaceMode string

// WorktreeBasePolicy defines a package boundary type.
type WorktreeBasePolicy string

// WorktreeCleanupPolicy defines a package boundary type.
type WorktreeCleanupPolicy string

// DefaultWorkWorkspaceConfig executes the corresponding package operation.
func DefaultWorkWorkspaceConfig() WorkWorkspaceConfig {
	return WorkWorkspaceConfig{
		Mode:            WorkWorkspaceModeProject,
		WorktreeRoot:    DefaultWorktreeRoot,
		WorktreeBase:    WorktreeBasePolicyRemoteIntegrationBranch,
		WorktreeCleanup: WorktreeCleanupPolicyAfterMerge,
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

// FillWorkWorkspaceConfigDefaults preserves the existing field defaults.
func FillWorkWorkspaceConfigDefaults(config WorkWorkspaceConfig) WorkWorkspaceConfig {
	defaults := DefaultWorkWorkspaceConfig()

	if config.Mode == "" {
		config.Mode = defaults.Mode
	}

	if config.WorktreeRoot == "" {
		config.WorktreeRoot = defaults.WorktreeRoot
	}

	if config.WorktreeBase == "" {
		config.WorktreeBase = defaults.WorktreeBase
	}

	if config.WorktreeCleanup == "" {
		config.WorktreeCleanup = defaults.WorktreeCleanup
	}

	return config
}

const (
	// WorkWorkspaceModeProject defines a package boundary constant.
	WorkWorkspaceModeProject WorkWorkspaceMode = "project"
	// WorkWorkspaceModeWorktree defines a package boundary constant.
	WorkWorkspaceModeWorktree WorkWorkspaceMode = "worktree"
	// WorktreeBasePolicyRemoteIntegrationBranch defines a package boundary constant.
	WorktreeBasePolicyRemoteIntegrationBranch WorktreeBasePolicy = "remote_integration_branch"
	// WorktreeBasePolicyLocalIntegrationBranch defines a package boundary constant.
	WorktreeBasePolicyLocalIntegrationBranch WorktreeBasePolicy = "local_integration_branch"
	// WorktreeBasePolicyCurrentBranch defines a package boundary constant.
	WorktreeBasePolicyCurrentBranch WorktreeBasePolicy = "current_branch"
	// WorktreeCleanupPolicyAfterMerge defines a package boundary constant.
	WorktreeCleanupPolicyAfterMerge WorktreeCleanupPolicy = "after_merge"
	// WorktreeCleanupPolicyManual defines a package boundary constant.
	WorktreeCleanupPolicyManual WorktreeCleanupPolicy = "manual"
	// WorktreeCleanupPolicyNever defines a package boundary constant.
	WorktreeCleanupPolicyNever WorktreeCleanupPolicy = "never"
	// DefaultWorktreeRoot defines a package boundary constant.
	DefaultWorktreeRoot = "ops/default/worktree"
)
