package app

import (
	"context"
	"errors"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ExecuteConfirmedWorkExtent admits the displayed snapshot and policy under
// the canonical guard, then performs only the requested bounded checkpoint.
func (s service) ExecuteConfirmedWorkExtent(ctx context.Context, input workflowstate.StartInput, steps int) (AgentWorkResult, error) {
	if input.Admission == nil || len(input.Selection.SnapshotHash) != 64 || input.Selection.Offset != 0 || steps < 1 || steps > workflowstate.PageLimit {
		return AgentWorkResult{}, workflowstate.ErrConflict
	}

	s.ctx = ctx
	if err := s.checkConfirmedWorkExtent(ctx, input); err != nil {
		return AgentWorkResult{}, err
	}

	store, err := s.workflowStore()
	if err != nil {
		return AgentWorkResult{}, err
	}

	request := workflowcore.StartRequest{StartInput: input, ResolveHarness: s.prepareWorkflowHarness}

	started, err := store.StartWorkflowGuarded(ctx, request, s.boundAdmissionCheck(input.Selection, nil))
	if err != nil {
		return AgentWorkResult{}, err
	}

	progress, err := s.ContinueWorkExtent(ctx, started.Invocation.ID, steps, started.Invocation.Revision)
	if progress.Invocation.ID == "" {
		progress.Invocation = started.Invocation
	}

	result, readErr := s.nativeAgentWorkResult(progress.Invocation)
	result.Message += fmt.Sprintf(" Accepted steps: %d; checkpoint: %s.", progress.Steps, progress.Reason)

	return result, errors.Join(err, readErr)
}

func (s service) checkConfirmedWorkExtent(ctx context.Context, input workflowstate.StartInput) error {
	preview, err := s.PreviewWorkExtent(ctx, input.Selection)
	if err != nil {
		return err
	}

	if input.Admission.Policy != preview.Policy || input.Admission.ExecutionPolicy != preview.ExecutionPolicy {
		return workflowstate.ErrConflict
	}

	return nil
}
