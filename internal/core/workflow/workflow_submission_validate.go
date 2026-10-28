package workflow

import (
	"fmt"
	"path/filepath"
	"slices"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowSubmissionBinding(run workflowInvocationFile) error {
	b := run.Submission
	if b == nil {
		return nil
	}

	if run.Lane == nil || b.Unit != run.Lane.Unit || b.JournalKey != workflowSubmissionJournalKey(run) || !validWorkflowSubmissionProjection(*b) {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowSubmissionProjection(b SubmissionBinding) bool {
	return len(b.JournalHash) == 64 && len(b.TargetKey) == 64 && b.ID != "" && b.Attempt >= 1 && b.Attempt <= 10 && validWorkflowSubmissionState(b.State) && (!b.ExternalDelivery || b.State == "confirmed")
}

func validWorkflowSubmissionState(state string) bool {
	return slices.Contains([]string{"awaiting_observation", "unpublished", "confirmed"}, state)
}

func validateWorkflowSubmissionJournal(journal workflowSubmissionJournal) error {
	if !validWorkflowSubmissionJournalHeader(journal) {
		return workflowstate.ErrConflict
	}

	run := workflowInvocationFile{OwnerRoot: journal.OwnerRoot, Invocation: Invocation{Identity: Identity{ID: journal.InvocationID, Namespace: journal.Namespace}, InvocationExecution: InvocationExecution{Lane: &Lane{Unit: journal.Unit}}}}
	if journal.Key != workflowSubmissionJournalKey(run) {
		return workflowstate.ErrConflict
	}

	for i, submission := range journal.Attempts {
		if submission.Attempt != i+1 || submission.InvocationID != journal.InvocationID || submission.Unit != journal.Unit {
			return workflowstate.ErrConflict
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
	return slices.Contains([]string{"chr-integration-journal/v1", "chr-integration-journal/v2"}, journal.Schema) && filepath.IsAbs(journal.OwnerRoot) && opsnamespace.Valid(journal.Namespace) && journal.InvocationID != "" && journal.Unit.ID != "" && len(journal.Attempts) >= 1 && len(journal.Attempts) <= 10
}

func validateWorkflowSubmission(submission Submission, journal workflowSubmissionJournal) error {
	guard := submission.Guard
	if !validWorkflowSubmissionIdentity(submission, journal) {
		return workflowstate.ErrConflict
	}

	target, err := workflowSubmissionTargetKey(guard.PullRequest)
	if err != nil || target != journal.TargetKey || !validWorkflowLane(submission.PreparedLane) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowSubmissionGuard(guard); err != nil {
		return err
	}

	if !slices.Contains([]string{"", "accepted", "unaccepted", "error_unknown"}, submission.Response) || submission.ResponseCommit != "" && !workflowCommitID(submission.ResponseCommit) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeSubmission(submission, journal.Schema); err != nil {
		return err
	}

	return validateWorkflowSubmissionProof(submission)
}

func validWorkflowSubmissionIdentity(submission Submission, journal workflowSubmissionJournal) bool {
	return submission.ID == workflowSubmissionID(journal.Key, submission.Attempt, submission.Guard.PreparationHash) && submission.CandidateID != "" && len(submission.CandidateHash) == 64 && validWorkflowSubmissionState(submission.State) && !submission.SubmittedAt.IsZero() && !submission.UpdatedAt.Before(submission.SubmittedAt)
}

func validateWorkflowSubmissionGuard(guard PublicationGuard) error {
	if guard.PreparationRef == "" || len(guard.PreparationHash) != 64 || len(guard.ReviewHash) != 64 || guard.PullRequest.HeadCommit != guard.Git.SourceCommit || guard.PullRequest.BaseCommit != guard.Git.BaseCommit || guard.PullRequest.BaseBranch != guard.Git.TargetBranch {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowReviewPolicy(guard.Policy); err != nil {
		return err
	}

	if err := validateWorkflowPublicationCapability(ReviewInput{PullRequest: guard.PullRequest, Policy: guard.Policy}, guard.Capability); err != nil {
		return err
	}

	if err := checkWorkflowPreparationSigning(guard.Git, guard.Capability.RequiredSignatures, guard.Capability.SigningProfile); err != nil {
		return err
	}

	return validateWorkflowIntegrationProof(guard.Git.IntegrationQuery, guard.Git)
}

func validateWorkflowSubmissionProof(submission Submission) error {
	if submission.State != "confirmed" {
		if submission.Proof != nil || submission.ObservedPR != nil || submission.ExternalDelivery {
			return workflowstate.ErrConflict
		}

		return validateWorkflowUnpublishedProof(submission)
	}

	if !validWorkflowConfirmedSubmissionHeader(submission) {
		return workflowstate.ErrConflict
	}

	proof := *submission.Proof

	external := proof.ResultCommit != submission.Guard.Git.ResultCommit || proof.BaseCommit != submission.Guard.Git.BaseCommit || proof.HistoryHash != submission.Guard.Git.HistoryHash
	if external != submission.ExternalDelivery || submission.ObservedPR.ResultCommit != proof.ResultCommit {
		return workflowstate.ErrConflict
	}

	query := submission.Guard.Git.IntegrationQuery
	query.ResultCommit = proof.ResultCommit

	return validateWorkflowIntegrationProof(query, proof)
}

func validWorkflowConfirmedSubmissionHeader(submission Submission) bool {
	return submission.Unpublished == nil && submission.Proof != nil && submission.ObservedPR != nil && submission.ObservedPR.Merged && workflowSubmissionPRMatches(submission, *submission.ObservedPR)
}

func validateWorkflowUnpublishedProof(submission Submission) error {
	if submission.State == "unpublished" {
		if submission.Unpublished == nil || submission.Unpublished.SubmissionID != submission.ID || !submission.Unpublished.Unpublished {
			return workflowstate.ErrConflict
		}
	} else if submission.Unpublished != nil {
		return workflowstate.ErrConflict
	}

	return nil
}
