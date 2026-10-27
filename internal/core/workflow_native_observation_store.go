package core

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) workflowNativeObservationBinding(ctx context.Context, submission WorkflowSubmission, request WorkflowGitPublication) (binding WorkflowNativePublicationBinding, err error) {
	err = s.withLock(ctx, func() error {
		_, owner, err := s.locateWorkflowInvocation(ctx, submission.InvocationID)
		if err != nil {
			return err
		}

		return owner.readWorkflowNativeObservationBinding(submission, request, &binding)
	})

	return binding, err
}

func (s *WorkflowStore) readWorkflowNativeObservationBinding(submission WorkflowSubmission, request WorkflowGitPublication, binding *WorkflowNativePublicationBinding) error {
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
		var observation WorkflowNativeGitObservation
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

func (s *WorkflowStore) retainWorkflowNativeCanonicalObservation(ctx context.Context, update WorkflowNativeSubmissionUpdate, expected WorkflowSubmission, observation WorkflowNativePublicationObservation) (submission WorkflowSubmission, next WorkflowIntegrationBudgetGuard, err error) {
	submission = expected
	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, update.InvocationID)
		if err != nil {
			return err
		}

		if err := owner.checkWorkflowNativeCanonicalCallback(ctx, run, update); err != nil {
			return err
		}

		submission, err = owner.recordWorkflowNativeCanonicalObservation(ctx, run, update, expected, observation)
		if err != nil {
			return err
		}

		run, err = owner.readWorkflowInvocation(run.ID)
		if err != nil {
			return err
		}

		entry, err := owner.readWorkflowIntegrationEntry(run, *run.Integration)
		if err == nil {
			next = WorkflowIntegrationBudgetGuard{InvocationID: run.ID, InvocationRevision: run.Revision, BudgetRevision: run.IntegrationBudget.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration}
		}

		return err
	})

	return submission, next, err
}

func (s *WorkflowStore) recordWorkflowNativeCanonicalObservation(ctx context.Context, run workflowInvocationFile, update WorkflowNativeSubmissionUpdate, expected WorkflowSubmission, observation WorkflowNativePublicationObservation) (WorkflowSubmission, error) {
	ref, err := s.writeWorkflowNativeDispatchEvidence(expected, observation)
	if err != nil {
		return expected, err
	}

	if err := appendWorkflowNativeCanonicalObservation(&update.Facts, ref); err != nil {
		return expected, err
	}

	return s.retainWorkflowNativeSubmission(ctx, run, update)
}

func appendWorkflowNativeCanonicalObservation(native *WorkflowNativeSubmission, ref WorkflowEvidenceRef) error {
	refs := native.ObservationRefs
	if len(refs) > 0 && refs[len(refs)-1] == ref {
		return nil
	}

	if len(refs) >= WorkflowPageLimit {
		return ErrWorkflowObservationRequired
	}

	native.ObservationRefs = append(refs, ref)

	return nil
}

func (s *WorkflowStore) checkWorkflowNativeCanonicalCallback(ctx context.Context, run workflowInvocationFile, update WorkflowNativeSubmissionUpdate) error {
	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, update.WorkflowIntegrationBudgetGuard)
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
