package core

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReconcileWorkflowNativeGitSubmission observes the same submission without
// dispatching. Actual refs and method proof classify Git, never forge delivery.
func (s *WorkflowStore) ReconcileWorkflowNativeGitSubmission(ctx context.Context, guard WorkflowIntegrationBudgetGuard, inspector WorkflowGitPublicationInspector, git WorkflowPreparedIntegrationGit) (WorkflowSubmission, error) {
	if inspector == nil || git == nil {
		return WorkflowSubmission{}, ErrWorkflowAtomicGuard
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	submission, request, update, err := s.workflowNativeReconciliationInput(ctx, guard)
	if err != nil || submission.Native.ApplicationState == "not_applied" {
		return submission, err
	}

	observation, err := observeWorkflowNativeGit(ctx, submission, request, inspector, git)
	if err != nil {
		return submission, err
	}

	return s.retainWorkflowNativeGitObservation(ctx, update, submission, observation)
}

func (s *WorkflowStore) workflowNativeReconciliationInput(ctx context.Context, guard WorkflowIntegrationBudgetGuard) (submission WorkflowSubmission, request WorkflowGitPublication, update WorkflowNativeSubmissionUpdate, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if err != nil {
			return err
		}

		if err := owner.recoverWorkflowNativeReceiptProjection(ctx, &run, guard); err != nil {
			return err
		}

		entry, budget, err := owner.currentWorkflowIntegrationBudget(ctx, run, guard)
		if err != nil {
			return err
		}

		if err := guardWorkflowIntegrationBudget(run, entry, budget, guard); err != nil {
			return err
		}

		journal, err := owner.readWorkflowSubmissionJournal(ctx, run)
		if err != nil {
			return err
		}

		submission = journal.Attempts[len(journal.Attempts)-1]
		if !workflowNativeSubmissionEntryMatches(submission, entry) {
			return workflowstate.ErrConflict
		}

		var candidate WorkflowIntegrationCandidate
		if err := owner.readWorkflowNativeEvidence(run.ID, submission.Native.CandidateRef, &candidate); err != nil {
			return err
		}

		request, err = owner.workflowNativeDispatchRequest(candidate)
		update = WorkflowNativeSubmissionUpdate{WorkflowIntegrationBudgetGuard: guard, SubmissionID: submission.ID, JournalHash: journal.sourceHash, Facts: *submission.Native}

		return err
	})

	return submission, request, update, err
}

func observeWorkflowNativeGit(ctx context.Context, submission WorkflowSubmission, request WorkflowGitPublication, inspector WorkflowGitPublicationInspector, git WorkflowPreparedIntegrationGit) (WorkflowNativeGitObservation, error) {
	advertisement, err := inspector.InspectWorkflowGitPublication(ctx, request)
	if err != nil && !errors.Is(err, workflowstate.ErrConflict) && !errors.Is(err, ErrWorkflowObservationRequired) && !errors.Is(err, ErrWorkflowAtomicGuard) {
		return WorkflowNativeGitObservation{}, err
	}

	observation := WorkflowNativeGitObservation{SchemaVersion: 1, SubmissionID: submission.ID, CandidateRevision: request.CandidateRevision, SourceCommit: advertisement.SourceCommit, TargetCommit: advertisement.TargetCommit, Condition: "result_unverified"}
	if err := validateWorkflowNativeGitObservation(submission, observation); err != nil {
		return observation, err
	}

	query := request.Git.WorkflowIntegrationQuery
	query.SourceDirectory = request.Directory

	proof, proofErr := git.VerifyPreparedWorkflowIntegration(ctx, query, request.Git)
	if err := ctx.Err(); err != nil {
		return observation, err
	}

	if proofErr == nil {
		observation.Git = &proof
		observation.Condition = "result_verified"
	}

	return observation, validateWorkflowNativeGitObservation(submission, observation)
}

func (s *WorkflowStore) retainWorkflowNativeGitObservation(ctx context.Context, update WorkflowNativeSubmissionUpdate, submission WorkflowSubmission, observation WorkflowNativeGitObservation) (retained WorkflowSubmission, err error) {
	retained = submission
	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, update.InvocationID)
		if err != nil {
			return err
		}

		if len(update.Facts.GitObservationRefs) >= workflowstate.WorkflowPageLimit {
			return ErrWorkflowObservationRequired
		}

		ref, err := owner.writeWorkflowNativeDispatchEvidence(submission, observation)
		if err != nil {
			return err
		}

		update.Facts.GitObservationRefs = append(update.Facts.GitObservationRefs, ref)
		if err := owner.classifyWorkflowNativeGitObservation(submission, observation, &update.Facts); err != nil {
			return err
		}

		retained, err = owner.retainWorkflowNativeSubmission(ctx, run, update)

		return err
	})

	return retained, err
}

func (s *WorkflowStore) classifyWorkflowNativeGitObservation(submission WorkflowSubmission, observation WorkflowNativeGitObservation, native *WorkflowNativeSubmission) error {
	if native.TransportResponseRef == nil {
		return nil
	}

	var response WorkflowGitPublicationResult
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.TransportResponseRef, &response); err != nil {
		return err
	}

	if workflowNativeObservedApplication(*native, response, observation) {
		native.ApplicationState = "git_applied"

		return nil
	}

	if !workflowNativeMaySealRejection(*native, response, observation) {
		return nil
	}

	changed := observation.SourceCommit != native.OriginalSource || observation.TargetCommit != native.ValidatedBase
	proof := WorkflowNativeNonApplicationProof{SchemaVersion: 1, SubmissionID: submission.ID, CannotApply: true, InputChanged: changed, ObservedSource: observation.SourceCommit, ObservedTarget: observation.TargetCommit}

	ref, err := s.writeWorkflowNativeDispatchEvidence(submission, proof)
	if err != nil {
		return err
	}

	native.NonApplicationRef, native.ApplicationState = &ref, "not_applied"

	return nil
}

func workflowNativeObservedApplication(native WorkflowNativeSubmission, response WorkflowGitPublicationResult, observation WorkflowNativeGitObservation) bool {
	return observation.Git != nil && (response.State == "applied" || response.State == "unknown" && observation.SourceCommit == native.PreparedResult)
}

func workflowNativeMaySealRejection(native WorkflowNativeSubmission, response WorkflowGitPublicationResult, observation WorkflowNativeGitObservation) bool {
	return native.DispatchCompletionRef != nil && observation.Git == nil && observation.SourceCommit != native.PreparedResult && observation.TargetCommit != native.PreparedResult && workflowNativeDefiniteResponse(native, response)
}
