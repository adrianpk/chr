package workflow

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PublishWorkflowNativeSubmission journals the host dispatch before invoking the
// exact atomic Git publisher once. A transport result is not a delivery receipt.
func (s *Store) PublishWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, provider PublicationGuardProvider, publisher GitPublisher) (Submission, GitPublicationResult, error) {
	if provider == nil || publisher == nil {
		return Submission{}, GitPublicationResult{}, ErrWorkflowAtomicGuard
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	submission, request, err := s.prepareWorkflowNativeDispatch(ctx, guard, provider)
	if err != nil {
		return submission, GitPublicationResult{}, err
	}

	if err := ctx.Err(); err != nil {
		return submission, GitPublicationResult{}, err
	}

	request.BeforeSend = func(ctx context.Context, advertisement GitPublicationAdvertisement) error {
		return s.retainWorkflowNativeSend(ctx, submission, request, advertisement)
	}

	watched, stop := s.watchWorkflowStop(ctx, &Claim{InvocationID: guard.InvocationID})
	defer stop()

	result, err := publisher.PublishWorkflowGit(watched, request)

	finished, finishErr := s.completeWorkflowNativeDispatch(ctx, submission, result)

	return finished, result, errors.Join(err, finishErr)
}

func (s *Store) prepareWorkflowNativeDispatch(ctx context.Context, guard IntegrationBudgetGuard, provider PublicationGuardProvider) (submission Submission, request GitPublication, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if readErr != nil {
			return readErr
		}

		var fresh bool

		submission, fresh, readErr = owner.beginWorkflowNativeSubmission(ctx, run, guard, provider)
		if readErr != nil {
			return readErr
		}

		if !fresh {
			return ErrWorkflowObservationRequired
		}

		submission, request, readErr = owner.authorizeWorkflowNativeDispatch(ctx, submission, provider)

		return readErr
	})

	return submission, request, err
}

func (s *Store) authorizeWorkflowNativeDispatch(ctx context.Context, submission Submission, provider PublicationGuardProvider) (Submission, GitPublication, error) {
	run, err := s.readWorkflowInvocation(submission.InvocationID)
	if err != nil {
		return submission, GitPublication{}, err
	}

	guard, err := s.checkWorkflowNativeDispatchGuard(ctx, run, submission, provider)
	if err != nil {
		return submission, GitPublication{}, err
	}

	if !workflowNativeJSONEqual(guard, submission.Guard) {
		return submission, GitPublication{}, workflowstate.ErrConflict
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil {
		return submission, GitPublication{}, err
	}

	request, err := s.workflowNativeDispatchRequest(candidate)
	if err != nil {
		return submission, request, err
	}

	dispatched, err := s.retainWorkflowNativeDispatch(ctx, run, submission)

	return dispatched, request, err
}

func (s *Store) workflowNativeDispatchRequest(candidate IntegrationCandidate) (GitPublication, error) {
	if candidate.Native == nil {
		return GitPublication{}, workflowstate.ErrConflict
	}

	binding := candidate.Native.NativeRecordBinding

	var capability NativeCapability
	if err := s.readWorkflowNativeEvidence(binding.InvocationID, binding.CapabilityRef, &capability); err != nil {
		return GitPublication{}, err
	}

	request := GitPublication{Repository: binding.SourceRepository, SourceRepository: binding.SourceRepository, Directory: candidate.Lane.Worktree, SourceRef: binding.SourceRef, TargetRef: binding.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: capability.OwnershipRevision, Git: candidate.Input.Git}

	var err error

	request.CandidateRevision, err = GitPublicationRevision(request)

	return request, err
}
