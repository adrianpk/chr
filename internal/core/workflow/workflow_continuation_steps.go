package workflow

import (
	"context"
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowContinuationState struct {
	Invocation        workflowstate.Invocation
	MembershipChanged bool
}

// workflowContinuationHost owns the bounded accepted-step loop.
type workflowContinuationHost struct{ store *Store }

func (host workflowContinuationHost) run(ctx context.Context, id string, request ContinuationRequest, ports ContinuationPorts, limit int) (workflowContinuationStep, error) {
	for steps := 0; steps < limit; steps++ {
		observed, err := readWorkflowContinuation(ctx, host.store, id)
		if err != nil {
			return workflowContinuationStep{Steps: steps, Reason: "blocked"}, err
		}

		if observed.MembershipChanged {
			continue
		}

		run := observed.Invocation

		if workflowContinuationPreparationRetry(run, request, steps) {
			retry, err := retryWorkflowContinuationPreparation(ctx, host.store, id, ports)

			return retry, err
		}

		if reason := workflowContinuationBefore(run); reason != "" {
			return workflowContinuationStep{Steps: steps, Reason: reason}, nil
		}

		if err := advanceWorkflowContinuation(ctx, host.store, run, request, ports); err != nil {
			return workflowContinuationStep{Steps: steps, Reason: workflowContinuationFailureReason(err)}, err
		}

		after, err := host.store.ReadWorkflow(ctx, id)
		if err != nil {
			return workflowContinuationStep{Steps: steps, Reason: "blocked"}, err
		}

		accepted := workflowContinuationAfter(run, after)
		if accepted.Reason != "" {
			return workflowContinuationStep{Steps: steps + accepted.Steps, Reason: accepted.Reason}, nil
		}
	}

	return workflowContinuationStep{Steps: limit, Reason: "step_limit"}, nil
}

func workflowContinuationFailureReason(err error) string {
	if errors.Is(err, ErrWorkflowCapacity) {
		return "waiting_capacity"
	}

	return "blocked"
}

func readWorkflowContinuation(ctx context.Context, store *Store, id string) (workflowContinuationState, error) {
	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowContinuationState{Invocation: run}, err
	}

	refreshed, err := store.ReconcileWorkflowMembership(ctx, id)

	return workflowContinuationState{Invocation: run, MembershipChanged: refreshed.Revision != run.Revision}, err
}

func workflowContinuationAfter(before, after workflowstate.Invocation) workflowContinuationStep {
	if after.Revision == before.Revision {
		return workflowContinuationStep{Reason: "awaiting_authority"}
	}

	return workflowContinuationStep{Steps: 1, Reason: workflowContinuationBarrier(after)}
}

func workflowContinuationBarrier(run workflowstate.Invocation) string {
	if workflowstate.InvocationTerminal(run.Status) {
		return string(run.Status)
	}

	if run.StopIntent != "" || run.Status == workflowstate.InvocationPaused {
		return "stopped"
	}

	if run.Outcome == "set_closure_required" {
		return "set_closure_required"
	}

	if run.Outcome == "closing_validation_failed" || run.Outcome == "interrupted" {
		return run.Outcome
	}

	if run.Status == workflowstate.InvocationBlocked || run.Status == workflowstate.InvocationWaitingCapacity {
		return string(run.Status)
	}

	return ""
}

func advanceWorkflowContinuation(ctx context.Context, store *Store, run workflowstate.Invocation, request ContinuationRequest, ports ContinuationPorts) error {
	state := workflowInvocationFile{Invocation: run}
	restoreWorkflowCapacityWait(&state)

	run = state.Invocation

	if workflowContinuationClosing(run) {
		_, err := CloseInvocationSet(ctx, store, run.ID, request.PullRequest.BaseBranch, FrontierPorts{Provider: ports.Publication.Provider, Observer: ports.Publication.Observer, Target: ports.Target})

		return err
	}

	if run.DeliveryRef != "" {
		return advanceWorkflowContinuationDelivery(ctx, store, run, ports)
	}

	if run.CandidateRef != "" {
		if workflowCandidateNeedsPR(run) {
			return ensureWorkflowContinuationPR(ctx, store, run, request, ports)
		}

		_, err := AdvanceWorkflowSliceIntegration(ctx, store, run.ID, ports.Publication)

		return err
	}

	if run.Lane != nil && run.Lane.Unit.Kind == "ticket" && workflowSliceValidationReady(run) {
		_, err := store.PrepareWorkflowTicketCandidate(ctx, run.ID)

		return err
	}

	if workflowSliceValidationReady(run) {
		_, err := ValidateSlice(ctx, store, run.ID)

		return err
	}

	return executeWorkflowSliceTask(ctx, store, run.ID, request.Worker)
}

func advanceWorkflowContinuationDelivery(ctx context.Context, store *Store, run workflowstate.Invocation, ports ContinuationPorts) error {
	if run.Integration != nil && run.ConfirmedCursor < run.Cursor {
		_, err := AdvanceWorkflowSliceIntegration(ctx, store, run.ID, ports.Publication)

		return err
	}

	_, err := store.AdvanceWorkflowFrontier(ctx, run.ID, FrontierPorts{Provider: ports.Publication.Provider, Observer: ports.Publication.Observer, Target: ports.Target})

	return err
}

func ensureWorkflowContinuationPR(ctx context.Context, store *Store, run workflowstate.Invocation, request ContinuationRequest, ports ContinuationPorts) error {
	if ports.Source != nil && (run.PullRequest == nil || run.PullRequest.State == "prepared" || run.PullRequest.State == "bound") {
		if err := store.PublishWorkflowSource(ctx, run.ID, ports.Source); err != nil {
			return err
		}
	}

	_, err := store.EnsureWorkflowSlicePullRequest(ctx, run.ID, request.PullRequest, ports.Publication.Provider)

	return err
}

func workflowContinuationClosing(run workflowstate.Invocation) bool {
	return !workflowstate.InvocationTerminal(run.Status) && run.StopIntent == "" && run.Status != workflowstate.InvocationPaused && (run.Outcome == "set_closure_required" || run.Outcome == "closing_validation_failed" || run.Outcome == "interrupted" || run.ClosingSet != "")
}

func workflowContinuationBefore(run workflowstate.Invocation) string {
	if run.Status == workflowstate.InvocationWaitingCapacity && run.StopIntent == "" {
		return ""
	}

	if workflowContinuationClosing(run) {
		return ""
	}

	return workflowContinuationBarrier(run)
}

func workflowCandidateNeedsPR(run workflowstate.Invocation) bool {
	return run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.CandidateHash != run.CandidateHash
}
