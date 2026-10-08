package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// PreviewWorkflow exposes one adapter-neutral, read-only selection operation.
func (s service) PreviewWorkflow(ctx context.Context, request core.WorkflowPreviewRequest) (core.WorkflowPreview, error) {
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir})
	if err != nil {
		return core.WorkflowPreview{}, err
	}

	return core.NewWorkflowStore(layout, s.projectContext.ProjectID).Preview(ctx, request)
}
