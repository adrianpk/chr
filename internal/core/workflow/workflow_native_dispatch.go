package workflow

import (
	"context"
	"errors"
	"time"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PublishWorkflowNativeSubmission journals the host dispatch before invoking the
// exact atomic Git publisher once. A transport result is not a delivery receipt.
func (s *Store) PublishWorkflowNativeSubmission(ctx context.Context, guard IntegrationBudgetGuard, provider PublicationGuardProvider, publisher GitPublisher) (NativePublicationResult, error) {
	if provider == nil || publisher == nil {
		return NativePublicationResult{Submission: workflowstate.Submission{}, Transport: workflowstate.GitPublicationResult{}}, workflowstate.ErrAtomicGuard
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	nativeDispatch, err := s.prepareWorkflowNativeDispatch(ctx, guard, provider)

	submission, request := nativeDispatch.Submission, nativeDispatch.Request
	if err != nil {
		return NativePublicationResult{Submission: submission, Transport: workflowstate.GitPublicationResult{}}, err
	}

	if err := ctx.Err(); err != nil {
		return NativePublicationResult{Submission: submission, Transport: workflowstate.GitPublicationResult{}}, err
	}

	request.BeforeSend = func(ctx context.Context, advertisement workflowstate.GitPublicationAdvertisement) error {
		return s.retainWorkflowNativeSend(ctx, submission, request, advertisement)
	}

	watched, stop := context.WithCancel(ctx)

	s.watchWorkflowStop(watched, workflowStopWatch{Claim: &Claim{InvocationID: guard.InvocationID}, Cancel: stop})
	defer stop()

	result, err := publisher.PublishWorkflowGit(watched, request)

	finished, finishErr := s.completeWorkflowNativeDispatch(ctx, submission, result)

	return NativePublicationResult{Submission: finished, Transport: result}, errors.Join(err, finishErr)
}

func (s *Store) prepareWorkflowNativeDispatch(ctx context.Context, guard IntegrationBudgetGuard, provider PublicationGuardProvider) (result workflowNativeDispatch, err error) {
	var (
		submission workflowstate.Submission
		request    GitPublication
	)

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, guard.InvocationID)
		if readErr != nil {
			return readErr
		}

		var fresh bool

		submissionIntent, readErr := invocationLocation.Owner.beginWorkflowNativeSubmission(ctx, invocationLocation.Run, guard, provider)
		submission, fresh = submissionIntent.Submission, submissionIntent.Dispatch

		if readErr != nil {
			return readErr
		}

		if !fresh {
			return workflowstate.ErrObservationRequired
		}

		nativeDispatch2, readErr := invocationLocation.Owner.authorizeWorkflowNativeDispatch(ctx, submission, provider)
		submission, request = nativeDispatch2.Submission, nativeDispatch2.Request

		return readErr
	})

	return workflowNativeDispatch{Submission: submission, Request: request}, err
}

func (s *Store) authorizeWorkflowNativeDispatch(ctx context.Context, submission workflowstate.Submission, provider PublicationGuardProvider) (workflowNativeDispatch, error) {
	run, err := s.readWorkflowInvocation(submission.InvocationID)
	if err != nil {
		return workflowNativeDispatch{Submission: submission, Request: GitPublication{}}, err
	}

	guard, err := s.checkWorkflowNativeDispatchGuard(ctx, run, submission, provider)
	if err != nil {
		return workflowNativeDispatch{Submission: submission, Request: GitPublication{}}, err
	}

	if !workflowNativeJSONEqual(guard, submission.Guard) {
		return workflowNativeDispatch{Submission: submission, Request: GitPublication{}}, workflowstate.ErrConflict
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil {
		return workflowNativeDispatch{Submission: submission, Request: GitPublication{}}, err
	}

	request, err := s.workflowNativeDispatchRequest(candidate)
	if err != nil {
		return workflowNativeDispatch{Submission: submission, Request: request}, err
	}

	dispatched, err := s.retainWorkflowNativeDispatch(ctx, run, submission)

	return workflowNativeDispatch{Submission: dispatched, Request: request}, err
}

func (s *Store) workflowNativeDispatchRequest(candidate workflowstate.IntegrationCandidate) (GitPublication, error) {
	if candidate.Native == nil {
		return GitPublication{}, workflowstate.ErrConflict
	}

	binding := candidate.Native.NativeRecordBinding

	var capability workflowstate.NativeCapability
	if err := s.readWorkflowNativeEvidence(binding.InvocationID, binding.CapabilityRef, &capability); err != nil {
		return GitPublication{}, err
	}

	request := GitPublication{Repository: binding.SourceRepository, SourceRepository: binding.SourceRepository, Directory: candidate.Lane.Worktree, SourceRef: binding.SourceRef, TargetRef: binding.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: capability.OwnershipRevision, Git: candidate.Input.Git}

	var err error

	request.CandidateRevision, err = workflowforge.PublicationRevision(workflowPublicationIdentity(request))

	return request, err
}
