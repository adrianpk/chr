package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) reuseWorkflowValidatedTasks(ctx context.Context, run *workflowInvocationFile) error {
	if !slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationPending, workflowstate.InvocationPaused, workflowstate.InvocationWaitingCapacity}, run.Status) {
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

func (s *Store) reuseWorkflowTask(ctx context.Context, run *workflowInvocationFile, record workflowstate.Record) error {
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

func workflowTaskBoundary(run *workflowInvocationFile, record workflowstate.Record) bool {
	run.Outcome = "task_validated"
	if run.Selection.Selector.Extent == "ticket" && record.Task.SliceID != "" {
		run.Status = workflowstate.InvocationCompleted

		return true
	}

	if record.Task.SliceID == "" || run.Cursor == len(run.Members) || run.Members[run.Cursor].Admission.SliceID != record.Task.SliceID {
		run.Status = workflowstate.InvocationWaitingIntegration

		return true
	}

	return false
}

func (s *Store) checkWorkflowReusableTask(ctx context.Context, run workflowInvocationFile, record workflowstate.Record, lane workflowstate.Lane) error {
	return s.checkWorkflowReusableTaskPrepared(ctx, run, record, lane, (*Store).recoverTransition)
}

func (s *Store) checkWorkflowReusableTaskPrepared(ctx context.Context, run workflowInvocationFile, record workflowstate.Record, lane workflowstate.Lane, prepare func(*Store) error) error {
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

func workflowAttemptValidatesTask(attempt workflowstate.Attempt, record workflowstate.Record) bool {
	return attempt.Purpose == "" && attempt.TaskKey == record.Task.Key && attempt.ContractHash == record.Task.ContractHash && attempt.Outcome == "validated" && slices.Contains(record.Task.Evidence, attempt.ResultRef)
}

func (s *Store) workflowRetainedResultMatches(ctx context.Context, run workflowInvocationFile, attempt workflowstate.Attempt, record workflowstate.Record, reader *workflowProofReader) bool {
	owner, err := s.workflowOwnerStore(run)
	if err != nil {
		return false
	}

	path, err := owner.safePath(attempt.ResultRef)
	if err != nil {
		return false
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != attempt.ResultHash {
		return false
	}

	var result workflowstate.TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return false
	}

	return result.AttemptID == attempt.ID && result.Generation == attempt.Generation && result.InputHash == attempt.InputHash && owner.workflowRetainedInputMatches(attempt, record) && reader.validTaskProof(ctx, record, result.TaskValidation)
}

func (s *Store) workflowRetainedInputMatches(attempt workflowstate.Attempt, record workflowstate.Record) bool {
	path, err := s.safePath(attempt.InputRef)
	if err != nil {
		return false
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != attempt.InputHash {
		return false
	}

	var input TaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		return false
	}

	return input.AttemptID == attempt.ID && input.ContractHash == record.Task.ContractHash && input.Body == string(record.Body)
}
