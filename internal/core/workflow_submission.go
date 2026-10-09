package core

import (
	"context"
	"errors"
	"time"
)

// ErrWorkflowObservationRequired retains target exclusion until a submitted
// operation is verified delivered or definitively sealed without publication.
var ErrWorkflowObservationRequired = errors.New("workflow submission requires canonical observation")

// ErrWorkflowIntegrationBudget requires an explicit operator decision after the
// per-unit, per-invocation submission budget is consumed.
var ErrWorkflowIntegrationBudget = errors.New("workflow integration submission budget exhausted")

// WorkflowSubmission freezes publication intent before any provider mutation.
// State is awaiting_observation, unpublished or confirmed; responses alone never
// release exclusion or complete tasks.
type WorkflowSubmission struct {
	ID             string                         `json:"id"`
	InvocationID   string                         `json:"invocation_id"`
	Unit           WorkflowRef                    `json:"unit"`
	Attempt        int                            `json:"attempt"`
	CandidateID    string                         `json:"candidate_id"`
	CandidateHash  string                         `json:"candidate_hash"`
	Guard          WorkflowPublicationGuard       `json:"guard"`
	PreparedLane   WorkflowLane                   `json:"prepared_lane"`
	State          string                         `json:"state"`
	SubmittedAt    time.Time                      `json:"submitted_at"`
	UpdatedAt      time.Time                      `json:"updated_at"`
	Response       string                         `json:"response,omitempty"`
	ResponseCommit string                         `json:"response_commit,omitempty"`
	ObservedPR     *WorkflowPullRequestSnapshot   `json:"observed_pr,omitempty"`
	Proof          *WorkflowIntegrationProof      `json:"proof,omitempty"`
	Unpublished    *WorkflowSubmissionObservation `json:"unpublished_observation,omitempty"`
}

// WorkflowSubmissionBinding projects the repository journal into an invocation.
// The common-directory journal remains authoritative after interrupted updates.
type WorkflowSubmissionBinding struct {
	JournalKey  string      `json:"journal_key"`
	JournalHash string      `json:"journal_hash"`
	TargetKey   string      `json:"target_key"`
	ID          string      `json:"id"`
	Unit        WorkflowRef `json:"unit"`
	Attempt     int         `json:"attempt"`
	State       string      `json:"state"`
}

// WorkflowSubmissionResponse records an untrusted command response for later
// observation. Acceptance and a result SHA are not delivery proof.
type WorkflowSubmissionResponse struct {
	Accepted     bool
	ResultCommit string
}

// WorkflowSubmissionObservation is a trusted server observation of one stable
// submission ID. Unpublished guarantees that operation cannot publish later;
// an open PR, timeout, lease expiry or missing result does not establish this.
type WorkflowSubmissionObservation struct {
	SubmissionID string `json:"submission_id"`
	Unpublished  bool   `json:"unpublished"`
}

// WorkflowPublicationProvider submits the exact prepared result with both guards
// in one server transaction. It never falls back to an unguarded merge endpoint.
type WorkflowPublicationProvider interface {
	WorkflowPublicationGuardProvider
	SubmitWorkflowIntegration(context.Context, WorkflowSubmission) (WorkflowSubmissionResponse, error)
	ObserveWorkflowSubmission(context.Context, WorkflowSubmission) (WorkflowSubmissionObservation, error)
}

// SubmitWorkflowSliceIntegration journals intent and exclusion before provider
// IO. An existing uncertain intent is observed, never submitted again blindly.
func (s *WorkflowStore) SubmitWorkflowSliceIntegration(ctx context.Context, id string, provider WorkflowPublicationProvider, git WorkflowIntegrationGit) (submission WorkflowSubmission, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	submission, submit, err := s.beginWorkflowSubmission(ctx, id, provider, git)
	if err != nil || !submit {
		if err == nil && submission.State == "awaiting_observation" {
			return s.ReconcileWorkflowSubmission(ctx, id, provider, git)
		}

		return submission, err
	}

	watched, stop := s.watchWorkflowStop(ctx, &WorkflowClaim{InvocationID: id})
	defer stop()

	response, submitErr := provider.SubmitWorkflowIntegration(watched, submission)

	cleanup, finish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finish()

	submission, err = s.retainWorkflowSubmissionResponse(cleanup, id, submission.ID, response, submitErr)

	return submission, errors.Join(submitErr, err)
}

// ReconcileWorkflowSubmission observes actual canonical delivery or a sealed
// unpublished operation without consuming another attempt or running a worker.
func (s *WorkflowStore) ReconcileWorkflowSubmission(ctx context.Context, id string, provider WorkflowPublicationProvider, git WorkflowIntegrationGit) (WorkflowSubmission, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	submission, err := s.ReadWorkflowSubmission(ctx, id)
	if err != nil {
		return submission, err
	}

	observed, err := observeWorkflowSubmission(ctx, submission, provider, git)
	if err != nil {
		return submission, err
	}

	return s.finishWorkflowSubmissionObservation(ctx, id, observed)
}

// ReadWorkflowSubmission reads the durable common-directory journal, including
// intent that survived a crash before its invocation projection was updated.
func (s *WorkflowStore) ReadWorkflowSubmission(ctx context.Context, id string) (submission WorkflowSubmission, err error) {
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
