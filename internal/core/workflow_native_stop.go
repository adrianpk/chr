package core

import (
	"context"
	"errors"
	"os"
	"time"
)

func (s *WorkflowStore) stopWorkflowNativeSubmission(ctx context.Context, run *workflowInvocationFile, intent string) (bool, error) {
	journal, err := s.readWorkflowSubmissionJournal(ctx, *run)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.Native == nil || last.State == "unpublished" {
		return false, nil
	}

	if last.State == "confirmed" && run.DeliveryRef != "" {
		entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
		if err != nil {
			return true, err
		}

		if entry.State == "integrated" {
			return true, s.finishWorkflowNativeStop(ctx, run, intent)
		}
	}

	// Repair the common intent's projection before recording stop. A crash may
	// have retained publication facts without their invocation/queue pointers.
	if err := s.projectWorkflowNativeSubmission(run, journal); err != nil {
		return true, err
	}

	return true, s.retainWorkflowNativeStop(run, intent)
}

func (s *WorkflowStore) retainWorkflowNativeStop(run *workflowInvocationFile, intent string) error {
	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return err
	}

	if run.StopIntent == intent && entry.StopIntent == intent && run.Status != workflowStopStatus(intent) {
		return nil
	}

	before := run.Integration.EntryRef.Hash
	run.StopIntent, entry.StopIntent = intent, intent
	entry.Revision++
	entry.StateRevision++
	entry.UpdatedAt = time.Now().UTC()
	projectWorkflowIntegrationPhase(run, entry)

	return s.commitWorkflowIntegrationEntry(run, entry, before)
}

func (s *WorkflowStore) finishWorkflowNativeStop(ctx context.Context, run *workflowInvocationFile, intent string) error {
	if _, err := s.readWorkflowSliceReceipt(ctx, *run); err != nil {
		return err
	}

	if workflowInvocationTerminal(run.Status) || run.StopIntent == intent && run.Status == workflowStopStatus(intent) {
		return nil
	}

	run.StopIntent, run.Status = intent, workflowStopStatus(intent)

	return s.commitWorkflowState(run)
}
