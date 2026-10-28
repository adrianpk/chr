package workflow

import (
	"context"
	"os"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) retainWorkflowNativeDispatch(ctx context.Context, run workflowInvocationFile, submission Submission) (Submission, error) {
	entry, err := s.workflowNativeDispatchEntry(run, submission)
	if err != nil {
		return submission, err
	}

	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return submission, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.ID != submission.ID || last.Native == nil || last.Native.DispatchState != "not_started" {
		return last, ErrWorkflowObservationRequired
	}

	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		return submission, err
	}

	native := *last.Native
	native.DispatchIdentity, native.DispatchState, native.ApplicationState = &identity, "dispatching", "unknown"
	guard := IntegrationBudgetGuard{InvocationID: run.ID, InvocationRevision: run.Revision, BudgetRevision: run.IntegrationBudget.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration}
	request := NativeSubmissionUpdate{IntegrationBudgetGuard: guard, SubmissionID: last.ID, JournalHash: journal.sourceHash, Facts: native}

	return s.retainWorkflowNativeSubmission(ctx, run, request)
}

func (s *Store) workflowNativeDispatchEntry(run workflowInvocationFile, submission Submission) (IntegrationEntry, error) {
	if run.Integration == nil || run.IntegrationBudget == nil {
		return IntegrationEntry{}, workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return entry, err
	}

	if !workflowNativeSubmissionEntryMatches(submission, entry) || entry.State != "publishing" {
		return entry, workflowstate.ErrConflict
	}

	return entry, nil
}
