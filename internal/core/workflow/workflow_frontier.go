package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// UnitDelivery retains a delivered unit before its active bindings move.
// EvidenceRef binds the exact frontier, target and receipt to an immutable image.
type UnitDelivery struct {
	Schema            string                      `json:"schema"`
	InvocationID      string                      `json:"invocation_id"`
	Unit              workflowstate.Ref           `json:"unit"`
	SetID             string                      `json:"set_id,omitempty"`
	Cursor            int                         `json:"cursor"`
	Target            workflowstate.PreviewTarget `json:"target"`
	NextTarget        workflowstate.PreviewTarget `json:"next_target"`
	Lane              Lane                        `json:"lane"`
	CandidateRef      string                      `json:"candidate_ref"`
	CandidateHash     string                      `json:"candidate_hash"`
	PullRequest       *PullRequestBinding         `json:"pull_request_binding"`
	DeliveryProofRef  string                      `json:"delivery_proof_ref"`
	DeliveryProofHash string                      `json:"delivery_proof_hash"`
	DeliveryRef       string                      `json:"delivery_ref"`
	DeliveryHash      string                      `json:"delivery_hash"`
	Preparation       *PreparationBinding         `json:"preparation_binding,omitempty"`
	Review            *ReviewBinding              `json:"review_binding,omitempty"`
	Submission        *SubmissionBinding          `json:"submission_binding,omitempty"`
	Proof             IntegrationProof            `json:"proof"`
	EvidenceRef       string                      `json:"evidence_ref,omitempty"`
	EvidenceHash      string                      `json:"evidence_hash,omitempty"`
}

// FrontierPorts supplies trusted canonical observation and target hydration.
type FrontierPorts struct {
	Provider PullRequestProvider
	Observer IntegrationGit
	Target   TargetGit
}

// ExecutionSelection preserves admission identity while refreshing execution base.
func (run Invocation) ExecutionSelection() workflowstate.PreviewRequest {
	selection := run.Selection
	if run.ExecutionTarget != nil {
		selection.Target = *run.ExecutionTarget
	}

	return selection
}

// AdvanceWorkflowFrontier releases one finalized integration unit without work.
// It verifies canonical delivery, journals retained proof, and hydrates the next
// base. Grouped set boundaries wait for closing validation, including the last
// set; loose units return to their frozen queue for fresh selection.
func (s *Store) AdvanceWorkflowFrontier(ctx context.Context, id string, ports FrontierPorts) (result Invocation, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		readErr = owner.advanceWorkflowFrontier(ctx, &run, ports)
		result = run.Invocation

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

	target, proof, err := s.observeWorkflowFrontier(ctx, *run, receipt, candidate, ports)

	err = errors.Join(err, s.refreshWorkflowNativeFrontier(run, receipt))
	if err != nil {
		return fmt.Errorf("observe canonical frontier delivery: %w", err)
	}

	if err := s.checkWorkflowInputs(ctx, *run); err != nil {
		return err
	}

	if err := s.commitWorkflowFrontier(run, target, proof); err != nil {
		return fmt.Errorf("retain verified frontier: %w", err)
	}

	return nil
}

func (s *Store) refreshWorkflowNativeFrontier(run *workflowInvocationFile, receipt Record) error {
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
	if run.StopIntent != "" || run.Status == "paused" || run.Status == "cancelled" {
		return false, ErrWorkflowStop
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
		return false, ErrWorkflowObservationRequired
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
