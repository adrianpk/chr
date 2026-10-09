package core

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// WorkflowPublicationPorts supplies trusted canonical IO and the separate role.
// Read-only delivery reconciliation does not require a guarded publisher.
type WorkflowPublicationPorts struct {
	Provider WorkflowPullRequestProvider
	Preparer WorkflowPreparationGit
	Observer WorkflowIntegrationGit
	Reviewer []string
}

// WorkflowIntegrationProgress exposes the retained barrier or shared receipt.
type WorkflowIntegrationProgress struct {
	Invocation WorkflowInvocation
	Receipt    *WorkflowRecord
}

// AdvanceWorkflowSliceIntegration observes before preparing, reviewing or
// submitting. One call processes only the retained slice and performs at most
// one combined gate, one review and one submission. It never starts task work.
func (s *WorkflowStore) AdvanceWorkflowSliceIntegration(ctx context.Context, id string, ports WorkflowPublicationPorts) (WorkflowIntegrationProgress, error) {
	if ports.Provider == nil || ports.Observer == nil {
		return WorkflowIntegrationProgress{}, fmt.Errorf("canonical provider and Git observer required")
	}

	receipt, delivered, err := s.observeWorkflowPublicationCheckpoint(ctx, id, ports)
	if err != nil || delivered {
		return s.workflowIntegrationProgress(ctx, id, receipt, err)
	}

	publisher, ok := ports.Provider.(WorkflowPublicationProvider)
	if !ok || ports.Preparer == nil {
		return s.workflowIntegrationProgress(ctx, id, nil, s.blockWorkflowMissingPublisher(ctx, id))
	}

	if _, err := s.PrepareWorkflowSliceIntegration(ctx, id, publisher, ports.Preparer); err != nil {
		return s.workflowIntegrationProgress(ctx, id, nil, err)
	}

	review, err := s.ReviewWorkflowSlice(ctx, id, publisher, ports.Reviewer)
	if err != nil || review.State != "authorized" {
		return s.workflowIntegrationProgress(ctx, id, nil, err)
	}

	_, err = s.SubmitWorkflowSliceIntegration(ctx, id, publisher, ports.Observer)

	return s.workflowIntegrationProgress(ctx, id, nil, err)
}

func (s *WorkflowStore) workflowIntegrationProgress(ctx context.Context, id string, receipt *WorkflowRecord, failure error) (WorkflowIntegrationProgress, error) {
	// A lost response may cancel the caller context; retain its latest checkpoint.
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	run, err := s.ReadWorkflow(cleanup, id)

	return WorkflowIntegrationProgress{Invocation: run, Receipt: receipt}, errors.Join(failure, err)
}

func (s *WorkflowStore) blockWorkflowMissingPublisher(ctx context.Context, id string) error {
	return s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		if run.StopIntent != "" || workflowInvocationTerminal(run.Status) || run.Status == "paused" {
			return ErrWorkflowStop
		}

		return owner.blockWorkflowPublication(&run, ErrWorkflowAtomicGuard)
	})
}

func (s *WorkflowStore) observeWorkflowPublicationCheckpoint(ctx context.Context, id string, ports WorkflowPublicationPorts) (receipt *WorkflowRecord, delivered bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		receipt, delivered, readErr = owner.reconcileWorkflowPublicationCheckpoint(ctx, run, ports)

		return readErr
	})

	return receipt, delivered, err
}

func (s *WorkflowStore) reconcileWorkflowPublicationCheckpoint(ctx context.Context, run workflowInvocationFile, ports WorkflowPublicationPorts) (*WorkflowRecord, bool, error) {
	if run.DeliveryRef != "" {
		receipt, err := s.readWorkflowSliceReceipt(ctx, run)
		if err != nil {
			return nil, true, err
		}

		return &receipt, true, nil
	}

	journal, pr, err := s.readWorkflowPublicationCheckpoint(ctx, run, ports.Provider)
	if err != nil {
		return nil, false, err
	}

	if pr.Merged {
		return s.finalizeWorkflowObservedPublication(ctx, run, journal, ports)
	}

	if err := s.reconcileWorkflowUnmergedSubmission(ctx, &run, journal, ports); err != nil {
		return nil, false, err
	}

	if run.StopIntent != "" || workflowInvocationTerminal(run.Status) || run.Status == "paused" {
		return nil, false, ErrWorkflowStop
	}

	return nil, false, nil
}

func (s *WorkflowStore) finalizeWorkflowObservedPublication(ctx context.Context, run workflowInvocationFile, journal workflowSubmissionJournal, ports WorkflowPublicationPorts) (*WorkflowRecord, bool, error) {
	policy, err := s.workflowReviewPolicy(run)
	if len(journal.Attempts) > 0 && workflowSubmissionMatchesCandidate(run, journal.Attempts[len(journal.Attempts)-1]) {
		policy, err = journal.Attempts[len(journal.Attempts)-1].Guard.Policy, nil
	}

	if err != nil {
		return nil, true, err
	}

	if _, err := s.verifyWorkflowSliceIntegration(ctx, run, policy.Method, ports.Provider, ports.Observer); err != nil {
		return nil, true, err
	}

	run, err = s.readWorkflowInvocation(run.ID)
	if err != nil {
		return nil, true, err
	}

	receipt, err := s.finalizeWorkflowSlice(ctx, run, ports.Provider, ports.Observer)
	if err != nil {
		return nil, true, err
	}

	return &receipt, true, nil
}

func (s *WorkflowStore) reconcileWorkflowUnmergedSubmission(ctx context.Context, run *workflowInvocationFile, journal workflowSubmissionJournal, ports WorkflowPublicationPorts) error {
	if len(journal.Attempts) == 0 {
		return nil
	}

	publisher, ok := ports.Provider.(WorkflowPublicationProvider)
	if !ok {
		return ErrWorkflowObservationRequired
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if !workflowSubmissionMatchesCandidate(*run, last) {
		if err := checkWorkflowSupersededSubmission(ctx, *run, last, publisher); err != nil {
			return err
		}

		return s.projectWorkflowSubmission(run, journal)
	}

	observed, err := observeWorkflowSubmission(ctx, last, publisher, ports.Observer)
	if err != nil {
		return err
	}

	if err := s.retainWorkflowSubmissionObservation(ctx, &journal, observed); err != nil {
		return err
	}

	if err := s.projectWorkflowSubmission(run, journal); err != nil {
		return err
	}

	// Delivery can occur while the unpublished observation is being checked.
	if observed.State == "confirmed" {
		return ErrWorkflowObservationRequired
	}

	return nil
}

func (s *WorkflowStore) readWorkflowPublicationCheckpoint(ctx context.Context, run workflowInvocationFile, provider WorkflowPullRequestProvider) (workflowSubmissionJournal, WorkflowPullRequestSnapshot, error) {
	if run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.URL == "" {
		return workflowSubmissionJournal{}, WorkflowPullRequestSnapshot{}, fmt.Errorf("publication requires a bound canonical slice PR")
	}

	journal, err := s.loadWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return journal, WorkflowPullRequestSnapshot{}, err
	}

	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return journal, pr, err
	}

	if reason := workflowPullRequestMismatch(*run.PullRequest, []WorkflowPullRequestSnapshot{pr}); reason != "" {
		return journal, pr, fmt.Errorf("%w: %s", ErrWorkflowConflict, reason)
	}

	return journal, pr, nil
}
