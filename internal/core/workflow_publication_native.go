package core

import "context"

func (s *WorkflowStore) workflowNativeContinuationProfile(ctx context.Context, id string) (native bool, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		config, err := LoadWorkspaceConfig(owner.layout.ProjectDir())
		native = run.Integration != nil || config.Work.Integration.PublicationProfile == WorkflowPublicationGitAtomicDualRef

		return err
	})

	return native, err
}

// Continue one retained phase through the same guarded owners used by explicit
// operations. An uncertain submission always enters observation, never dispatch.
func (s *WorkflowStore) advanceWorkflowNativeIntegration(ctx context.Context, id string, ports WorkflowPublicationPorts) (WorkflowIntegrationProgress, error) {
	run, err := s.ReadWorkflow(ctx, id)
	if err != nil {
		return WorkflowIntegrationProgress{}, err
	}

	if run.DeliveryRef != "" {
		receipt, err := s.FinalizeWorkflowSlice(ctx, id, nil, nil)

		return s.workflowIntegrationProgress(ctx, id, &receipt, err)
	}

	provider, ok := ports.Provider.(WorkflowNativePreparationProvider)
	if !ok || ports.Preparer == nil || ports.NativePublisher == nil || ports.NativeObserver == nil {
		return s.workflowIntegrationProgress(ctx, id, nil, ErrWorkflowAtomicGuard)
	}

	if run.Integration == nil {
		return s.workflowIntegrationProgress(ctx, id, nil, s.admitWorkflowNativeContinuation(ctx, run))
	}

	entry, guard, err := s.workflowNativeContinuationOwner(ctx, id)
	if err == nil {
		err = s.continueWorkflowNativePhase(ctx, entry, guard, provider, ports)
	}

	return s.workflowIntegrationProgress(ctx, id, nil, err)
}

func (s *WorkflowStore) admitWorkflowNativeContinuation(ctx context.Context, run WorkflowInvocation) error {
	if run.Lane == nil || run.PullRequest == nil || run.PullRequest.State != "bound" {
		return ErrWorkflowConflict
	}

	_, err := s.AdmitWorkflowIntegration(ctx, WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: run.Lane.Unit, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/" + run.PullRequest.Request.BaseBranch})

	return err
}

func (s *WorkflowStore) workflowNativeContinuationOwner(ctx context.Context, id string) (entry WorkflowIntegrationEntry, guard WorkflowIntegrationBudgetGuard, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		if run.Integration == nil || run.IntegrationBudget == nil {
			return ErrWorkflowConflict
		}

		entry, err = owner.readWorkflowIntegrationEntry(run, *run.Integration)
		guard = WorkflowIntegrationBudgetGuard{InvocationID: id, InvocationRevision: run.Revision, BudgetRevision: run.IntegrationBudget.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration}

		return err
	})

	return entry, guard, err
}

func (s *WorkflowStore) continueWorkflowNativePhase(ctx context.Context, entry WorkflowIntegrationEntry, guard WorkflowIntegrationBudgetGuard, provider WorkflowNativePreparationProvider, ports WorkflowPublicationPorts) error {
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
		_, _, err := s.PublishWorkflowNativeSubmission(ctx, guard, provider, ports.NativePublisher)

		return err
	case "publishing", "observing":
		return s.observeWorkflowNativeContinuation(ctx, guard, ports)
	case "finalizing":
		_, err := s.FinalizeWorkflowSlice(ctx, guard.InvocationID, nil, nil)

		return err
	default:
		return ErrWorkflowObservationRequired
	}
}

func (s *WorkflowStore) reviewWorkflowNativeContinuation(ctx context.Context, guard WorkflowIntegrationBudgetGuard, provider WorkflowNativePreparationProvider, ports WorkflowPublicationPorts) error {
	review, err := s.ReviewWorkflowSlice(ctx, guard.InvocationID, provider, ports.Reviewer)
	if err != nil || review.State != "authorized" {
		return err
	}

	entry, guard, err := s.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	if err != nil {
		return err
	}

	return s.transitionWorkflowNativeContinuation(ctx, entry, guard, "ready_to_publish")
}

func (s *WorkflowStore) transitionWorkflowNativeContinuation(ctx context.Context, entry WorkflowIntegrationEntry, guard WorkflowIntegrationBudgetGuard, state string) error {
	hash := ""
	if entry.CandidateRef != nil {
		hash = entry.CandidateRef.Hash
	}

	_, err := s.TransitionWorkflowIntegration(ctx, WorkflowIntegrationTransition{InvocationID: guard.InvocationID, EntryID: entry.ID, InvocationRevision: guard.InvocationRevision, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, CandidateHash: hash, State: state})

	return err
}

func (s *WorkflowStore) observeWorkflowNativeContinuation(ctx context.Context, guard WorkflowIntegrationBudgetGuard, ports WorkflowPublicationPorts) error {
	return s.observeWorkflowNativeIntegration(ctx, guard, ports, true)
}

func (s *WorkflowStore) observeWorkflowNativeIntegration(ctx context.Context, guard WorkflowIntegrationBudgetGuard, ports WorkflowPublicationPorts, poll bool) error {
	inspector, ok := ports.NativePublisher.(WorkflowGitPublicationInspector)
	git, prepared := ports.Observer.(WorkflowPreparedIntegrationGit)

	if !ok || !prepared || ports.NativeObserver == nil {
		return ErrWorkflowAtomicGuard
	}

	if _, err := s.ReconcileWorkflowNativeGitSubmission(ctx, guard, inspector, git); err != nil {
		return err
	}

	_, guard, err := s.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	if err != nil {
		return err
	}

	if poll {
		_, err = s.PollWorkflowNativeSubmission(ctx, guard, ports.NativeObserver, ports.Observer)
	} else {
		_, err = s.ObserveWorkflowNativeSubmission(ctx, guard, ports.NativeObserver, ports.Observer)
	}

	if err != nil {
		return err
	}

	_, guard, err = s.workflowNativeContinuationOwner(ctx, guard.InvocationID)
	if err != nil {
		return err
	}

	_, err = s.ConfirmWorkflowNativeReceipt(ctx, guard)

	return err
}
