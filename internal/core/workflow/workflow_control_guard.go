package workflow

import (
	"context"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func checkWorkflowRevision(run Invocation, expected []uint64) error {
	if len(expected) == 0 {
		return nil
	}

	if len(expected) != 1 || expected[0] == 0 || expected[0] != run.Revision {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowSelectionGuard(ctx context.Context, request StartRequest) error {
	if request.Selected == nil {
		return nil
	}

	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return err
	}

	guard := request.Selected

	record := selection.records[guard.Key]
	if guard.Revision == 0 || record.Task == nil || record.Revision != guard.Revision {
		return workflowstate.ErrConflict
	}

	ticket := ticketcore.Indexed{IndexedIdentity: ticketcore.IndexedIdentity{Key: guard.Key}, IndexedClassification: ticketcore.IndexedClassification{UseCase: record.Task.UseCase}}
	if !selection.matches(request.Selection.Selector, ticket, record) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) acceptWorkflowContinuation(ctx context.Context, id string, expected uint64, checks ...AdmissionCheck) (result Invocation, err error) {
	if len(checks) > 1 {
		return result, workflowstate.ErrConflict
	}

	var check AdmissionCheck
	if len(checks) == 1 {
		check = checks[0]
	}

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if err := checkWorkflowRevision(invocationLocation.Run.Invocation, []uint64{expected}); err != nil {
			return err
		}

		if workflowInvocationTerminal(invocationLocation.Run.Status) {
			result = invocationLocation.Run.Invocation

			return nil
		}

		if err := invocationLocation.Owner.checkWorkflowInputs(ctx, invocationLocation.Run); err != nil {
			return err
		}

		if err := workflowCheckAuthorization(ctx, check); err != nil {
			return err
		}

		// Record acceptance before leaving exclusion. A second UI request with
		// this revision cannot independently start the same continuation.
		if err := invocationLocation.Owner.commitWorkflowContinuation(&invocationLocation.Run, 0, "accepted"); err != nil {
			return err
		}

		result = invocationLocation.Run.Invocation

		return nil
	})

	return result, err
}

func checkWorkflowResumeRevision(run Invocation, ports []ResumePorts) error {
	if len(ports) == 1 && ports[0].ExpectedRevision != 0 {
		return checkWorkflowRevision(run, []uint64{ports[0].ExpectedRevision})
	}

	return nil
}

func workflowCheckAuthorization(ctx context.Context, check AdmissionCheck) error {
	if check == nil {
		return nil
	}

	return check(ctx)
}
