package core

import (
	"context"
	"slices"
	"strings"
	"time"
)

// WorkflowInspection observes retained authorization and current task records.
// Adapters project it without exposing process identities or owner credentials.
type WorkflowInspection struct {
	Integration *WorkflowIntegrationInspection
	Invocation  WorkflowInvocation
	Policy      WorkflowReviewPolicy
	Tasks       []WorkflowRecord
}

// InspectWorkflow does not replay transitions, reconcile ownership or admit work.
func (s *WorkflowStore) InspectWorkflow(ctx context.Context, id string) (result WorkflowInspection, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withWorkflowInspection(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		result.Invocation = run.WorkflowInvocation

		result.Policy, readErr = owner.workflowReviewPolicy(run)
		if readErr != nil {
			return readErr
		}

		snapshot, readErr := owner.readSnapshot(ctx)
		result.Tasks = snapshot.Records

		if readErr != nil {
			return readErr
		}

		result.Integration, readErr = owner.inspectWorkflowIntegration(ctx, run)

		return readErr
	})

	return result, err
}

// InspectWorkflowPolicy reports current publication policy without changing state.
func (s *WorkflowStore) InspectWorkflowPolicy(ctx context.Context) (policy WorkflowReviewPolicy, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withWorkflowInspection(ctx, func() error {
		var readErr error

		policy, readErr = s.workflowReviewPolicy(workflowInvocationFile{})

		return readErr
	})

	return policy, err
}

// InspectWorkflowRecords reads native selection identities without recovery writes.
func (s *WorkflowStore) InspectWorkflowRecords(ctx context.Context) (snapshot WorkflowSnapshot, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withWorkflowInspection(ctx, func() error {
		var readErr error

		snapshot, readErr = s.readSnapshot(ctx)

		return readErr
	})

	return snapshot, err
}

// InspectWorkflowPage returns retained invocations in stable ID order.
// It does not reconcile owners, replay journals or start work.
func (s *WorkflowStore) InspectWorkflowPage(ctx context.Context, offset, limit int) (page []WorkflowInvocation, next int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	next = -1
	if offset < 0 || offset > WorkflowRecordLimit || limit < 1 || limit > WorkflowPageLimit {
		return nil, next, ErrWorkflowConflict
	}

	err = s.withWorkflowInspection(ctx, func() error {
		runs, readErr := s.repositoryWorkflowInvocations(ctx)
		if readErr != nil {
			return readErr
		}

		runs = slices.DeleteFunc(runs, func(run workflowInvocationFile) bool { return run.Namespace != s.layout.Namespace() })

		slices.SortFunc(runs, func(a, b workflowInvocationFile) int { return strings.Compare(a.ID, b.ID) })

		start, end := min(offset, len(runs)), min(offset+limit, len(runs))
		page = []WorkflowInvocation{}

		for _, run := range runs[start:end] {
			page = append(page, run.WorkflowInvocation)
		}

		if end < len(runs) {
			next = end
		}

		return nil
	})

	return page, next, err
}

func (s *WorkflowStore) withWorkflowInspection(ctx context.Context, action func() error) error {
	return s.withRepositoryLock(ctx, func() error {
		layouts, err := s.workflowRepositoryLayouts(ctx)
		if err != nil {
			return err
		}

		for _, layout := range layouts {
			store := NewWorkflowStore(layout, s.projectID)
			if err := store.requireNoWorkflowTransition(); err != nil {
				return err
			}
		}

		return nil
	}, action)
}
