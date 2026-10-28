package workflow

import (
	"context"
	"errors"
	"math"
	"os"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) transitionWorkflowIntegration(ctx context.Context, run workflowInvocationFile, request IntegrationTransition) (IntegrationEntry, error) {
	if run.Integration == nil || run.Integration.EntryID != request.EntryID || run.Revision != request.InvocationRevision {
		return IntegrationEntry{}, workflowIntegrationConflict("callback has a stale invocation pointer")
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return entry, err
	}

	if err := guardWorkflowIntegrationTransition(entry, request); err != nil {
		return entry, err
	}

	if err := s.checkWorkflowIntegrationTransition(ctx, run, entry, request); err != nil {
		return entry, err
	}

	if err := s.guardWorkflowPreparationTransition(ctx, run, request); err != nil {
		return entry, err
	}

	before := run.Integration.EntryRef.Hash

	updateWorkflowIntegrationEntry(&entry, request)
	bindWorkflowIntegrationPublication(&entry, run)
	retainWorkflowIntegrationStop(&entry, run, request)
	projectWorkflowIntegrationPhase(&run, entry)

	return entry, s.commitWorkflowIntegrationEntry(&run, entry, before)
}

func guardWorkflowIntegrationTransition(entry IntegrationEntry, request IntegrationTransition) error {
	hash := ""
	if entry.CandidateRef != nil {
		hash = entry.CandidateRef.Hash
	}

	if entry.Revision != request.EntryRevision || entry.ClaimGeneration != request.ClaimGeneration || hash != request.CandidateHash {
		return workflowIntegrationConflict("callback has a stale revision, generation or candidate")
	}

	if entry.Revision == math.MaxUint64 || entry.StateRevision == math.MaxUint64 || entry.ClaimGeneration == math.MaxUint64 {
		return workflowIntegrationConflict("revision or generation exhausted")
	}

	if workflowIntegrationPublicationStop(entry, request) {
		return nil
	}

	return workflowIntegrationNextState(entry, request.State)
}

func workflowIntegrationPublicationStop(entry IntegrationEntry, request IntegrationTransition) bool {
	return request.State == entry.State && request.StopIntent != "" && request.CandidateRef == nil && request.Blocker == nil && slices.Contains([]string{"publishing", "observing"}, entry.State)
}

func (s *Store) checkWorkflowIntegrationTransition(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, request IntegrationTransition) error {
	if err := s.guardWorkflowDecisionTransition(ctx, run, entry, request); err != nil {
		return err
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return err
	}

	if err := guardWorkflowIntegrationStop(run, entry, request); err != nil {
		return err
	}

	next := entry
	updateWorkflowIntegrationEntry(&next, request)
	bindWorkflowIntegrationPublication(&next, run)
	retainWorkflowIntegrationStop(&next, run, request)

	if err := validateWorkflowIntegrationEntry(next); err != nil {
		return err
	}

	if err := s.checkWorkflowIntegrationCandidate(next); err != nil {
		return err
	}

	if slices.Contains([]string{"ready_to_publish", "publishing"}, next.State) {
		if err := s.checkWorkflowIntegrationAuthorization(run, next); err != nil {
			return err
		}
	}

	if err := s.checkWorkflowIntegrationPublication(ctx, run, entry, next); err != nil {
		return err
	}

	return s.checkWorkflowIntegrationRetainedRecords(ctx, run)
}

func updateWorkflowIntegrationEntry(entry *IntegrationEntry, request IntegrationTransition) {
	if request.CandidateRef != nil {
		entry.CandidateRef = request.CandidateRef
	}

	if entry.State == "queued" && request.State == "preparing" {
		entry.ClaimGeneration++
	}

	entry.State, entry.Blocker = request.State, request.Blocker
	if request.StopIntent != "" {
		entry.StopIntent = request.StopIntent
	}

	if request.CandidateRef != nil || slices.Contains([]string{"preparing", "merging", "validating"}, request.State) {
		entry.PreparationReserved = true
	}

	if workflowIntegrationTerminal(request.State) {
		entry.PreparationReserved = false
	}

	entry.Revision++
	entry.StateRevision++
	entry.UpdatedAt = time.Now().UTC()
}

func (s *Store) checkWorkflowIntegrationCandidate(entry IntegrationEntry) error {
	if entry.CandidateRef == nil {
		return nil
	}

	var candidate workflowstate.IntegrationCandidate
	if err := s.readWorkflowNativeEvidence(entry.InvocationID, *entry.CandidateRef, &candidate); err != nil {
		return err
	}

	if candidate.Native == nil || candidate.Native.EntryID != entry.ID || candidate.InvocationID != entry.InvocationID || candidate.Native.Unit != entry.Unit || candidate.Native.SourceRepository != entry.Repository || candidate.Native.TargetRef != entry.TargetRef {
		return workflowIntegrationConflict("candidate has a foreign unit or target")
	}

	return s.checkWorkflowNativeCandidate(candidate)
}

func (s *Store) checkWorkflowIntegrationAuthorization(run workflowInvocationFile, entry IntegrationEntry) error {
	if entry.CandidateRef == nil || run.Preparation == nil || run.Review == nil || run.Review.State != "authorized" || run.Preparation.State != "validated" {
		return workflowIntegrationConflict("requires retained preparation and exact review authorization")
	}

	if run.Preparation.Unit != entry.Unit || run.Review.Unit != entry.Unit || run.Preparation.ResultRef != entry.CandidateRef.Path || run.Preparation.ResultHash != entry.CandidateRef.Hash {
		return workflowIntegrationConflict("authorization names another candidate")
	}

	_, err := s.readWorkflowPreparation(run)

	return err
}

func (s *Store) checkWorkflowIntegrationPublication(ctx context.Context, run workflowInvocationFile, old, next IntegrationEntry) error {
	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if errors.Is(err, os.ErrNotExist) {
		if run.Submission != nil || next.SubmissionRef != nil || slices.Contains([]string{"publishing", "observing", "finalizing", "integrated"}, next.State) {
			return workflowIntegrationConflict("has no native submission journal")
		}

		return nil
	}

	if err != nil {
		return err
	}

	latest := journal.Attempts[len(journal.Attempts)-1]
	if workflowIntegrationCandidateExcluded(next, latest) {
		return ErrWorkflowObservationRequired
	}

	if err := guardWorkflowIntegrationSubmissionPhase(next, latest); err != nil {
		return err
	}

	if err := checkWorkflowIntegrationSubmissionBinding(run, next, journal); err != nil {
		return err
	}

	return s.checkWorkflowIntegrationFinalization(ctx, run, old, next, latest)
}

func guardWorkflowIntegrationSubmissionPhase(entry IntegrationEntry, submission workflowstate.Submission) error {
	if submission.State == "awaiting_observation" && !slices.Contains([]string{"publishing", "observing", "blocked"}, entry.State) {
		return ErrWorkflowObservationRequired
	}

	if submission.State == "confirmed" && !slices.Contains([]string{"observing", "finalizing", "integrated", "blocked"}, entry.State) {
		return workflowIntegrationConflict("confirmed publication cannot be requeued or cancelled")
	}

	return nil
}

func checkWorkflowIntegrationSubmissionBinding(run workflowInvocationFile, entry IntegrationEntry, journal workflowSubmissionJournal) error {
	if entry.State != "publishing" && entry.State != "observing" {
		return nil
	}

	if journal.Schema != "chr-integration-journal/v2" || entry.SubmissionRef == nil || run.Submission == nil || *entry.SubmissionRef != *run.Submission {
		return workflowIntegrationConflict("requires native publication facts and their current projection")
	}

	return nil
}

func (s *Store) checkWorkflowIntegrationFinalization(ctx context.Context, run workflowInvocationFile, old, next IntegrationEntry, submission workflowstate.Submission) error {
	if workflowIntegrationNeedsRequeueProof(next, submission) {
		if submission.State != "unpublished" {
			return ErrWorkflowObservationRequired
		}

		return s.checkWorkflowNativeNonApplication(ctx, submission)
	}

	if old.State == "publishing" && next.State == "queued" {
		return workflowIntegrationConflict("requires native dispatch reconciliation before requeue")
	}

	if next.State != "finalizing" && next.State != "integrated" {
		return nil
	}

	if !workflowIntegrationReceiptBindingMatches(run, next, submission) {
		return workflowIntegrationConflict("receipt differs from confirmed publication")
	}

	if err := s.checkWorkflowNativeReceiptContinuation(submission); err != nil {
		return err
	}

	return s.checkWorkflowIntegrationReceiptImages(ctx, run, next)
}

func (s *Store) checkWorkflowIntegrationReceiptImages(ctx context.Context, run workflowInvocationFile, next IntegrationEntry) error {
	receipt, err := s.readWorkflowFinalizationReceipt(ctx, run)
	if err != nil {
		return err
	}

	if next.State == "integrated" {
		if err := s.checkWorkflowFinalizationImages(ctx, run, receipt); err != nil {
			return err
		}
	}

	if !workflowIntegrationReceiptOwnerMatches(receipt.Receipt, run, next) {
		return workflowIntegrationConflict("requires native verified receipt")
	}

	return nil
}

func workflowIntegrationReceiptBindingMatches(run workflowInvocationFile, entry IntegrationEntry, submission workflowstate.Submission) bool {
	return submission.State == "confirmed" && entry.ReceiptRef != nil && entry.ReceiptRef.Path == run.DeliveryRef && entry.ReceiptRef.Hash == run.DeliveryHash
}

func projectWorkflowIntegrationPhase(run *workflowInvocationFile, entry IntegrationEntry) {
	checkOwned := workflowIntegrationCheckOwned(*run)

	switch entry.State {
	case "queued":
		run.Status = workflowstate.InvocationPending
	case "waiting_review":
		run.Status = workflowstate.InvocationWaitingReview
	case "blocked":
		run.Status = workflowstate.InvocationBlocked
		run.Blocker = &Blocker{Dependency: entry.Unit, Reason: entry.Blocker.Reason, NextActionOwner: "operator"}
	case "paused", "cancelled":
		run.Status = workflowstate.InvocationStatus(entry.State)
	case "integrated":
		// Verified unit completion is distinct from the existing set frontier.
	default:
		run.Status = workflowstate.InvocationWaitingIntegration
	}

	if entry.State != "blocked" {
		run.Blocker = nil
	}

	if checkOwned {
		// A queue projection cannot replace a live gate or review credential. Its owner
		// must retain stopped observations before releasing shared capacity.
		run.Status = workflowstate.InvocationRunning
	}

	run.StopIntent = entry.StopIntent
}

func workflowIntegrationCheckOwned(run workflowInvocationFile) bool {
	if run.Status != workflowstate.InvocationRunning || len(run.Attempts) == 0 {
		return false
	}

	attempt := run.Attempts[len(run.Attempts)-1]

	return slices.Contains([]string{workflowIntegrationValidationPurpose, workflowSliceReviewPurpose}, attempt.Purpose) && attempt.Outcome == "running"
}

func bindWorkflowIntegrationPublication(entry *IntegrationEntry, run workflowInvocationFile) {
	if slices.Contains([]string{"publishing", "observing", "finalizing", "integrated"}, entry.State) {
		entry.SubmissionRef = run.Submission
	}

	if entry.State == "finalizing" || entry.State == "integrated" {
		entry.ReceiptRef = &workflowstate.EvidenceRef{Path: run.DeliveryRef, Hash: run.DeliveryHash}
	}
}

func guardWorkflowIntegrationStop(run workflowInvocationFile, entry IntegrationEntry, request IntegrationTransition) error {
	if request.ClearStopIntent && entry.State != "paused" {
		return workflowIntegrationConflict("stop intent may be cleared only at a paused checkpoint")
	}

	stopped := run.StopIntent != "" || entry.StopIntent != "" || request.StopIntent != ""
	if stopped && !request.ClearStopIntent && (request.State == "queued" || request.State == "preparing") {
		return workflowIntegrationConflict("retained stop intent blocks execution")
	}

	return nil
}

func retainWorkflowIntegrationStop(entry *IntegrationEntry, run workflowInvocationFile, request IntegrationTransition) {
	if request.ClearStopIntent {
		entry.StopIntent = ""

		return
	}

	if run.StopIntent == "cancel" || entry.StopIntent == "cancel" {
		entry.StopIntent = "cancel"
	} else if entry.StopIntent == "" {
		entry.StopIntent = run.StopIntent
	}
}

func workflowIntegrationCandidateExcluded(entry IntegrationEntry, submission workflowstate.Submission) bool {
	if submission.Native == nil || submission.State == "unpublished" {
		return false
	}

	return !workflowNativeSubmissionEntryMatches(submission, entry)
}

func workflowIntegrationReceiptOwnerMatches(receipt *Receipt, run workflowInvocationFile, entry IntegrationEntry) bool {
	return receipt.Native != nil && receipt.Native.EntryID == entry.ID && receipt.Native.InvocationID == run.ID
}

func workflowIntegrationNeedsRequeueProof(entry IntegrationEntry, submission workflowstate.Submission) bool {
	return entry.State == "queued" && submission.Native != nil
}

func (s *Store) guardWorkflowPreparationTransition(ctx context.Context, run workflowInvocationFile, request IntegrationTransition) error {
	if request.CandidateRef == nil && !slices.Contains([]string{"preparing", "merging", "validating"}, request.State) {
		return nil
	}

	snapshot, err := s.readWorkflowIntegrationSnapshot(ctx)
	if err != nil {
		return err
	}

	return s.guardWorkflowPreparedCapacity(ctx, snapshot, run)
}
