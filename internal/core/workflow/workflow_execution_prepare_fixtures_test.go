package workflow

import "context"

func (s *Store) prepareWorkflowTask(ctx context.Context, claim *Claim, worker []string) (prepared workflowPreparedTask, err error) {
	return s.prepareWorkflowTaskMode(ctx, claim, worker, "automatic")
}
