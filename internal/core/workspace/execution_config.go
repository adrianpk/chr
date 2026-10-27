package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ExecutionConfig describes workspace ownership for agent execution.
type ExecutionConfig struct {
	Mode            ExecutionMode
	WorktreeRoot    string
	WorktreeBase    BasePolicy
	WorktreeCleanup CleanupPolicy
}

// ExecutionMode defines a package boundary type.
type ExecutionMode string

// BasePolicy defines a package boundary type.
type BasePolicy string

// CleanupPolicy defines a package boundary type.
type CleanupPolicy string

// DefaultExecutionConfig executes the corresponding package operation.
func DefaultExecutionConfig() ExecutionConfig {
	return ExecutionConfig{
		Mode:            ExecutionProject,
		WorktreeRoot:    DefaultWorktreeRoot,
		WorktreeBase:    BaseRemoteIntegrationBranch,
		WorktreeCleanup: CleanupAfterMerge,
	}
}

// ValidateExecutionConfig validates the corresponding package boundary value.
func ValidateExecutionConfig(config ExecutionConfig) error {
	if config == (ExecutionConfig{}) {
		return nil
	}

	if !ValidExecutionMode(config.Mode) {
		return fmt.Errorf("invalid work workspace_mode: %s", config.Mode)
	}

	if config.Mode != ExecutionWorktree {
		return nil
	}

	if strings.TrimSpace(config.WorktreeRoot) == "" {
		return fmt.Errorf("work worktree_root is required")
	}

	if filepath.ToSlash(filepath.Clean(config.WorktreeRoot)) != DefaultWorktreeRoot {
		return fmt.Errorf("work worktree_root must be %s", DefaultWorktreeRoot)
	}

	if !ValidBasePolicy(config.WorktreeBase) {
		return fmt.Errorf("invalid work worktree_base: %s", config.WorktreeBase)
	}

	if !ValidCleanupPolicy(config.WorktreeCleanup) {
		return fmt.Errorf("invalid work worktree_cleanup: %s", config.WorktreeCleanup)
	}

	return nil
}

// ValidExecutionMode executes the corresponding package operation.
func ValidExecutionMode(mode ExecutionMode) bool {
	switch ExecutionMode(strings.TrimSpace(string(mode))) {
	case ExecutionProject, ExecutionWorktree:
		return true
	default:
		return false
	}
}

// ValidBasePolicy executes the corresponding package operation.
func ValidBasePolicy(policy BasePolicy) bool {
	switch BasePolicy(strings.TrimSpace(string(policy))) {
	case BaseRemoteIntegrationBranch,
		BaseLocalIntegrationBranch,
		BaseCurrentBranch:
		return true
	default:
		return false
	}
}

// ValidCleanupPolicy executes the corresponding package operation.
func ValidCleanupPolicy(policy CleanupPolicy) bool {
	switch CleanupPolicy(strings.TrimSpace(string(policy))) {
	case CleanupAfterMerge,
		CleanupManual,
		CleanupNever:
		return true
	default:
		return false
	}
}

// FillExecutionDefaults preserves the existing field defaults.
func FillExecutionDefaults(config ExecutionConfig) ExecutionConfig {
	defaults := DefaultExecutionConfig()

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
	// ExecutionProject defines a package boundary constant.
	ExecutionProject ExecutionMode = "project"
	// ExecutionWorktree defines a package boundary constant.
	ExecutionWorktree ExecutionMode = "worktree"
	// BaseRemoteIntegrationBranch defines a package boundary constant.
	BaseRemoteIntegrationBranch BasePolicy = "remote_integration_branch"
	// BaseLocalIntegrationBranch defines a package boundary constant.
	BaseLocalIntegrationBranch BasePolicy = "local_integration_branch"
	// BaseCurrentBranch defines a package boundary constant.
	BaseCurrentBranch BasePolicy = "current_branch"
	// CleanupAfterMerge defines a package boundary constant.
	CleanupAfterMerge CleanupPolicy = "after_merge"
	// CleanupManual defines a package boundary constant.
	CleanupManual CleanupPolicy = "manual"
	// CleanupNever defines a package boundary constant.
	CleanupNever CleanupPolicy = "never"
	// DefaultWorktreeRoot defines a package boundary constant.
	DefaultWorktreeRoot = "ops/default/worktree"
)
