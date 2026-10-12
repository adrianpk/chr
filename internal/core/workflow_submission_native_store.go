package core

import (
	"context"
	"time"
)

func (s *WorkflowStore) beginWorkflowNativeSubmission(ctx context.Context, run workflowInvocationFile, fence WorkflowIntegrationBudgetGuard, provider WorkflowPublicationGuardProvider) (WorkflowSubmission, bool, error) {
	if err := s.checkWorkflowIntegrationRetainedRecords(ctx, run); err != nil {
		return WorkflowSubmission{}, false, err
	}

	journal, err := s.loadWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return WorkflowSubmission{}, false, err
	}

	if len(journal.Attempts) > 0 && journal.Attempts[len(journal.Attempts)-1].State != "unpublished" {
		return s.retryWorkflowNativeSubmission(&run, journal)
	}

	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, fence)
	if err != nil {
		return WorkflowSubmission{}, false, err
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, fence); err != nil {
		return WorkflowSubmission{}, false, err
	}

	if !workflowNativeSubmissionReady(run, entry, journal) {
		return WorkflowSubmission{}, false, ErrWorkflowConflict
	}

	if err := s.checkWorkflowIntegrationAuthorization(run, entry); err != nil {
		return WorkflowSubmission{}, false, err
	}

	return s.authorizeWorkflowNativeIntent(ctx, run, entry, journal, fence, provider)
}

func (s *WorkflowStore) retryWorkflowNativeSubmission(run *workflowInvocationFile, journal workflowSubmissionJournal) (WorkflowSubmission, bool, error) {
	last := journal.Attempts[len(journal.Attempts)-1]
	if last.Native == nil {
		return last, false, ErrWorkflowObservationRequired
	}

	return last, false, s.projectWorkflowNativeSubmission(run, journal)
}

func workflowNativeSubmissionReady(run workflowInvocationFile, entry WorkflowIntegrationEntry, journal workflowSubmissionJournal) bool {
	return entry.State == "ready_to_publish" && entry.StopIntent == "" && run.StopIntent == "" && len(journal.Attempts) < 10
}

func (s *WorkflowStore) authorizeWorkflowNativeIntent(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, journal workflowSubmissionJournal, fence WorkflowIntegrationBudgetGuard, provider WorkflowPublicationGuardProvider) (WorkflowSubmission, bool, error) {
	guard, err := s.checkWorkflowPublicationGuard(ctx, run, provider)
	if err != nil {
		return WorkflowSubmission{}, false, err
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil || candidate.Native == nil {
		return WorkflowSubmission{}, false, errorsOrConflict(err)
	}

	if journal.Key == "" {
		journal, err = newWorkflowSubmissionJournal(run, guard)
		if err != nil {
			return WorkflowSubmission{}, false, err
		}
	}

	if err := s.guardWorkflowSubmissionTarget(ctx, journal.Key, journal.TargetKey); err != nil {
		return WorkflowSubmission{}, false, err
	}

	return s.persistWorkflowNativeIntent(ctx, run, entry, journal, guard, candidate, fence)
}

func (s *WorkflowStore) persistWorkflowNativeIntent(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, journal workflowSubmissionJournal, guard WorkflowPublicationGuard, candidate WorkflowIntegrationCandidate, fence WorkflowIntegrationBudgetGuard) (WorkflowSubmission, bool, error) {
	number := len(journal.Attempts) + 1
	id := workflowSubmissionID(journal.Key, number, guard.PreparationHash)

	reservation := WorkflowIntegrationReservation{ID: id, Kind: "submission", EntryID: entry.ID, Unit: entry.Unit, ClaimGeneration: entry.ClaimGeneration, AttemptID: id, State: "reserved"}
	if _, err := s.reserveWorkflowIntegration(ctx, run, WorkflowIntegrationReserve{WorkflowIntegrationBudgetGuard: fence, Reservation: reservation}); err != nil {
		return WorkflowSubmission{}, false, err
	}

	run, err := s.readWorkflowInvocation(run.ID)
	if err != nil {
		return WorkflowSubmission{}, false, err
	}

	native := workflowNativeSubmissionIntent(candidate, run, entry)
	submission := WorkflowSubmission{ID: id, InvocationID: run.ID, Unit: entry.Unit, Attempt: number, CandidateID: run.PullRequest.CandidateID, CandidateHash: run.CandidateHash, Guard: guard, PreparedLane: candidate.Lane, State: "awaiting_observation", SubmittedAt: native.IntentTime, UpdatedAt: native.IntentTime, Native: &native}
	journal.Schema = "chr-integration-journal/v2"

	journal.Attempts = append(journal.Attempts, submission)
	if err := s.checkWorkflowNativeSubmissionEvidence(ctx, submission); err != nil {
		return submission, false, err
	}

	if err := s.writeWorkflowSubmissionJournal(ctx, &journal); err != nil {
		return submission, false, err
	}

	return submission, true, s.projectWorkflowNativeSubmission(&run, journal)
}

func (s *WorkflowStore) projectWorkflowNativeSubmission(run *workflowInvocationFile, journal workflowSubmissionJournal) error {
	if run.Integration == nil {
		return ErrWorkflowConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowNativeSubmissionEntryMatches(last, entry) {
		return ErrWorkflowConflict
	}

	binding := workflowSubmissionProjection(journal)
	if workflowNativeSubmissionProjected(*run, entry, binding) {
		return nil
	}

	run.Submission, entry.SubmissionRef = &binding, &binding
	projectWorkflowNativeSubmissionPhase(&entry, last)

	entry.Revision++
	entry.StateRevision++
	entry.UpdatedAt = time.Now().UTC()
	retainWorkflowIntegrationStop(&entry, *run, WorkflowIntegrationTransition{})
	projectWorkflowIntegrationPhase(run, entry)

	return s.commitWorkflowIntegrationEntry(run, entry, run.Integration.EntryRef.Hash)
}

func workflowNativeSubmissionEntryMatches(submission WorkflowSubmission, entry WorkflowIntegrationEntry) bool {
	native := submission.Native

	return native != nil && native.EntryID == entry.ID && entry.CandidateRef != nil && native.CandidateRef == *entry.CandidateRef && native.ClaimGeneration == entry.ClaimGeneration
}

func workflowNativeSubmissionProjected(run workflowInvocationFile, entry WorkflowIntegrationEntry, binding WorkflowSubmissionBinding) bool {
	return entry.SubmissionRef != nil && *entry.SubmissionRef == binding && run.Submission != nil && *run.Submission == binding
}

func projectWorkflowNativeSubmissionPhase(entry *WorkflowIntegrationEntry, last WorkflowSubmission) {
	if entry.State == "blocked" {
		return
	}

	entry.Blocker = nil

	switch last.Native.ApplicationState {
	case "not_started":
		entry.State = "publishing"
	case "not_applied":
		entry.State = "blocked"
		entry.Blocker = &WorkflowIntegrationBlocker{Reason: "native_not_applied", RequiredAction: "reconcile changed input and dispatch quiescence", ResumeState: "queued"}
	default:
		entry.State = "observing"
	}
}

func (s *WorkflowStore) retainWorkflowNativeSubmission(ctx context.Context, run workflowInvocationFile, request WorkflowNativeSubmissionUpdate) (WorkflowSubmission, error) {
	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, request.WorkflowIntegrationBudgetGuard)
	if err != nil {
		return WorkflowSubmission{}, err
	}

	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return WorkflowSubmission{}, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowNativeSubmissionCallbackMatches(last, entry, request) {
		return last, ErrWorkflowConflict
	}

	if workflowNativeJSONEqual(*last.Native, request.Facts) {
		return last, s.projectWorkflowNativeSubmission(&run, journal)
	}

	if err := checkWorkflowNativeSubmissionCallback(run, entry, budget, last, journal.sourceHash, request); err != nil {
		return last, err
	}

	last.Native = &request.Facts

	last.UpdatedAt = time.Now().UTC()
	if err := s.applyWorkflowNativeSubmissionEvidence(ctx, &last); err != nil {
		return last, err
	}

	journal.Attempts[len(journal.Attempts)-1] = last
	if err := s.writeWorkflowSubmissionJournal(ctx, &journal); err != nil {
		return last, err
	}

	return last, s.projectWorkflowNativeSubmission(&run, journal)
}

func workflowNativeSubmissionCallbackMatches(last WorkflowSubmission, entry WorkflowIntegrationEntry, request WorkflowNativeSubmissionUpdate) bool {
	return last.Native != nil && last.ID == request.SubmissionID && last.Native.EntryID == entry.ID
}

func workflowNativeDispatchStopped(native WorkflowNativeSubmission, run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return native.DispatchState == "not_started" && workflowIntegrationReservationStopped(run, entry)
}

func checkWorkflowNativeSubmissionCallback(run workflowInvocationFile, entry WorkflowIntegrationEntry, budget WorkflowIntegrationBudget, last WorkflowSubmission, hash string, request WorkflowNativeSubmissionUpdate) error {
	if hash != request.JournalHash {
		return ErrWorkflowConflict
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.WorkflowIntegrationBudgetGuard); err != nil {
		return err
	}

	if workflowNativeDispatchStopped(*last.Native, run, entry) {
		return ErrWorkflowStop
	}

	return guardWorkflowNativeSubmissionUpdate(*last.Native, request.Facts)
}
