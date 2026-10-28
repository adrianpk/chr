package workflow

import (
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Capacity waits retain their previous phase without consuming an attempt or
// changing task/review/preparation evidence. An explicit retry checks capacity.
func (s *Store) retainWorkflowCapacityWait(observed workflowInvocationFile, failure error) error {
	if !errors.Is(failure, ErrWorkflowCapacity) || observed.Revision == 0 {
		return failure
	}

	run, err := s.readWorkflowInvocation(observed.ID)
	if err != nil {
		return errors.Join(failure, err)
	}

	if run.Status == workflowstate.InvocationWaitingCapacity {
		return failure
	}

	run.CapacityResumeStatus, run.Status = run.Status, workflowstate.InvocationWaitingCapacity

	return errors.Join(failure, s.commitWorkflowState(&run))
}

func restoreWorkflowCapacityWait(run *workflowInvocationFile) {
	if run.Status != workflowstate.InvocationWaitingCapacity || run.StopIntent != "" {
		return
	}

	run.Status = run.CapacityResumeStatus
	if run.Status == workflowstate.InvocationUnspecified {
		run.Status = workflowstate.InvocationPending
	}

	run.CapacityResumeStatus = workflowstate.InvocationUnspecified
}
