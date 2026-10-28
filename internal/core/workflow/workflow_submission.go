package workflow

import (
	"context"
	"errors"
	"time"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrWorkflowIntegrationBudget requires an explicit operator decision after the
// per-unit, per-invocation submission budget is consumed.
var ErrWorkflowIntegrationBudget = errors.New("workflow integration submission budget exhausted")

// PublicationProvider submits the exact prepared result with both guards
// in one server transaction. It never falls back to an unguarded merge endpoint.
type PublicationProvider interface {
	PublicationGuardProvider
	workflowforge.Submitter
	ObserveWorkflowSubmission(context.Context, workflowstate.Submission) (workflowstate.SubmissionObservation, error)
}

// SubmitIntegration journals intent and exclusion before provider
// IO. An existing uncertain intent is observed, never submitted again blindly.
func SubmitIntegration(ctx context.Context, store *Store, id string, provider PublicationProvider, git IntegrationGit) (submission workflowstate.Submission, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	submissionIntent, err := store.beginWorkflowSubmission(ctx, id, provider, git)

	submission, submit := submissionIntent.Submission, submissionIntent.Dispatch
	if err != nil || !submit {
		if err == nil && submission.State == workflowstate.SubmissionAwaitingObservation {
			return ReconcileSubmission(ctx, store, id, provider, git)
		}

		return submission, err
	}

	watched, stop := context.WithCancel(ctx)

	watchWorkflowStop(watched, store, workflowStopWatch{Claim: &Claim{InvocationID: id}, Cancel: stop})
	defer stop()

	responses := workflowSubmissionResponses{store: store, invocationID: id, submissionID: submission.ID}

	return workflowforge.Submit(ctx, watched, submission, provider, responses)
}

// ReconcileSubmission observes actual canonical delivery or a sealed
// unpublished operation without consuming another attempt or running a worker.
func ReconcileSubmission(ctx context.Context, store *Store, id string, provider PublicationProvider, git IntegrationGit) (workflowstate.Submission, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	submission, err := store.ReadWorkflowSubmission(ctx, id)
	if err != nil {
		return submission, err
	}

	if submission.Native != nil {
		return submission, workflowstate.ErrObservationRequired
	}

	observed, err := observeWorkflowSubmission(ctx, submission, provider, git)
	if err != nil {
		return submission, err
	}

	return store.finishWorkflowSubmissionObservation(ctx, id, observed)
}

// ReadWorkflowSubmission reads the durable common-directory journal, including
// intent that survived a crash before its invocation projection was updated.
func (s *Store) ReadWorkflowSubmission(ctx context.Context, id string) (submission workflowstate.Submission, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		journal, readErr := invocationLocation.Owner.readWorkflowSubmissionJournal(ctx, invocationLocation.Run)
		if readErr == nil {
			submission = journal.Attempts[len(journal.Attempts)-1]
		}

		return readErr
	})

	return submission, err
}
