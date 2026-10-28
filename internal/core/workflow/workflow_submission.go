package workflow

import (
	"context"
	"errors"
	"time"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrWorkflowObservationRequired retains target exclusion until a submitted
// operation is verified delivered or definitively sealed without publication.
var ErrWorkflowObservationRequired = errors.New("workflow submission requires canonical observation")

// ErrWorkflowIntegrationBudget requires an explicit operator decision after the
// per-unit, per-invocation submission budget is consumed.
var ErrWorkflowIntegrationBudget = errors.New("workflow integration submission budget exhausted")

// SubmissionBinding projects the repository journal into an invocation.
// The common-directory journal remains authoritative after interrupted updates.
type SubmissionBinding struct {
	JournalKey       string            `json:"journal_key"`
	JournalHash      string            `json:"journal_hash"`
	TargetKey        string            `json:"target_key"`
	ID               string            `json:"id"`
	Unit             workflowstate.Ref `json:"unit"`
	Attempt          int               `json:"attempt"`
	State            string            `json:"state"`
	ExternalDelivery bool              `json:"external_delivery,omitempty"`
}

// PublicationProvider submits the exact prepared result with both guards
// in one server transaction. It never falls back to an unguarded merge endpoint.
type PublicationProvider interface {
	PublicationGuardProvider
	SubmitWorkflowIntegration(context.Context, workflowstate.Submission) (workflowstate.SubmissionResponse, error)
	ObserveWorkflowSubmission(context.Context, workflowstate.Submission) (workflowstate.SubmissionObservation, error)
}

// SubmitWorkflowSliceIntegration journals intent and exclusion before provider
// IO. An existing uncertain intent is observed, never submitted again blindly.
func (s *Store) SubmitWorkflowSliceIntegration(ctx context.Context, id string, provider PublicationProvider, git IntegrationGit) (submission workflowstate.Submission, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	submission, submit, err := s.beginWorkflowSubmission(ctx, id, provider, git)
	if err != nil || !submit {
		if err == nil && submission.State == "awaiting_observation" {
			return s.ReconcileWorkflowSubmission(ctx, id, provider, git)
		}

		return submission, err
	}

	watched, stop := s.watchWorkflowStop(ctx, &Claim{InvocationID: id})
	defer stop()

	response, submitErr := provider.SubmitWorkflowIntegration(watched, submission)

	cleanup, finish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finish()

	var responses workflowforge.Responses = workflowSubmissionResponses{store: s, invocationID: id, submissionID: submission.ID}

	submission, err = responses.Retain(cleanup, workflowforge.Response{Reply: response, Failure: submitErr})

	return submission, errors.Join(submitErr, err)
}

// ReconcileWorkflowSubmission observes actual canonical delivery or a sealed
// unpublished operation without consuming another attempt or running a worker.
func (s *Store) ReconcileWorkflowSubmission(ctx context.Context, id string, provider PublicationProvider, git IntegrationGit) (workflowstate.Submission, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	submission, err := s.ReadWorkflowSubmission(ctx, id)
	if err != nil {
		return submission, err
	}

	if submission.Native != nil {
		return submission, ErrWorkflowObservationRequired
	}

	observed, err := observeWorkflowSubmission(ctx, submission, provider, git)
	if err != nil {
		return submission, err
	}

	return s.finishWorkflowSubmissionObservation(ctx, id, observed)
}

// ReadWorkflowSubmission reads the durable common-directory journal, including
// intent that survived a crash before its invocation projection was updated.
func (s *Store) ReadWorkflowSubmission(ctx context.Context, id string) (submission workflowstate.Submission, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		journal, readErr := owner.readWorkflowSubmissionJournal(ctx, run)
		if readErr == nil {
			submission = journal.Attempts[len(journal.Attempts)-1]
		}

		return readErr
	})

	return submission, err
}
