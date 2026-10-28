package closeout

import (
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckInvocation validates ordered closure cursors and the active closing set.
func CheckInvocation(run workflowstate.Invocation) error {
	if len(run.Closures) > len(run.SetImages) {
		return workflowstate.ErrConflict
	}

	cursor := 0
	for i, closure := range run.Closures {
		if !validInvocationClosure(run, i, cursor, closure) {
			return workflowstate.ErrConflict
		}

		cursor = closure.Cursor
	}

	if run.ClosingSet != "" && (run.ClosingSet != NextSet(run) || run.Status != workflowstate.InvocationRunning) {
		return workflowstate.ErrConflict
	}

	return nil
}

func validInvocationClosure(run workflowstate.Invocation, index, cursor int, closure workflowstate.InvocationClosure) bool {
	return closure.SetID == SetID(run.SetImages[index]) && closure.Cursor >= cursor && closure.Cursor <= run.ConfirmedCursor && closure.Target.Repository == run.Selection.Target.Repository && workflowstate.ValidCommitID(closure.Target.Commit) && len(closure.Receipt.Hash) == 64
}

// CheckInvocationEvidence checks each retained successful closure image in order.
func CheckInvocationEvidence(closures []workflowstate.InvocationClosure, evidence Evidence) error {
	for _, closure := range closures {
		data, err := evidence.Read(closure.Receipt.Ref)
		if err != nil || contenthash.Sum(data) != closure.Receipt.Hash {
			return proofError(err)
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
