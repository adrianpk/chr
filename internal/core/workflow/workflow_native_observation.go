package workflow

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NativePublicationObserver reads canonical metadata and fetched Git
// proof through the existing forge adapters. It has no publication authority.
type NativePublicationObserver interface {
	ObserveWorkflowNativePublication(context.Context, NativePublicationBinding, IntegrationGit) (NativePublicationObservation, error)
}

// ObserveWorkflowNativeSubmission reconciles the retained intent once, including
// an explicit observation after its polling deadline. It never issues receipts.
func (s *Store) ObserveWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, observer NativePublicationObserver, git IntegrationGit) (workflowstate.Submission, error) {
	submission, _, _, err := s.observeWorkflowNativeSubmission(ctx, guard, observer, git, false)

	return submission, err
}

// PollWorkflowNativeSubmission polls with exponential backoff within the same
// persisted deadline. Metadata lag retains exclusion and consumes no submission.
func (s *Store) PollWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, observer NativePublicationObserver, git IntegrationGit) (workflowstate.Submission, error) {
	delay := 250 * time.Millisecond

	for {
		submission, observation, next, err := s.observeWorkflowNativeSubmission(ctx, guard, observer, git, true)
		if err != nil || observation.State != "observing" {
			return submission, err
		}

		guard = next

		remaining := time.Until(submission.Native.ObservationDeadline)
		if remaining <= 0 {
			continue
		}

		timer := time.NewTimer(min(delay, remaining))
		select {
		case <-ctx.Done():
			timer.Stop()

			return submission, ctx.Err()
		case <-timer.C:
		}

		delay = min(2*delay, 8*time.Second)
	}
}

func (s *Store) observeWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, observer NativePublicationObserver, git IntegrationGit, poll bool) (submission workflowstate.Submission, observation NativePublicationObservation, next IntegrationBudgetGuard, err error) {
	if observer == nil || git == nil {
		return submission, observation, next, workflowstate.ErrAtomicGuard
	}

	submission, request, update, err := s.workflowNativeReconciliationInput(ctx, guard)
	if err != nil {
		return submission, observation, next, err
	}

	binding, err := s.workflowNativeObservationBinding(ctx, submission, request)
	if err != nil {
		return submission, observation, next, err
	}

	if poll && !binding.ReceiptConfirmed && !time.Now().Before(binding.ObservationDeadline) {
		observation = workflowNativeDeadlineObservation(binding)
	} else {
		observation, err = observeWorkflowNativeBounded(ctx, observer, binding, git)
	}

	// IO completion may coincide with cancellation; only the unchanged owner may
	// retain its sanitized outcome. Cleanup cannot authorize another publication.
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	retained, next, retainErr := s.retainWorkflowNativeCanonicalObservation(cleanup, update, submission, observation)

	return retained, observation, next, errors.Join(err, retainErr)
}

func observeWorkflowNativeBounded(ctx context.Context, observer NativePublicationObserver, binding NativePublicationBinding, git IntegrationGit) (NativePublicationObservation, error) {
	deadline := time.Now().Add(3 * time.Minute)
	if time.Now().Before(binding.ObservationDeadline) && binding.ObservationDeadline.Before(deadline) {
		deadline = binding.ObservationDeadline
	}

	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	observation, err := observer.ObserveWorkflowNativePublication(ctx, binding, git)
	if err != nil {
		if !time.Now().Before(binding.ObservationDeadline) && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return workflowNativeDeadlineObservation(binding), errors.Join(err, ctx.Err())
		}

		if observation.Condition == "" {
			observation.State, observation.Condition = "reconciliation_required", "canonical_observation_unavailable"
		}
	}

	observation.CanonicalPR = binding.PullRequest.URL
	observation.OriginalSourceCommit, observation.PublishedSourceCommit = binding.Git.SourceCommit, binding.Git.ResultCommit

	return observation, err
}

func workflowNativeDeadlineObservation(binding NativePublicationBinding) NativePublicationObservation {
	return NativePublicationObservation{CanonicalPR: binding.PullRequest.URL, State: "reconciliation_required", Condition: "canonical_pr_observation_deadline", PullRequest: binding.PullRequest, OriginalSourceCommit: binding.Git.SourceCommit, PublishedSourceCommit: binding.Git.ResultCommit}
}
