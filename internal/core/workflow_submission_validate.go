package core

import (
	"fmt"
	"path/filepath"
	"slices"
)

func validateWorkflowSubmissionBinding(run workflowInvocationFile) error {
	b := run.Submission
	if b == nil {
		return nil
	}

	if run.Lane == nil || b.Unit != run.Lane.Unit || b.JournalKey != workflowSubmissionJournalKey(run) || !validWorkflowSubmissionProjection(*b) {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowSubmissionProjection(b WorkflowSubmissionBinding) bool {
	return len(b.JournalHash) == 64 && len(b.TargetKey) == 64 && b.ID != "" && b.Attempt >= 1 && b.Attempt <= 10 && validWorkflowSubmissionState(b.State) && (!b.ExternalDelivery || b.State == "confirmed")
}

func validWorkflowSubmissionState(state string) bool {
	return slices.Contains([]string{"awaiting_observation", "unpublished", "confirmed"}, state)
}

func validateWorkflowSubmissionJournal(journal workflowSubmissionJournal) error {
	if !validWorkflowSubmissionJournalHeader(journal) {
		return ErrWorkflowConflict
	}

	run := workflowInvocationFile{OwnerRoot: journal.OwnerRoot, WorkflowInvocation: WorkflowInvocation{WorkflowIdentity: WorkflowIdentity{ID: journal.InvocationID, Namespace: journal.Namespace}, Lane: &WorkflowLane{Unit: journal.Unit}}}
	if journal.Key != workflowSubmissionJournalKey(run) {
		return ErrWorkflowConflict
	}

	for i, submission := range journal.Attempts {
		if submission.Attempt != i+1 || submission.InvocationID != journal.InvocationID || submission.Unit != journal.Unit {
			return ErrWorkflowConflict
		}

		if i < len(journal.Attempts)-1 && submission.State != "unpublished" {
			return fmt.Errorf("another submission requires sealed unpublished proof")
		}

		if err := validateWorkflowSubmission(submission, journal); err != nil {
			return err
		}
	}

	return nil
}

func validWorkflowSubmissionJournalHeader(journal workflowSubmissionJournal) bool {
	return slices.Contains([]string{"chr-integration-journal/v1", "chr-integration-journal/v2"}, journal.Schema) && filepath.IsAbs(journal.OwnerRoot) && IsValidOpsNamespace(journal.Namespace) && journal.InvocationID != "" && journal.Unit.ID != "" && len(journal.Attempts) >= 1 && len(journal.Attempts) <= 10
}

func validateWorkflowSubmission(submission WorkflowSubmission, journal workflowSubmissionJournal) error {
	guard := submission.Guard
	if !validWorkflowSubmissionIdentity(submission, journal) {
		return ErrWorkflowConflict
	}

	target, err := workflowSubmissionTargetKey(guard.PullRequest)
	if err != nil || target != journal.TargetKey || !validWorkflowLane(submission.PreparedLane) {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowSubmissionGuard(guard); err != nil {
		return err
	}

	if !slices.Contains([]string{"", "accepted", "unaccepted", "error_unknown"}, submission.Response) || submission.ResponseCommit != "" && !workflowCommitID(submission.ResponseCommit) {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowNativeSubmission(submission, journal.Schema); err != nil {
		return err
	}

	return validateWorkflowSubmissionProof(submission)
}

func validWorkflowSubmissionIdentity(submission WorkflowSubmission, journal workflowSubmissionJournal) bool {
	return submission.ID == workflowSubmissionID(journal.Key, submission.Attempt, submission.Guard.PreparationHash) && submission.CandidateID != "" && len(submission.CandidateHash) == 64 && validWorkflowSubmissionState(submission.State) && !submission.SubmittedAt.IsZero() && !submission.UpdatedAt.Before(submission.SubmittedAt)
}

func validateWorkflowSubmissionGuard(guard WorkflowPublicationGuard) error {
	if guard.PreparationRef == "" || len(guard.PreparationHash) != 64 || len(guard.ReviewHash) != 64 || guard.PullRequest.HeadCommit != guard.Git.SourceCommit || guard.PullRequest.BaseCommit != guard.Git.BaseCommit || guard.PullRequest.BaseBranch != guard.Git.TargetBranch {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowReviewPolicy(guard.Policy); err != nil {
		return err
	}

	if err := validateWorkflowPublicationCapability(WorkflowReviewInput{PullRequest: guard.PullRequest, Policy: guard.Policy}, guard.Capability); err != nil {
		return err
	}

	return validateWorkflowIntegrationProof(guard.Git.WorkflowIntegrationQuery, guard.Git)
}

func validateWorkflowSubmissionProof(submission WorkflowSubmission) error {
	if submission.State != "confirmed" {
		if submission.Proof != nil || submission.ObservedPR != nil || submission.ExternalDelivery {
			return ErrWorkflowConflict
		}

		return validateWorkflowUnpublishedProof(submission)
	}

	if !validWorkflowConfirmedSubmissionHeader(submission) {
		return ErrWorkflowConflict
	}

	proof := *submission.Proof

	external := proof.ResultCommit != submission.Guard.Git.ResultCommit || proof.BaseCommit != submission.Guard.Git.BaseCommit || proof.HistoryHash != submission.Guard.Git.HistoryHash
	if external != submission.ExternalDelivery || submission.ObservedPR.ResultCommit != proof.ResultCommit {
		return ErrWorkflowConflict
	}

	query := submission.Guard.Git.WorkflowIntegrationQuery
	query.ResultCommit = proof.ResultCommit

	return validateWorkflowIntegrationProof(query, proof)
}

func validWorkflowConfirmedSubmissionHeader(submission WorkflowSubmission) bool {
	return submission.Unpublished == nil && submission.Proof != nil && submission.ObservedPR != nil && submission.ObservedPR.Merged && workflowSubmissionPRMatches(submission, *submission.ObservedPR)
}

func validateWorkflowUnpublishedProof(submission WorkflowSubmission) error {
	if submission.State == "unpublished" {
		if submission.Unpublished == nil || submission.Unpublished.SubmissionID != submission.ID || !submission.Unpublished.Unpublished {
			return ErrWorkflowConflict
		}
	} else if submission.Unpublished != nil {
		return ErrWorkflowConflict
	}

	return nil
}
