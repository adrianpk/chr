package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// ApplyWorkflowRecordMutation preserves caller-neutral request identity through
// the shared record owner. Namespace may be frozen by a caller's retained scope.
func (s service) ApplyWorkflowRecordMutation(ctx context.Context, namespace string, request core.WorkflowMutationRequest) (core.WorkflowMutationResult, error) {
	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: namespace})
	if err != nil {
		return core.WorkflowMutationResult{}, err
	}

	return core.NewWorkflowStore(layout, s.projectContext.ProjectID).ApplyRecordMutation(ctx, request)
}
