package workflow

import "errors"

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

	if run.Status == "waiting_capacity" {
		return failure
	}

	run.CapacityResumeStatus, run.Status = run.Status, "waiting_capacity"

	return errors.Join(failure, s.commitWorkflowState(&run))
}

func restoreWorkflowCapacityWait(run *workflowInvocationFile) {
	if run.Status != "waiting_capacity" || run.StopIntent != "" {
		return
	}

	run.Status = run.CapacityResumeStatus
	if run.Status == "" {
		run.Status = "pending"
	}

	run.CapacityResumeStatus = ""
}
