package closeout

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Execution binds the frozen closing input to its retained evidence. The host
// supplies the watched deadline context and keeps the mutable claim private.
type Execution struct {
	Input    workflowstate.ClosingInput
	Evidence workflowstate.ClosingEvidence
}

// Execute retains each command's partial result before checking failure, then
// checks the exact checkout. The host must still confirm the returned receipt
// against current authority and inputs in its guarded closing transition.
func Execute(ctx context.Context, execution Execution, workspace Workspace, commands Commands) (workflowstate.ClosingReceipt, error) {
	receipt := workflowstate.ClosingReceipt{ClosingInput: execution.Input, Input: execution.Evidence, Waiver: execution.Input.Validation.WaiverReason}
	receipt.Schema, receipt.Outcome = "chr-set-closure/v1", "closing_validation_failed"

	directory, err := workspace.Prepare(ctx)
	if err != nil {
		return receipt, err
	}

	for _, command := range execution.Input.Validation.Commands {
		result, err := commands.Run(ctx, directory, command)

		receipt.Validation = append(receipt.Validation, result)
		if err != nil {
			return receipt, err
		}

		if result.ExitCode != 0 {
			return receipt, workflowstate.ErrConflict
		}
	}

	if err := workspace.Check(ctx, directory); err != nil {
		return receipt, err
	}

	receipt.Outcome = "completed"

	return receipt, ctx.Err()
}
