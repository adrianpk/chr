package core

import (
	"context"
	"slices"
)

func (s *WorkflowStore) checkWorkflowIntegrationRetainedRecords(ctx context.Context, run workflowInvocationFile) error {
	if err := s.checkWorkflowIntegrationFiles(run); err != nil {
		return integrationRecordError(workflowInvocationPath(run.ID), err)
	}

	if len(run.IntegrationEntries) == 0 && run.IntegrationBudget == nil {
		return nil
	}

	entries := []WorkflowIntegrationEntry{}

	for _, binding := range run.IntegrationEntries {
		entry, err := s.readWorkflowIntegrationEntry(run, binding)
		if err != nil {
			return integrationRecordError(binding.EntryRef.Path, err)
		}

		if err := s.checkWorkflowIntegrationEntryRecords(ctx, run, entry); err != nil {
			return integrationRecordError(binding.EntryRef.Path, err)
		}

		entries = append(entries, entry)
	}

	budget, err := s.readWorkflowIntegrationBudget(run)
	if err == nil {
		err = s.checkWorkflowIntegrationBudgetRecords(ctx, run, budget, entries)
	}

	return integrationRecordError(workflowIntegrationBudgetPath(run.ID), err)
}

func (s *WorkflowStore) checkWorkflowIntegrationBudgetRecords(ctx context.Context, run workflowInvocationFile, budget WorkflowIntegrationBudget, entries []WorkflowIntegrationEntry) error {
	usage, err := s.workflowIntegrationUsage(ctx, run)
	if err != nil {
		return err
	}

	if err := checkWorkflowIntegrationUsage(budget.Usage, usage); err != nil {
		return err
	}

	counters, err := s.workflowIntegrationLegacySubmissions(ctx, run)
	if err != nil {
		return err
	}

	if err := checkWorkflowIntegrationLegacySubmissions(budget, counters); err != nil {
		return err
	}

	return checkWorkflowIntegrationBudgetOwnership(run, budget, entries)
}

func checkWorkflowIntegrationBudgetOwnership(run workflowInvocationFile, budget WorkflowIntegrationBudget, entries []WorkflowIntegrationEntry) error {
	for _, reservation := range budget.Reservations {
		if reservation.AttemptInput != nil {
			if !workflowAgentReservationOwned(run, reservation) {
				return ErrWorkflowConflict
			}

			continue
		}

		index := slices.IndexFunc(entries, func(e WorkflowIntegrationEntry) bool { return e.ID == reservation.EntryID })
		if index < 0 || reservation.ClaimGeneration > entries[index].ClaimGeneration || entries[index].Unit != reservation.Unit {
			return workflowIntegrationConflict("reservation has missing ownership or future generation")
		}
	}

	return checkWorkflowIntegrationCounterOwners(run, budget, entries)
}

func checkWorkflowIntegrationCounterOwners(run workflowInvocationFile, budget WorkflowIntegrationBudget, entries []WorkflowIntegrationEntry) error {
	for _, entry := range entries {
		if workflowIntegrationEntryInvocation(run, entry) && !workflowIntegrationTerminal(entry.State) && !workflowIntegrationCurrentEntry(run, entry) {
			return workflowIntegrationConflict("active unit has no current invocation pointer")
		}

		if workflowIntegrationEntryInvocation(run, entry) && !slices.ContainsFunc(budget.UnitCounters, func(c WorkflowIntegrationCounter) bool { return c.Unit == entry.Unit }) {
			return workflowIntegrationConflict("entry has no retained unit counter")
		}
	}

	return checkWorkflowIntegrationCounterHistory(run, budget, entries)
}

func checkWorkflowIntegrationCounterHistory(run workflowInvocationFile, budget WorkflowIntegrationBudget, entries []WorkflowIntegrationEntry) error {
	for _, counter := range budget.UnitCounters {
		if !workflowIntegrationCounterHasOwner(run, counter, entries) {
			return workflowIntegrationConflict("counter has no retained unit owner")
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowIntegrationEntryRecords(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry) error {
	if workflowIntegrationTargetChanged(run, entry) {
		return workflowIntegrationConflict("target differs from the owning original PR")
	}

	if err := s.checkWorkflowIntegrationCandidate(entry); err != nil {
		return err
	}

	if workflowIntegrationCurrentEntry(run, entry) && entry.State == "ready_to_publish" {
		if err := s.checkWorkflowIntegrationAuthorization(run, entry); err != nil {
			return err
		}
	}

	if err := s.checkWorkflowIntegrationEntryReceipt(run, entry); err != nil {
		return err
	}

	if entry.State == "integrated" && workflowIntegrationCurrentEntry(run, entry) {
		_, err := s.readWorkflowSliceReceipt(ctx, run)

		return err
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowIntegrationJournals(ctx context.Context, snapshot workflowIntegrationSnapshot) error {
	for _, journal := range snapshot.Journals {
		for _, submission := range journal.Attempts {
			if submission.Native == nil {
				continue
			}

			if err := s.checkWorkflowIntegrationJournalOwner(ctx, snapshot, journal, submission); err != nil {
				return integrationRecordError(journal.Key+".json", err)
			}
		}
	}

	if err := checkWorkflowIntegrationNativeProjections(snapshot); err != nil {
		return err
	}

	return checkWorkflowIntegrationProjections(snapshot)
}

func (s *WorkflowStore) checkWorkflowIntegrationJournalOwner(ctx context.Context, snapshot workflowIntegrationSnapshot, journal workflowSubmissionJournal, submission WorkflowSubmission) error {
	index := slices.IndexFunc(snapshot.Runs, func(r workflowInvocationFile) bool {
		return r.ID == journal.InvocationID && r.Namespace == journal.Namespace && r.OwnerRoot == journal.OwnerRoot
	})
	if index < 0 {
		return workflowIntegrationConflict("native publication has no authoritative invocation owner")
	}

	run := snapshot.Runs[index]

	owner, err := s.workflowOwnerStore(run)
	if err != nil {
		return err
	}

	view := slices.IndexFunc(snapshot.Views, func(v WorkflowIntegrationView) bool { return v.Entry.ID == submission.Native.EntryID })
	if view < 0 || !workflowIntegrationJournalMatches(snapshot.Views[view], submission) {
		return workflowIntegrationConflict("publication differs from entry or consumed reservation")
	}

	return owner.checkWorkflowNativeSubmissionEvidence(ctx, submission)
}

func workflowIntegrationJournalMatches(view WorkflowIntegrationView, submission WorkflowSubmission) bool {
	native, entry := submission.Native, view.Entry
	if native.Unit != entry.Unit || native.SourceRepository != entry.Repository || native.TargetRef != entry.TargetRef || native.ClaimGeneration > entry.ClaimGeneration {
		return false
	}

	if submission.State != "unpublished" && !workflowNativeSubmissionEntryMatches(submission, entry) {
		return false
	}

	return slices.ContainsFunc(view.Budget.Reservations, func(r WorkflowIntegrationReservation) bool {
		return r.ID == submission.ID && r.Kind == "submission" && r.EntryID == entry.ID && r.ClaimGeneration == native.ClaimGeneration
	})
}

func checkWorkflowIntegrationProjections(snapshot workflowIntegrationSnapshot) error {
	for _, view := range snapshot.Views {
		binding := view.Entry.SubmissionRef
		if binding == nil {
			continue
		}

		index := slices.IndexFunc(snapshot.Journals, func(j workflowSubmissionJournal) bool { return j.Key == binding.JournalKey })
		if index < 0 {
			return ErrWorkflowConflict
		}

		journal := snapshot.Journals[index]

		projection := workflowSubmissionProjection(journal)
		if projection != *binding {
			return workflowIntegrationConflict("publication projection requires reconciliation")
		}
	}

	return nil
}

func checkWorkflowIntegrationNativeProjections(snapshot workflowIntegrationSnapshot) error {
	for _, journal := range snapshot.Journals {
		last := journal.Attempts[len(journal.Attempts)-1]
		if last.Native == nil || last.State == "unpublished" {
			continue
		}

		index := slices.IndexFunc(snapshot.Views, func(v WorkflowIntegrationView) bool { return v.Entry.ID == last.Native.EntryID })
		if index < 0 {
			return ErrWorkflowConflict
		}

		binding := snapshot.Views[index].Entry.SubmissionRef
		if binding == nil || *binding != workflowSubmissionProjection(journal) {
			return workflowIntegrationConflict("common intent requires owner projection reconciliation")
		}

		if !workflowIntegrationInvocationPublicationMatches(snapshot, snapshot.Views[index].Entry, *binding) {
			return workflowIntegrationConflict("invocation publication projection requires reconciliation")
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowIntegrationEntryReceipt(run workflowInvocationFile, entry WorkflowIntegrationEntry) error {
	if entry.ReceiptRef == nil {
		return nil
	}

	var record WorkflowRecord
	if err := s.readWorkflowNativeEvidence(run.ID, *entry.ReceiptRef, &record); err != nil {
		return err
	}

	if !workflowIntegrationEntryReceiptMatches(run, entry, record) {
		return ErrWorkflowConflict
	}

	if err := s.checkWorkflowNativeReceipt(record); err != nil {
		return err
	}

	return nil
}

func workflowIntegrationEntryReceiptMatches(run workflowInvocationFile, entry WorkflowIntegrationEntry, record WorkflowRecord) bool {
	if record.ProjectID != run.ProjectID || record.Namespace != run.Namespace || record.Receipt == nil || record.Receipt.Native == nil || entry.CandidateRef == nil {
		return false
	}

	native := record.Receipt.Native

	return workflowIntegrationReceiptEntryScopeMatches(native, run, entry)
}

func workflowIntegrationTargetChanged(run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return run.PullRequest != nil && run.PullRequest.Unit == entry.Unit && entry.TargetRef != "refs/heads/"+run.PullRequest.Request.BaseBranch
}

func workflowIntegrationCurrentEntry(run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return run.Integration != nil && run.Integration.EntryID == entry.ID
}

func workflowIntegrationEntryInvocation(run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return entry.InvocationID == run.ID && entry.Namespace == run.Namespace && entry.ProjectID == run.ProjectID
}

func workflowIntegrationReceiptEntryScopeMatches(native *WorkflowNativeReceipt, run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return native.EntryID == entry.ID && native.InvocationID == run.ID && native.Unit == entry.Unit && native.SourceRepository == entry.Repository && native.TargetRef == entry.TargetRef && native.CandidateRef == *entry.CandidateRef
}

func workflowIntegrationInvocationPublicationMatches(snapshot workflowIntegrationSnapshot, entry WorkflowIntegrationEntry, binding WorkflowSubmissionBinding) bool {
	index := slices.IndexFunc(snapshot.Runs, func(run workflowInvocationFile) bool { return workflowIntegrationEntryInvocation(run, entry) })
	if index < 0 {
		return false
	}

	run := snapshot.Runs[index]
	if !workflowIntegrationCurrentEntry(run, entry) {
		return true
	}

	return run.Submission != nil && *run.Submission == binding
}

func workflowIntegrationCounterHasOwner(run workflowInvocationFile, counter WorkflowIntegrationCounter, entries []WorkflowIntegrationEntry) bool {
	return counter.ExistingSubmissions > 0 || workflowAgentCounterOwner(run, counter) || slices.ContainsFunc(entries, func(entry WorkflowIntegrationEntry) bool {
		return entry.Unit == counter.Unit && workflowIntegrationEntryInvocation(run, entry)
	})
}

func workflowAgentCounterOwner(run workflowInvocationFile, counter WorkflowIntegrationCounter) bool {
	return counter.Merger == 0 && counter.Drift == 0 && counter.Submission == 0 && workflowIntegrationRetainedUnit(run, counter.Unit)
}

func workflowAgentReservationOwned(run workflowInvocationFile, r WorkflowIntegrationReservation) bool {
	return r.Kind == "agent_call" && workflowAgentUsageRecorded(run, r) && workflowIntegrationRetainedUnit(run, r.Unit) && slices.ContainsFunc(run.Attempts, func(a WorkflowAttempt) bool {
		return a.ID == r.AttemptID && a.Generation == r.ClaimGeneration && a.InputRef == r.AttemptInput.Path && a.InputHash == r.AttemptInput.Hash && a.Lane != nil && a.Lane.Unit == r.Unit
	})
}
