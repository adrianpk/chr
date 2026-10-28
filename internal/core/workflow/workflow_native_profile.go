package workflow

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// NativeCapabilityRequest supplies the explicit profile and exact source.
type NativeCapabilityRequest struct {
	Config      workspaceconfig.WorkIntegrationConfig
	PullRequest workflowstate.PullRequestSnapshot
	Publication GitPublication
}
