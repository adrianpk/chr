package core

import "context"

// WorkflowSelectionGuard binds admission to the selected native ticket image.
// It is admission input, not a retained credential or an execution override.
type WorkflowSelectionGuard struct {
	Key      string
	Revision uint64
}

func checkWorkflowRevision(run WorkflowInvocation, expected []uint64) error {
	if len(expected) == 0 {
		return nil
	}

	if len(expected) != 1 || expected[0] == 0 || expected[0] != run.Revision {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowSelectionGuard(ctx context.Context, request WorkflowStartRequest) error {
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
		return ErrWorkflowConflict
	}

	ticket := IndexedTicket{IndexedTicketIdentity: IndexedTicketIdentity{Key: guard.Key}, IndexedTicketClassification: IndexedTicketClassification{UseCase: record.Task.UseCase}}
	if !selection.matches(request.Selection.Selector, ticket, record) {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) acceptWorkflowContinuation(ctx context.Context, id string, expected uint64, checks ...WorkflowAdmissionCheck) (result WorkflowInvocation, err error) {
	if len(checks) > 1 {
		return result, ErrWorkflowConflict
	}

	var check WorkflowAdmissionCheck
	if len(checks) == 1 {
		check = checks[0]
	}

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if err := checkWorkflowRevision(run.WorkflowInvocation, []uint64{expected}); err != nil {
			return err
		}

		if workflowInvocationTerminal(run.Status) {
			result = run.WorkflowInvocation

			return nil
		}

		if err := owner.checkWorkflowInputs(ctx, run); err != nil {
			return err
		}

		if err := workflowCheckAuthorization(ctx, check); err != nil {
			return err
		}

		// Record acceptance before leaving exclusion. A second UI request with
		// this revision cannot independently start the same continuation.
		if err := owner.commitWorkflowContinuation(&run, 0, "accepted"); err != nil {
			return err
		}

		result = run.WorkflowInvocation

		return nil
	})

	return result, err
}

func checkWorkflowResumeRevision(run WorkflowInvocation, ports []WorkflowResumePorts) error {
	if len(ports) == 1 && ports[0].ExpectedRevision != 0 {
		return checkWorkflowRevision(run, []uint64{ports[0].ExpectedRevision})
	}

	return nil
}

func workflowCheckAuthorization(ctx context.Context, check WorkflowAdmissionCheck) error {
	if check == nil {
		return nil
	}

	return check(ctx)
}
