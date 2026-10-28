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
	Invocation workflowstate.Invocation
	Receipt    *workflowstate.Record
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
	publicationCheckpoint, err := s.observeWorkflowPublicationCheckpoint(ctx, id, ports)

	receipt, delivered := publicationCheckpoint.Receipt, publicationCheckpoint.Observed
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
	if err != nil || review.State != workflowstate.ReviewAuthorized {
		return s.workflowIntegrationProgress(ctx, id, nil, err)
	}

	_, err = s.SubmitWorkflowSliceIntegration(ctx, id, publisher, ports.Observer)

	return s.workflowIntegrationProgress(ctx, id, nil, err)
}

func (s *Store) workflowIntegrationProgress(ctx context.Context, id string, receipt *workflowstate.Record, failure error) (IntegrationProgress, error) {
	// A lost response may cancel the caller context; retain its latest checkpoint.
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	run, err := s.ReadWorkflow(cleanup, id)

	return IntegrationProgress{Invocation: run, Receipt: receipt}, errors.Join(failure, err)
}

func (s *Store) blockWorkflowMissingPublisher(ctx context.Context, id string) error {
	return s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		if invocationLocation.Run.StopIntent != "" || workflowstate.InvocationTerminal(invocationLocation.Run.Status) || invocationLocation.Run.Status == workflowstate.InvocationPaused {
			return workflowstate.ErrStop
		}

		return invocationLocation.Owner.blockWorkflowPublication(&invocationLocation.Run, workflowstate.ErrAtomicGuard)
	})
}

func (s *Store) observeWorkflowPublicationCheckpoint(ctx context.Context, id string, ports PublicationPorts) (result workflowPublicationCheckpoint, err error) {
	var (
		receipt   *workflowstate.Record
		delivered bool
	)

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation2, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		publicationCheckpoint2, readErr := invocationLocation2.Owner.reconcileWorkflowPublicationCheckpoint(ctx, invocationLocation2.Run, ports)
		receipt, delivered = publicationCheckpoint2.Receipt, publicationCheckpoint2.Observed

		return readErr
	})

	return workflowPublicationCheckpoint{Receipt: receipt, Observed: delivered}, err
}

func (s *Store) reconcileWorkflowPublicationCheckpoint(ctx context.Context, run workflowInvocationFile, ports PublicationPorts) (workflowPublicationCheckpoint, error) {
	if run.DeliveryRef != "" {
		receipt, err := s.readWorkflowSliceReceipt(ctx, run)
		if err != nil {
			return workflowPublicationCheckpoint{Receipt: nil, Observed: true}, err
		}

		return workflowPublicationCheckpoint{Receipt: &receipt, Observed: true}, nil
	}

	publicationObservation, err := s.readWorkflowPublicationCheckpoint(ctx, run, ports.Provider)
	journal, pr := publicationObservation.Journal, publicationObservation.PullRequest

	if err != nil {
		return workflowPublicationCheckpoint{Receipt: nil, Observed: false}, err
	}

	if pr.Merged {
		return s.finalizeWorkflowObservedPublication(ctx, run, journal, ports)
	}

	if err := s.reconcileWorkflowUnmergedSubmission(ctx, &run, journal, ports); err != nil {
		return workflowPublicationCheckpoint{Receipt: nil, Observed: false}, err
	}

	if run.StopIntent != "" || workflowstate.InvocationTerminal(run.Status) || run.Status == workflowstate.InvocationPaused {
		return workflowPublicationCheckpoint{Receipt: nil, Observed: false}, workflowstate.ErrStop
	}

	return workflowPublicationCheckpoint{Receipt: nil, Observed: false}, nil
}

func (s *Store) finalizeWorkflowObservedPublication(ctx context.Context, run workflowInvocationFile, journal workflowSubmissionJournal, ports PublicationPorts) (workflowPublicationCheckpoint, error) {
	policy, err := s.workflowReviewPolicy(run)
	if len(journal.Attempts) > 0 && workflowSubmissionMatchesCandidate(run, journal.Attempts[len(journal.Attempts)-1]) {
		policy, err = journal.Attempts[len(journal.Attempts)-1].Guard.Policy, nil
	}

	if err != nil {
		return workflowPublicationCheckpoint{Receipt: nil, Observed: true}, err
	}

	if _, err := s.verifyWorkflowSliceIntegration(ctx, run, policy.Method, ports.Provider, ports.Observer); err != nil {
		return workflowPublicationCheckpoint{Receipt: nil, Observed: true}, err
	}

	run, err = s.readWorkflowInvocation(run.ID)
	if err != nil {
		return workflowPublicationCheckpoint{Receipt: nil, Observed: true}, err
	}

	receipt, err := s.finalizeWorkflowSlice(ctx, run, ports.Provider, ports.Observer)
	if err != nil {
		return workflowPublicationCheckpoint{Receipt: nil, Observed: true}, err
	}

	return workflowPublicationCheckpoint{Receipt: &receipt, Observed: true}, nil
}

func (s *Store) reconcileWorkflowUnmergedSubmission(ctx context.Context, run *workflowInvocationFile, journal workflowSubmissionJournal, ports PublicationPorts) error {
	if len(journal.Attempts) == 0 {
		return nil
	}

	publisher, ok := ports.Provider.(PublicationProvider)
	if !ok {
		return workflowstate.ErrObservationRequired
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
	if observed.State == workflowstate.SubmissionConfirmed {
		return workflowstate.ErrObservationRequired
	}

	return nil
}

func (s *Store) readWorkflowPublicationCheckpoint(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider) (workflowPublicationObservation, error) {
	if run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.URL == "" {
		return workflowPublicationObservation{Journal: workflowSubmissionJournal{}, PullRequest: workflowstate.PullRequestSnapshot{}}, fmt.Errorf("publication requires a bound canonical integration-unit PR")
	}

	journal, err := s.loadWorkflowSubmissionJournal(ctx, run)
	if err != nil {
		return workflowPublicationObservation{Journal: journal, PullRequest: workflowstate.PullRequestSnapshot{}}, err
	}

	pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
	if err != nil {
		return workflowPublicationObservation{Journal: journal, PullRequest: pr}, err
	}

	if reason := workflowPullRequestMismatch(*run.PullRequest, []workflowstate.PullRequestSnapshot{pr}); reason != "" {
		return workflowPublicationObservation{Journal: journal, PullRequest: pr}, fmt.Errorf("%w: %s", workflowstate.ErrConflict, reason)
	}

	return workflowPublicationObservation{Journal: journal, PullRequest: pr}, nil
}
