package workflow

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowNativeObservationBinding(ctx context.Context, submission workflowstate.Submission, request GitPublication) (binding NativePublicationBinding, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, submission.InvocationID)
		if err != nil {
			return err
		}

		return invocationLocation.Owner.readWorkflowNativeObservationBinding(submission, request, &binding)
	})

	return binding, err
}

func (s *Store) readWorkflowNativeObservationBinding(submission workflowstate.Submission, request GitPublication, binding *NativePublicationBinding) error {
	native := submission.Native
	if native.ApplicationState != "git_applied" || native.TransportResponseRef == nil {
		return ErrWorkflowObservationRequired
	}

	binding.PullRequest, binding.Git, binding.SourceDirectory = submission.Guard.PullRequest, request.Git, request.Directory

	binding.ObservationDeadline = native.ObservationDeadline

	binding.ReceiptConfirmed = native.ReceiptRef != nil
	if binding.ReceiptConfirmed {
		binding.ObservationDeadline = time.Time{}
	}

	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CapabilityRef, &binding.Capability); err != nil {
		return err
	}

	if err := s.readWorkflowNativeEvidence(native.InvocationID, *native.TransportResponseRef, &binding.Publication); err != nil {
		return err
	}

	for _, ref := range native.GitObservationRefs {
		var observation NativeGitObservation
		if err := s.readWorkflowNativeEvidence(native.InvocationID, ref, &observation); err != nil {
			return err
		}

		if observation.Git != nil && observation.SourceCommit == native.PreparedResult {
			binding.GitApplication = &observation

			break
		}
	}

	return validateWorkflowNativeBinding(*binding)
}

func (s *Store) retainWorkflowNativeCanonicalObservation(ctx context.Context, update NativeSubmissionUpdate, expected workflowstate.Submission, observation NativePublicationObservation) (submission workflowstate.Submission, next IntegrationBudgetGuard, err error) {
	submission = expected
	err = s.withLock(ctx, func() error {
		invocationLocation2, err := s.locateWorkflowInvocation(ctx, update.InvocationID)
		if err != nil {
			return err
		}

		if err := invocationLocation2.Owner.checkWorkflowNativeCanonicalCallback(ctx, invocationLocation2.Run, update); err != nil {
			return err
		}

		submission, err = invocationLocation2.Owner.recordWorkflowNativeCanonicalObservation(ctx, invocationLocation2.Run, update, expected, observation)
		if err != nil {
			return err
		}

		invocationLocation2.Run, err = invocationLocation2.Owner.readWorkflowInvocation(invocationLocation2.Run.ID)
		if err != nil {
			return err
		}

		entry, err := invocationLocation2.Owner.readWorkflowIntegrationEntry(invocationLocation2.Run, *invocationLocation2.Run.Integration)
		if err == nil {
			next = IntegrationBudgetGuard{InvocationID: invocationLocation2.Run.ID, InvocationRevision: invocationLocation2.Run.Revision, BudgetRevision: invocationLocation2.Run.IntegrationBudget.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration}
		}

		return err
	})

	return submission, next, err
}

func (s *Store) recordWorkflowNativeCanonicalObservation(ctx context.Context, run workflowInvocationFile, update NativeSubmissionUpdate, expected workflowstate.Submission, observation NativePublicationObservation) (workflowstate.Submission, error) {
	ref, err := s.writeWorkflowNativeDispatchEvidence(expected, observation)
	if err != nil {
		return expected, err
	}

	if err := appendWorkflowNativeCanonicalObservation(&update.Facts, ref); err != nil {
		return expected, err
	}

	return s.retainWorkflowNativeSubmission(ctx, run, update)
}

func appendWorkflowNativeCanonicalObservation(native *workflowstate.NativeSubmission, ref workflowstate.EvidenceRef) error {
	refs := native.ObservationRefs
	if len(refs) > 0 && refs[len(refs)-1] == ref {
		return nil
	}

	if len(refs) >= workflowstate.PageLimit {
		return ErrWorkflowObservationRequired
	}

	native.ObservationRefs = append(refs, ref)

	return nil
}

func (s *Store) checkWorkflowNativeCanonicalCallback(ctx context.Context, run workflowInvocationFile, update NativeSubmissionUpdate) error {
	budgetObservation, err := s.currentWorkflowIntegrationBudget(ctx, run, update.IntegrationBudgetGuard)

	entry, budget := budgetObservation.Entry, budgetObservation.Budget

	if err != nil {
		return err
	}

	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowNativeSubmissionCallbackMatches(last, entry, update) {
		return workflowstate.ErrConflict
	}

	return checkWorkflowNativeSubmissionCallback(run, entry, budget, last, journal.sourceHash, update)
}
