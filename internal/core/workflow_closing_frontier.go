package core

import (
	"encoding/json"
)

func validateWorkflowInvocationClosures(run workflowInvocationFile) error {
	if len(run.Closures) > len(run.SetImages) {
		return ErrWorkflowConflict
	}

	cursor := 0
	for i, closure := range run.Closures {
		if !validWorkflowInvocationClosure(run, i, cursor, closure) {
			return ErrWorkflowConflict
		}

		cursor = closure.Cursor
	}

	if run.ClosingSet != "" && (run.ClosingSet != workflowNextClosingSet(run.WorkflowInvocation) || run.Status != "running") {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowInvocationClosureEvidence(run workflowInvocationFile) error {
	for _, closure := range run.Closures {
		data, err := s.workflowClosingRead(closure.Receipt.Ref)
		if err != nil || workflowHash(data) != closure.Receipt.Hash {
			return errorsOrConflict(err)
		}

		var receipt WorkflowClosingReceipt
		if err := json.Unmarshal(data, &receipt); err != nil {
			return err
		}

		if receipt.Schema != "chr-set-closure/v1" || receipt.SetID != closure.SetID || receipt.Target != closure.Target || receipt.Outcome != "completed" {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func validWorkflowInvocationClosure(run workflowInvocationFile, index, cursor int, closure WorkflowInvocationClosure) bool {
	return closure.SetID == workflowInvocationSetID(run.SetImages[index]) && closure.Cursor >= cursor && closure.Cursor <= run.ConfirmedCursor && closure.Target.Repository == run.Selection.Target.Repository && workflowCommitID(closure.Target.Commit) && len(closure.Receipt.Hash) == 64
}
