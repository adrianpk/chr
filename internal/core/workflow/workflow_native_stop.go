package workflow

import (
	"context"
	"errors"
	"os"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) stopWorkflowNativeSubmission(ctx context.Context, run *workflowInvocationFile, intent string, control *workflowControlCommit) (bool, error) {
	journal, err := s.readWorkflowSubmissionJournal(ctx, *run)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.Native == nil || last.State == workflowstate.SubmissionUnpublished {
		return false, nil
	}

	if last.State == workflowstate.SubmissionConfirmed && run.DeliveryRef != "" {
		entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
		if err != nil {
			return true, err
		}

		if entry.State == "integrated" {
			return true, s.finishWorkflowNativeStop(ctx, run, intent, control)
		}
	}

	projection, err := s.prepareWorkflowNativeProjection(run, journal)
	if err != nil {
		return true, err
	}

	return true, s.retainWorkflowNativeStop(ctx, run, intent, control, projection)
}

func (s *Store) retainWorkflowNativeStop(ctx context.Context, run *workflowInvocationFile, intent string, control *workflowControlCommit, projection workflowNativeProjection) error {
	entry := projection.entry
	if !projection.changed && run.StopIntent == intent && entry.StopIntent == intent && run.Status != workflowstate.StopStatus(intent) {
		return s.retainWorkflowControlNoop(ctx, *run, control)
	}

	before := run.Integration.EntryRef.Hash
	run.StopIntent, entry.StopIntent = intent, intent
	entry.Revision++
	entry.StateRevision++
	entry.UpdatedAt = time.Now().UTC()
	projectWorkflowIntegrationPhase(run, entry)

	files, err := control.files(ctx, run.Revision+1)
	if err != nil {
		return err
	}

	return s.commitWorkflowIntegrationEntryFiles(run, entry, before, append(projection.files, files...))
}

func (s *Store) finishWorkflowNativeStop(ctx context.Context, run *workflowInvocationFile, intent string, control *workflowControlCommit) error {
	if _, err := s.readWorkflowSliceReceipt(ctx, *run); err != nil {
		return err
	}

	if workflowstate.InvocationTerminal(run.Status) || run.StopIntent == intent && run.Status == workflowstate.StopStatus(intent) {
		return s.retainWorkflowControlNoop(ctx, *run, control)
	}

	run.StopIntent, run.Status = intent, workflowstate.StopStatus(intent)

	return s.commitWorkflowControlState(ctx, run, control)
}
