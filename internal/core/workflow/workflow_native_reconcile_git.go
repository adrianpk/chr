package workflow

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReconcileWorkflowNativeGitSubmission observes the same submission without
// dispatching. Actual refs and method proof classify Git, never forge delivery.
func (s *Store) ReconcileWorkflowNativeGitSubmission(ctx context.Context, guard IntegrationBudgetGuard, inspector GitPublicationInspector, git PreparedIntegrationGit) (workflowstate.Submission, error) {
	if inspector == nil || git == nil {
		return workflowstate.Submission{}, workflowstate.ErrAtomicGuard
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	nativeReconciliation, err := s.workflowNativeReconciliationInput(ctx, guard)

	submission, request, update := nativeReconciliation.Submission, nativeReconciliation.Request, nativeReconciliation.Update
	if err != nil || submission.Native.ApplicationState == "not_applied" {
		return submission, err
	}

	observation, err := observeWorkflowNativeGit(ctx, submission, request, inspector, git)
	if err != nil {
		return submission, err
	}

	return s.retainWorkflowNativeGitObservation(ctx, update, submission, observation)
}

func (s *Store) workflowNativeReconciliationInput(ctx context.Context, guard IntegrationBudgetGuard) (result workflowNativeReconciliation, err error) {
	var (
		submission workflowstate.Submission
		request    GitPublication
		update     NativeSubmissionUpdate
	)

	err = s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if err != nil {
			return err
		}

		if err := invocationLocation.Owner.recoverWorkflowNativeReceiptProjection(ctx, &invocationLocation.Run, guard); err != nil {
			return err
		}

		budgetObservation, err := invocationLocation.Owner.currentWorkflowIntegrationBudget(ctx, invocationLocation.Run, guard)

		entry, budget := budgetObservation.Entry, budgetObservation.Budget

		if err != nil {
			return err
		}

		if err := guardWorkflowIntegrationBudget(invocationLocation.Run, entry, budget, guard); err != nil {
			return err
		}

		journal, err := invocationLocation.Owner.readWorkflowSubmissionJournal(ctx, invocationLocation.Run)
		if err != nil {
			return err
		}

		submission = journal.Attempts[len(journal.Attempts)-1]
		if !workflowNativeSubmissionEntryMatches(submission, entry) {
			return workflowstate.ErrConflict
		}

		var candidate workflowstate.IntegrationCandidate
		if err := invocationLocation.Owner.readWorkflowNativeEvidence(invocationLocation.Run.ID, submission.Native.CandidateRef, &candidate); err != nil {
			return err
		}

		request, err = invocationLocation.Owner.workflowNativeDispatchRequest(candidate)
		update = NativeSubmissionUpdate{IntegrationBudgetGuard: guard, SubmissionID: submission.ID, JournalHash: journal.sourceHash, Facts: *submission.Native}

		return err
	})

	return workflowNativeReconciliation{Submission: submission, Request: request, Update: update}, err
}

func observeWorkflowNativeGit(ctx context.Context, submission workflowstate.Submission, request GitPublication, inspector GitPublicationInspector, git PreparedIntegrationGit) (workflowstate.NativeGitObservation, error) {
	advertisement, err := inspector.InspectWorkflowGitPublication(ctx, request)
	if err != nil && !errors.Is(err, workflowstate.ErrConflict) && !errors.Is(err, ErrWorkflowObservationRequired) && !errors.Is(err, workflowstate.ErrAtomicGuard) {
		return workflowstate.NativeGitObservation{}, err
	}

	observation := workflowstate.NativeGitObservation{SchemaVersion: 1, SubmissionID: submission.ID, CandidateRevision: request.CandidateRevision, SourceCommit: advertisement.SourceCommit, TargetCommit: advertisement.TargetCommit, Condition: "result_unverified"}
	if err := validateWorkflowNativeGitObservation(submission, observation); err != nil {
		return observation, err
	}

	query := request.Git.IntegrationQuery
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

func (s *Store) retainWorkflowNativeGitObservation(ctx context.Context, update NativeSubmissionUpdate, submission workflowstate.Submission, observation workflowstate.NativeGitObservation) (retained workflowstate.Submission, err error) {
	retained = submission
	err = s.withLock(ctx, func() error {
		invocationLocation2, err := s.locateWorkflowInvocation(ctx, update.InvocationID)
		if err != nil {
			return err
		}

		if len(update.Facts.GitObservationRefs) >= workflowstate.PageLimit {
			return ErrWorkflowObservationRequired
		}

		ref, err := invocationLocation2.Owner.writeWorkflowNativeDispatchEvidence(submission, observation)
		if err != nil {
			return err
		}

		update.Facts.GitObservationRefs = append(update.Facts.GitObservationRefs, ref)
		if err := invocationLocation2.Owner.classifyWorkflowNativeGitObservation(submission, observation, &update.Facts); err != nil {
			return err
		}

		retained, err = invocationLocation2.Owner.retainWorkflowNativeSubmission(ctx, invocationLocation2.Run, update)

		return err
	})

	return retained, err
}

func (s *Store) classifyWorkflowNativeGitObservation(submission workflowstate.Submission, observation workflowstate.NativeGitObservation, native *workflowstate.NativeSubmission) error {
	if native.TransportResponseRef == nil {
		return nil
	}

	var response workflowstate.GitPublicationResult
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
	proof := NativeNonApplicationProof{SchemaVersion: 1, SubmissionID: submission.ID, CannotApply: true, InputChanged: changed, ObservedSource: observation.SourceCommit, ObservedTarget: observation.TargetCommit}

	ref, err := s.writeWorkflowNativeDispatchEvidence(submission, proof)
	if err != nil {
		return err
	}

	native.NonApplicationRef, native.ApplicationState = &ref, "not_applied"

	return nil
}

func workflowNativeObservedApplication(native workflowstate.NativeSubmission, response workflowstate.GitPublicationResult, observation workflowstate.NativeGitObservation) bool {
	return observation.Git != nil && (response.State == "applied" || response.State == "unknown" && observation.SourceCommit == native.PreparedResult)
}

func workflowNativeMaySealRejection(native workflowstate.NativeSubmission, response workflowstate.GitPublicationResult, observation workflowstate.NativeGitObservation) bool {
	return native.DispatchCompletionRef != nil && observation.Git == nil && observation.SourceCommit != native.PreparedResult && observation.TargetCommit != native.PreparedResult && workflowNativeDefiniteResponse(native, response)
}
