package core

import (
	"encoding/json"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// WorkflowAttemptCheckpoint preserves a stopped attempt without validation proof.
type WorkflowAttemptCheckpoint struct {
	Schema     string              `json:"schema"`
	AttemptID  string              `json:"attempt_id"`
	Generation uint64              `json:"generation"`
	Reason     string              `json:"reason"`
	FinishedAt time.Time           `json:"finished_at"`
	Processes  []WorkflowProcess   `json:"processes"`
	Result     *WorkflowTaskResult `json:"result"`
}

func (s *WorkflowStore) checkpointWorkflowAttempt(run workflowInvocationFile, record WorkflowRecord, reason string, result *WorkflowTaskResult, extra ...workflowFileChange) error {
	attempt := &run.Attempts[len(run.Attempts)-1]
	if err := workflowAttemptQuiescent(*attempt); err != nil {
		return err
	}

	if err := s.retainWorkflowProcessOutputs(attempt); err != nil {
		return err
	}

	now := time.Now().UTC()
	checkpoint := WorkflowAttemptCheckpoint{Schema: "chr-attempt-checkpoint/v1", AttemptID: attempt.ID, Generation: attempt.Generation, Reason: reason, FinishedAt: now, Processes: attempt.Processes, Result: result}

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return fmt.Errorf("workflow checkpoint exceeds file bounds: %w", errorsOrConflict(err))
	}

	data = append(data, '\n')
	ref := workflowAttemptFile(run.ID, attempt.ID, "checkpoint.json")
	attempt.CheckpointRef, attempt.CheckpointHash = ref, contenthash.Sum(data)
	attempt.Outcome, attempt.FinishedAt = reason, now
	record.Task.AttemptID = ""

	switch attempt.Purpose {
	case workflowIntegrationValidationPurpose:
		run.ResumeStatus = "waiting_integration"
		run.Preparation.State = "stale"
	case workflowSliceReviewPurpose:
		run.ResumeStatus = "waiting_review"
		run.Review.State = "stale"
	case workflowSliceValidationPurpose:
		run.ResumeStatus = "waiting_integration"
	default:
		record.Task.Phase, record.Task.Status = "pending", TicketStatusReadyForDev
		run.ResumeStatus = "pending"
	}

	run.Status = workflowStopStatus(run.StopIntent)
	run.Outcome = reason

	return s.commitWorkflowClaim(run, record, append(extra, workflowFileChange{Path: ref, After: data})...)
}

func (s *WorkflowStore) retainWorkflowProcessOutputs(attempt *WorkflowAttempt) error {
	for i := range attempt.Processes {
		process := &attempt.Processes[i]
		if err := s.hashWorkflowProcessOutput(process); err != nil {
			return err
		}

		process.State = "stopped"
	}

	return nil
}
