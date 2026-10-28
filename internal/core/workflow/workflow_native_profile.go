package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// NativePolicyReader rechecks trusted ownership, credential roles and
// provider observation certification before each admission or intent check.
type NativePolicyReader interface {
	ReadWorkflowNativePolicy(context.Context, workflowstate.PullRequestSnapshot, string) (workflowstate.NativePolicy, error)
}

// NativeCapabilityRequest supplies the explicit profile and exact source.
type NativeCapabilityRequest struct {
	Config      workspaceconfig.WorkIntegrationConfig
	PullRequest workflowstate.PullRequestSnapshot
	Publication GitPublication
}
