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
	ObserveWorkflowNativePublication(context.Context, workflowstate.NativePublicationBinding, IntegrationGit) (workflowstate.NativePublicationObservation, error)
}

// ObserveWorkflowNativeSubmission reconciles the retained intent once, including
// an explicit observation after its polling deadline. It never issues receipts.
func (s *Store) ObserveWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, observer NativePublicationObserver, git IntegrationGit) (workflowstate.Submission, error) {
	nativeObservation, err := s.observeWorkflowNativeSubmission(ctx, guard, workflowNativeObservers{Canonical: observer, Git: git}, false)
	submission := nativeObservation.Submission

	return submission, err
}

// PollWorkflowNativeSubmission polls with exponential backoff within the same
// persisted deadline. Metadata lag retains exclusion and consumes no submission.
func (s *Store) PollWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, observer NativePublicationObserver, git IntegrationGit) (workflowstate.Submission, error) {
	delay := 250 * time.Millisecond

	for {
		nativeObservation2, err := s.observeWorkflowNativeSubmission(ctx, guard, workflowNativeObservers{Canonical: observer, Git: git}, true)

		submission, observation, next := nativeObservation2.Submission, nativeObservation2.Observation, nativeObservation2.Next
		if err != nil || observation.State != workflowstate.NativeObserving {
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

func (s *Store) observeWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, observers workflowNativeObservers, poll bool) (result workflowNativeObservation, err error) {
	observer, git := observers.Canonical, observers.Git

	var (
		submission  workflowstate.Submission
		observation workflowstate.NativePublicationObservation
		next        IntegrationBudgetGuard
	)

	if observer == nil || git == nil {
		return workflowNativeObservation{Submission: submission, Observation: observation, Next: next}, workflowstate.ErrAtomicGuard
	}

	nativeReconciliation, err := s.workflowNativeReconciliationInput(ctx, guard)

	submission, request, update := nativeReconciliation.Submission, nativeReconciliation.Request, nativeReconciliation.Update
	if err != nil {
		return workflowNativeObservation{Submission: submission, Observation: observation, Next: next}, err
	}

	binding, err := s.workflowNativeObservationBinding(ctx, submission, request)
	if err != nil {
		return workflowNativeObservation{Submission: submission, Observation: observation, Next: next}, err
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

	nativeRetention, retainErr := s.retainWorkflowNativeCanonicalObservation(cleanup, update, submission, observation)
	retained, next := nativeRetention.Submission, nativeRetention.Next

	return workflowNativeObservation{Submission: retained, Observation: observation, Next: next}, errors.Join(err, retainErr)
}

func observeWorkflowNativeBounded(ctx context.Context, observer NativePublicationObserver, binding workflowstate.NativePublicationBinding, git IntegrationGit) (workflowstate.NativePublicationObservation, error) {
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
			observation.State, observation.Condition = workflowstate.NativeReconciliationRequired, "canonical_observation_unavailable"
		}
	}

	observation.CanonicalPR = binding.PullRequest.URL
	observation.OriginalSourceCommit, observation.PublishedSourceCommit = binding.Git.SourceCommit, binding.Git.ResultCommit

	return observation, err
}

func workflowNativeDeadlineObservation(binding workflowstate.NativePublicationBinding) workflowstate.NativePublicationObservation {
	return workflowstate.NativePublicationObservation{CanonicalPR: binding.PullRequest.URL, State: workflowstate.NativeReconciliationRequired, Condition: "canonical_pr_observation_deadline", PullRequest: binding.PullRequest, OriginalSourceCommit: binding.Git.SourceCommit, PublishedSourceCommit: binding.Git.ResultCommit}
}
