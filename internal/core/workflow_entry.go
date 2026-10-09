package core

import (
	"context"
	"fmt"
	"strings"
)

// FindWorkflowEntry attaches a compatibility entrypoint to equivalent unfinished
// work without replacing its original target with a newly observed branch head.
// StartWorkflow still verifies the full frozen request and owns all mutations.
func (s *WorkflowStore) FindWorkflowEntry(ctx context.Context, request WorkflowStartRequest) (result WorkflowInvocation, found bool, err error) {
	request.Selection.Target.Commit = strings.Repeat("0", 40)

	request, err = normalizeWorkflowStart(request)
	if err != nil {
		return result, false, err
	}

	err = s.withLock(ctx, func() error {
		runs, readErr := s.repositoryWorkflowInvocations(ctx)
		if readErr != nil {
			return readErr
		}

		result, found, readErr = s.matchWorkflowEntry(runs, request)

		return readErr
	})

	return result, found, err
}

// WorkflowTickets provides a bounded filesystem-owned display view. It neither
// refreshes SQLite nor grants execution authority to unadopted ticket rows.
func (s *WorkflowStore) WorkflowTickets(ctx context.Context) (tickets []IndexedTicket, err error) {
	err = s.withLock(ctx, func() error {
		selection, readErr := s.readWorkflowSelection(ctx)
		tickets = selection.tickets

		return readErr
	})

	return tickets, err
}

// CheckLegacyWorkKey refuses an adopted key before legacy cache/runtime setup.
// This preliminary read is followed by CheckLegacyWorkTicket under the legacy
// operation lock, which excludes adoption through the actual worker start.
func (s *WorkflowStore) CheckLegacyWorkKey(ctx context.Context, key string) error {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	for _, record := range snapshot.Records {
		if record.Task != nil && record.Task.Key == key {
			return ErrWorkflowAuthority
		}
	}

	return nil
}

func (s *WorkflowStore) matchWorkflowEntry(runs []workflowInvocationFile, request WorkflowStartRequest) (result WorkflowInvocation, found bool, err error) {
	for _, run := range runs {
		if run.Namespace != s.layout.Namespace() || workflowInvocationTerminal(run.Status) || run.Selection.Target.Repository != request.Selection.Target.Repository {
			continue
		}

		request.Selection.Target = run.Selection.Target

		hash, hashErr := workflowStartHash(request)
		if hashErr != nil {
			return result, found, hashErr
		}

		if run.RequestHash == hash {
			if found {
				return result, found, fmt.Errorf("%w: ambiguous unfinished entrypoint", ErrWorkflowConflict)
			}

			result, found = run.WorkflowInvocation, true
		}
	}

	return result, found, nil
}
