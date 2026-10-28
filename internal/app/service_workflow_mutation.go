package app

import (
	"context"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// ApplyWorkflowRecordMutation preserves caller-neutral request identity through
// the shared record owner. Namespace may be frozen by a caller's retained scope.
func (s service) ApplyWorkflowRecordMutation(ctx context.Context, namespace string, request workflowcore.MutationRequest) (workflowcore.MutationResult, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: namespace})
	if err != nil {
		return workflowcore.MutationResult{}, err
	}

	return workflowcore.NewWorkflowStore(layout, s.projectContext.ProjectID).ApplyRecordMutation(ctx, request)
}
