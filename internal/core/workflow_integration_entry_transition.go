package core

import (
	"context"
	"errors"
	"math"
	"os"
	"slices"
	"time"
)

func (s *WorkflowStore) transitionWorkflowIntegration(ctx context.Context, run workflowInvocationFile, request WorkflowIntegrationTransition) (WorkflowIntegrationEntry, error) {
	if run.Integration == nil || run.Integration.EntryID != request.EntryID || run.Revision != request.InvocationRevision {
		return WorkflowIntegrationEntry{}, workflowIntegrationConflict("callback has a stale invocation pointer")
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

	before := run.Integration.EntryRef.Hash

	updateWorkflowIntegrationEntry(&entry, request)
	bindWorkflowIntegrationPublication(&entry, run)
	retainWorkflowIntegrationStop(&entry, run, request)
	projectWorkflowIntegrationPhase(&run, entry)

	return entry, s.commitWorkflowIntegrationEntry(&run, entry, before)
}

func guardWorkflowIntegrationTransition(entry WorkflowIntegrationEntry, request WorkflowIntegrationTransition) error {
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

func workflowIntegrationPublicationStop(entry WorkflowIntegrationEntry, request WorkflowIntegrationTransition) bool {
	return request.State == entry.State && request.StopIntent != "" && request.CandidateRef == nil && request.Blocker == nil && slices.Contains([]string{"publishing", "observing"}, entry.State)
}

func (s *WorkflowStore) checkWorkflowIntegrationTransition(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, request WorkflowIntegrationTransition) error {
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

	return s.checkWorkflowIntegrationPublication(ctx, run, entry, next)
}

func updateWorkflowIntegrationEntry(entry *WorkflowIntegrationEntry, request WorkflowIntegrationTransition) {
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

	entry.Revision++
	entry.StateRevision++
	entry.UpdatedAt = time.Now().UTC()
}

func (s *WorkflowStore) checkWorkflowIntegrationCandidate(entry WorkflowIntegrationEntry) error {
	if entry.CandidateRef == nil {
		return nil
	}

	var candidate WorkflowIntegrationCandidate
	if err := s.readWorkflowNativeEvidence(entry.InvocationID, *entry.CandidateRef, &candidate); err != nil {
		return err
	}

	if candidate.Native == nil || candidate.Native.EntryID != entry.ID || candidate.InvocationID != entry.InvocationID || candidate.Native.Unit != entry.Unit || candidate.Native.SourceRepository != entry.Repository || candidate.Native.TargetRef != entry.TargetRef {
		return workflowIntegrationConflict("candidate has a foreign unit or target")
	}

	return s.checkWorkflowNativeCandidate(candidate)
}

func (s *WorkflowStore) checkWorkflowIntegrationAuthorization(run workflowInvocationFile, entry WorkflowIntegrationEntry) error {
	if entry.CandidateRef == nil || run.Preparation == nil || run.Review == nil || run.Review.State != "authorized" || run.Preparation.State != "validated" {
		return workflowIntegrationConflict("requires retained preparation and exact review authorization")
	}

	if run.Preparation.Unit != entry.Unit || run.Review.Unit != entry.Unit || run.Preparation.ResultRef != entry.CandidateRef.Path || run.Preparation.ResultHash != entry.CandidateRef.Hash {
		return workflowIntegrationConflict("authorization names another candidate")
	}

	_, err := s.readWorkflowPreparation(run)

	return err
}

func (s *WorkflowStore) checkWorkflowIntegrationPublication(ctx context.Context, run workflowInvocationFile, old, next WorkflowIntegrationEntry) error {
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
	if err := guardWorkflowIntegrationSubmissionPhase(next, latest); err != nil {
		return err
	}

	if err := checkWorkflowIntegrationSubmissionBinding(run, next, journal); err != nil {
		return err
	}

	return s.checkWorkflowIntegrationFinalization(ctx, run, old, next, latest)
}

func guardWorkflowIntegrationSubmissionPhase(entry WorkflowIntegrationEntry, submission WorkflowSubmission) error {
	if submission.State == "awaiting_observation" && !slices.Contains([]string{"publishing", "observing", "blocked"}, entry.State) {
		return ErrWorkflowObservationRequired
	}

	if submission.State == "confirmed" && !slices.Contains([]string{"observing", "finalizing", "integrated", "blocked"}, entry.State) {
		return workflowIntegrationConflict("confirmed publication cannot be requeued or cancelled")
	}

	return nil
}

func checkWorkflowIntegrationSubmissionBinding(run workflowInvocationFile, entry WorkflowIntegrationEntry, journal workflowSubmissionJournal) error {
	if entry.State != "publishing" && entry.State != "observing" {
		return nil
	}

	if journal.Schema != "chr-integration-journal/v2" || entry.SubmissionRef == nil || run.Submission == nil || *entry.SubmissionRef != *run.Submission {
		return workflowIntegrationConflict("requires native publication facts and their current projection")
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowIntegrationFinalization(ctx context.Context, run workflowInvocationFile, old, next WorkflowIntegrationEntry, submission WorkflowSubmission) error {
	if old.State == "publishing" && next.State == "queued" {
		// Unpublished proof alone does not establish changed input and a dead
		// dispatch. Native journal reconciliation owns that composite transition.
		return workflowIntegrationConflict("requires native dispatch reconciliation before requeue")
	}

	if next.State != "finalizing" && next.State != "integrated" {
		return nil
	}

	if !workflowIntegrationReceiptBindingMatches(run, next, submission) {
		return workflowIntegrationConflict("receipt differs from confirmed publication")
	}

	receipt, err := s.readWorkflowSliceReceipt(ctx, run)
	if err != nil {
		return err
	}

	if receipt.Receipt.Native == nil || receipt.Receipt.Native.EntryID != next.ID || receipt.Receipt.Native.InvocationID != run.ID {
		return workflowIntegrationConflict("requires native verified receipt")
	}

	return nil
}

func workflowIntegrationReceiptBindingMatches(run workflowInvocationFile, entry WorkflowIntegrationEntry, submission WorkflowSubmission) bool {
	return submission.State == "confirmed" && entry.ReceiptRef != nil && entry.ReceiptRef.Path == run.DeliveryRef && entry.ReceiptRef.Hash == run.DeliveryHash
}

func projectWorkflowIntegrationPhase(run *workflowInvocationFile, entry WorkflowIntegrationEntry) {
	switch entry.State {
	case "queued":
		run.Status = "pending"
	case "waiting_review":
		run.Status = "waiting_review"
	case "blocked":
		run.Status = "blocked"
		run.Blocker = &WorkflowBlocker{Dependency: entry.Unit, Reason: entry.Blocker.Reason, NextActionOwner: "operator"}
	case "paused", "cancelled":
		run.Status = entry.State
	case "integrated":
		// Verified unit completion is distinct from the existing set frontier.
	default:
		run.Status = "waiting_integration"
	}

	if entry.State != "blocked" {
		run.Blocker = nil
	}

	run.StopIntent = entry.StopIntent
}

func bindWorkflowIntegrationPublication(entry *WorkflowIntegrationEntry, run workflowInvocationFile) {
	if slices.Contains([]string{"publishing", "observing", "finalizing", "integrated"}, entry.State) {
		entry.SubmissionRef = run.Submission
	}

	if entry.State == "finalizing" || entry.State == "integrated" {
		entry.ReceiptRef = &WorkflowEvidenceRef{Path: run.DeliveryRef, Hash: run.DeliveryHash}
	}
}

func guardWorkflowIntegrationStop(run workflowInvocationFile, entry WorkflowIntegrationEntry, request WorkflowIntegrationTransition) error {
	if request.ClearStopIntent && entry.State != "paused" {
		return workflowIntegrationConflict("stop intent may be cleared only at a paused checkpoint")
	}

	stopped := run.StopIntent != "" || entry.StopIntent != "" || request.StopIntent != ""
	if stopped && !request.ClearStopIntent && (request.State == "queued" || request.State == "preparing") {
		return workflowIntegrationConflict("retained stop intent blocks execution")
	}

	return nil
}

func retainWorkflowIntegrationStop(entry *WorkflowIntegrationEntry, run workflowInvocationFile, request WorkflowIntegrationTransition) {
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
