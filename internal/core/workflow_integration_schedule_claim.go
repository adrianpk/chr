package core

import "context"

// ClaimNextWorkflowIntegration claims the earliest eligible admitted task.
// Selection and the existing task/capacity transaction share the repository lock.
// It does not admit another unit, resume human waits or replace unknown owners.
func (s *WorkflowStore) ClaimNextWorkflowIntegration(ctx context.Context, q WorkflowIntegrationScheduleQuery) (claim *WorkflowClaim, err error) {
	if err := validateWorkflowIntegrationQueueQuery(q.queueQuery()); err != nil {
		return nil, err
	}

	if q.Offset != 0 {
		return nil, ErrWorkflowConflict
	}

	q.Limit = 1
	err = s.withLock(ctx, func() error {
		if _, err := s.repositoryWorkflowInvocations(ctx); err != nil {
			return err
		}

		page, err := s.workflowIntegrationSchedule(ctx, q)
		if err != nil {
			return err
		}

		if len(page.Items) == 0 {
			return ErrWorkflowOwned
		}

		claim, err = s.claimWorkflowIntegrationSelection(ctx, page.Items[0])

		return err
	})

	return claim, err
}

func (s *WorkflowStore) claimWorkflowIntegrationSelection(ctx context.Context, item WorkflowIntegrationSelection) (*WorkflowClaim, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return nil, err
	}

	run, err := workflowIntegrationScheduleRun(runs, item.Entry)
	if err != nil {
		return nil, err
	}

	owner, err := s.workflowOwnerStore(run)
	if err != nil {
		return nil, err
	}

	if run.Revision != item.InvocationRevision || run.Namespace != item.Entry.Namespace {
		return nil, ErrWorkflowConflict
	}

	return owner.claimWorkflow(ctx, run)
}

func (s *WorkflowStore) guardWorkflowIntegrationExecutor(ctx context.Context, run workflowInvocationFile) error {
	if run.Integration == nil {
		return nil
	}

	if err := s.checkWorkflowIntegrationRetainedRecords(ctx, run); err != nil {
		return err
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if entry.State != "queued" || entry.StopIntent != "" || !workflowIntegrationExecutorFrontier(run, entry) {
		return ErrWorkflowOwned
	}

	return nil
}
