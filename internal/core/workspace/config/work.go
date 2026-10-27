package config

import (
	"fmt"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
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

	if err := executioncore.ValidateWorkBranchPolicy(config.BranchPolicy); err != nil {
		return err
	}

	return workspacecore.ValidateExecutionConfig(config.Workspace)
}
