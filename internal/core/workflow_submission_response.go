package core

import (
	"context"
	"time"
)

func (s *WorkflowStore) retainWorkflowSubmissionResponse(ctx context.Context, id, submissionID string, response WorkflowSubmissionResponse, submitErr error) (submission WorkflowSubmission, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		journal, readErr := owner.readWorkflowSubmissionJournal(ctx, run)
		if readErr != nil {
			return readErr
		}

		submission = journal.Attempts[len(journal.Attempts)-1]
		if submission.ID != submissionID || submission.State != "awaiting_observation" {
			return nil // A concurrent definitive observation wins over a late response.
		}

		submission.Response, submission.UpdatedAt = workflowSubmissionResponseState(response, submitErr), time.Now().UTC()
		if workflowCommitID(response.ResultCommit) {
			submission.ResponseCommit = response.ResultCommit
		}

		journal.Attempts[len(journal.Attempts)-1] = submission
		if readErr := owner.writeWorkflowSubmissionJournal(ctx, &journal); readErr != nil {
			return readErr
		}

		return owner.projectWorkflowSubmission(&run, journal)
	})

	return submission, err
}

func workflowSubmissionResponseState(response WorkflowSubmissionResponse, err error) string {
	if err != nil {
		return "error_unknown"
	}

	if response.Accepted {
		return "accepted"
	}

	return "unaccepted"
}

func (s *WorkflowStore) finishWorkflowSubmissionObservation(ctx context.Context, id string, observed WorkflowSubmission) (submission WorkflowSubmission, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		journal, readErr := owner.readWorkflowSubmissionJournal(ctx, run)
		if readErr != nil {
			return readErr
		}

		submission = journal.Attempts[len(journal.Attempts)-1]
		if submission.ID != observed.ID {
			return ErrWorkflowConflict
		}

		if submission.State == "awaiting_observation" {
			submission.State, submission.ObservedPR, submission.Proof = observed.State, observed.ObservedPR, observed.Proof
			submission.Unpublished = observed.Unpublished
			submission.UpdatedAt = time.Now().UTC()

			journal.Attempts[len(journal.Attempts)-1] = submission
			if readErr := owner.writeWorkflowSubmissionJournal(ctx, &journal); readErr != nil {
				return readErr
			}
		}

		return owner.projectWorkflowSubmission(&run, journal)
	})

	return submission, err
}
