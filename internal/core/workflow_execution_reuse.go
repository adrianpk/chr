package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) reuseWorkflowValidatedTasks(ctx context.Context, run *workflowInvocationFile) error {
	if !slices.Contains([]string{"pending", "paused", "waiting_capacity"}, run.Status) {
		return nil
	}

	if err := s.checkWorkflowInputs(ctx, *run); err != nil {
		return err
	}

	originalCursor := run.Cursor
	for run.Cursor < len(run.Members) {
		record, err := s.readRecord(run.Members[run.Cursor].Admission.Path)
		if err != nil {
			return err
		}

		if !slices.Contains([]string{"validated", "waiting_slice"}, record.Task.Phase) {
			break
		}

		if err := s.reuseWorkflowTask(ctx, run, record); err != nil {
			return err
		}

		run.Cursor++
		if workflowTaskBoundary(run, record) {
			break
		}
	}

	if run.Cursor != originalCursor {
		return s.commitWorkflowState(run)
	}

	return nil
}

func (s *WorkflowStore) reuseWorkflowTask(ctx context.Context, run *workflowInvocationFile, record WorkflowRecord) error {
	if err := s.checkWorkflowTarget(ctx, run.ExecutionSelection().Target); err != nil {
		return err
	}

	lane, err := s.retainedWorkflowLane(ctx, *run, record)
	if err != nil || lane == nil {
		return fmt.Errorf("%w: validated task has no retained lane", workflowstate.ErrConflict)
	}

	if err := s.checkWorkflowReusableTask(ctx, *run, record, *lane); err != nil {
		return err
	}

	run.Lane = lane

	return nil
}

func workflowTaskBoundary(run *workflowInvocationFile, record WorkflowRecord) bool {
	run.Outcome = "task_validated"
	if run.Selection.Selector.Extent == "ticket" && record.Task.SliceID != "" {
		run.Status = "completed"

		return true
	}

	if record.Task.SliceID == "" || run.Cursor == len(run.Members) || run.Members[run.Cursor].Admission.SliceID != record.Task.SliceID {
		run.Status = "waiting_integration"

		return true
	}

	return false
}

func (s *WorkflowStore) checkWorkflowReusableTask(ctx context.Context, run workflowInvocationFile, record WorkflowRecord, lane WorkflowLane) error {
	return s.checkWorkflowReusableTaskPrepared(ctx, run, record, lane, (*WorkflowStore).recoverTransition)
}

func (s *WorkflowStore) checkWorkflowReusableTaskPrepared(ctx context.Context, run workflowInvocationFile, record WorkflowRecord, lane WorkflowLane, prepare func(*WorkflowStore) error) error {
	if _, err := s.observeWorkflowLane(ctx, lane); err != nil {
		return err
	}

	runs, err := s.repositoryWorkflowInvocationsPrepared(ctx, prepare)
	if err != nil {
		return err
	}

	request := run.ExecutionSelection()
	request.CumulativeBranch, request.CumulativeCommit = lane.Branch, lane.HeadCommit
	reader := workflowProofReader{store: s, request: request}

	for _, previous := range runs {
		if previous.Namespace != run.Namespace {
			continue
		}

		for _, attempt := range previous.Attempts {
			if !workflowAttemptValidatesTask(attempt, record) {
				continue
			}

			if s.workflowRetainedResultMatches(ctx, previous, attempt, record, &reader) {
				return nil
			}
		}
	}

	return fmt.Errorf("%w: reusable task lacks unchanged host evidence", workflowstate.ErrConflict)
}

func workflowAttemptValidatesTask(attempt WorkflowAttempt, record WorkflowRecord) bool {
	return attempt.Purpose == "" && attempt.TaskKey == record.Task.Key && attempt.ContractHash == record.Task.ContractHash && attempt.Outcome == "validated" && slices.Contains(record.Task.Evidence, attempt.ResultRef)
}

func (s *WorkflowStore) workflowRetainedResultMatches(ctx context.Context, run workflowInvocationFile, attempt WorkflowAttempt, record WorkflowRecord, reader *workflowProofReader) bool {
	owner, err := s.workflowOwnerStore(run)
	if err != nil {
		return false
	}

	path, err := owner.safePath(attempt.ResultRef)
	if err != nil {
		return false
	}

	data, err := readWorkflowFile(path)
	if err != nil || contenthash.Sum(data) != attempt.ResultHash {
		return false
	}

	var result WorkflowTaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return false
	}

	return result.AttemptID == attempt.ID && result.Generation == attempt.Generation && result.InputHash == attempt.InputHash && owner.workflowRetainedInputMatches(attempt, record) && reader.validTaskProof(ctx, record, result.WorkflowTaskValidation)
}

func (s *WorkflowStore) workflowRetainedInputMatches(attempt WorkflowAttempt, record WorkflowRecord) bool {
	path, err := s.safePath(attempt.InputRef)
	if err != nil {
		return false
	}

	data, err := readWorkflowFile(path)
	if err != nil || contenthash.Sum(data) != attempt.InputHash {
		return false
	}

	var input WorkflowTaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		return false
	}

	return input.AttemptID == attempt.ID && input.ContractHash == record.Task.ContractHash && input.Body == string(record.Body)
}
