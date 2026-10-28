package workflow

import (
	"context"
	"time"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) retainWorkflowSubmissionResponse(ctx context.Context, operation workflowSubmissionResponses, reply workflowforge.Response) (submission workflowstate.Submission, err error) {
	id, submissionID := operation.invocationID, operation.submissionID
	response, submitErr := reply.Reply, reply.Failure

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		journal, readErr := invocationLocation.Owner.readWorkflowSubmissionJournal(ctx, invocationLocation.Run)
		if readErr != nil {
			return readErr
		}

		submission = journal.Attempts[len(journal.Attempts)-1]
		if submission.ID != submissionID || submission.State != workflowstate.SubmissionAwaitingObservation {
			return nil // A concurrent definitive observation wins over a late response.
		}

		submission.Response, submission.UpdatedAt = workflowforge.ResponseState(response, submitErr), time.Now().UTC()
		if workflowstate.ValidCommitID(response.ResultCommit) {
			submission.ResponseCommit = response.ResultCommit
		}

		journal.Attempts[len(journal.Attempts)-1] = submission
		if readErr := invocationLocation.Owner.writeWorkflowSubmissionJournal(ctx, &journal); readErr != nil {
			return readErr
		}

		return invocationLocation.Owner.projectWorkflowSubmission(&invocationLocation.Run, journal)
	})

	return submission, err
}

func (s *Store) finishWorkflowSubmissionObservation(ctx context.Context, id string, observed workflowstate.Submission) (submission workflowstate.Submission, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation2, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		journal, readErr := invocationLocation2.Owner.readWorkflowSubmissionJournal(ctx, invocationLocation2.Run)
		if readErr != nil {
			return readErr
		}

		submission = journal.Attempts[len(journal.Attempts)-1]
		if submission.ID != observed.ID {
			return workflowstate.ErrConflict
		}

		if readErr := invocationLocation2.Owner.retainWorkflowSubmissionObservation(ctx, &journal, observed); readErr != nil {
			return readErr
		}

		submission = journal.Attempts[len(journal.Attempts)-1]

		return invocationLocation2.Owner.projectWorkflowSubmission(&invocationLocation2.Run, journal)
	})

	return submission, err
}
