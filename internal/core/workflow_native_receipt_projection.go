package core

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) projectWorkflowNativeReceiptObservation(run *workflowInvocationFile, entry *WorkflowIntegrationEntry, submission WorkflowSubmission) error {
	refs := submission.Native.ObservationRefs
	if len(refs) == 0 {
		return workflowstate.ErrConflict
	}

	var observation WorkflowNativePublicationObservation
	if err := s.readWorkflowNativeEvidence(run.ID, refs[len(refs)-1], &observation); err != nil {
		return err
	}

	if entry.State == "blocked" && !workflowNativeCanonicalBlocker(entry.Blocker) {
		return nil
	}

	if observation.State != "merged" {
		entry.State = "blocked"
		entry.Blocker = &WorkflowIntegrationBlocker{Reason: observation.Condition, RequiredAction: "reconcile the original canonical PR and retained receipt history", ResumeState: "finalizing"}

		return nil
	}

	entry.State, entry.Blocker = "finalizing", nil
	if run.ConfirmedCursor == run.Cursor {
		entry.State = "integrated"

		workflowRetainDeliveryStatus(run, run.Outcome, run.Cursor == len(run.Members) && (run.Selection.Selector.Extent == "slice" || run.Selection.Selector.Extent == "ticket"))
	}

	return nil
}

func retainWorkflowNativeReceiptStop(run *workflowInvocationFile, entry WorkflowIntegrationEntry) {
	if entry.ReceiptRef == nil || run.StopIntent == "" || run.ConfirmedCursor != run.Cursor {
		return
	}

	run.Status = workflowstate.StopStatus(run.StopIntent)
}

func (s *WorkflowStore) checkWorkflowNativeReceiptContinuation(submission WorkflowSubmission) error {
	if submission.Native == nil || submission.Native.ReceiptRef == nil {
		return nil
	}

	refs := submission.Native.ObservationRefs
	if len(refs) == 0 {
		return ErrWorkflowObservationRequired
	}

	var observation WorkflowNativePublicationObservation
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, refs[len(refs)-1], &observation); err != nil {
		return err
	}

	if observation.State != "merged" {
		return ErrWorkflowObservationRequired
	}

	return nil
}

func (s *WorkflowStore) recoverWorkflowNativeReceiptProjection(ctx context.Context, run *workflowInvocationFile, guard WorkflowIntegrationBudgetGuard) error {
	if run.Integration == nil || run.Submission == nil || run.Submission.State != "confirmed" {
		return nil
	}

	journal, err := s.readWorkflowSubmissionJournal(ctx, *run)
	if err != nil {
		return err
	}

	if *run.Submission == workflowSubmissionProjection(journal) {
		return nil
	}

	return s.repairWorkflowNativeReceiptOwner(ctx, run, guard, journal)
}

func (s *WorkflowStore) repairWorkflowNativeReceiptOwner(ctx context.Context, run *workflowInvocationFile, guard WorkflowIntegrationBudgetGuard, journal workflowSubmissionJournal) error {
	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return err
	}

	budget, err := s.readWorkflowIntegrationBudget(*run)
	if err != nil {
		return err
	}

	if err := guardWorkflowIntegrationBudget(*run, entry, budget, guard); err != nil {
		return err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.State != "confirmed" || !workflowNativeSubmissionEntryMatches(last, entry) {
		return workflowstate.ErrConflict
	}

	if err := s.checkWorkflowNativeSubmissionEvidence(ctx, last); err != nil {
		return err
	}

	if err := s.projectWorkflowNativeSubmission(run, journal); err != nil {
		return err
	}

	// The caller must obtain the repaired owner revision before starting new IO.
	return workflowstate.ErrConflict
}
