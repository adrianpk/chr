package workflow

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) retainWorkflowSubmissionResponse(ctx context.Context, id, submissionID string, response workflowstate.SubmissionResponse, submitErr error) (submission workflowstate.Submission, err error) {
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

func workflowSubmissionResponseState(response workflowstate.SubmissionResponse, err error) string {
	if err != nil {
		return "error_unknown"
	}

	if response.Accepted {
		return "accepted"
	}

	return "unaccepted"
}

func (s *Store) finishWorkflowSubmissionObservation(ctx context.Context, id string, observed workflowstate.Submission) (submission workflowstate.Submission, err error) {
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
			return workflowstate.ErrConflict
		}

		if readErr := owner.retainWorkflowSubmissionObservation(ctx, &journal, observed); readErr != nil {
			return readErr
		}

		submission = journal.Attempts[len(journal.Attempts)-1]

		return owner.projectWorkflowSubmission(&run, journal)
	})

	return submission, err
}
