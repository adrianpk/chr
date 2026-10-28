package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// FrontierPorts supplies trusted canonical observation and target hydration.
type FrontierPorts struct {
	Provider PullRequestProvider
	Observer IntegrationGit
	Target   TargetGit
}

// AdvanceWorkflowFrontier releases one finalized integration unit without work.
// It verifies canonical delivery, journals retained proof, and hydrates the next
// base. Grouped set boundaries wait for closing validation, including the last
// set; loose units return to their frozen queue for fresh selection.
func (s *Store) AdvanceWorkflowFrontier(ctx context.Context, id string, ports FrontierPorts) (result workflowstate.Invocation, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		readErr = invocationLocation.Owner.advanceWorkflowFrontier(ctx, &invocationLocation.Run, ports)
		result = invocationLocation.Run.Invocation

		return readErr
	})

	return result, err
}

func (s *Store) advanceWorkflowFrontier(ctx context.Context, run *workflowInvocationFile, ports FrontierPorts) error {
	ready, err := s.workflowFrontierReady(ctx, *run)
	if err != nil || !ready {
		return err
	}

	receipt, err := s.readWorkflowSliceReceipt(ctx, *run)
	if err != nil {
		return fmt.Errorf("read finalized frontier receipt: %w", err)
	}

	candidate, err := s.readWorkflowFrontierCandidate(ctx, *run)
	if err != nil {
		return fmt.Errorf("read retained frontier candidate: %w", err)
	}

	observed, err := s.observeWorkflowFrontier(ctx, workflowFrontierUnit{Run: *run, Receipt: receipt, Candidate: candidate}, ports)

	err = errors.Join(err, s.refreshWorkflowNativeFrontier(run, receipt))
	if err != nil {
		return fmt.Errorf("observe canonical frontier delivery: %w", err)
	}

	if err := s.checkWorkflowInputs(ctx, *run); err != nil {
		return err
	}

	if err := s.commitWorkflowFrontier(run, observed.Target, observed.Proof); err != nil {
		return fmt.Errorf("retain verified frontier: %w", err)
	}

	return nil
}

func (s *Store) refreshWorkflowNativeFrontier(run *workflowInvocationFile, receipt workflowstate.Record) error {
	if receipt.Receipt.Native == nil {
		return nil
	}

	// Canonical observations can retain a blocker as well as verified delivery.
	// Return the retained owner revision even when remote verification failed.
	current, err := s.readWorkflowInvocation(run.ID)
	if err == nil {
		*run = current
	}

	return err
}

func (s *Store) workflowFrontierReady(ctx context.Context, run workflowInvocationFile) (bool, error) {
	if run.StopIntent != "" || run.Status == workflowstate.InvocationPaused || run.Status == workflowstate.InvocationCancelled {
		return false, workflowstate.ErrStop
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return false, err
	}

	if run.Selection.Selector.Extent != "slice_set" && run.Selection.Selector.Extent != "use_case" && run.Selection.Selector.Extent != "loose" {
		return false, nil
	}

	if run.Lane == nil {
		return false, nil
	}

	if run.DeliveryRef == "" {
		return false, workflowstate.ErrObservationRequired
	}

	return true, nil
}

func (s *Store) readWorkflowFrontierCandidate(ctx context.Context, run workflowInvocationFile) (SliceCandidate, error) {
	if run.Lane.Unit.Kind == "ticket" {
		observation, err := s.readWorkflowTicketMembership(run)
		if err != nil {
			return SliceCandidate{}, err
		}

		return s.readWorkflowSliceCandidate(ctx, run, observation)
	}

	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return SliceCandidate{}, err
	}

	dependencies, err := newWorkflowDependencies(selection)
	if err != nil {
		return SliceCandidate{}, err
	}

	tasks := dependencies.tasks[run.Lane.Unit.ID]
	if len(tasks) == 0 {
		return SliceCandidate{}, fmt.Errorf("%w: finalized slice has no members", workflowstate.ErrConflict)
	}

	observation := workflowSliceObservation{descriptor: selection.slices[run.Lane.Unit.ID], set: selection.sets[tasks[0].Task.SetID], tasks: tasks}

	return s.readWorkflowSliceCandidate(ctx, run, observation)
}
