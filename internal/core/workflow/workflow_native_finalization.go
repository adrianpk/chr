package workflow

import (
	"context"
	"errors"
	"math"
	"os"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) finalizeWorkflowNativeUnit(ctx context.Context, run workflowInvocationFile) (bool, Record, error) {
	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if errors.Is(err, os.ErrNotExist) {
		return false, Record{}, nil
	}

	if err != nil {
		return true, Record{}, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.Native == nil {
		return false, Record{}, nil
	}

	if last.State != "confirmed" {
		return true, Record{}, ErrWorkflowObservationRequired
	}

	if err := s.checkWorkflowNativeSubmissionEvidence(ctx, last); err != nil {
		return true, Record{}, err
	}

	if err := s.projectWorkflowNativeSubmission(&run, journal); err != nil {
		return true, Record{}, err
	}

	receipt, err := s.finalizeWorkflowNativeReceipt(ctx, &run)

	return true, receipt, err
}

func (s *Store) finalizeWorkflowNativeReceipt(ctx context.Context, run *workflowInvocationFile) (Record, error) {
	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return Record{}, err
	}

	if entry.State == "integrated" {
		return s.readWorkflowSliceReceipt(ctx, *run)
	}

	if entry.State != "finalizing" {
		return Record{}, workflowstate.ErrConflict
	}

	receipt, err := s.readWorkflowFinalizationReceipt(ctx, *run)
	if err != nil {
		return receipt, err
	}

	deliveryPreparation, err := s.workflowDeliveryCandidate(ctx, *run)

	observation, _ := deliveryPreparation.Observation, deliveryPreparation.Candidate

	if err != nil {
		return receipt, err
	}

	if err := s.commitWorkflowNativeFanout(ctx, run, entry, observation, receipt); err != nil {
		return receipt, err
	}

	return s.readWorkflowSliceReceipt(ctx, *run)
}

func (s *Store) commitWorkflowNativeFanout(ctx context.Context, run *workflowInvocationFile, entry IntegrationEntry, observation workflowSliceObservation, receipt Record) error {
	if entry.Revision == math.MaxUint64 || entry.StateRevision == math.MaxUint64 {
		return workflowstate.IntegrationConflict("finalization revision exhausted")
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

func finishWorkflowNativeFanout(run *workflowInvocationFile, entry *IntegrationEntry) {
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
