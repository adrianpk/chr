package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PublicationPorts supplies trusted canonical IO and the separate role.
// Read-only delivery reconciliation does not require a guarded publisher.
type PublicationPorts struct {
	Provider        PullRequestProvider
	Preparer        PreparationGit
	Observer        IntegrationGit
	Reviewer        []string
	NativePublisher GitPublisher
	NativeObserver  NativePublicationObserver
}

// IntegrationProgress exposes the retained barrier or shared receipt.
type IntegrationProgress struct {
	Invocation Invocation
	Receipt    *Record
}

// AdvanceWorkflowSliceIntegration observes before preparing, reviewing or
// submitting. One call processes only the retained slice and performs at most
// one combined gate, one review and one submission. It never starts task work.
func (s *Store) AdvanceWorkflowSliceIntegration(ctx context.Context, id string, ports PublicationPorts) (IntegrationProgress, error) {
	if ports.Provider == nil || ports.Observer == nil {
		return IntegrationProgress{}, fmt.Errorf("canonical provider and Git observer required")
	}

	native, err := s.workflowNativeContinuationProfile(ctx, id)
	if err != nil {
		return s.workflowIntegrationProgress(ctx, id, nil, err)
	}

	if native {
		return s.advanceWorkflowNativeIntegration(ctx, id, ports)
	}

	return s.advanceWorkflowLegacyIntegration(ctx, id, ports)
}

func (s *Store) advanceWorkflowLegacyIntegration(ctx context.Context, id string, ports PublicationPorts) (IntegrationProgress, error) {
	receipt, delivered, err := s.observeWorkflowPublicationCheckpoint(ctx, id, ports)
	if err != nil || delivered {
		return s.workflowIntegrationProgress(ctx, id, receipt, err)
	}

	publisher, ok := ports.Provider.(PublicationProvider)
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

func (s *Store) workflowIntegrationProgress(ctx context.Context, id string, receipt *Record, failure error) (IntegrationProgress, error) {
	// A lost response may cancel the caller context; retain its latest checkpoint.
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	run, err := s.ReadWorkflow(cleanup, id)

	return IntegrationProgress{Invocation: run, Receipt: receipt}, errors.Join(failure, err)
}

func (s *Store) blockWorkflowMissingPublisher(ctx context.Context, id string) error {
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

func (s *Store) observeWorkflowPublicationCheckpoint(ctx context.Context, id string, ports PublicationPorts) (receipt *Record, delivered bool, err error) {
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

func (s *Store) reconcileWorkflowPublicationCheckpoint(ctx context.Context, run workflowInvocationFile, ports PublicationPorts) (*Record, bool, error) {
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

func (s *Store) finalizeWorkflowObservedPublication(ctx context.Context, run workflowInvocationFile, journal workflowSubmissionJournal, ports PublicationPorts) (*Record, bool, error) {
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

func (s *Store) reconcileWorkflowUnmergedSubmission(ctx context.Context, run *workflowInvocationFile, journal workflowSubmissionJournal, ports PublicationPorts) error {
	if len(journal.Attempts) == 0 {
		return nil
	}

	publisher, ok := ports.Provider.(PublicationProvider)
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

func (s *Store) readWorkflowPublicationCheckpoint(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider) (workflowSubmissionJournal, PullRequestSnapshot, error) {
	if run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.URL == "" {
		return workflowSubmissionJournal{}, PullRequestSnapshot{}, fmt.Errorf("publication requires a bound canonical integration-unit PR")
	}

	journal, err := s.loadWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return journal, PullRequestSnapshot{}, err
	}

	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return journal, pr, err
	}

	if reason := workflowPullRequestMismatch(*run.PullRequest, []PullRequestSnapshot{pr}); reason != "" {
		return journal, pr, fmt.Errorf("%w: %s", workflowstate.ErrConflict, reason)
	}

	return journal, pr, nil
}
