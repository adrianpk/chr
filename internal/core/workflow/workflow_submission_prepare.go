package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s *Store) beginWorkflowSubmission(ctx context.Context, id string, provider PublicationProvider, git IntegrationGit) (submission workflowstate.Submission, submit bool, err error) {
	if provider == nil || git == nil {
		return submission, false, fmt.Errorf("guarded publisher and canonical Git observer required")
	}

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		submission, submit, readErr = owner.prepareWorkflowSubmission(ctx, &run, provider, git)

		return readErr
	})

	return submission, submit, err
}

func (s *Store) prepareWorkflowSubmission(ctx context.Context, run *workflowInvocationFile, provider PublicationProvider, git IntegrationGit) (workflowstate.Submission, bool, error) {
	journal, err := s.loadWorkflowSubmissionJournal(ctx, *run)
	if err != nil {
		return workflowstate.Submission{}, false, err
	}

	if journal.Schema == "chr-integration-journal/v2" {
		return workflowstate.Submission{}, false, ErrWorkflowObservationRequired
	}

	previous, retry, err := s.checkWorkflowSubmissionRetry(ctx, run, &journal, provider, git)
	if err != nil || !retry {
		return previous, false, err
	}

	if err := s.checkWorkflowSubmissionAdmission(run, journal); err != nil {
		return workflowstate.Submission{}, false, err
	}

	guard, err := s.checkWorkflowLegacyPublicationGuard(ctx, run, provider)
	if err != nil {
		return workflowstate.Submission{}, false, err
	}

	if journal.Key == "" {
		journal, err = newWorkflowSubmissionJournal(*run, guard)
		if err != nil {
			return workflowstate.Submission{}, false, err
		}
	}

	return s.persistWorkflowSubmissionIntent(ctx, run, journal, guard)
}

func (s *Store) loadWorkflowSubmissionJournal(ctx context.Context, run workflowInvocationFile) (workflowSubmissionJournal, error) {
	journal, err := s.readWorkflowSubmissionJournal(ctx, run)
	if errors.Is(err, os.ErrNotExist) {
		if run.Submission != nil && run.Lane != nil && run.Submission.Unit == run.Lane.Unit {
			return journal, workflowstate.ErrConflict
		}

		return workflowSubmissionJournal{}, nil
	}

	return journal, err
}

func (s *Store) checkWorkflowSubmissionRetry(ctx context.Context, run *workflowInvocationFile, journal *workflowSubmissionJournal, provider PublicationProvider, git IntegrationGit) (workflowstate.Submission, bool, error) {
	if len(journal.Attempts) == 0 {
		return workflowstate.Submission{}, true, nil
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.State == "awaiting_observation" {
		return last, false, s.projectWorkflowSubmission(run, *journal)
	}

	if !workflowSubmissionMatchesCandidate(*run, last) {
		if err := checkWorkflowSupersededSubmission(ctx, *run, last, provider); err != nil {
			return last, false, err
		}

		return last, true, s.projectWorkflowSubmission(run, *journal)
	}

	if last.State == "confirmed" && run.Submission != nil && *run.Submission == workflowSubmissionProjection(*journal) {
		return last, false, nil
	}

	return s.observeWorkflowSubmissionRetry(ctx, run, journal, provider, git)
}

func (s *Store) observeWorkflowSubmissionRetry(ctx context.Context, run *workflowInvocationFile, journal *workflowSubmissionJournal, provider PublicationProvider, git IntegrationGit) (workflowstate.Submission, bool, error) {
	last := journal.Attempts[len(journal.Attempts)-1]

	observed, err := observeWorkflowSubmission(ctx, last, provider, git)
	if err != nil {
		return last, false, err
	}

	if observed.State == "unpublished" {
		return last, true, s.projectWorkflowSubmission(run, *journal)
	}

	observed.UpdatedAt = time.Now().UTC()

	journal.Attempts[len(journal.Attempts)-1] = observed
	if err := s.writeWorkflowSubmissionJournal(ctx, journal); err != nil {
		return observed, false, err
	}

	return observed, false, s.projectWorkflowSubmission(run, *journal)
}

func (s *Store) checkWorkflowSubmissionAdmission(run *workflowInvocationFile, journal workflowSubmissionJournal) error {
	if run.StopIntent != "" || workflowInvocationTerminal(run.Status) || run.Status == "paused" {
		return ErrWorkflowStop
	}

	if len(journal.Attempts) >= workflowIntegrationAttemptLimit(run.Policy) {
		return s.blockWorkflowSubmissionBudget(run)
	}

	return nil
}

func workflowIntegrationAttemptLimit(policy workflowstate.ExecutionPolicy) int {
	if policy.MaxIntegrationAttempts == 0 {
		return 3
	}

	return policy.MaxIntegrationAttempts
}

func newWorkflowSubmissionJournal(run workflowInvocationFile, guard workflowstate.PublicationGuard) (workflowSubmissionJournal, error) {
	target, err := workflowSubmissionTargetKey(guard.PullRequest)

	return workflowSubmissionJournal{Schema: "chr-integration-journal/v1", Key: workflowSubmissionJournalKey(run), OwnerRoot: run.OwnerRoot, Namespace: run.Namespace, InvocationID: run.ID, Unit: run.Lane.Unit, TargetKey: target}, err
}

func (s *Store) persistWorkflowSubmissionIntent(ctx context.Context, run *workflowInvocationFile, journal workflowSubmissionJournal, guard workflowstate.PublicationGuard) (workflowstate.Submission, bool, error) {
	if err := s.guardWorkflowSubmissionTarget(ctx, journal.Key, journal.TargetKey); err != nil {
		return workflowstate.Submission{}, false, err
	}

	candidate, err := s.readWorkflowPreparation(*run)
	if err != nil {
		return workflowstate.Submission{}, false, err
	}

	now := time.Now().UTC()
	number := len(journal.Attempts) + 1
	submission := workflowstate.Submission{ID: workflowSubmissionID(journal.Key, number, guard.PreparationHash), InvocationID: run.ID, Unit: run.Lane.Unit, Attempt: number, CandidateID: run.PullRequest.CandidateID, CandidateHash: run.CandidateHash, Guard: guard, PreparedLane: candidate.Lane, State: "awaiting_observation", SubmittedAt: now, UpdatedAt: now}
	journal.Attempts = append(journal.Attempts, submission)
	// This single atomic image contains intent, consumed budget and exclusion.
	// The invocation projection is written afterwards; a crash cannot free it.
	if err := s.writeWorkflowSubmissionJournal(ctx, &journal); err != nil {
		return submission, false, err
	}

	return submission, true, s.projectWorkflowSubmission(run, journal)
}

func workflowSubmissionID(key string, number int, preparationHash string) string {
	return "chr-submission-" + contenthash.Sum([]byte(key+":"+strconv.Itoa(number)+":"+preparationHash))
}

func (s *Store) projectWorkflowSubmission(run *workflowInvocationFile, journal workflowSubmissionJournal) error {
	last := journal.Attempts[len(journal.Attempts)-1]

	binding := workflowSubmissionProjection(journal)
	if run.Submission != nil && *run.Submission == binding {
		return nil
	}

	run.Submission = &binding
	if run.StopIntent == "" && !workflowInvocationTerminal(run.Status) && run.Status != "paused" {
		run.Status, run.Outcome = "waiting_integration", last.State
	}

	return s.commitWorkflowState(run)
}

func (s *Store) blockWorkflowSubmissionBudget(run *workflowInvocationFile) error {
	run.Status, run.Outcome = "blocked", "integration_budget_exhausted"
	run.Blocker = &Blocker{Dependency: run.Lane.Unit, Reason: "integration_submission_budget_exhausted", NextActionOwner: "workflow_operator"}

	return errors.Join(ErrWorkflowIntegrationBudget, s.commitWorkflowState(run))
}

func workflowSubmissionProjection(journal workflowSubmissionJournal) SubmissionBinding {
	last := journal.Attempts[len(journal.Attempts)-1]

	return SubmissionBinding{JournalKey: journal.Key, JournalHash: journal.sourceHash, TargetKey: journal.TargetKey, ID: last.ID, Unit: journal.Unit, Attempt: last.Attempt, State: last.State, ExternalDelivery: last.ExternalDelivery}
}

func (s *Store) checkWorkflowLegacyPublicationGuard(ctx context.Context, run *workflowInvocationFile, provider PublicationGuardProvider) (workflowstate.PublicationGuard, error) {
	guard, err := s.checkWorkflowPublicationGuard(ctx, *run, provider)
	if err != nil {
		if errors.Is(err, ErrWorkflowAtomicGuard) {
			err = s.blockWorkflowPublication(run, err)
		}

		return guard, err
	}

	if guard.Capability.Mechanism == workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return guard, ErrWorkflowObservationRequired
	}

	return guard, nil
}
