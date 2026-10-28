package workflow

import (
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowInvocationClosures(run workflowInvocationFile) error {
	if len(run.Closures) > len(run.SetImages) {
		return workflowstate.ErrConflict
	}

	cursor := 0
	for i, closure := range run.Closures {
		if !validWorkflowInvocationClosure(run, i, cursor, closure) {
			return workflowstate.ErrConflict
		}

		cursor = closure.Cursor
	}

	if run.ClosingSet != "" && (run.ClosingSet != workflowNextClosingSet(run.Invocation) || run.Status != workflowstate.InvocationRunning) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowInvocationClosureEvidence(run workflowInvocationFile) error {
	for _, closure := range run.Closures {
		data, err := s.workflowClosingRead(closure.Receipt.Ref)
		if err != nil || contenthash.Sum(data) != closure.Receipt.Hash {
			return errorsOrConflict(err)
		}

		var receipt workflowstate.ClosingReceipt
		if err := json.Unmarshal(data, &receipt); err != nil {
			return err
		}

		if receipt.Schema != "chr-set-closure/v1" || receipt.SetID != closure.SetID || receipt.Target != closure.Target || receipt.Outcome != "completed" {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func validWorkflowInvocationClosure(run workflowInvocationFile, index, cursor int, closure InvocationClosure) bool {
	return closure.SetID == workflowInvocationSetID(run.SetImages[index]) && closure.Cursor >= cursor && closure.Cursor <= run.ConfirmedCursor && closure.Target.Repository == run.Selection.Target.Repository && workflowCommitID(closure.Target.Commit) && len(closure.Receipt.Hash) == 64
}
