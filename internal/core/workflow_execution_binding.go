package core

import (
	"context"
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowExecutionScope struct {
	RequestHash string
	Members     []WorkflowMember
	Sets        []WorkflowInputImage
	Cursor      int
	Lane        *WorkflowLane
}

func workflowExecutionScopeHash(run workflowInvocationFile) (string, error) {
	data, err := json.Marshal(workflowExecutionScope{RequestHash: run.RequestHash, Members: run.Members, Sets: run.SetImages, Cursor: run.Cursor, Lane: run.Lane})

	return contenthash.Sum(data), err
}

func (s *WorkflowStore) checkWorkflowPreparedClaim(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask) error {
	return s.withLock(ctx, func() error {
		run, record, err := s.checkWorkflowClaim(ctx, claim)
		if err != nil {
			return err
		}

		if run.StopIntent != "" {
			return ErrWorkflowStop
		}

		if input := prepared.input.Preparation; input != nil {
			if err := claim.store.checkWorkflowNativePreparationFence(run, input.Native); err != nil {
				return err
			}
		}

		if input := prepared.input.Review; input != nil {
			if err := claim.store.checkWorkflowNativeReviewFence(run, input); err != nil {
				return err
			}
		}

		return checkWorkflowExecutionBinding(run, record, prepared)
	})
}

func checkWorkflowExecutionBinding(run workflowInvocationFile, record WorkflowRecord, prepared workflowPreparedTask) error {
	if err := checkWorkflowTaskHarness(run, prepared.input); err != nil {
		return err
	}

	hash, err := workflowExecutionScopeHash(run)
	if err != nil || hash != prepared.input.ScopeHash {
		return fmt.Errorf("%w: frozen execution checkpoint changed", workflowstate.ErrConflict)
	}

	if record.Task.Key != prepared.input.TaskKey || record.SourceHash != prepared.input.SourceHash {
		return fmt.Errorf("%w: prepared task image changed", workflowstate.ErrConflict)
	}

	attempt := run.Attempts[len(run.Attempts)-1]
	if !workflowPreparedAttemptMatches(attempt, prepared) {
		return fmt.Errorf("%w: frozen attempt binding changed", workflowstate.ErrConflict)
	}

	return nil
}

func workflowPreparedAttemptMatches(attempt WorkflowAttempt, prepared workflowPreparedTask) bool {
	input := prepared.input

	return attempt.InputRef == prepared.inputRef && attempt.InputHash == prepared.inputHash && attempt.InputRevision == input.InputRevision && attempt.ContractHash == input.ContractHash && attempt.LeaseUntil.Equal(prepared.leaseUntil) && attempt.Lane != nil && *attempt.Lane == input.Lane
}
