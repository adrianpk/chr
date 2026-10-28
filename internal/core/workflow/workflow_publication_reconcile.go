package workflow

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// workflowRetainDeliveryStatus records a fact without clearing execution stops.
// A paused invocation resumes at its delivery barrier, never at old task work.
func workflowRetainDeliveryStatus(run *workflowInvocationFile, outcome string, complete bool) {
	run.Outcome, run.Blocker = outcome, nil
	if run.StopIntent != "" || run.Status == workflowstate.InvocationPaused || run.Status == workflowstate.InvocationCancelled {
		if run.Status == workflowstate.InvocationPaused {
			run.ResumeStatus = workflowstate.InvocationWaitingIntegration
		}

		return
	}

	run.Status = workflowstate.InvocationWaitingIntegration
	if complete {
		run.Status = workflowstate.InvocationCompleted
	}
}

func workflowDeliverySubmissionID(run workflowInvocationFile) string {
	if run.Submission != nil && run.Submission.State == "confirmed" && !run.Submission.ExternalDelivery {
		return run.Submission.ID
	}

	return run.PullRequest.ID
}

func checkWorkflowVerifiedPR(ctx context.Context, run workflowInvocationFile, proof workflowstate.IntegrationProof, provider PullRequestProvider) error {
	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return err
	}

	if workflowPullRequestMismatch(*run.PullRequest, []workflowstate.PullRequestSnapshot{pr}) != "" || !pr.Merged || pr.ResultCommit != proof.ResultCommit || pr.BaseCommit != proof.ObservedTarget {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) reconcileWorkflowPublishedSubmission(ctx context.Context, run *workflowInvocationFile, pr workflowstate.PullRequestSnapshot, observers workflowDeliveryObservers) error {
	journal, err := s.loadWorkflowSubmissionJournal(ctx, *run)
	if err != nil || len(journal.Attempts) == 0 {
		return err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowSubmissionMatchesCandidate(*run, last) {
		// A permanent unpublished seal remains history after a repaired candidate.
		// This delivery must not be attributed to that older guarded command.
		if last.State != "unpublished" {
			return workflowstate.ErrConflict
		}

		return s.projectWorkflowSubmission(run, journal)
	}

	observed, err := observeWorkflowSubmittedDelivery(ctx, last, pr, observers)
	if err != nil {
		return err
	}

	if err := s.retainWorkflowSubmissionObservation(ctx, &journal, observed); err != nil {
		return err
	}

	return s.projectWorkflowSubmission(run, journal)
}

func (s *Store) retainWorkflowSubmissionObservation(ctx context.Context, journal *workflowSubmissionJournal, observed workflowstate.Submission) error {
	last := &journal.Attempts[len(journal.Attempts)-1]
	if last.ID != observed.ID {
		return workflowstate.ErrConflict
	}

	if last.State == "confirmed" {
		if observed.State != "confirmed" || last.Proof.HistoryHash != observed.Proof.HistoryHash {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if last.State == "unpublished" && observed.State == "unpublished" {
		return nil
	}

	last.State, last.ObservedPR, last.Proof = observed.State, observed.ObservedPR, observed.Proof
	last.Unpublished, last.ExternalDelivery = observed.Unpublished, observed.ExternalDelivery
	last.UpdatedAt = time.Now().UTC()

	return s.writeWorkflowSubmissionJournal(ctx, journal)
}

// New preparation, review or repair must not invalidate a pending publication.
func (s *Store) guardWorkflowPublicationWork(ctx context.Context, run workflowInvocationFile) error {
	if err := s.guardWorkflowIntegrationRoleOwner(run); err != nil {
		return err
	}

	return s.guardWorkflowPublicationJournal(ctx, run)
}

func (s *Store) guardWorkflowPublicationJournal(ctx context.Context, run workflowInvocationFile) error {
	if run.DeliveryProofRef != "" || run.DeliveryRef != "" {
		return ErrWorkflowOwned
	}

	journal, err := s.loadWorkflowSubmissionJournal(ctx, run)
	if err != nil || len(journal.Attempts) == 0 {
		return err
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.State != "unpublished" || run.Submission == nil || *run.Submission != workflowSubmissionProjection(journal) {
		return ErrWorkflowObservationRequired
	}

	return nil
}

func (s *Store) checkWorkflowDeliveryEvidence(run workflowInvocationFile) error {
	if run.Preparation != nil && run.Preparation.ResultRef != "" {
		if _, err := s.readWorkflowPreparation(run); err != nil {
			return err
		}
	}

	if run.Review != nil {
		if _, err := s.readWorkflowReview(run); err != nil {
			return err
		}
	}

	return nil
}

func workflowSubmissionMatchesCandidate(run workflowInvocationFile, submission workflowstate.Submission) bool {
	return run.PullRequest != nil && submission.CandidateID == run.PullRequest.CandidateID && submission.CandidateHash == run.CandidateHash
}

func checkWorkflowSupersededSubmission(ctx context.Context, run workflowInvocationFile, submission workflowstate.Submission, provider PublicationProvider) error {
	if submission.State != "unpublished" {
		return ErrWorkflowObservationRequired
	}

	sealed, err := provider.ObserveWorkflowSubmission(ctx, submission)
	if err != nil {
		return err
	}

	if sealed.SubmissionID != submission.ID || !sealed.Unpublished {
		return ErrWorkflowObservationRequired
	}

	return checkWorkflowSupersededTarget(ctx, run, provider)
}

func checkWorkflowSupersededTarget(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider) error {
	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return err
	}

	if workflowPullRequestMismatch(*run.PullRequest, []workflowstate.PullRequestSnapshot{pr}) != "" || pr.Merged {
		return ErrWorkflowObservationRequired
	}

	target, err := provider.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)
	if err != nil {
		return err
	}

	if !workflowstate.ValidCommitID(target) || target != pr.BaseCommit {
		return workflowstate.ErrConflict
	}

	return nil
}
