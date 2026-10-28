package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s *Store) workflowNativeContinuationProfile(ctx context.Context, id string) (native bool, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		config, err := workspaceconfig.Load(invocationLocation.Owner.layout.ProjectDir())
		native = invocationLocation.Run.Integration != nil || config.Work.Integration.PublicationProfile == workspaceconfig.WorkflowPublicationGitAtomicDualRef

		return err
	})

	return native, err
}

// Continue one retained phase through the same guarded owners used by explicit
// operations. An uncertain submission always enters observation, never dispatch.
func (s *Store) advanceWorkflowNativeIntegration(ctx context.Context, id string, ports PublicationPorts) (IntegrationProgress, error) {
	run, err := s.ReadWorkflow(ctx, id)
	if err != nil {
		return IntegrationProgress{}, err
	}

	if run.DeliveryRef != "" {
		receipt, err := s.FinalizeWorkflowSlice(ctx, id, nil, nil)

		return s.workflowIntegrationProgress(ctx, id, &receipt, err)
	}

	provider, ok := ports.Provider.(NativePreparationProvider)
	if !ok || ports.Preparer == nil || ports.NativePublisher == nil || ports.NativeObserver == nil {
		return s.workflowIntegrationProgress(ctx, id, nil, workflowstate.ErrAtomicGuard)
	}

	if run.Integration == nil {
		return s.workflowIntegrationProgress(ctx, id, nil, s.admitWorkflowNativeContinuation(ctx, run))
	}

	nativeContinuation, err := s.workflowNativeContinuationOwner(ctx, id)

	entry, guard := nativeContinuation.Entry, nativeContinuation.Guard
	if err == nil {
		err = s.continueWorkflowNativePhase(ctx, workflowNativeContinuation{Entry: entry, Guard: guard}, provider, ports)
	}

	return s.workflowIntegrationProgress(ctx, id, nil, err)
}

func (s *Store) admitWorkflowNativeContinuation(ctx context.Context, run Invocation) error {
	if run.Lane == nil || run.PullRequest == nil || run.PullRequest.State != "bound" {
		return workflowstate.ErrConflict
	}

	_, err := s.AdmitWorkflowIntegration(ctx, IntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: run.Lane.Unit, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/" + run.PullRequest.Request.BaseBranch})

	return err
}

func (s *Store) workflowNativeContinuationOwner(ctx context.Context, id string) (result workflowNativeContinuation, err error) {
	var (
		entry IntegrationEntry
		guard IntegrationBudgetGuard
	)

	err = s.withLock(ctx, func() error {
		invocationLocation2, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		if invocationLocation2.Run.Integration == nil || invocationLocation2.Run.IntegrationBudget == nil {
			return workflowstate.ErrConflict
		}

		entry, err = invocationLocation2.Owner.readWorkflowIntegrationEntry(invocationLocation2.Run, *invocationLocation2.Run.Integration)
		guard = IntegrationBudgetGuard{InvocationID: id, InvocationRevision: invocationLocation2.Run.Revision, BudgetRevision: invocationLocation2.Run.IntegrationBudget.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration}

		return err
	})

	return workflowNativeContinuation{Entry: entry, Guard: guard}, err
}

func (s *Store) continueWorkflowNativePhase(ctx context.Context, owner workflowNativeContinuation, provider NativePreparationProvider, ports PublicationPorts) error {
	entry, guard := owner.Entry, owner.Guard

	switch entry.State {
	case "queued":
		return s.transitionWorkflowNativeContinuation(ctx, entry, guard, "preparing")
	case "preparing":
		return s.transitionWorkflowNativeContinuation(ctx, entry, guard, "validating")
	case "validating":
		_, err := s.PrepareWorkflowSliceIntegration(ctx, guard.InvocationID, provider, ports.Preparer)

		return err
	case "reviewing", "waiting_review":
		return s.reviewWorkflowNativeContinuation(ctx, guard, provider, ports)
	case "ready_to_publish":
		_, err := s.PublishWorkflowNativeSubmission(ctx, guard, provider, ports.NativePublisher)

		return err
	case "publishing", "observing":
		return s.observeWorkflowNativeContinuation(ctx, guard, ports)
	case "finalizing":
		_, err := s.FinalizeWorkflowSlice(ctx, guard.InvocationID, nil, nil)

		return err
	default:
		return workflowstate.ErrObservationRequired
	}
}

func (s *Store) reviewWorkflowNativeContinuation(ctx context.Context, guard IntegrationBudgetGuard, provider NativePreparationProvider, ports PublicationPorts) error {
	review, err := s.ReviewWorkflowSlice(ctx, guard.InvocationID, provider, ports.Reviewer)
	if err != nil || review.State != workflowstate.ReviewAuthorized {
		return err
	}

	nativeContinuation2, err := s.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	entry, guard := nativeContinuation2.Entry, nativeContinuation2.Guard

	if err != nil {
		return err
	}

	return s.transitionWorkflowNativeContinuation(ctx, entry, guard, "ready_to_publish")
}

func (s *Store) transitionWorkflowNativeContinuation(ctx context.Context, entry IntegrationEntry, guard IntegrationBudgetGuard, state IntegrationState) error {
	hash := ""
	if entry.CandidateRef != nil {
		hash = entry.CandidateRef.Hash
	}

	_, err := s.TransitionWorkflowIntegration(ctx, IntegrationTransition{InvocationID: guard.InvocationID, EntryID: entry.ID, InvocationRevision: guard.InvocationRevision, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, CandidateHash: hash, State: state})

	return err
}

func (s *Store) observeWorkflowNativeContinuation(ctx context.Context, guard IntegrationBudgetGuard, ports PublicationPorts) error {
	return s.observeWorkflowNativeIntegration(ctx, guard, ports, workflowNativeObservationPoll)
}

func (s *Store) observeWorkflowNativeIntegration(ctx context.Context, guard IntegrationBudgetGuard, ports PublicationPorts, mode workflowNativeObservationMode) error {
	inspector, ok := ports.NativePublisher.(GitPublicationInspector)
	git, prepared := ports.Observer.(PreparedIntegrationGit)

	if !ok || !prepared || ports.NativeObserver == nil {
		return workflowstate.ErrAtomicGuard
	}

	if _, err := s.ReconcileWorkflowNativeGitSubmission(ctx, guard, inspector, git); err != nil {
		return err
	}

	nativeContinuation3, err := s.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	guard = nativeContinuation3.Guard

	if err != nil {
		return err
	}

	if mode == workflowNativeObservationPoll {
		_, err = s.PollWorkflowNativeSubmission(ctx, guard, ports.NativeObserver, ports.Observer)
	} else {
		_, err = s.ObserveWorkflowNativeSubmission(ctx, guard, ports.NativeObserver, ports.Observer)
	}

	if err != nil {
		return err
	}

	nativeContinuation4, err := s.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	guard = nativeContinuation4.Guard

	if err != nil {
		return err
	}

	_, err = s.ConfirmWorkflowNativeReceipt(ctx, guard)

	return err
}

// The selected observation operation changes repetition only. Both modes retain
// the same reconciled intent, current guard, caller context and receipt checks.
type workflowNativeObservationMode uint8

const (
	workflowNativeObservationOnce workflowNativeObservationMode = iota
	workflowNativeObservationPoll
)
