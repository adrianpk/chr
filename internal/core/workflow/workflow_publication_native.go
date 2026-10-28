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
func advanceWorkflowNativeIntegration(ctx context.Context, store *Store, id string, ports PublicationPorts) (IntegrationProgress, error) {
	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return IntegrationProgress{}, err
	}

	if run.DeliveryRef != "" {
		receipt, err := store.FinalizeWorkflowSlice(ctx, id, nil, nil)

		return workflowIntegrationProgress(ctx, store, id, &receipt, err)
	}

	provider, ok := ports.Provider.(NativePreparationProvider)
	if !ok || ports.Preparer == nil || ports.NativePublisher == nil || ports.NativeObserver == nil {
		return workflowIntegrationProgress(ctx, store, id, nil, workflowstate.ErrAtomicGuard)
	}

	if run.Integration == nil {
		return workflowIntegrationProgress(ctx, store, id, nil, admitWorkflowNativeContinuation(ctx, store, run))
	}

	nativeContinuation, err := store.workflowNativeContinuationOwner(ctx, id)

	entry, guard := nativeContinuation.Entry, nativeContinuation.Guard
	if err == nil {
		err = continueWorkflowNativePhase(ctx, store, workflowNativeContinuation{Entry: entry, Guard: guard}, provider, ports)
	}

	return workflowIntegrationProgress(ctx, store, id, nil, err)
}

func admitWorkflowNativeContinuation(ctx context.Context, store *Store, run workflowstate.Invocation) error {
	if run.Lane == nil || run.PullRequest == nil || run.PullRequest.State != "bound" {
		return workflowstate.ErrConflict
	}

	_, err := store.AdmitWorkflowIntegration(ctx, IntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: run.Lane.Unit, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/" + run.PullRequest.Request.BaseBranch})

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

func continueWorkflowNativePhase(ctx context.Context, store *Store, owner workflowNativeContinuation, provider NativePreparationProvider, ports PublicationPorts) error {
	entry, guard := owner.Entry, owner.Guard

	switch entry.State {
	case "queued":
		return transitionWorkflowNativeContinuation(ctx, store, entry, guard, "preparing")
	case "preparing":
		return transitionWorkflowNativeContinuation(ctx, store, entry, guard, "validating")
	case "validating":
		_, err := PrepareIntegration(ctx, store, guard.InvocationID, provider, ports.Preparer)

		return err
	case "reviewing", "waiting_review":
		return reviewWorkflowNativeContinuation(ctx, store, guard, provider, ports)
	case "ready_to_publish":
		_, err := PublishNativeSubmission(ctx, store, guard, provider, ports.NativePublisher)

		return err
	case "publishing", "observing":
		return observeWorkflowNativeIntegration(ctx, store, guard, ports, workflowNativeObservationPoll)
	case "finalizing":
		_, err := store.FinalizeWorkflowSlice(ctx, guard.InvocationID, nil, nil)

		return err
	default:
		return workflowstate.ErrObservationRequired
	}
}

func reviewWorkflowNativeContinuation(ctx context.Context, store *Store, guard IntegrationBudgetGuard, provider NativePreparationProvider, ports PublicationPorts) error {
	review, err := ReviewSlice(ctx, store, guard.InvocationID, provider, ports.Reviewer)
	if err != nil || review.State != workflowstate.ReviewAuthorized {
		return err
	}

	nativeContinuation2, err := store.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	entry, guard := nativeContinuation2.Entry, nativeContinuation2.Guard

	if err != nil {
		return err
	}

	return transitionWorkflowNativeContinuation(ctx, store, entry, guard, "ready_to_publish")
}

func transitionWorkflowNativeContinuation(ctx context.Context, store *Store, entry IntegrationEntry, guard IntegrationBudgetGuard, state IntegrationState) error {
	hash := ""
	if entry.CandidateRef != nil {
		hash = entry.CandidateRef.Hash
	}

	_, err := store.TransitionWorkflowIntegration(ctx, IntegrationTransition{InvocationID: guard.InvocationID, EntryID: entry.ID, InvocationRevision: guard.InvocationRevision, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, CandidateHash: hash, State: state})

	return err
}

func observeWorkflowNativeIntegration(ctx context.Context, store *Store, guard IntegrationBudgetGuard, ports PublicationPorts, mode workflowNativeObservationMode) error {
	inspector, ok := ports.NativePublisher.(GitPublicationInspector)
	git, prepared := ports.Observer.(PreparedIntegrationGit)

	if !ok || !prepared || ports.NativeObserver == nil {
		return workflowstate.ErrAtomicGuard
	}

	if _, err := ReconcileNativeGitSubmission(ctx, store, guard, inspector, git); err != nil {
		return err
	}

	nativeContinuation3, err := store.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	guard = nativeContinuation3.Guard

	if err != nil {
		return err
	}

	if mode == workflowNativeObservationPoll {
		_, err = PollNativeSubmission(ctx, store, guard, ports.NativeObserver, ports.Observer)
	} else {
		_, err = ObserveNativeSubmission(ctx, store, guard, ports.NativeObserver, ports.Observer)
	}

	if err != nil {
		return err
	}

	nativeContinuation4, err := store.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	guard = nativeContinuation4.Guard

	if err != nil {
		return err
	}

	_, err = store.ConfirmWorkflowNativeReceipt(ctx, guard)

	return err
}

// The selected observation operation changes repetition only. Both modes retain
// the same reconciled intent, current guard, caller context and receipt checks.
type workflowNativeObservationMode uint8

const (
	workflowNativeObservationOnce workflowNativeObservationMode = iota
	workflowNativeObservationPoll
)
