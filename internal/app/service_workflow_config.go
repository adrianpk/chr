package app

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// WorkflowConfiguration reports current values and sources without state changes.
func (s service) WorkflowConfiguration(ctx context.Context, restriction workflowstate.ReviewMode) (workspaceconfig.WorkflowConfiguration, error) {
	if err := ctx.Err(); err != nil {
		return workspaceconfig.WorkflowConfiguration{}, err
	}

	return workspaceconfig.ResolveWorkflowConfiguration(s.projectContext.ProjectDir, restriction)
}
