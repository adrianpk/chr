package workflow

import (
	"context"
	"slices"
	"strings"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Inspection observes retained authorization and current task records.
// Adapters project it without exposing process identities or owner credentials.
type Inspection struct {
	Integration *IntegrationInspection
	Invocation  Invocation
	Policy      workflowstate.ReviewPolicy
	Tasks       []Record
}

// InspectWorkflow does not replay transitions, reconcile ownership or admit work.
func (s *Store) InspectWorkflow(ctx context.Context, id string) (result Inspection, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withWorkflowInspection(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		result.Invocation = run.Invocation

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
func (s *Store) InspectWorkflowPolicy(ctx context.Context) (policy workflowstate.ReviewPolicy, err error) {
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
func (s *Store) InspectWorkflowRecords(ctx context.Context) (snapshot Snapshot, err error) {
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
func (s *Store) InspectWorkflowPage(ctx context.Context, offset, limit int) (page []Invocation, next int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	next = -1
	if offset < 0 || offset > workflowstate.RecordLimit || limit < 1 || limit > workflowstate.PageLimit {
		return nil, next, workflowstate.ErrConflict
	}

	err = s.withWorkflowInspection(ctx, func() error {
		runs, readErr := s.repositoryWorkflowInvocations(ctx)
		if readErr != nil {
			return readErr
		}

		runs = slices.DeleteFunc(runs, func(run workflowInvocationFile) bool { return run.Namespace != s.layout.Namespace() })

		slices.SortFunc(runs, func(a, b workflowInvocationFile) int { return strings.Compare(a.ID, b.ID) })

		start, end := min(offset, len(runs)), min(offset+limit, len(runs))
		page = []Invocation{}

		for _, run := range runs[start:end] {
			page = append(page, run.Invocation)
		}

		if end < len(runs) {
			next = end
		}

		return nil
	})

	return page, next, err
}

func (s *Store) withWorkflowInspection(ctx context.Context, action func() error) error {
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
