package core

import (
	"context"
	"errors"
	"math"
	"os"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) finalizeWorkflowNativeUnit(ctx context.Context, run workflowInvocationFile) (bool, WorkflowRecord, error) {
	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if errors.Is(err, os.ErrNotExist) {
		return false, WorkflowRecord{}, nil
	}

	if err != nil {
		return true, WorkflowRecord{}, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.Native == nil {
		return false, WorkflowRecord{}, nil
	}

	if last.State != "confirmed" {
		return true, WorkflowRecord{}, ErrWorkflowObservationRequired
	}

	if err := s.checkWorkflowNativeSubmissionEvidence(ctx, last); err != nil {
		return true, WorkflowRecord{}, err
	}

	if err := s.projectWorkflowNativeSubmission(&run, journal); err != nil {
		return true, WorkflowRecord{}, err
	}

	receipt, err := s.finalizeWorkflowNativeReceipt(ctx, &run)

	return true, receipt, err
}

func (s *WorkflowStore) finalizeWorkflowNativeReceipt(ctx context.Context, run *workflowInvocationFile) (WorkflowRecord, error) {
	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if entry.State == "integrated" {
		return s.readWorkflowSliceReceipt(ctx, *run)
	}

	if entry.State != "finalizing" {
		return WorkflowRecord{}, workflowstate.ErrConflict
	}

	receipt, err := s.readWorkflowFinalizationReceipt(ctx, *run)
	if err != nil {
		return receipt, err
	}

	observation, _, err := s.workflowDeliveryCandidate(ctx, *run)
	if err != nil {
		return receipt, err
	}

	if err := s.commitWorkflowNativeFanout(ctx, run, entry, observation, receipt); err != nil {
		return receipt, err
	}

	return s.readWorkflowSliceReceipt(ctx, *run)
}

func (s *WorkflowStore) commitWorkflowNativeFanout(ctx context.Context, run *workflowInvocationFile, entry WorkflowIntegrationEntry, observation workflowSliceObservation, receipt WorkflowRecord) error {
	if entry.Revision == math.MaxUint64 || entry.StateRevision == math.MaxUint64 {
		return workflowIntegrationConflict("finalization revision exhausted")
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	files, err := s.workflowFinalizationTaskImages(run, &snapshot, observation.tasks, receipt)
	if err != nil {
		return err
	}

	set, err := s.workflowFinalizationSetImages(run, &snapshot, observation.set)
	if err != nil {
		return err
	}

	if err := snapshot.Validate(); err != nil {
		return err
	}

	finishWorkflowNativeFanout(run, &entry)

	// Receipt bytes already precede this transaction. Member and set images precede
	// the integrated queue image and invocation frontier in the existing journal.
	return s.commitWorkflowIntegrationEntryFiles(run, entry, run.Integration.EntryRef.Hash, append(files, set...))
}

func finishWorkflowNativeFanout(run *workflowInvocationFile, entry *WorkflowIntegrationEntry) {
	run.ConfirmedCursor = run.Cursor

	outcome := "slice_integrated"
	if entry.Unit.Kind == "ticket" {
		outcome = "ticket_integrated"
	}

	workflowRetainDeliveryStatus(run, outcome, run.Cursor == len(run.Members) && (run.Selection.Selector.Extent == "slice" || run.Selection.Selector.Extent == "ticket"))

	entry.State, entry.Blocker, entry.PreparationReserved = "integrated", nil, false
	entry.Revision++
	entry.StateRevision++
	entry.UpdatedAt = time.Now().UTC()

	if run.StopIntent != "" {
		run.Status = workflowstate.StopStatus(run.StopIntent)
	}
}
