package workflow

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) beginWorkflowNativeSubmission(ctx context.Context, run workflowInvocationFile, fence IntegrationBudgetGuard, provider PublicationGuardProvider) (workflowstate.Submission, bool, error) {
	if err := s.checkWorkflowIntegrationRetainedRecords(ctx, run); err != nil {
		return workflowstate.Submission{}, false, err
	}

	journal, err := s.loadWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return workflowstate.Submission{}, false, err
	}

	if len(journal.Attempts) > 0 && journal.Attempts[len(journal.Attempts)-1].State != "unpublished" {
		return s.retryWorkflowNativeSubmission(&run, journal)
	}

	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, fence)
	if err != nil {
		return workflowstate.Submission{}, false, err
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, fence); err != nil {
		return workflowstate.Submission{}, false, err
	}

	if !workflowNativeSubmissionReady(run, entry, journal) {
		return workflowstate.Submission{}, false, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowIntegrationAuthorization(run, entry); err != nil {
		return workflowstate.Submission{}, false, err
	}

	return s.authorizeWorkflowNativeIntent(ctx, run, entry, journal, fence, provider)
}

func (s *Store) retryWorkflowNativeSubmission(run *workflowInvocationFile, journal workflowSubmissionJournal) (workflowstate.Submission, bool, error) {
	last := journal.Attempts[len(journal.Attempts)-1]
	if last.Native == nil {
		return last, false, ErrWorkflowObservationRequired
	}

	return last, false, s.projectWorkflowNativeSubmission(run, journal)
}

func workflowNativeSubmissionReady(run workflowInvocationFile, entry IntegrationEntry, journal workflowSubmissionJournal) bool {
	return entry.State == "ready_to_publish" && entry.StopIntent == "" && run.StopIntent == "" && len(journal.Attempts) < 10
}

func (s *Store) authorizeWorkflowNativeIntent(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, journal workflowSubmissionJournal, fence IntegrationBudgetGuard, provider PublicationGuardProvider) (workflowstate.Submission, bool, error) {
	guard, err := s.checkWorkflowPublicationGuard(ctx, run, provider)
	if err != nil {
		return workflowstate.Submission{}, false, err
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil || candidate.Native == nil {
		return workflowstate.Submission{}, false, errorsOrConflict(err)
	}

	if journal.Key == "" {
		journal, err = newWorkflowSubmissionJournal(run, guard)
		if err != nil {
			return workflowstate.Submission{}, false, err
		}
	}

	if err := s.guardWorkflowSubmissionTarget(ctx, journal.Key, journal.TargetKey); err != nil {
		return workflowstate.Submission{}, false, err
	}

	return s.persistWorkflowNativeIntent(ctx, run, entry, journal, guard, candidate, fence)
}

func (s *Store) persistWorkflowNativeIntent(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, journal workflowSubmissionJournal, guard workflowstate.PublicationGuard, candidate workflowstate.IntegrationCandidate, fence IntegrationBudgetGuard) (workflowstate.Submission, bool, error) {
	number := len(journal.Attempts) + 1
	id := workflowSubmissionID(journal.Key, number, guard.PreparationHash)

	reservation := IntegrationReservation{ID: id, Kind: "submission", EntryID: entry.ID, Unit: entry.Unit, ClaimGeneration: entry.ClaimGeneration, AttemptID: id, State: "reserved"}
	if _, err := s.reserveWorkflowIntegration(ctx, run, IntegrationReserve{IntegrationBudgetGuard: fence, Reservation: reservation}); err != nil {
		return workflowstate.Submission{}, false, err
	}

	run, err := s.readWorkflowInvocation(run.ID)
	if err != nil {
		return workflowstate.Submission{}, false, err
	}

	native := workflowNativeSubmissionIntent(candidate, run, entry)
	submission := workflowstate.Submission{ID: id, InvocationID: run.ID, Unit: entry.Unit, Attempt: number, CandidateID: run.PullRequest.CandidateID, CandidateHash: run.CandidateHash, Guard: guard, PreparedLane: candidate.Lane, State: "awaiting_observation", SubmittedAt: native.IntentTime, UpdatedAt: native.IntentTime, Native: &native}
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

func (s *Store) projectWorkflowNativeSubmission(run *workflowInvocationFile, journal workflowSubmissionJournal) error {
	projection, err := s.prepareWorkflowNativeProjection(run, journal)
	if err != nil || !projection.changed {
		return err
	}

	return s.commitWorkflowIntegrationEntryFiles(run, projection.entry, run.Integration.EntryRef.Hash, projection.files)
}

// Preparing projection files separately lets stop intent and its receipt share
// the same commit as publication-pointer repair.
type workflowNativeProjection struct {
	entry   IntegrationEntry
	files   []workflowstate.FileChange
	changed bool
}

func (s *Store) prepareWorkflowNativeProjection(run *workflowInvocationFile, journal workflowSubmissionJournal) (workflowNativeProjection, error) {
	if run.Integration == nil {
		return workflowNativeProjection{}, workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return workflowNativeProjection{}, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowNativeSubmissionEntryMatches(last, entry) {
		return workflowNativeProjection{}, workflowstate.ErrConflict
	}

	binding := workflowSubmissionProjection(journal)
	if workflowNativeSubmissionProjected(*run, entry, binding) {
		return s.workflowNativeReceiptProjection(run, entry, last)
	}

	run.Submission, entry.SubmissionRef = &binding, &binding
	if err := s.projectWorkflowNativeSubmissionPhase(&entry, last); err != nil {
		return workflowNativeProjection{}, err
	}

	files, err := s.projectWorkflowNativeReceipt(run, &entry, last)
	if err != nil {
		return workflowNativeProjection{}, err
	}

	entry.Revision++
	entry.StateRevision++
	entry.UpdatedAt = time.Now().UTC()
	retainWorkflowIntegrationStop(&entry, *run, IntegrationTransition{})
	projectWorkflowIntegrationPhase(run, entry)
	retainWorkflowNativeReceiptStop(run, entry)

	return workflowNativeProjection{entry: entry, files: files, changed: true}, nil
}

func workflowNativeSubmissionEntryMatches(submission workflowstate.Submission, entry IntegrationEntry) bool {
	native := submission.Native

	return native != nil && native.EntryID == entry.ID && entry.CandidateRef != nil && native.CandidateRef == *entry.CandidateRef && native.ClaimGeneration == entry.ClaimGeneration
}

func workflowNativeSubmissionProjected(run workflowInvocationFile, entry IntegrationEntry, binding SubmissionBinding) bool {
	return entry.SubmissionRef != nil && *entry.SubmissionRef == binding && run.Submission != nil && *run.Submission == binding
}

func (s *Store) projectWorkflowNativeSubmissionPhase(entry *IntegrationEntry, last workflowstate.Submission) error {
	if last.State == "confirmed" {
		return nil
	}

	if projectWorkflowNativeStop(entry, last) {
		return nil
	}

	if len(last.Native.ObservationRefs) > 0 {
		if err := s.projectWorkflowNativeCanonicalObservation(entry, last); err != nil {
			return err
		}
	}

	if entry.State == "blocked" {
		return nil
	}

	entry.Blocker = nil

	switch last.Native.ApplicationState {
	case "not_started":
		entry.State = "publishing"
	case "not_applied":
		entry.State = "blocked"

		entry.Blocker = &IntegrationBlocker{Reason: "native_not_applied", RequiredAction: "reconcile changed input and dispatch quiescence", ResumeState: "queued"}
		if last.State != "unpublished" {
			entry.Blocker = &IntegrationBlocker{Reason: "native_publication_rejected", RequiredAction: "resolve publication permissions or capability before new authorization", ResumeState: "observing"}
		}
	default:
		entry.State = "observing"
	}

	return nil
}

func (s *Store) retainWorkflowNativeSubmission(ctx context.Context, run workflowInvocationFile, request NativeSubmissionUpdate) (workflowstate.Submission, error) {
	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, request.IntegrationBudgetGuard)
	if err != nil {
		return workflowstate.Submission{}, err
	}

	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return workflowstate.Submission{}, err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowNativeSubmissionCallbackMatches(last, entry, request) {
		return last, workflowstate.ErrConflict
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

func workflowNativeSubmissionCallbackMatches(last workflowstate.Submission, entry IntegrationEntry, request NativeSubmissionUpdate) bool {
	return last.Native != nil && last.ID == request.SubmissionID && last.Native.EntryID == entry.ID
}

func workflowNativeDispatchStopped(native workflowstate.NativeSubmission, run workflowInvocationFile, entry IntegrationEntry) bool {
	return native.DispatchState == "not_started" && workflowIntegrationReservationStopped(run, entry)
}

func checkWorkflowNativeSubmissionCallback(run workflowInvocationFile, entry IntegrationEntry, budget IntegrationBudget, last workflowstate.Submission, hash string, request NativeSubmissionUpdate) error {
	if hash != request.JournalHash {
		return workflowstate.ErrConflict
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.IntegrationBudgetGuard); err != nil {
		return err
	}

	if workflowNativeDispatchStopped(*last.Native, run, entry) {
		return ErrWorkflowStop
	}

	return guardWorkflowNativeSubmissionUpdate(*last.Native, request.Facts)
}

func projectWorkflowNativeStop(entry *IntegrationEntry, last workflowstate.Submission) bool {
	if last.State != "unpublished" || entry.StopIntent == "" {
		return false
	}

	entry.State, entry.Blocker = workflowstate.StopStatus(entry.StopIntent), nil
	if entry.State == "cancelled" {
		entry.PreparationReserved = false
	}

	return true
}
