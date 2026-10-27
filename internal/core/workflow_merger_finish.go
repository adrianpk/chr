package core

import (
	"context"
	"errors"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) finishWorkflowMerger(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, provider WorkflowPullRequestProvider, attempt *WorkflowMergerAttempt, executionErr error) error {
	freshErr := s.checkWorkflowMergerPrepared(ctx, claim, prepared, provider)

	return s.withLock(ctx, func() error {
		run, record, err := s.checkWorkflowCredential(ctx, claim, false)
		if err != nil {
			return err
		}

		if err := workflowAttemptQuiescent(run.Attempts[len(run.Attempts)-1]); err != nil {
			return err
		}

		bindingErr := checkWorkflowExecutionBinding(run, record, prepared)
		if err := s.checkWorkflowPreparedInput(prepared); err != nil {
			bindingErr = errors.Join(bindingErr, err)
		}

		failure := errors.Join(executionErr, freshErr, bindingErr)
		if failure != nil {
			attempt.Proposal, attempt.Failure = nil, workflowFailureText(failure)
			rejectWorkflowMergerAcceptance(attempt)
		}

		if err := boundWorkflowMergerAttempt(attempt); err != nil {
			failure = errors.Join(failure, err)
		}

		if err := s.persistWorkflowMergerFinish(ctx, &run, record, claim, *attempt); err != nil {
			return errors.Join(failure, err)
		}

		return errors.Join(failure, claim.releaseReservation())
	})
}

func (s *WorkflowStore) persistWorkflowMergerFinish(ctx context.Context, run *workflowInvocationFile, record WorkflowRecord, claim *WorkflowClaim, result WorkflowMergerAttempt) error {
	data, err := encodeWorkflowMergerAttempt(result)
	if err != nil {
		return err
	}

	budget, err := s.readWorkflowIntegrationBudget(*run)
	if err != nil {
		return err
	}

	index := slices.IndexFunc(budget.Reservations, func(r WorkflowIntegrationReservation) bool {
		return r.ID == result.Input.BudgetReservationID && r.AttemptID == claim.AttemptID
	})
	if index < 0 || !slices.Contains([]string{"reserved", "launched"}, budget.Reservations[index].State) {
		return workflowstate.ErrConflict
	}
	// Dispatch uncertainty and unmetered calls retain their charge. Whole-call
	// observations remain available to the existing budget settlement owner.
	budget.Reservations[index].State = "unknown"

	settlement, err := workflowMergerTokenSettlement(&budget.Reservations[index], result)
	if err != nil {
		return err
	}

	resultFile := workflowMergerFinishedRun(run, claim, result, data)
	record.Task.AttemptID = ""

	files, err := s.workflowClaimRecordChanges(run, &record)
	if err != nil {
		return err
	}

	file, err := s.workflowIntegrationBudgetChange(ctx, run, &budget)
	if err != nil {
		return err
	}

	decisionFiles, err := s.workflowMergerDecisionFiles(ctx, run, result)
	if err != nil {
		return err
	}

	files = append(files, decisionFiles...)

	files = append(files, file)
	files = appendWorkflowMergerSettlement(files, settlement)

	return s.commitWorkflowPullRequestFiles(run, files, resultFile)
}

func workflowMergerFinishedRun(run *workflowInvocationFile, claim *WorkflowClaim, result WorkflowMergerAttempt, data []byte) workflowstate.FileChange {
	current := &run.Attempts[len(run.Attempts)-1]
	current.Outcome, current.FinishedAt = "checkpointed", time.Now().UTC()
	current.ResultRef, current.ResultHash = workflowAttemptFile(run.ID, claim.AttemptID, "result.json"), contenthash.Sum(data)

	run.Status, run.Outcome = "waiting_integration", "merger_proposal_pending"
	if result.Acceptance != nil && result.Acceptance.Outcome == "accepted_pending_validation" {
		run.Outcome = "merger_repair_pending_validation"
	}

	if result.Proposal == nil || result.Proposal.Outcome == "failed" {
		current.Outcome, run.Outcome = "failed", "merger_failed"
	}

	if run.StopIntent != "" {
		run.Status, current.Outcome = workflowstate.StopStatus(run.StopIntent), workflowstate.StopStatus(run.StopIntent)
	}

	if run.Status == "paused" {
		run.ResumeStatus = "waiting_integration"
	}

	return workflowstate.FileChange{Path: current.ResultRef, After: data}
}

func rejectWorkflowMergerAcceptance(attempt *WorkflowMergerAttempt) {
	if attempt.Acceptance != nil {
		attempt.Acceptance.Outcome, attempt.Acceptance.Failure = "rejected", attempt.Failure
	}
}

func boundWorkflowMergerAttempt(attempt *WorkflowMergerAttempt) error {
	if _, err := encodeWorkflowMergerAttempt(*attempt); err == nil {
		return nil
	}

	attempt.Proposal, attempt.Failure = nil, "Merger proposal exceeds the retained attempt image bound."
	if attempt.Acceptance != nil {
		attempt.Acceptance.Next, attempt.Acceptance.Repair = nil, nil
	}

	rejectWorkflowMergerAcceptance(attempt)

	attempt.Harness.Stdout, attempt.Harness.Stderr, attempt.Harness.Truncated = "", "", true

	return workflowstate.ErrConflict
}

func appendWorkflowMergerSettlement(files []workflowstate.FileChange, settlement workflowstate.FileChange) []workflowstate.FileChange {
	if settlement.Path != "" {
		return append(files, settlement)
	}

	return files
}
