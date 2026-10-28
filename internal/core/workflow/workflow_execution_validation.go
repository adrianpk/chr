package workflow

import (
	"context"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) validateWorkflowTask(ctx context.Context, claim *Claim, prepared workflowPreparedTask, result *TaskResult) error {
	commit, tree, err := claim.store.workflowLaneCandidate(ctx, prepared.input.Lane)
	if err != nil {
		return err
	}

	result.Commit, result.Tree = commit, tree

	gate := prepared.input.Contract.Validation
	if !gate.Required {
		result.WaiverReason = gate.WaiverReason

		return s.checkWorkflowValidationImage(ctx, claim, prepared, *result)
	}

	seconds := workflowValidationTimeout(gate.TimeoutSeconds, prepared.input.ValidationTimeoutSeconds)

	validationContext, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()

	for _, command := range gate.Commands {
		if err := s.checkWorkflowValidationImage(validationContext, claim, prepared, *result); err != nil {
			return err
		}

		observed, err := workflowRunCommand(validationContext, claim, prepared, command)
		result.Validation = append(result.Validation, observed)

		if err != nil {
			return fmt.Errorf("workflow host validation: %w", err)
		}
	}

	return s.checkWorkflowValidationImage(validationContext, claim, prepared, *result)
}

func (s *Store) checkWorkflowValidationImage(ctx context.Context, claim *Claim, prepared workflowPreparedTask, result TaskResult) error {
	if err := s.checkWorkflowPreparedClaim(ctx, claim, prepared); err != nil {
		return err
	}

	if err := claim.store.checkWorkflowPreparedInput(prepared); err != nil {
		return err
	}

	commit, tree, err := claim.store.workflowLaneCandidate(ctx, prepared.input.Lane)
	if err != nil {
		return err
	}

	if commit != result.Commit || tree != result.Tree {
		return fmt.Errorf("%w: validation changed its immutable Git candidate", workflowstate.ErrConflict)
	}

	return nil
}
