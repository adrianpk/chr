package core

import "context"

func (s *WorkflowStore) prepareWorkflowTask(ctx context.Context, claim *WorkflowClaim, worker []string) (prepared workflowPreparedTask, err error) {
	return s.prepareWorkflowTaskMode(ctx, claim, worker, "automatic")
}
