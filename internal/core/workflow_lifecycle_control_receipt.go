package core

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) newWorkflowControlCommit(request WorkflowControlMutationRequest, digest string, check WorkflowAdmissionCheck) *workflowControlCommit {
	result := WorkflowControlResult{RequestID: request.Binding.RequestID, Action: request.Action, InvocationID: request.ID, ArgumentDigest: digest, ReceiptRef: workflowstate.MutationReceiptPath(request.Binding.RequestID), PriorRevision: request.ExpectedRevision, Revision: request.ExpectedRevision + 1}

	return &workflowControlCommit{receipt: workflowControlReceipt{Schema: "chr-workflow-control/v1", State: "confirmed", ProjectID: s.projectID, Namespace: s.layout.Namespace(), Operator: request.Binding.Operator, Result: result}, check: check}
}

func (s *WorkflowStore) commitWorkflowControlState(ctx context.Context, run *workflowInvocationFile, control *workflowControlCommit) error {
	files, err := control.files(ctx, run.Revision+1)
	if err != nil {
		return err
	}

	return s.commitWorkflowState(run, files...)
}

func (s *WorkflowStore) retainWorkflowClosingControlNoop(ctx context.Context, control *workflowControlCommit) error {
	if control == nil {
		return nil
	}

	run, err := s.readWorkflowInvocation(control.receipt.Result.InvocationID)
	if err != nil {
		return err
	}

	return s.retainWorkflowControlNoop(ctx, run, control)
}

func (s *WorkflowStore) workflowClosingControlFiles(ctx context.Context, claim *workflowClosingClaim, control *workflowControlCommit) ([]workflowstate.FileChange, error) {
	if control == nil {
		return nil, nil
	}

	if claim.state.InvocationID != control.receipt.Result.InvocationID {
		return nil, workflowstate.ErrConflict
	}

	run, err := s.readWorkflowInvocation(claim.state.InvocationID)
	if err != nil {
		return nil, err
	}

	files, err := control.files(ctx, run.Revision+1)
	if err != nil || !workflowInvocationTerminal(run.Status) {
		return files, err
	}

	// Closing recovery keeps terminal invocation state. Its bound receipt still
	// needs a revision in the same transaction as the recovered closing record.
	run.Revision++
	run.UpdatedAt = time.Now().UTC()
	change, err := workflowInvocationChange(run)

	return append(files, change), err
}

func (s *WorkflowStore) commitWorkflowControlResume(ctx context.Context, run *workflowInvocationFile, changes []workflowstate.FileChange, control *workflowControlCommit) error {
	files, err := control.files(ctx, run.Revision+1)
	if err != nil {
		return err
	}

	return s.commitWorkflowResume(run, append(changes, files...))
}
