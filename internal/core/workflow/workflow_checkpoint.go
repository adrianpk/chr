package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// AttemptCheckpoint preserves a stopped attempt without validation proof.
type AttemptCheckpoint struct {
	Schema     string      `json:"schema"`
	AttemptID  string      `json:"attempt_id"`
	Generation uint64      `json:"generation"`
	Reason     string      `json:"reason"`
	FinishedAt time.Time   `json:"finished_at"`
	Processes  []Process   `json:"processes"`
	Result     *TaskResult `json:"result"`
}

func (s *Store) checkpointWorkflowAttempt(run workflowInvocationFile, record Record, reason string, result *TaskResult, extra ...workflowstate.FileChange) error {
	attempt := &run.Attempts[len(run.Attempts)-1]
	if err := workflowAttemptQuiescent(*attempt); err != nil {
		return err
	}

	if err := s.retainWorkflowProcessOutputs(attempt); err != nil {
		return err
	}

	now := time.Now().UTC()
	checkpoint := AttemptCheckpoint{Schema: "chr-attempt-checkpoint/v1", AttemptID: attempt.ID, Generation: attempt.Generation, Reason: reason, FinishedAt: now, Processes: attempt.Processes, Result: result}

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
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
		record.Task.Phase, record.Task.Status = "pending", string(ticketcore.StatusReadyForDev)
		run.ResumeStatus = "pending"
	}

	run.Status = workflowstate.StopStatus(run.StopIntent)
	run.Outcome = reason

	return s.commitWorkflowClaim(run, record, append(extra, workflowstate.FileChange{Path: ref, After: data})...)
}

func (s *Store) retainWorkflowProcessOutputs(attempt *Attempt) error {
	for i := range attempt.Processes {
		process := &attempt.Processes[i]
		if err := s.hashWorkflowProcessOutput(process); err != nil {
			return err
		}

		process.State = "stopped"
	}

	return nil
}
