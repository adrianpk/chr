package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkflowConfiguration reports current values and sources without state changes.
func (s service) WorkflowConfiguration(ctx context.Context, restriction core.ReviewWorkflowMode) (core.WorkflowConfiguration, error) {
	if err := ctx.Err(); err != nil {
		return core.WorkflowConfiguration{}, err
	}

	return core.ResolveWorkflowConfiguration(s.projectContext.ProjectDir, restriction)
}
